package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/hugr-lab/mcp/pkg/indexer"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
	"github.com/mark3labs/mcp-go/mcp"
)

var discoveryDataObjectFieldValuesTool = mcp.NewTool("discovery-data_object_field_values",
	mcp.WithDescription("Return field values stats for a specific field in a data object"),
	mcp.WithInputSchema[schemaDataObjectFieldValuesInput](),
	mcp.WithOutputSchema[DataObjectFieldValueStat](),
)

type schemaDataObjectFieldValuesInput struct {
	ObjectName     string         `json:"object_name" jsonschema_description:"The name of the data object (GraphQL type) to query"`
	FieldName      string         `json:"field_name" jsonschema_description:"The name of the field within the data object to get stats for"`
	Offset         int            `json:"offset,omitempty" jsonschema_description:"The number of distinct values to skip before starting to collect the result set" jsonschema:"minimum=0,default=0"`
	Limit          int            `json:"limit" jsonschema_description:"The number of top distinct values to return" jsonschema:"minimum=1,default=10,maximum=100"`
	CalculateStats bool           `json:"calculate_stats,omitempty" jsonschema_description:"Whether to calculate and return statistical summaries (min, max, avg, distinct count, null count) for the field" jsonschema:"default=false"`
	Filter         map[string]any `json:"filter,omitempty" jsonschema_description:"Optional filter to apply when querying field stats. The filter should be a JSON object that represents the GraphQL filter input for the data object. For example, to filter on a users table by age greater than 30, you might use: {\"age\": {\"gt\": 30}}"`
	Args           map[string]any `json:"args,omitempty" jsonschema_description:"Optional arguments to pass to the data object (if it is parameterized view). The args should be a JSON object that represents the GraphQL arguments input for the data object."`
}

type DataObjectFieldValueStat struct {
	Min      any   `json:"min,omitempty"`
	Max      any   `json:"max,omitempty"`
	Avg      any   `json:"avg,omitempty"`
	Distinct int   `json:"distinct,omitempty"`
	Values   []any `json:"values,omitempty"`
}

func (s *Service) discoveryDataObjectFieldValuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &schemaDataObjectFieldValuesInput{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}
	if strings.Contains(input.FieldName, ".") {
		return mcp.NewToolResultError("field_name should be a top-level field, nested fields are not supported"), nil
	}

	// get data object queries info
	tqq, err := s.indexer.DataObjectQueriesInfo(ctx, input.ObjectName)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get type info", err), nil
	}
	if tqq == nil {
		return mcp.NewToolResultError("data object not found"), nil
	}

	// prepare query
	stats, err := s.queryDataObjectFieldStats(ctx, tqq, *input)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to query field stats", err), nil
	}

	out := mcp.NewToolResultStructuredOnly(stats)
	return out, nil
}

const (
	dataObjectFieldStatsMinMaxTemplate = `stats: %s %s @cache(ttl: $ttl) {field: %s {min, max, avg, distinct: count }}`
	dataObjectFieldStatsOtherTemplate  = `stats: %s %s @cache(ttl: $ttl)  {field: %s { distinct: count }}`
	dataObjectFieldValuesTemplate      = `values: %s(%s limit: $limit offset: $offset distinct_on: ["field"]) @cache(ttl: $ttl) { field: %s{ list }}`
)

func (s *Service) queryDataObjectFieldStats(ctx context.Context, info *indexer.DataObjectQueriesInfo, req schemaDataObjectFieldValuesInput) (DataObjectFieldValueStat, error) {
	aggQuery := ""
	for _, q := range info.Queries {
		if q.Type == string(metainfo.QueryTypeAggregate) {
			aggQuery = q.Name
			break
		}
	}
	if aggQuery == "" {
		return DataObjectFieldValueStat{}, fmt.Errorf("data object %q does not support aggregation queries", info.Name)
	}

	// field type to decide which stats to calculate
	fieldType, err := s.indexer.DataObjectFieldType(ctx, req.ObjectName, req.FieldName)
	if err != nil {
		return DataObjectFieldValueStat{}, fmt.Errorf("get data object field type: %w", err)
	}
	if fieldType == "" {
		return DataObjectFieldValueStat{}, fmt.Errorf("field %q not found in data object %q", req.FieldName, req.ObjectName)
	}

	if info.ArgsType == "" {
		req.Args = nil // no args supported
	}

	calcValues := req.Limit > 0
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	args := ""
	if info.ArgsType != "" && len(req.Args) != 0 {
		args = "args: $args"
	}
	if info.FilterType != "" && len(req.Filter) != 0 {
		if args != "" {
			args += " "
		}
		args += "filter: $filter"
	}

	query := ""
	if req.CalculateStats {
		args := args
		if args != "" {
			args = "(" + args + ")"
		}
		switch fieldType {
		case "Int", "Float", "BigInt", "Timestamp", "Date", "Time":
			query = fmt.Sprintf(dataObjectFieldStatsMinMaxTemplate, aggQuery, args, req.FieldName)
		case "String", "JSON", "Boolean", "Geometry":
			query = fmt.Sprintf(dataObjectFieldStatsOtherTemplate, aggQuery, args, req.FieldName)
		}
	}
	if calcValues {
		if query != "" {
			query += "\n"
		}
		query += fmt.Sprintf(dataObjectFieldValuesTemplate, aggQuery, args, req.FieldName)
	}
	if query == "" {
		return DataObjectFieldValueStat{}, fmt.Errorf("nothing to calculate for field type %q", fieldType)
	}

	// add module path if not core
	pp := strings.Split(info.Module, ".")
	var pre, post string
	for _, p := range pp {
		pre += p + " {"
		post += "}"
	}
	query = pre + " " + query + " " + post

	args = "$ttl: Int! $limit: Int! $offset: Int!"
	vars := map[string]any{
		"ttl":    s.cfg.ttl,
		"limit":  req.Limit,
		"offset": req.Offset,
	}
	if info.ArgsType != "" {
		args += " $args: " + info.ArgsType + "!"
		vars["args"] = req.Args
	}
	if info.FilterType != "" {
		args += " $filter: " + info.FilterType + "!"
		vars["filter"] = req.Filter
	}

	// add query definition
	query = "query fieldValues(" + args + ") {\n" + query + "\n}"
	res, err := s.hugr.Query(ctx, query, vars)
	if err != nil {
		return DataObjectFieldValueStat{}, err
	}
	defer res.Close()
	if res.Err() != nil {
		return DataObjectFieldValueStat{}, res.Err()
	}

	// parse result
	var stats DataObjectFieldValueStat
	path := ""
	if info.Module != "" {
		path = info.Module + "."
	}
	if req.CalculateStats {
		var data struct {
			Stats DataObjectFieldValueStat `json:"field"`
		}

		err = res.ScanData(path+"stats", &data)
		if err != nil {
			return DataObjectFieldValueStat{}, fmt.Errorf("failed to parse stats result: %w", err)
		}
		stats = data.Stats
	}
	if !calcValues {
		return stats, nil
	}
	var vals struct {
		Field struct {
			List []any `json:"list"`
		} `json:"field"`
	}
	err = res.ScanData(path+"values", &vals)
	if err != nil {
		return DataObjectFieldValueStat{}, fmt.Errorf("failed to parse values result: %w", err)
	}
	stats.Values = vals.Field.List
	return stats, nil
}
