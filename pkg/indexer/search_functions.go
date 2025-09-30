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

type FunctionSearchItem struct {
	Name        string                       `json:"name" jsonschema_description:"Name of the function"`
	Description string                       `json:"description" jsonschema_description:"Description of the function"`
	IsMutation  bool                         `json:"is_mutation" jsonschema_description:"Indicates if the function is a mutation"`
	IsList      bool                         `json:"is_list" jsonschema_description:"Indicates if the function returns a list"`
	Score       float64                      `json:"score" jsonschema_description:"Relevance score of the function for the search query"`
	Arguments   []FunctionSearchItemArgument `json:"arguments,omitempty" jsonschema_description:"Arguments of the function"`
	Returns     FunctionSearchItemResult     `json:"returns" jsonschema_description:"Return type of the function"`
}

type FunctionSearchItemArgument struct {
	Name        string `json:"name" jsonschema_description:"Name of the argument"`
	Description string `json:"description" jsonschema_description:"Description of the argument"`
	Type        string `json:"type" jsonschema_description:"Type of the argument"`
	IsList      bool   `json:"is_list" jsonschema_description:"Indicates if the argument is a list"`
	Required    bool   `json:"required" jsonschema_description:"Indicates if the argument is required"`
}

type FunctionSearchItemResult struct {
	TypeName       string                          `json:"type_name" jsonschema_description:"Name of the return type"`
	IsList         bool                            `json:"is_list" jsonschema_description:"Indicates if the function returns a list"`
	Fields         []FunctionSearchItemResultField `json:"fields,omitempty" jsonschema_description:"Fields of the return type"`
	TruncateFields bool                            `json:"fields_truncated" jsonschema_description:"Indicates if the fields list is truncated (not full)"`
}

type FunctionSearchItemResultField struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsList      bool   `json:"is_list"`
}

type SearchFunctionsRequest struct {
	Module            string  `json:"module" jsonschema_description:"The name of the module to search within"`
	Query             string  `json:"query" jsonschema_description:"The natural-language query to search for relevant functions"`
	TopK              int     `json:"top_k" jsonschema_description:"The number of top relevant functions to return" jsonschema:"minimum=1,default=5,maximum=50"`
	MinScore          float64 `json:"min_score" jsonschema_description:"Minimum relevance score threshold (between 0 and 1) to filter the results" jsonschema:"minimum=0,maximum=1,default=0"`
	IncludeSubModules bool    `json:"include_sub_modules" jsonschema_description:"Whether to include functions from sub-modules of the specified module" jsonschema:"default=false"`
	IncludeMutations  bool    `json:"include_mutations" jsonschema_description:"Whether to include mutation functions in the search results" jsonschema:"default=false"`
}

const maxReturnsFields = 10

func (s *Service) SearchModuleFunctions(ctx context.Context, req *SearchFunctionsRequest) (*SearchResult[FunctionSearchItem], error) {
	if req.TopK < 1 || req.TopK > 50 {
		req.TopK = 5
	}
	// get module
	mm, err := s.modulesByNameCached(auth.CtxWithAdmin(ctx), req.Module, req.IncludeSubModules)
	if err != nil {
		return nil, err
	}
	if mm == nil {
		return nil, fmt.Errorf("module %q not found", req.Module)
	}
	out := &SearchResult[FunctionSearchItem]{}
	moduleMap := make(map[string]*Module)
	moduleTypeMap := make(map[string]*TypeIntro)
	for _, m := range mm {
		if m.FunctionRoot == "" && (!req.IncludeMutations || m.MutationRoot == "") {
			// skip modules without functions
			continue
		}
		moduleMap[m.Name] = &m
		if m.FunctionRoot != "" {
			moduleTypeMap[m.FunctionRoot], err = s.typeIntroShort(ctx, m.FunctionRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to get function root type intro for module %s: %w", m.Name, err)
			}
		}
		if req.IncludeMutations && m.MutFunctionRoot != "" {
			moduleTypeMap[m.MutFunctionRoot], err = s.typeIntroShort(ctx, m.MutFunctionRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to get mutation root type intro for module %s: %w", m.Name, err)
			}
		}
	}

	q := searchModuleFunctionsQuery
	if s.c.EmbeddingsEnabled && req.Query != "" {
		q = searchModuleFunctionsQueryWithEmbedding
	}

	res, err := s.h.Query(auth.CtxWithAdmin(ctx), q, map[string]any{
		"filter": buildModuleFunctionFilter(req.Module, req.IncludeSubModules, req.IncludeMutations),
		"query":  req.Query,
		"ttl":    s.c.ttl,
		"ft":     base.HugrTypeFieldFunction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search functions for module %s: %w", req.Module, err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to search functions for module %s: %w", req.Module, res.Err())
	}

	var items []searchFunctionsResult
	err = res.ScanData("core.mcp.fields", &items)
	if errors.Is(err, types.ErrNoData) {
		return out, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode functions for module %s: %w", req.Module, err)
	}
	for _, it := range items {
		m, ok := moduleMap[it.RootType.Module]
		if !ok {
			// should not happen
			continue
		}
		isMutation := it.RootType.Name == m.MutFunctionRoot
		if isMutation && !req.IncludeMutations {
			continue
		}
		moduleFuncIntro := moduleTypeMap[m.FunctionRoot]
		moduleMutFuncIntro := moduleTypeMap[m.MutFunctionRoot]
		// check if function is accessible from the module
		if isMutation &&
			(moduleMutFuncIntro == nil ||
				!slices.ContainsFunc(moduleMutFuncIntro.Fields, func(f FieldIntro) bool {
					return f.Name == it.Name
				})) {
			continue
		}
		if !isMutation &&
			(moduleFuncIntro == nil ||
				!slices.ContainsFunc(moduleFuncIntro.Fields, func(f FieldIntro) bool {
					return f.Name == it.Name
				})) {
			continue
		}
		fr := FunctionSearchItem{
			Name:        it.Name,
			Description: it.Description,
			IsMutation:  isMutation,
			IsList:      it.IsList,
			Score:       1 - it.Score,
		}
		if fr.Score < req.MinScore {
			continue
		}
		for _, arg := range it.Arguments {
			fr.Arguments = append(fr.Arguments, FunctionSearchItemArgument{
				Name:        arg.Name,
				Description: arg.Description,
				Type:        arg.Type,
				IsList:      arg.IsList,
				Required:    arg.IsNotNull,
			})
		}
		if it.ReturnType.Name != "" {
			fr.Returns.TypeName = it.ReturnType.Name
			fr.Returns.IsList = it.IsList
			for _, f := range it.ReturnType.Fields {
				if f.McpExclude {
					continue
				}
				if len(fr.Returns.Fields) >= maxReturnsFields {
					fr.Returns.TruncateFields = true
					break
				}
				fr.Returns.Fields = append(fr.Returns.Fields, FunctionSearchItemResultField{
					Name:        f.Name,
					Description: f.Description,
					Type:        f.Type,
					IsList:      f.IsList,
				})
			}
		}
		out.Items = append(out.Items, fr)
	}
	out.Total = len(out.Items)
	if req.TopK > 0 && len(out.Items) > req.TopK {
		out.Items = out.Items[:req.TopK]
	}

	return out, nil
}

