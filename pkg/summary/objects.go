package summary

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hugr-lab/mcp/pkg/pool"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
)

//go:embed templates/table.tmpl
var tableInfoTmpl string

//go:embed templates/system.txt
var systemPrompt string

func (s *Service) SummarizeDataObject(ctx context.Context, schema *metainfo.SchemaInfo, object *metainfo.DataObjectInfo) (*DataObjectSummary, error) {
	// Use the tableInfoTmpl and systemPrompt to generate the summary
	input, err := prepareDataObjectInput(schema, object)
	if err != nil {
		return nil, fmt.Errorf("prepare table input: %w", err)
	}

	// to template object
	var data DataObjectDescribeTemplateData
	b, err := json.Marshal(input.Object)
	if err != nil {
		return nil, fmt.Errorf("marshal input object: %w", err)
	}
	data.ObjectJSON = string(b)
	b, err = json.Marshal(input.Columns)
	if err != nil {
		return nil, fmt.Errorf("marshal input columns: %w", err)
	}
	data.ColumnsJSON = string(b)
	b, err = json.Marshal(input.References)
	if err != nil {
		return nil, fmt.Errorf("marshal input references: %w", err)
	}
	data.ReferencesJSON = string(b)
	b, err = json.Marshal(input.Subqueries)
	if err != nil {
		return nil, fmt.Errorf("marshal input subqueries: %w", err)
	}
	data.SubqueriesJSON = string(b)
	b, err = json.Marshal(input.FunctionCalls)
	if err != nil {
		return nil, fmt.Errorf("marshal input function calls: %w", err)
	}
	data.FunctionCallsJSON = string(b)
	b, err = json.Marshal(input.Mutations)
	if err != nil {
		return nil, fmt.Errorf("marshal input mutations: %w", err)
	}
	data.MutationsJSON = string(b)
	b, err = json.Marshal(input.Queries)
	if err != nil {
		return nil, fmt.Errorf("marshal input queries: %w", err)
	}
	data.QueriesJSON = string(b)
	b, err = json.Marshal(input.Arguments)
	if err != nil {
		return nil, fmt.Errorf("marshal input arguments: %w", err)
	}
	data.ArgumentsJSON = string(b)
	b, err = json.Marshal(input.RelatedGraph.Nodes)
	if err != nil {
		return nil, fmt.Errorf("marshal input related graph: %w", err)
	}
	data.RelatedGraphNodesJSON = string(b)
	b, err = json.Marshal(input.RelatedGraph.Edges)
	if err != nil {
		return nil, fmt.Errorf("marshal input related graph edges: %w", err)
	}
	data.RelatedGraphEdgesJSON = string(b)
	data.RelatedGraphMaxDepth = input.RelatedGraph.MaxDepth

	b, err = json.Marshal(input.ModuleContext)
	if err != nil {
		return nil, fmt.Errorf("marshal input module context: %w", err)
	}
	data.ModuleContextJSON = string(b)

	b, err = json.Marshal(input.DataSourceContext)
	if err != nil {
		return nil, fmt.Errorf("marshal input data source context: %w", err)
	}
	data.DataSourceContextJSON = string(b)

	c, err := s.pool.Connection(ctx)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	defer c.Close()
	out, err := c.Summarize(ctx, &pool.SummarizationTask{
		SystemPrompt:       systemPrompt,
		UserPromptTemplate: tableInfoTmpl,
		Data:               data,
		Temperature:        0.3,
		MaxTokens:          16384,
	})
	if err != nil {
		return nil, fmt.Errorf("summarize table: %w", err)
	}

	summary := DataObjectSummary{}
	err = json.Unmarshal([]byte(out), &summary)
	if err != nil {
		return nil, errors.Join(ErrSummarizationOutputFormat, err)
	}
	return &summary, nil
}

