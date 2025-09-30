package indexer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hugr-lab/query-engine/pkg/types"
)

func (s *Service) IndexDataSources(ctx context.Context, summarized bool) error {
	filter := map[string]any{}
	if summarized {
		filter["is_summarized"] = map[string]any{"eq": true}
	}
	res, err := s.h.Query(ctx, `query ($filter: mcp_data_sources_filter) {
		core {
			mcp {
				data_sources(
					filter: $filter
					order_by: [{field: "name"}]
				) {
					name
					description
					long_description
				}
			}
		}
	}`, map[string]any{
		"filter": filter,
	})
	if err != nil {
		return fmt.Errorf("failed to query data sources for indexing: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to query data sources for indexing: %w", res.Err())
	}

	var dss []DataSource
	err = res.ScanData("core.mcp.data_sources", &dss)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return fmt.Errorf("failed to scan data sources for indexing: %w", err)
	}
	if len(dss) == 0 {
		fmt.Println("No data sources to index")
		return nil
	}

	for _, ds := range dss {
		summary := ds.LongDescription
		if summary == "" {
			summary = ds.Description
		}
		if summary == "" {
			summary = ds.Name
		}
		if err := s.IndexDataSource(ctx, ds.Name, summary); err != nil {
			log.Printf("failed to index data source %s: %v", ds.Name, err)
			continue
		}
		log.Printf("data source %s indexed", ds.Name)
	}
	return nil
}

func (s *Service) IndexDataSource(ctx context.Context, name, summary string) error {
	res, err := s.h.Query(ctx, `mutation updateDataSourceDescription($name: String!, $summary: String!) {
		core {
			mcp {
				update_data_sources(
					filter: { name: { eq: $name }}
					summary: $summary
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"name":    name,
		"summary": summary,
	})
	if err != nil {
		return fmt.Errorf("failed to update data source description: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update data source description: %w", res.Err())
	}
	return nil
}

func (s *Service) IndexModules(ctx context.Context, summarized bool) error {
	filter := map[string]any{}
	if summarized {
		filter["is_summarized"] = map[string]any{"eq": true}
	}
	res, err := s.h.Query(ctx, `query ($isSummarized: Boolean!) {
		core {
			mcp {
				modules(
					filter: {is_summarized: {eq: $isSummarized}}
					order_by: [{field: "name"}]
				){
					name
					description
				}
			}
		}
	}`, map[string]any{
		"isSummarized": summarized,
	})
	if err != nil {
		return fmt.Errorf("failed to query modules for indexing: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to query modules for indexing: %w", res.Err())
	}

	var mods []Module
	err = res.ScanData("core.mcp.modules", &mods)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return fmt.Errorf("failed to scan modules for indexing: %w", err)
	}
	if len(mods) == 0 {
		fmt.Println("No modules to index")
		return nil
	}

	for _, mod := range mods {
		summary := mod.Description
		if summary == "" {
			summary = mod.Name
		}
		if err := s.IndexModule(ctx, mod.Name, summary); err != nil {
			log.Printf("failed to index module %s: %v", mod.Name, err)
			continue
		}
		log.Printf("module %s indexed", mod.Name)
	}
	return nil
}

func (s *Service) IndexModule(ctx context.Context, name, desc string) error {
	long := desc
	if len(desc) > 1000 {
		long = desc[:1000]
	}
	res, err := s.h.Query(ctx, `mutation updateModuleDescription($name: String!, $desc: String!, $isSummarized: Boolean!) {
		core {
			mcp {
				update_modules(
					filter: { name: { eq: $name }}
					data: {
						description: $desc
						is_summarized: $isSummarized
					}
					summary: $desc
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"name":         name,
		"desc":         long,
		"isSummarized": true,
	})
	if err != nil {
		return fmt.Errorf("failed to update module description: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update module description: %w", res.Err())
	}
	return nil
}

func (s *Service) IndexFields(ctx context.Context, summarized bool) error {
	filter := map[string]any{}
	if summarized {
		filter["is_summarized"] = map[string]any{"eq": true}
	}
	res, err := s.h.Query(ctx, `query ($filter: mcp_fields_filter) {
		core {
			mcp {
				fields(
					filter: $filter
					order_by: [
						{field: "type_name"}
						{field: "name"}
					]
				) {
					name
					type_name
					description
				}
			}
		}
	}`, map[string]any{
		"filter": filter,
	})
	if err != nil {
		return fmt.Errorf("failed to query fields for indexing: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to query fields for indexing: %w", res.Err())
	}

	var fields []Field
	err = res.ScanData("core.mcp.fields", &fields)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return fmt.Errorf("failed to scan fields for indexing: %w", err)
	}
	if len(fields) == 0 {
		fmt.Println("No fields to index")
		return nil
	}

	for _, field := range fields {
		if field.Description == "" {
			field.Description = field.Name
		}
		if err := s.IndexField(ctx, field.TypeName, field.Name, field.Description); err != nil {
			log.Printf("failed to index field %s.%s: %v", field.TypeName, field.Name, err)
			continue
		}
		log.Printf("field %s.%s indexed", field.TypeName, field.Name)
	}
	return nil
}

func (s *Service) IndexField(ctx context.Context, typeName, fieldName, summary string) error {
	res, err := s.h.Query(ctx, `mutation updateFieldSummary($typeName: String!, $fieldName: String!, $summary: String!) {
		core {
			mcp {
				update_fields(
					filter: { 
						type_name: { eq: $typeName }
						name: { eq: $fieldName }
					}
					summary: $summary
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"typeName":  typeName,
		"fieldName": fieldName,
		"summary":   summary,
	})
	if err != nil {
		return fmt.Errorf("failed to update field summary: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update field summary: %w", res.Err())
	}
	return nil
}

func (s *Service) IndexTypes(ctx context.Context, summarized bool) error {
	filter := map[string]any{}
	if summarized {
		filter["is_summarized"] = map[string]any{"eq": true}
	}
	res, err := s.h.Query(ctx, `query ($filter: mcp_types_filter) {
		core {
			mcp {
				types(
					filter: $filter
					order_by: [
						{field: "module"}
						{field: "name"}
					]
				) {
					name
					module
					description
					long_description
				}
			}
		}
	}`, map[string]any{
		"filter": filter,
	})
	if err != nil {
		return fmt.Errorf("failed to query types for indexing: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to query types for indexing: %w", res.Err())
	}

	var tt []Type
	err = res.ScanData("core.mcp.types", &tt)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return fmt.Errorf("failed to scan types for indexing: %w", err)
	}
	if len(tt) == 0 {
		fmt.Println("No types to index")
		return nil
	}

	for _, t := range tt {
		summary := t.Long
		if summary == "" {
			summary = t.Description
		}
		if summary == "" {
			summary = t.Name
		}
		if err := s.IndexType(ctx, t.Name, summary); err != nil {
			log.Printf("failed to index type %s.%s: %v", t.Module, t.Name, err)
			continue
		}
		log.Printf("type %s.%s indexed", t.Module, t.Name)
	}
	return nil
}

func (s *Service) IndexType(ctx context.Context, name, summary string) error {
	res, err := s.h.Query(ctx, `mutation updateTypeDescription($name: String!, $summary: String!) {
		core {
			mcp {
				update_types(
					filter: { name: { eq: $name }}
					summary: $summary
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"name":    name,
		"summary": summary,
	})
	if err != nil {
		return fmt.Errorf("failed to update type description: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update type description: %w", res.Err())
	}
	return nil
}
