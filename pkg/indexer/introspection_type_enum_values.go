package indexer

import "context"

type EnumValueInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (s *Service) EnumValuesIntrospection(ctx context.Context, typeName string) ([]EnumValueInfo, error) {
	res, err := s.h.Query(ctx, `query enumValues($name: String!) {
		__type(name: $name) {
			enumValues {
				name
				description
			}
		}
	}`, map[string]any{
		"name": typeName,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, res.Err()
	}
	var data struct {
		EnumValues []EnumValueInfo `json:"enumValues"`
	}
	if err := res.ScanData("__type", &data); err != nil {
		return nil, err
	}
	return data.EnumValues, nil
}