type DataObjectSummary struct {
	Short                      string                               `json:"short"`
	Long                       string                               `json:"long"`
	AggregationTypeShort       string                               `json:"aggregation_type_short"`
	AggregationTypeLong        string                               `json:"aggregation_type_long"`
	SubAggregationTypeShort    string                               `json:"sub_aggregation_type_short"`
	SubAggregationTypeLong     string                               `json:"sub_aggregation_type_long"`
	BucketAggregationTypeShort string                               `json:"bucket_aggregation_type_short"`
	BucketAggregationTypeLong  string                               `json:"bucket_aggregation_type_long"`
	Fields                     map[string]string                    `json:"fields"` // name -> description
	ExtraFields                map[string]string                    `json:"extra_fields,omitempty"`
	Filter                     DataObjectFilterSummary              `json:"filter,omitempty"`
	References                 map[string]DataObjectSubQuerySummary `json:"references,omitempty"`
	SubQueries                 map[string]DataObjectSubQuerySummary `json:"subqueries,omitempty"`
	FunctionCalls              map[string]string                    `json:"function_calls,omitempty"`
	Arguments                  DataObjectArgumentsSummary           `json:"arguments,omitempty"`
	Queries                    map[string]string                    `json:"queries,omitempty"`
	Mutations                  map[string]string                    `json:"mutations,omitempty"`
}

type DataObjectSubQuerySummary struct {
	Short     string `json:"short"`
	Filter    string `json:"filter,omitempty"`
	Select    string `json:"select"`
	Agg       string `json:"select_agg,omitempty"`
	BucketAgg string `json:"select_bucket_agg,omitempty"`
}

type DataObjectFilterSummary struct {
	Row        string            `json:"row"`
	Fields     map[string]string `json:"fields"`
	References map[string]string `json:"references,omitempty"`
}

type DataObjectArgumentsSummary struct {
	Short  string            `json:"short"`
	Fields map[string]string `json:"fields"`
}

func prepareDataObjectInput(schema *metainfo.SchemaInfo, object *metainfo.DataObjectInfo) (*DataObjectDescribeInput, error) {
	input := &DataObjectDescribeInput{
		Object: DataObjectInfo{
			Name:                     object.Name,
			Type:                     object.Type,
			Description:              object.Description,
			HasPrimaryKey:            object.HasPrimaryKey,
			HasGeometry:              object.HasGeometry,
			IsM2M:                    object.IsM2M,
			IsCube:                   object.IsCube,
			IsHypertable:             object.IsHypertable,
			HasAggregationType:       object.AggregationType != "",
			HasSubAggregationType:    object.SubAggregationType != "",
			HasBucketAggregationType: object.BucketAggregationType != "",
		},
		FunctionCalls: object.FunctionCalls,
		Mutations:     object.Mutations,
		Arguments:     object.Arguments,
		RelatedGraph: RelatedGraph{
			Nodes:    []RelatedNode{},
			Edges:    []RelatedEdge{},
			MaxDepth: 2,
		},
	}
	// Columns
	for _, col := range object.Columns {
		column := FieldInfo{
			Name:         col.Name,
			Description:  col.Description,
			Type:         col.Type,
			IsPrimaryKey: col.IsPrimaryKey,
			IsArray:      col.ReturnsArray,
		}
		if col.IsCalculated {
			if column.Description != "" {
				column.Description += " "
			}
			column.Description += "(calculated)"
		}
		for _, extra := range col.ExtraFields {
			field := fmt.Sprintf("%s: %s", extra.Name, extra.Type)
			if extra.Description != "" {
				field += " (" + extra.Description + ")"
			}
			column.ExtraFields = append(column.ExtraFields, field)
		}
		input.Columns = append(input.Columns, column)
	}
	// references
	for _, ref := range object.References {
		reference := metainfo.SubqueryInfo{
			Name:                ref.Name,
			Type:                ref.Type,
			Module:              ref.Module,
			DataObject:          ref.DataObject,
			Description:         ref.Description,
			FieldDataQuery:      ref.FieldDataQuery,
			FieldAggQuery:       ref.FieldAggQuery,
			FieldBucketAggQuery: ref.FieldBucketAggQuery,
		}
		input.References = append(input.References, reference)
	}

	for _, ref := range object.Subqueries {
		sq := metainfo.SubqueryInfo{
			Name:                ref.Name,
			Type:                ref.Type,
			Module:              ref.Module,
			DataObject:          ref.DataObject,
			Description:         ref.Description,
			FieldDataQuery:      ref.FieldDataQuery,
			FieldAggQuery:       ref.FieldAggQuery,
			FieldBucketAggQuery: ref.FieldBucketAggQuery,
		}
		input.Subqueries = append(input.Subqueries, sq)
	}

	// queries
	for _, q := range object.Queries {
		query := metainfo.QueryInfo{
			Name:             q.Name,
			Description:      q.Description,
			Type:             q.Type,
			ReturnedTypeName: q.ReturnedTypeName,
			IsSingleRow:      q.IsSingleRow,
		}
		input.Queries = append(input.Queries, query)
	}

	err := input.RelatedGraph.addDataObject(schema, object, 0)
	if err != nil {
		return nil, err
	}

	ds := schema.DataSource(object.DataSource)
	if ds == nil {
		if !strings.HasPrefix(object.DataSource, "core") {
			return nil, fmt.Errorf("data source not found: %s", object.DataSource)
		}
		ds = &metainfo.DataSourceInfo{
			Name:        object.DataSource,
			Type:        "runtime",
			Description: "Core built-in data source for managed hugr",
		}
	}
	input.DataSourceContext = DataSourceContext{
		Name:        ds.Name,
		SummaryText: fmt.Sprintf("%s (%s)", ds.Description, ds.Type),
	}
	m := schema.Module(object.Module)
	if m == nil {
		return nil, fmt.Errorf("module not found: %s", object.Module)
	}
	input.ModuleContext = ModuleContext{
		Name:     m.Name,
		Overview: m.Description,
	}

	return input, nil
}

