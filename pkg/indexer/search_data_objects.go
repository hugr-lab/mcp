package indexer

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hugr-lab/mcp/pkg/auth"
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	"github.com/hugr-lab/query-engine/pkg/types"
)

type DataObjectSearchItem struct {
	Module          string                      `json:"module" jsonschema_description:"Module of the data object"`
	Name            string                      `json:"name" jsonschema_description:"Name of the data object"`
	Description     string                      `json:"description" jsonschema_description:"Description of the data object"`
	LongDescription string                      `json:"long_description" jsonschema_description:"Long description of the data object"`
	Type            string                      `json:"type" jsonschema_description:"Type of the data object"`
	Score           float64                     `json:"score" jsonschema_description:"Search score of the data object"`
	Fields          []DataObjectSearchItemField `json:"fields" jsonschema_description:"Fields of the data object"`
	FieldsTruncated bool                        `json:"fields_truncated" jsonschema_description:"Indicates if the fields were truncated"`
	Queries         []DataObjectSearchItemQuery `json:"queries" jsonschema_description:"Queries associated with the data object"`
}

type DataObjectSearchItemField struct {
	Name         string  `json:"name" jsonschema_description:"Name of the field"`
	Type         string  `json:"type" jsonschema_description:"Type of the field"`
	Description  string  `json:"description" jsonschema_description:"Description of the field"`
	HugrType     string  `json:"hugr_type" jsonschema_description:"Hugr type of the field"`
	IsPrimaryKey bool    `json:"is_primary_key" jsonschema_description:"Indicates if the field is a primary key"`
	IsList       bool    `json:"is_list" jsonschema_description:"Indicates if the field is a list"`
	IsNotNull    bool    `json:"is_not_null" jsonschema_description:"Indicates if the field is not null"`
	Score        float64 `json:"score" jsonschema_description:"Search score of the field"`
}

type DataObjectSearchItemQuery struct {
	Name        string                    `json:"name" jsonschema_description:"Name of the query"`
	Description string                    `json:"description" jsonschema_description:"Description of the query"`
	Type        string                    `json:"type" jsonschema_description:"Type of the query"`
	Returns     string                    `json:"returns" jsonschema_description:"Return type of the query"`
	IsList      bool                      `json:"is_list" jsonschema_description:"Indicates if the query returns a list"`
	Args        []DataObjectSearchItemArg `json:"args" jsonschema_description:"Arguments of the query"`
}

type DataObjectSearchItemArg struct {
	Name        string `json:"name" jsonschema_description:"Name of the argument"`
	Description string `json:"description" jsonschema_description:"Description of the argument"`
	Type        string `json:"type" jsonschema_description:"Type of the argument"`
	IsList      bool   `json:"is_list" jsonschema_description:"Indicates if the argument is a list"`
	Required    bool   `json:"required" jsonschema_description:"Indicates if the argument is required"`
}

type SearchDataObjectsRequest struct {
	Module            string  `json:"module" jsonschema_description:"The name of the module to search within"`
	Query             string  `json:"query" jsonschema_description:"The natural-language query to search for relevant data objects"`
	FieldsQuery       string  `json:"fields_query" jsonschema_description:"The natural-language query to search for relevant fields within the data objects"`
	TopK              int     `json:"top_k" jsonschema_description:"The number of top relevant data objects to return" jsonschema:"minimum=1,default=5,maximum=50"`
	TopKField         int     `json:"top_k_field" jsonschema_description:"The number of top relevant fields to return for each data object" jsonschema:"minimum=1,default=5,maximum=50"`
	MinScore          float64 `json:"min_score" jsonschema_description:"Minimum relevance score threshold (between 0 and 1) to filter the data object results" jsonschema:"minimum=0,maximum=1,default=0"`
	MinFieldScore     float64 `json:"min_field_score" jsonschema_description:"Minimum relevance score threshold (between 0 and 1) to filter the field results within each data object" jsonschema:"minimum=0,maximum=1,default=0"`
	IncludeSubModules bool    `json:"include_submodules" jsonschema_description:"Whether to include data objects from submodules of the specified module"`
}

