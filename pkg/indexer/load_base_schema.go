package indexer

import (
	"context"
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
)

// fillBaseSchema initial schema from Hugr
func (s *Service) fillBaseSchema(ctx context.Context) error {
	// 1. fetch schema types
	schema, err := s.fetchSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch schema: %w", err)
	}

	// 2. Clear db
	err = s.Clear(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear db: %w", err)
	}

	// 3. Add all types from schema
	// add unknown type
	err = s.AddType(ctx, Type{
		Name:        "Unknown",
		Description: "Unknown type",
		Kind:        "SCALAR",
	})
	if err != nil {
		return fmt.Errorf("failed to add unknown type: %w", err)
	}
	var ff []Field
	var aa []Argument
	am := map[string]struct{}{}
	for _, st := range schema.Types {
		t := Type{
			Name:        st.Name,
			Description: st.Description,
			Kind:        st.Kind,
			HugrType:    st.HugrType,
			Module:      st.Module,
			Catalog:     st.Catalog,
		}
		err := s.AddType(ctx, t)
		if err != nil {
			return fmt.Errorf("failed to add type %q: %w", st.Name, err)
		}
		fields := st.Fields
		if st.Kind == string(ast.InputObject) {
			fields = st.InputFields
		}
		for _, f := range fields {
			field := Field{
				Name:        f.Name,
				Description: f.Description,
				TypeName:    st.Name,
				HugrType:    f.HugrType,
				Catalog:     f.Catalog,
				Exclude:     f.Exclude,
				Type:        f.Type.TypeName(),
				IsList:      f.Type.IsList(),
				IsNotNull:   f.Type.IsNotNull(),
			}
			ff = append(ff, field)
			for _, a := range f.Args {
				key := fmt.Sprintf("%s.%s.%s", st.Name, f.Name, a.Name)
				if _, ok := am[key]; ok {
					return fmt.Errorf("duplicate argument %q in hugr schema", key)
				}
				am[key] = struct{}{}
				arg := Argument{
					Name:        a.Name,
					FieldName:   f.Name,
					TypeName:    st.Name,
					Description: a.Description,
					Type:        a.Type.TypeName(),
					IsList:      a.Type.IsList(),
					IsNotNull:   a.Type.IsNotNull(),
				}
				aa = append(aa, arg)
			}
		}
	}

	// 4. fields
	for _, f := range ff {
		err := s.AddField(ctx, f)
		if err != nil {
			return fmt.Errorf("failed to add field %q.%q: %w", f.TypeName, f.Name, err)
		}
	}
	// 5. arguments
	for _, a := range aa {
		err := s.AddArgument(ctx, a)
		if err != nil {
			return fmt.Errorf("failed to add argument %q: %w", a.Name, err)
		}
	}

	meta, err := s.fetchSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch schema summary: %w", err)
	}
	// 6. Modules
	for _, m := range meta.Modules() {
		err := s.AddModule(ctx, Module{
			Name:            m.Name,
			Description:     m.Description,
			QueryRoot:       m.QueryType,
			MutationRoot:    m.MutationType,
			FunctionRoot:    m.FunctionType,
			MutFunctionRoot: m.MutationFunctionType,
		})
		if err != nil {
			return fmt.Errorf("failed to add module %q: %w", m.Name, err)
		}
	}

	// 7. Data sources
	for _, ds := range meta.DataSources {
		err := s.AddDataSource(ctx, DataSource{
			Name:        ds.Name,
			Description: ds.Description,
			Type:        ds.Type,
			Prefix:      ds.Prefix,
			AsModule:    ds.AsModule,
			ReadOnly:    ds.ReadOnly,
		})
		if err != nil {
			return fmt.Errorf("failed to add data source %q: %w", ds.Name, err)
		}
	}

	// 8. Data objects
	for _, do := range meta.DataObjects() {
		m := meta.Module(do.Module)
		if m == nil || m.QueryType == "" {
			return fmt.Errorf("module %q not found for data object %q", do.Module, do.Name)
		}
		object := DataObject{
			Name:           do.Name,
			FilterTypeName: do.FilterType,
		}
		if do.Arguments != nil {
			object.ArgsTypeName = do.Arguments.Type
		}
		for _, q := range do.Queries {
			object.Queries = append(object.Queries, DataObjectQuery{
				Name:      q.Name,
				QueryType: string(q.Type),
				QueryRoot: m.QueryType,
			})
		}

		err := s.addDataObject(ctx, object)
		if err != nil {
			return fmt.Errorf("failed to add data object %q: %w", do.Name, err)
		}
	}

	return nil
}

func (s *Service) Clear(ctx context.Context) error {
	// 1. Clear types
	res, err := s.h.Query(ctx, `mutation {
		core {
			mcp {
				delete_arguments { success }
				delete_fields { success }
				delete_modules { success }
				delete_types { success }
				delete_data_sources { success }
				delete_data_object_queries { success }
				delete_data_objects { success }
			}
		}
	}`, nil)
	if err != nil {
		return fmt.Errorf("query clear types: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query clear types: %w", res.Err())
	}

	return nil
}