func (g *RelatedGraph) nodeExists(name string) bool {
	for _, node := range g.Nodes {
		if node.Name == name {
			return true
		}
	}
	return false
}

func (g *RelatedGraph) edgeExists(name string) bool {
	for _, edge := range g.Edges {
		if edge.Name == name {
			return true
		}
	}
	return false
}

func (g *RelatedGraph) addDataObject(schema *metainfo.SchemaInfo, object *metainfo.DataObjectInfo, depth int) error {
	if !g.nodeExists(object.Name) {
		brief := object.Description
		if brief != "" {
			brief += " | "
		}
		brief += "Fields: "
		for i, field := range object.Columns {
			if i > 0 {
				brief += ", "
			}
			brief += field.Name
			if field.Description != "" {
				brief += " (" + field.Description + ")"
			}
		}
		g.Nodes = append(g.Nodes, RelatedNode{
			Type:       string(object.Type),
			Name:       object.Name,
			Module:     object.Module,
			DataSource: object.DataSource,
			Brief:      brief,
		})
	}

	if depth >= g.MaxDepth {
		return nil
	}

	// references
	for _, ref := range object.References {
		refObject := schema.Table(ref.Module + "." + ref.DataObject)
		if refObject == nil {
			refObject = schema.View(ref.Module + "." + ref.DataObject)
		}
		if refObject == nil { // skip if references not found
			continue
		}
		err := g.addDataObject(schema, refObject, depth+1)
		if err != nil {
			return err
		}
		var kind RelatedEdgeKind
		switch ref.Type {
		case metainfo.ReferenceTypeOneToMany:
			kind = RelatedEdgeKindOneToMany
		case metainfo.ReferenceTypeManyToOne:
			kind = RelatedEdgeKindManyToOne
		case metainfo.ReferenceTypeManyToMany:
			kind = RelatedEdgeKindManyToMany
		}

		if !g.edgeExists(object.Name + ":" + ref.FieldDataQuery) {
			g.Edges = append(g.Edges, RelatedEdge{
				Name: object.Name + ":" + ref.FieldDataQuery,
				From: object.Name,
				To:   ref.DataObject,
				Kind: kind,
			})
		}
	}

	// subqueries
	for _, subq := range object.Subqueries {
		sqObject := schema.Table(subq.Module + "." + subq.DataObject)
		if sqObject == nil {
			sqObject = schema.View(subq.Module + "." + subq.DataObject)
		}
		if sqObject == nil { // skip if references not found
			continue
		}
		err := g.addDataObject(schema, sqObject, depth+1)
		if err != nil {
			return err
		}
		if !g.edgeExists(object.Name + ":" + subq.FieldDataQuery) {
			g.Edges = append(g.Edges, RelatedEdge{
				Name: object.Name + ":" + subq.FieldDataQuery,
				From: object.Name,
				To:   sqObject.Name,
				Kind: RelatedEdgeKindSubquery,
			})
		}
	}

	// function calls
	for _, fnCall := range object.FunctionCalls {
		err := g.addFunctionCall(&fnCall, depth+1)
		if err != nil {
			return err
		}
		if !g.edgeExists(object.Name + "->" + "fc:" + fnCall.FieldName) {
			g.Edges = append(g.Edges, RelatedEdge{
				Name: object.Name + "->" + "fc:" + fnCall.FieldName,
				From: object.Name,
				To:   "fc:" + fnCall.Name,
				Kind: RelatedEdgeKindFunctionCall,
			})
		}
	}

	return nil
}