type searchFunctionsResult struct {
	TypeName    string  `json:"type_name"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	IsList      bool    `json:"is_list"`
	RootType    struct {
		Name   string `json:"name"`
		Module string `json:"module"`
	} `json:"root_type"`
	Arguments []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		IsList      bool   `json:"is_list"`
		IsNotNull   bool   `json:"is_non_null"`
	} `json:"arguments"`
	ReturnType struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Fields      []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			McpExclude  bool   `json:"mcp_exclude"`
			IsList      bool   `json:"is_list"`
			IsNotNull   bool   `json:"is_non_null"`
		} `json:"fields"`
	} `json:"field_type"`
}

const searchModuleFunctionsQuery = `query ($module: [mcp_types_filter!], $ttl: Int!, $ft: String!) {
  core {
    mcp {
      fields(
        filter: {
          hugr_type: {eq: $ft}
          mcp_exclude: {eq: false}
          root_type: { _or: $filter }
		}
		order_by: [{field: "name"}]
      ) @cache(ttl: $ttl) {
        name
		type_name
        description
		root_type{
		  name
		  module
		}
        arguments{
          name
          description
          type
          is_non_null
          is_list
        }
        is_list
        field_type{
          name
          description
          fields{
            name
            description
			mcp_exclude
            type
            is_list
            is_non_null
          }
        }
      }
    }
  }
}`

const searchModuleFunctionsQueryWithEmbedding = `query ($filter: [mcp_types_filter!], $query: String!, $ttl: Int!, $ft: String!) {
  core {
    mcp {
      fields(
        filter: {
          hugr_type: {eq: $ft}
          mcp_exclude: {eq: false}
          root_type: { _or: $filter}
		}
		order_by: [{field: "score"}]
      ) @cache(ttl: $ttl) {
        name
		type_name
        description
		root_type{
		  name
		  module
		}
        arguments{
          name
          description
          type
          is_non_null
          is_list
        }
        is_list
		score: _distance_to_query(query: $query)
        field_type{
          name
          description
          fields{
            name
            description
			mcp_exclude
            type
            is_list
            is_non_null
          }
        }
      }
    }
  }
}`

func buildModuleFunctionFilter(module string, includeSubModules, includeMutations bool) []map[string]any {
	if includeSubModules && module == "" {
		if !includeMutations {
			return []map[string]any{
				{"module_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"is_null": false}}}},
			}
		}
		return []map[string]any{
			{"module_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"is_null": false}}}},
			{"module_mut_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"is_null": false}}}},
		}
	}
	filter := []map[string]any{
		{"module_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"eq": module}}}},
	}
	if includeMutations {
		filter = append(filter,
			map[string]any{"module_mut_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"eq": module}}}},
		)
	}
	if includeSubModules {
		filter = append(filter,
			map[string]any{"module_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"like": module + ".%"}}}},
		)
		if includeMutations {
			filter = append(filter,
				map[string]any{"module_mut_functions": map[string]any{"all_of": map[string]any{"name": map[string]any{"like": module + ".%"}}}},
			)
		}
	}
	return filter
}