const addDataSourceMutation = `mutation ($input: mcp_data_sources_mut_input_data!) {
		core {
			mcp {
				insert_data_sources(data: $input) {
					name
				}
			}
		}
	}`

const addDataSourceMutationWithEmbedding = `mutation ($input: mcp_data_sources_mut_input_data!, $summary: String!) {
		core {
			mcp {
				insert_data_sources(data: $input, summary: $summary) {
					name
				}
			}
		}
	}`

func (s *Service) AddDataSource(ctx context.Context, ds DataSource) error {
	desc := ds.LongDescription
	if desc == "" {
		desc = ds.Description
	}
	query := addDataSourceMutation
	if s.c.EmbeddingsEnabled {
		query = addDataSourceMutationWithEmbedding
	}
	// 1. Add data source to Hugr
	res, err := s.h.Query(ctx, query, map[string]any{
		"input":   ds,
		"summary": desc,
	})
	if err != nil {
		return fmt.Errorf("query add data source: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query add data source: %w", res.Err())
	}
	return nil
}

const addModuleMutation = `mutation ($input: mcp_modules_mut_input_data!) {
		core {
			mcp {
				insert_modules(data: $input) {
					name
				}
			}
		}
	}`

const addModuleMutationWithEmbedding = `mutation ($input: mcp_modules_mut_input_data!, $summary: String!) {
		core {
			mcp {
				insert_modules(data: $input, summary: $summary) {
					name
				}
			}
		}
	}`

func (s *Service) AddModule(ctx context.Context, m Module) error {
	query := addModuleMutation
	if s.c.EmbeddingsEnabled && m.Description != "" {
		query = addModuleMutationWithEmbedding
	}
	// 1. Add module to Hugr
	res, err := s.h.Query(ctx, query, map[string]any{
		"input":   m,
		"summary": m.Description,
	})
	if err != nil {
		return fmt.Errorf("query add module: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query add module: %w", res.Err())
	}
	return nil
}

const addTypeMutation = `mutation ($input: mcp_types_mut_input_data!) {
		core {
			mcp {
				insert_types(data: $input) {
					name
				}
			}
		}
	}`

const addTypeMutationWithEmbedding = `mutation ($input: mcp_types_mut_input_data!, $summary: String!) {
		core {
			mcp {
				insert_types(data: $input, summary: $summary) {
					name
				}
			}
		}
	}`

func (s *Service) AddType(ctx context.Context, t Type) error {
	desc := t.Long
	if desc == "" {
		desc = t.Description
	}
	query := addTypeMutation
	if s.c.EmbeddingsEnabled && desc != "" {
		query = addTypeMutationWithEmbedding
	}
	// 1. Add type to Hugr
	res, err := s.h.Query(ctx, query, map[string]any{
		"input":   t,
		"summary": desc,
	})
	if err != nil {
		return fmt.Errorf("query add type: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query add type: %w", res.Err())
	}
	return nil
}

const addFieldMutation = `mutation ($input: mcp_fields_mut_input_data!) {
		core {
			mcp {
				insert_fields(data: $input) {
					name
				}
			}
		}
	}`

const addFieldMutationWithEmbedding = `mutation ($input: mcp_fields_mut_input_data!, $summary: String!) {
		core {
			mcp {
				insert_fields(data: $input, summary: $summary) {
					name
				}
			}
		}
	}`

func (s *Service) AddField(ctx context.Context, f Field) error {
	query := addFieldMutation
	if s.c.EmbeddingsEnabled && f.Description != "" {
		query = addFieldMutationWithEmbedding
	}
	// 1. Add field to Hugr
	res, err := s.h.Query(ctx, query, map[string]any{
		"input":   f,
		"summary": f.Description,
	})
	if err != nil {
		return fmt.Errorf("query add field: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query add field: %w", res.Err())
	}
	return nil
}

func (s *Service) AddArgument(ctx context.Context, a Argument) error {
	// 1. Add argument to Hugr
	res, err := s.h.Query(ctx, `mutation ($input: mcp_arguments_mut_input_data!) {
		core {
			mcp {
				insert_arguments(data: $input) {
					name
				}
			}
		}
	}`, map[string]any{
		"input": a,
	})
	if err != nil {
		return fmt.Errorf("query add argument: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query add argument: %w", res.Err())
	}
	return nil
}

func (s *Service) addDataObject(ctx context.Context, do DataObject) error {
	// 1. Add data object to Hugr
	res, err := s.h.Query(ctx, `mutation ($name: String!, $input: mcp_data_objects_mut_input_data!) {
		core {
			mcp {
				delete_data_object_queries (filter: { object_name: { eq: $name }}) { success }
				delete_data_objects (filter: { name: { eq: $name }}) { success }
				insert_data_objects(data: $input) {
					name
				}
			}
		}
	}`, map[string]any{
		"name":  do.Name,
		"input": do,
	})
	if err != nil {
		return fmt.Errorf("query add data object: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query add data object: %w", res.Err())
	}
	return nil
}