func (g *RelatedGraph) addFunctionCall(fnCall *metainfo.FunctionCallInfo, depth int) error {
	if !g.nodeExists("fc:" + fnCall.Name) {
		g.Nodes = append(g.Nodes, RelatedNode{
			Type:       "function",
			Name:       "fc:" + fnCall.Name,
			Module:     fnCall.Module,
			DataSource: fnCall.DataSource,
			Brief:      fnCall.Description,
		})
	}

	return nil
}

// Related graph (deduped, bounded depth) for recursive awareness
type RelatedNode struct {
	Type       string `json:"type"` // "table" | "view" | "function"
	Name       string `json:"name"`
	Module     string `json:"module,omitempty"`
	DataSource string `json:"data_source,omitempty"`
	Brief      string `json:"brief,omitempty"`
}

type RelatedEdgeKind string

const (
	RelatedEdgeKindOneToMany    RelatedEdgeKind = "fk:one-to-many"
	RelatedEdgeKindManyToOne    RelatedEdgeKind = "fk:many-to-one"
	RelatedEdgeKindManyToMany   RelatedEdgeKind = "fk:many-to-many"
	RelatedEdgeKindSubquery     RelatedEdgeKind = "subquery"
	RelatedEdgeKindFunctionCall RelatedEdgeKind = "function_call"
)

type RelatedEdge struct {
	Name string          `json:"name"` // edge name
	From string          `json:"from"` // node name or qualified id
	To   string          `json:"to"`
	Kind RelatedEdgeKind `json:"kind"` // "fk:one-to-many" | "fk:many-to-one" | "fk:many-to-many" | "subquery" | "function_call"
}

type RelatedGraph struct {
	MaxDepth int           `json:"max_depth"`
	Nodes    []RelatedNode `json:"nodes"`
	Edges    []RelatedEdge `json:"edges,omitempty"`
}

type Hints struct {
	UserTask string   `json:"user_task,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

type DataObjectInfo struct {
	Name                     string                  `json:"name"`
	Type                     metainfo.DataObjectType `json:"type"`
	Description              string                  `json:"description"`
	HasPrimaryKey            bool                    `json:"has_primary_key"`
	HasGeometry              bool                    `json:"has_geometry"`
	IsM2M                    bool                    `json:"is_m2m"`
	IsCube                   bool                    `json:"is_cube"`
	IsHypertable             bool                    `json:"is_hypertable"`
	HasAggregationType       bool                    `json:"has_aggregation_type,omitempty"`
	HasSubAggregationType    bool                    `json:"has_sub_aggregation_type,omitempty"`
	HasBucketAggregationType bool                    `json:"has_bucket_aggregation_type,omitempty"`
}

// The full prompt input payload you will marshal and inject into the template.
type DataObjectDescribeInput struct {
	Object        DataObjectInfo              `json:"object"`
	Columns       []FieldInfo                 `json:"columns"`
	References    []metainfo.SubqueryInfo     `json:"relations"`
	Subqueries    []metainfo.SubqueryInfo     `json:"subqueries"`
	FunctionCalls []metainfo.FunctionCallInfo `json:"function_calls"`
	Queries       []metainfo.QueryInfo        `json:"queries"`
	Mutations     *metainfo.MutationInfo      `json:"mutations,omitempty"`
	Arguments     *metainfo.ArgumentInfo      `json:"arguments,omitempty"`

	DataSourceContext DataSourceContext `json:"data_source_context"`
	ModuleContext     ModuleContext     `json:"module_context"`

	RelatedGraph RelatedGraph `json:"related_graph"`
	Hints        *Hints       `json:"hints,omitempty"`
}

type FieldInfo struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	IsPrimaryKey bool        `json:"is_primary_key,omitempty"`
	IsArray      bool        `json:"is_array,omitempty"`
	ExtraFields  []string    `json:"extra_fields"`
	NestedFields []FieldInfo `json:"nested_fields,omitempty"`
}

type DataObjectDescribeTemplateData struct {
	ObjectJSON            string
	ColumnsJSON           string
	ReferencesJSON        string
	SubqueriesJSON        string
	FunctionCallsJSON     string
	QueriesJSON           string
	MutationsJSON         string
	ArgumentsJSON         string
	DataSourceContextJSON string
	ModuleContextJSON     string
	RelatedGraphMaxDepth  int
	RelatedGraphNodesJSON string
	RelatedGraphEdgesJSON string
	UserTaskJSON          string
	KeywordsJSON          string
}
