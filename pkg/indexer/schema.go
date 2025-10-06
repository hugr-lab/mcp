package indexer

import (
	"context"
	"fmt"

	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
)

func (s *Service) fetchSummary(ctx context.Context) (*metainfo.SchemaInfo, error) {
	res, err := s.h.Query(ctx, `query {
		function{
			core{
				meta{
					schema_summary
				}
			}
		}
	}`, nil)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, res.Err()
	}
	var info metainfo.SchemaInfo
	err = res.ScanData("function.core.meta.schema_summary", &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (s *Service) fetchSchema(ctx context.Context) (*SchemaIntro, error) {
	res, err := s.h.Query(ctx, `query schema {
		__schema{
			description
			queryType{
				...type_info
			}
			mutationType{
				...type_info
			}
			types{
				...type_info
			}
		}
	}

	fragment type_info on __Type {
		name
		description
		kind
		hugr_type
		module
		catalog
		enumValues{
			name
			description
		}
		inputFields{
			name
			description
			defaultValue
			type{
				name
				kind
				ofType{
					name
					kind
					ofType{
						name
						kind
						ofType{
							name
							kind
						}
					}
				}
			}
		}
		fields{
			name
			description
			hugr_type
			catalog
			args{
				name
				description
				defaultValue
				type{
					name
					kind
					ofType{
						name
						kind
						ofType{
							name
							kind
							ofType{
								name
								kind
							}
						}
					}
				}
			}
			type{
				name
				kind
				ofType{
					name
					kind
					description
					ofType{
						name
						kind
						ofType{
							name
							kind
						}
					}
				}
			}
		}
	}`, nil)
	if err != nil {
		return nil, fmt.Errorf("query schema: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("query schema: %w", res.Err())
	}

	var schema SchemaIntro

	err = res.ScanData("__schema", &schema)
	if err != nil {
		return nil, fmt.Errorf("scan schema: %w", err)
	}

	return &schema, nil
}

func (s *Service) typeIntroShort(ctx context.Context, typeName string) (*TypeIntro, error) {
	res, err := s.h.Query(ctx, `query types($name: String!) {
		__type(name: $name) {
			name
			kind
			hugr_type
			module
			catalog
			description
			inputFields{ name }
			fields{
				name
				args{ name }
			}
			enumValues{ name }
		}
	}`, map[string]any{
		"name": typeName,
	})
	if err != nil {
		return nil, fmt.Errorf("query type: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("query type: %w", res.Err())
	}

	var ti TypeIntro

	err = res.ScanData("__type", &ti)
	if err != nil {
		return nil, fmt.Errorf("scan type: %w", err)
	}

	return &ti, nil
}

type SchemaIntro struct {
	Description  string      `json:"description"`
	QueryType    *TypeIntro  `json:"queryType"`
	MutationType *TypeIntro  `json:"mutationType,omitempty"`
	Types        []TypeIntro `json:"types"`
}

func (s *SchemaIntro) TypeByName(name string) *TypeIntro {
	for i := range s.Types {
		if s.Types[i].Name == name {
			return &s.Types[i]
		}
	}
	return nil
}

type TypeIntro struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Kind        string           `json:"kind"`
	HugrType    base.HugrType    `json:"hugr_type"`
	Module      string           `json:"module"`
	Catalog     string           `json:"catalog,omitempty"`
	InputFields []FieldIntro     `json:"inputFields"`
	Fields      []FieldIntro     `json:"fields"`
	EnumValues  []EnumIntroValue `json:"enumValues"`
}

type EnumIntroValue struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type FieldIntro struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	HugrType    base.HugrTypeField `json:"hugr_type"`
	Catalog     string             `json:"catalog,omitempty"`
	Exclude     bool               `json:"mcp_exclude"`
	Args        []ArgIntro         `json:"args"`
	Type        TypeRefIntro       `json:"type"`
}

type ArgIntro struct {
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	DefaultValue string       `json:"defaultValue"`
	Type         TypeRefIntro `json:"type"`
}

type TypeRefIntro struct {
	Name   string        `json:"name"`
	Kind   string        `json:"kind"`
	OfType *TypeRefIntro `json:"ofType,omitempty"`
}

func (t *TypeRefIntro) TypeName() string {
	if t == nil {
		return "Unknown"
	}
	if t.Name == "" && t.OfType != nil {
		return t.OfType.TypeName()
	}
	if t.Name == "" {
		return "Unknown"
	}
	return t.Name
}

func (t *TypeRefIntro) IsList() bool {
	if t == nil {
		return false
	}
	if t.Kind == "LIST" {
		return true
	}
	if t.Name == "" && t.OfType != nil {
		return t.OfType.IsList()
	}
	return false
}

func (t *TypeRefIntro) IsNotNull() bool {
	if t == nil {
		return false
	}
	if t.Kind == "NON_NULL" {
		return true
	}
	return false
}
