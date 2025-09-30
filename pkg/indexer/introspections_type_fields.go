package indexer

import (
	"context"
	"fmt"
	"slices"

	"github.com/hugr-lab/mcp/pkg/auth"
	"github.com/vektah/gqlparser/v2/ast"
)

type TypeFieldInfo struct {
	Name               string                  `json:"name"`
	TypeName           string                  `json:"type_name"`
	HugrType           string                  `json:"hugr_type,omitempty"`
	IsList             bool                    `json:"is_list,omitempty"`
	Nullable           bool                    `json:"nullable,omitempty"`
	ArgumentsCount     int                     `json:"arguments_count,omitempty"`
	Arguments          []TypeFieldArgumentInfo `json:"arguments,omitempty"`
	DescriptionSnippet string                  `json:"description_snippet,omitempty"`
	Score              float64                 `json:"score,omitempty"`
}

type TypeFieldArgumentInfo struct {
	Name        string `json:"name"`
	TypeName    string `json:"type_name"`
	Description string `json:"description"`
	IsList      bool   `json:"is_list,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type TypeFieldsRequest struct {
	TypeName           string  `json:"type_name" jsonschema_description:"The name of the type to get fields for" jsonschema:"required"`
	Limit              int     `json:"limit,omitempty" jsonschema_description:"Maximum number of fields to return" jsonschema:"minimum=1,maximum=100,default=20"`
	Offset             int     `json:"offset,omitempty" jsonschema_description:"Number of fields to skip (for pagination)" jsonschema:"minimum=0,default=0"`
	RelevanceQuery     string  `json:"relevance_query,omitempty" jsonschema_description:"Optional natural-language query to rank fields by relevance"`
	TopK               int     `json:"top_k,omitempty" jsonschema_description:"Number of top relevant fields to return when relevance_query is provided" jsonschema:"minimum=1,maximum=50,default=5"`
	MinScore           float64 `json:"min_score,omitempty" jsonschema_description:"Minimum relevance score threshold (between 0 and 1) to filter the results" jsonschema:"minimum=0,maximum=1,default=0.3"`
	IncludeDescription bool    `json:"include_description,omitempty" jsonschema_description:"Whether to include description snippets for each field" jsonschema:"default=false"`
	IncludeArguments   bool    `json:"include_arguments,omitempty" jsonschema_description:"Whether to include the count of arguments for each field" jsonschema:"default=false"`
}

func (s *Service) TypeFieldsIntrospection(ctx context.Context, req *TypeFieldsRequest) (*SearchResult[TypeFieldInfo], error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.RelevanceQuery != "" && req.TopK <= 0 {
		req.TopK = 10
	}

	typeInfo, err := s.typeIntroShort(ctx, req.TypeName)
	if err != nil {
		return nil, err
	}
	if typeInfo == nil {
		return &SearchResult[TypeFieldInfo]{Total: 0, Items: []TypeFieldInfo{}}, nil
	}

	q := typeFieldsInfoQuery
	if req.RelevanceQuery != "" && s.c.EmbeddingModel != "" {
		q = typeFieldsInfoWithEmbeddingQuery
	}
	res, err := s.h.Query(auth.CtxWithAdmin(ctx), q, map[string]any{
		"name":  req.TypeName,
		"ttl":   s.c.ttl,
		"query": req.RelevanceQuery,
	})
	if err != nil {
		return nil, fmt.Errorf("query type fields: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("query type fields: %w", res.Err())
	}

	var items []typeFieldInfoResult
	err = res.ScanData("core.mcp.fields", &items)
	if err != nil {
		return nil, fmt.Errorf("scan type fields: %w", err)
	}

	out := &SearchResult[TypeFieldInfo]{}

	for _, it := range items {
		// check field accessible
		if typeInfo.Kind == string(ast.Object) &&
			!slices.ContainsFunc(typeInfo.Fields, func(f FieldIntro) bool { return f.Name == it.Name }) {
			continue
		}
		if typeInfo.Kind == string(ast.InputObject) &&
			!slices.ContainsFunc(typeInfo.InputFields, func(f FieldIntro) bool { return f.Name == it.Name }) {
			continue
		}
		// limit by relevance score
		if req.RelevanceQuery != "" && req.MinScore > 0 && it.Score < 1-req.MinScore {
			continue
		}

		f := TypeFieldInfo{
			Name:           it.Name,
			TypeName:       it.Type,
			HugrType:       it.HugrType,
			IsList:         it.IsList,
			Nullable:       !it.IsNonNull,
			ArgumentsCount: len(it.Arguments),
			Score:          1 - it.Score,
		}

		if req.IncludeDescription {
			f.DescriptionSnippet = it.DescriptionSnippet
		}
		if req.IncludeArguments {
			for _, a := range it.Arguments {
				f.Arguments = append(f.Arguments, TypeFieldArgumentInfo{
					Name:        a.Name,
					TypeName:    a.Type,
					Description: a.Description,
					IsList:      a.IsList,
					Required:    a.IsNonNull,
				})
			}
		}
		out.Items = append(out.Items, f)
	}

	// apply offset/limit or topK
	if req.RelevanceQuery != "" && req.TopK > 0 {
		if len(out.Items) > req.TopK {
			out.Items = out.Items[:req.TopK]
		}
	}
	out.Total = len(out.Items)
	if req.Limit > 0 && req.Offset >= 0 {
		if req.Offset < len(out.Items) {
			out.Items = out.Items[req.Offset:]
		} else {
			out.Items = []TypeFieldInfo{}
		}
		if len(out.Items) > req.Limit {
			out.Items = out.Items[:req.Limit]
		}
	}

	return out, nil
}

type typeFieldInfoResult struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	HugrType     string `json:"hugr_type,omitempty"`
	IsList       bool   `json:"is_list,omitempty"`
	IsNonNull    bool   `json:"is_non_null,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key,omitempty"`
	Arguments    []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		IsList      bool   `json:"is_list,omitempty"`
		IsNonNull   bool   `json:"is_non_null,omitempty"`
	} `json:"arguments,omitempty"`
	DescriptionSnippet string  `json:"description_snippet,omitempty"`
	Score              float64 `json:"score,omitempty"`
}

const typeFieldsInfoQuery = `query ($name: String!, $ttl: Int!) {
	core {
		mcp {
			fields(
				filter: { 
					type_name: { eq: $name } 
					mcp_exclude: { eq: false }
				}
				order_by: [{ field: "name" }]
			) @cache(ttl: $ttl) {
				name
				description
				type
				hugr_type
				is_list
				is_non_null
				is_primary_key
				arguments {
					name
					type
					description
					is_list
					is_non_null
				}
			}
		}
	}
}`

const typeFieldsInfoWithEmbeddingQuery = `query ($name: String!, $query: String!, $ttl: Int!) {
	core {
		mcp {
			fields(
				order_by: [{ field: "score" }]
				filter: { 
					type_name: { eq: $name } 
					mcp_exclude: { eq: false }
				}
			) @cache(ttl: $ttl) {
				name
				description
				type
				hugr_type
				is_list
				is_non_null
				is_primary_key
				arguments {
					name
					type
					description
					is_list
					is_non_null
				}
				score: _distance_to_query(query: $query)
			}
		}
	}
}`