func (s *Service) SearchModuleDataObjects(ctx context.Context, req *SearchDataObjectsRequest) (*SearchResult[DataObjectSearchItem], error) {
	out := &SearchResult[DataObjectSearchItem]{}
	// 1. Get module queries (introspection)
	mm, err := s.modulesByNameCached(auth.CtxWithAdmin(ctx), req.Module, req.IncludeSubModules)
	if err != nil {
		return nil, fmt.Errorf("failed to get module by name pattern: %w", err)
	}
	if len(mm) == 0 {
		return out, nil
	}
	moduleIntroMap := make(map[string]*TypeIntro)
	for _, mod := range mm {
		if mod.QueryRoot == "" {
			continue
		}
		if _, ok := moduleIntroMap[mod.Name]; ok {
			continue
		}
		mti, err := s.typeIntroShort(ctx, mod.QueryRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get module query type intro for module %s: %w", mod.Name, err)
		}
		if mti == nil {
			return nil, fmt.Errorf("module query type %q not found for module %s", mod.QueryRoot, mod.Name)
		}
		moduleIntroMap[mod.Name] = mti
	}
	hq := searchDataObjectsRankQuery
	if s.c.EmbeddingsEnabled {
		hq = searchDataObjectsRankQueryWithEmbeddings
	}
	filter := map[string]any{
		"hugr_type": map[string]any{"in": []base.HugrType{base.HugrTypeTable, base.HugrTypeView}},
	}
	if !req.IncludeSubModules {
		filter["module"] = map[string]any{"eq": req.Module}
	}
	if req.IncludeSubModules {
		filter["_or"] = []map[string]any{
			{"module": map[string]any{"eq": req.Module}},
			{"module": map[string]any{"like": req.Module + ".%"}},
		}
	}
	// 2. Perform search data objects query
	res, err := s.h.Query(auth.CtxWithAdmin(ctx), hq, map[string]any{
		"filter":      filter,
		"query":       req.Query,
		"fieldsQuery": req.FieldsQuery,
		"ttl":         s.c.ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query data objects: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query data objects: %w", res.Err())
	}
	var items []searchDataObjectResult
	err = res.ScanData("core.mcp.types", &items)
	if errors.Is(err, types.ErrNoData) {
		return out, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode data objects: %w", err)
	}
	// 3. Fill results
	for _, item := range items {
		if len(item.DataObject) == 0 {
			continue
		}
		sdi := DataObjectSearchItem{
			Module:          item.Module,
			Name:            item.Name,
			Description:     item.Description,
			LongDescription: item.LongDescription,
			Type:            item.Type,
			Score:           1 - item.Score,
		}
		for _, q := range item.DataObject[0].Queries {
			if q.Field.McpExclude {
				continue
			}
			mti, ok := moduleIntroMap[item.Module]
			if !ok || mti == nil {
				// should not happen
				continue
			}
			// check if query exists in module query type
			if !slices.ContainsFunc(mti.Fields, func(f FieldIntro) bool {
				return f.Name == q.Name
			}) {
				continue
			}
			doq := DataObjectSearchItemQuery{
				Name:        q.Name,
				Description: q.Field.Description,
				Type:        q.QueryType,
				Returns:     q.Field.Type,
				IsList:      q.Field.IsList,
			}
			for _, a := range q.Field.Arguments {
				doq.Args = append(doq.Args, DataObjectSearchItemArg{
					Name:        a.Name,
					Description: a.Description,
					Type:        a.Type,
					IsList:      a.IsList,
					Required:    a.IsNotNull,
				})
			}
			sdi.Queries = append(sdi.Queries, doq)
		}
		if len(sdi.Queries) == 0 {
			// skip data objects without module queries
			continue
		}
		// get data object type intro
		dti, err := s.typeIntroShort(ctx, item.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get data object type intro: %w", err)
		}
		if dti == nil {
			continue
		}
		for _, field := range item.Fields {
			if field.McpExclude {
				continue
			}
			// check if field exists in data object type
			if !slices.ContainsFunc(dti.Fields, func(f FieldIntro) bool {
				return f.Name == field.Name
			}) {
				continue
			}
			if (1-field.Score) < req.MinFieldScore && req.MinFieldScore > 0 && req.FieldsQuery != "" && s.c.EmbeddingsEnabled {
				continue
			}
			sdi.Fields = append(sdi.Fields, DataObjectSearchItemField{
				Name:         field.Name,
				Type:         field.Type,
				Description:  field.Description,
				HugrType:     field.HugrType,
				IsPrimaryKey: field.IsPrimaryKey,
				IsList:       field.IsList,
				IsNotNull:    field.IsNotNull,
				Score:        1 - field.Score,
			})
		}
		// skip data objects without fields
		if len(sdi.Fields) == 0 {
			continue
		}
		// truncate fields
		if req.TopKField > 0 && len(sdi.Fields) > req.TopKField {
			sdi.Fields = sdi.Fields[:req.TopKField]
			sdi.FieldsTruncated = true
		}
		if (1-item.Score) < req.MinScore && req.MinScore > 0 && req.Query != "" && s.c.EmbeddingsEnabled {
			continue
		}
		out.Items = append(out.Items, sdi)
	}
	// 4. Set total
	out.Total = len(out.Items)
	// 5. Apply topK
	if req.TopK > 0 && len(out.Items) > req.TopK {
		out.Items = out.Items[:req.TopK]
	}

	return out, nil
}

const searchDataObjectsRankQuery = `query ranked_data_objects($filter: mcp_types_filter!, $ttl: Int!) {
  core{
    mcp{
      types(
        filter: $filter
        order_by:[{field: "name"}]
      ) @cache(ttl: $ttl) {
        name
		module
        description
        long_description
		type: hugr_type
        fields(
          nested_order_by: [{field: "name"}]
        ){
          name
          description
          type
          hugr_type
		  mcp_exclude
          is_primary_key
          is_list
          is_non_null
        }
        data_object{
          args_type_name
          queries{
            name
            query_root
            query_type
            field{
			  mcp_exclude
              description
              type
              is_list
              arguments{
                name
                description
                type
                is_non_null
                is_list
              }
            }
          }
        }
      }
    }
  }
}`

const searchDataObjectsRankQueryWithEmbeddings = `query ranked_data_objects($filter: mcp_types_filter!, $query: String!, $fieldsQuery: String!, $ttl: Int!) {
  core{
    mcp{
      types(
        filter: $filter
        order_by:[{field: "score"}]
      ) @cache(ttl: $ttl) {
        name
		module
        description
        long_description
		type: hugr_type
        score: _distance_to_query(query: $query)
        fields(
          nested_order_by: [{field: "score"}]
        ){
          name
          description
          type
          hugr_type
		  mcp_exclude
          is_primary_key
          is_list
          is_non_null
          score: _distance_to_query(query: $fieldsQuery)
        }
        data_object{
          args_type_name
          queries{
            name
            query_root
            query_type
            field{
              description
			  mcp_exclude
              type
              is_list
              arguments{
                name
                description
                type
                is_non_null
                is_list
              }
            }
          }
        }
      }
    }
  }
}`

type searchDataObjectResult struct {
	Name            string  `json:"name"`
	Module          string  `json:"module"`
	Description     string  `json:"description"`
	LongDescription string  `json:"long_description"`
	Type            string  `json:"type"`
	Score           float64 `json:"score"`
	Fields          []struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		Type         string  `json:"type"`
		HugrType     string  `json:"hugr_type"`
		McpExclude   bool    `json:"mcp_exclude"`
		IsPrimaryKey bool    `json:"is_primary_key"`
		IsList       bool    `json:"is_list"`
		IsNotNull    bool    `json:"is_non_null"`
		Score        float64 `json:"score"`
	} `json:"fields"`
	DataObject []struct {
		ArgsTypeName string `json:"args_type_name"`
		Queries      []struct {
			Name      string `json:"name"`
			QueryRoot string `json:"query_root"`
			QueryType string `json:"query_type"`
			Field     struct {
				Description string `json:"description"`
				Type        string `json:"type"`
				IsList      bool   `json:"is_list"`
				McpExclude  bool   `json:"mcp_exclude"`
				Arguments   []struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Type        string `json:"type"`
					IsNotNull   bool   `json:"is_non_null"`
					IsList      bool   `json:"is_list"`
				} `json:"arguments"`
			} `json:"field"`
		} `json:"queries"`
	} `json:"data_object"`
}

func (s *Service) modulesByNameCached(ctx context.Context, module string, includeSubModules bool) ([]Module, error) {
	filter := map[string]any{
		"name": map[string]any{"eq": module},
	}
	if includeSubModules {
		filter = map[string]any{
			"_or": []map[string]any{
				filter,
				{"name": map[string]any{"like": module + ".%"}},
			},
		}
		if module == "" {
			filter = map[string]any{} // all modules
		}
	}
	filter["disabled"] = map[string]any{"eq": false}
	res, err := s.h.Query(ctx, `query ($filter: mcp_modules_filter!, $ttl: Int!) {
		core {
			mcp {
				modules(filter: $filter, order_by: [{field: "name" direction: DESC}]) @cache(ttl: $ttl) {
					name
					query_root
					mutation_root
					function_root
					mut_function_root
					is_summarized
				}
			}
		}
	}`, map[string]any{
		"filter": filter,
		"ttl":    s.c.ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get module by name: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get module by name: %w", res.Err())
	}
	var modules []Module
	if err := res.ScanData("core.mcp.modules", &modules); err != nil {
		return nil, fmt.Errorf("failed to decode module: %w", err)
	}
	return modules, nil

}
