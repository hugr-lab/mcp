package indexer

import (
	"context"
	"fmt"
)

// DataObjectQueriesInfo information about a data object queries
type DataObjectQueriesInfo struct {
	Name       string                `json:"name"`
	FilterType string                `json:"filter_type"`
	ArgsType   string                `json:"args_type"`
	Module     string                `json:"module"`
	Queries    []DataObjectQueryInfo `json:"queries,omitempty"`
}

type DataObjectQueryInfo struct {
	Name string `json:"name"`
	Type string `json:"query_type"`
}

func (s *Service) DataObjectQueriesInfo(ctx context.Context, objectName string) (*DataObjectQueriesInfo, error) {
	res, err := s.h.Query(ctx, `query ($name: String!, $ttl: Int!) {
		core {
			mcp {
				data_objects_by_pk(name: $name) @cache(ttl: $ttl) {
					name
					filter_type_name
					args_type_name
					type {
						name
						module
					}
					queries {
						name
						query_type
					}
				}
			}
		}
	}`, map[string]any{
		"name": objectName,
		"ttl":  s.c.ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("query data object queries info: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("query data object queries info: %w", res.Err())
	}

	var info struct {
		Name       string `json:"name"`
		FilterType string `json:"filter_type_name"`
		ArgsType   string `json:"args_type_name"`
		Type       struct {
			Name   string `json:"name"`
			Module string `json:"module"`
		} `json:"type"`
		Queries []DataObjectQueryInfo `json:"queries"`
	}
	err = res.ScanData("core.mcp.data_objects_by_pk", &info)
	if err != nil {
		return nil, fmt.Errorf("scan data object queries info: %w", err)
	}

	return &DataObjectQueriesInfo{
		Name:       info.Name,
		FilterType: info.FilterType,
		ArgsType:   info.ArgsType,
		Module:     info.Type.Module,
		Queries:    info.Queries,
	}, nil
}

func (s *Service) DataObjectFieldType(ctx context.Context, objectName, fieldName string) (string, error) {
	res, err := s.h.Query(ctx, `query ($objectName: String!, $fieldName: String!, $ttl: Int!) {
		core {
			mcp {
				data_objects_by_pk(name: $objectName) @cache(ttl: $ttl) {
					type {
						fields(filter: {name: {eq: $fieldName}}) {
							type
						}
					}
				}
			}
		}
	}`, map[string]any{
		"objectName": objectName,
		"fieldName":  fieldName,
		"ttl":        s.c.ttl,
	})
	if err != nil {
		return "", fmt.Errorf("query data object field type: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return "", fmt.Errorf("query data object field type: %w", res.Err())
	}

	var info struct {
		Type struct {
			Fields []struct {
				Type string `json:"type"`
			} `json:"fields"`
		} `json:"type"`
	}
	err = res.ScanData("core.mcp.data_objects_by_pk", &info)
	if err != nil {
		return "", fmt.Errorf("scan data object field type: %w", err)
	}
	if len(info.Type.Fields) == 0 {
		return "", fmt.Errorf("field %q not found in data object %q", fieldName, objectName)
	}

	return info.Type.Fields[0].Type, nil
}
