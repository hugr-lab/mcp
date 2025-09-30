package indexer

import (
	"context"
	"fmt"
	"slices"

	"github.com/hugr-lab/mcp/pkg/auth"
	"github.com/hugr-lab/query-engine/pkg/compiler"
	"github.com/vektah/gqlparser/v2/ast"
)

func (s *Service) TypeIntrospection(ctx context.Context, name string) (*TypeInfo, error) {
	// TODO: Implement type introspection logic
	ti, err := s.typeIntroShort(ctx, name)
	if err != nil {
		return nil, err
	}
	if ti == nil {
		return nil, nil
	}

	info := &TypeInfo{
		Name:            ti.Name,
		EnumValuesTotal: len(ti.EnumValues),
	}

	qi, err := s.typeInfo(auth.CtxWithAdmin(ctx), name)
	if err != nil && qi == nil {
		return nil, err
	}
	info.Kind = qi.Kind
	info.DescriptionSnippet = qi.Description
	if qi.LongDesc != "" {
		info.DescriptionSnippet = qi.LongDesc
	}
	if len(qi.DataObjectsFilter) > 0 {
		info.FilterForDataObject = qi.DataObjectsFilter[0].Name
	}
	info.Module = qi.Module
	info.HugrType = qi.HugrType
	info.Catalog = qi.Catalog
	for _, f := range qi.Fields {
		if f.McpExclude {
			continue
		}
		if qi.Kind == ast.Object && !slices.ContainsFunc(ti.Fields, func(ff FieldIntro) bool { return ff.Name == f.Name }) {
			continue
		}
		if qi.Kind == ast.InputObject && !slices.ContainsFunc(ti.InputFields, func(ff FieldIntro) bool { return ff.Name == f.Name }) {
			continue
		}
		if f.Type == compiler.GeometryTypeName && f.IsList != true {
			info.HasGeometryField = true
		}
		if len(f.Arguments) > 0 {
			info.HasFieldWithArguments = true
		}
		if qi.Kind == ast.Object {
			info.FieldsTotal++
		}
		if qi.Kind == ast.InputObject {
			info.InputFieldsTotal++
		}
	}

	return info, nil
}

func (s *Service) typeInfo(ctx context.Context, typeName string) (*typeQuickInfo, error) {
	res, err := s.h.Query(ctx, `query types($name: String!, $ttl: Int!) {
		core {
			mcp {
				types_by_pk(name: $name) @cache(ttl: $ttl) {
					name
					description
					long_description
					kind
					module
					hugr_type
					catalog
					filter_for_data_objects{
						name
					}
					fields {
						name
						type
						hugr_type
						mcp_exclude
						is_list
						arguments {
							name
						}
					}
				}
			}
		}
	}`, map[string]any{
		"name": typeName,
		"ttl":  s.c.ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("query type info: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("query type info: %w", res.Err())
	}

	var data typeQuickInfo
	err = res.ScanData("core.mcp.types_by_pk", &data)
	if err != nil {
		return nil, fmt.Errorf("scan type info: %w", err)
	}

	return &data, nil
}

type typeQuickInfo struct {
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	LongDesc          string             `json:"long_description"`
	Kind              ast.DefinitionKind `json:"kind"`
	Module            string             `json:"module,omitempty"`
	HugrType          string             `json:"hugr_type,omitempty"`
	Catalog           string             `json:"catalog,omitempty"`
	DataObjectsFilter []struct {
		Name string `json:"name"`
	} `json:"filter_for_data_objects,omitempty"`
	Fields []struct {
		Name       string `json:"name"`
		Type       string `json:"type"`
		HugrType   string `json:"hugr_type,omitempty"`
		IsList     bool   `json:"is_list,omitempty"`
		McpExclude bool   `json:"mcp_exclude,omitempty"`
		Arguments  []struct {
			Name string `json:"name"`
		} `json:"arguments,omitempty"`
	} `json:"fields,omitempty"`
}

type TypeInfo struct {
	Name                  string             `json:"name"`
	Kind                  ast.DefinitionKind `json:"kind"`
	Module                string             `json:"module,omitempty"`
	HugrType              string             `json:"hugr_type,omitempty"`
	Catalog               string             `json:"catalog,omitempty" jsonschema_description:"The data source this type belongs to"`
	FieldsTotal           int                `json:"fields_total,omitempty"`
	InputFieldsTotal      int                `json:"input_fields_total,omitempty"`
	EnumValuesTotal       int                `json:"enum_values_total,omitempty"`
	HasGeometryField      bool               `json:"has_geometry_field,omitempty"`
	HasFieldWithArguments bool               `json:"has_field_with_arguments,omitempty"`
	DescriptionSnippet    string             `json:"description_snippet,omitempty"`
	FilterForDataObject   string             `json:"filter_for_data_object,omitempty" jsonschema_description:"The name of the data object this type is used as a filter for"`
}
