package indexer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hugr-lab/mcp/pkg/summary"
	"github.com/hugr-lab/query-engine/pkg/types"
	"golang.org/x/sync/errgroup"
)

// Methods to work with summary
func (s *Service) Summarize(ctx context.Context) error {
	// 1. Get meta summary
	meta, err := s.fetchSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch meta summary: %w", err)
	}

	sum := summary.New(s.c.Summarize)

	// 2. Summarize Data Objects
	objects, err := s.DataObjectTypesForSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to summarize data objects: %w", err)
	}
	eg, ctxg := errgroup.WithContext(ctx)
	eg.SetLimit(s.c.Summarize.MaxConnections)
	for _, t := range objects {
		eg.Go(func() error {
			log.Printf("data object %s: scheduling summarization", t.Name)
			err := s.SummarizeDataObject(ctxg, sum, meta, t)
			if err != nil {
				log.Printf("data object %s: failed to summarize data object: %v", t.Name, err)
			}
			log.Printf("data object %s: summarization scheduled", t.Name)
			return nil
		})
	}

	// 3. Summarize Functions
	functions, err := s.FunctionFieldsForSummary(ctxg)
	for _, f := range functions {
		eg.Go(func() error {
			log.Printf("function %s: scheduling summarization", f.Name)
			err := s.SummarizeFunction(ctxg, sum, meta, f)
			if err != nil {
				log.Printf("function %s: failed to summarize function: %v", f.Name, err)
			}
			log.Printf("function %s: summarization scheduled", f.Name)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to summarize data objects and functions: %w", err)
	}

	log.Println("Data objects and functions summarization completed")

	// 4. Summarize Data Sources
	eg, ctxg = errgroup.WithContext(ctx)
	eg.SetLimit(s.c.Summarize.MaxConnections)
	dataSources, err := s.dataSourcesForSummary(ctxg)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return fmt.Errorf("failed to summarize data sources: %w", err)
	}
	for _, ds := range dataSources {
		eg.Go(func() error {
			log.Printf("data source %s: scheduling summarization", ds)
			err := s.SummarizeDataSource(ctxg, sum, ds)
			if err != nil {
				log.Printf("data source %s: failed to summarize data source: %v", ds, err)
			}
			log.Printf("data source %s: summarization scheduled", ds)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to summarize modules: %w", err)
	}
	// 5. Summarize Modules
	eg, ctxg = errgroup.WithContext(ctx)
	eg.SetLimit(s.c.Summarize.MaxConnections)
	modules, err := s.modulesForSummary(ctxg)
	if errors.Is(err, types.ErrNoData) {
		log.Println("No modules to summarize")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to summarize modules: %w", err)
	}
	for _, m := range modules {
		eg.Go(func() error {
			log.Printf("module %s: scheduling summarization", m)
			err := s.SummarizeModule(ctxg, sum, m)
			if err != nil {
				log.Printf("module %s: failed to summarize module: %v", m, err)
			}
			log.Printf("module %s: summarization scheduled", m)
			return nil
		})
	}

	return eg.Wait()
}

const updateDataSourceDescriptionQuery = `mutation ($name: String!, $desc: String!, $long: String!, $isSummarized: Boolean!) {
		core {
			mcp {
				update_data_sources(
					filter: { name: { eq: $name }}
					data: {
						description: $desc
						long_description: $long
						is_summarized: $isSummarized
					}
				) {
					success
				}
			}
		}
	}`

const updateDataSourceDescriptionWithEmbeddingQuery = `mutation ($name: String!, $desc: String!, $long: String!, $isSummarized: Boolean!) {
		core {
			mcp {
				update_data_sources(
					filter: { name: { eq: $name }}
					data: {
						description: $desc
						long_description: $long
						is_summarized: $isSummarized
					}
					summary: $long
				) {
					success
				}
			}
		}
	}`

func (s *Service) UpdateDataSourceDescription(ctx context.Context, name, desc, long string, isSummarized bool) error {
	query := updateDataSourceDescriptionQuery
	if s.c.EmbeddingsEnabled {
		query = updateDataSourceDescriptionWithEmbeddingQuery
	}
	res, err := s.h.Query(ctx, query, map[string]any{
		"name":         name,
		"desc":         desc,
		"long":         long,
		"isSummarized": isSummarized,
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

const updateModuleDescriptionQuery = `mutation (
	$name: String!
	$desc: String!
	$long: String!
	$isSummarized: Boolean!
) {
	core {
		mcp {
			update_modules(
				filter: { name: { eq: $name }}
				data: {
					description: $desc
					is_summarized: $isSummarized
				}
			) {
				success
			}
		}
	}
}`

const updateModuleDescriptionWithEmbeddingQuery = `mutation (
	$name: String! 
	$desc: String! 
	$long: String!
	$summary: String!
	$isSummarized: Boolean!
) {
	core {
		mcp {
			update_modules(
				filter: { name: { eq: $name }}
				data: {
					description: $desc
					long_description: $long
					is_summarized: $isSummarized
				}
				summary: $summary
			) {
				success
			}
		}
	}
}`

func (s *Service) UpdateModuleDescription(ctx context.Context, name, desc, long string, isSummarized bool) error {
	summary := long
	if summary == "" {
		summary = desc
	}
	if len(summary) > 1000 {
		summary = summary[:1000]
	}
	query := updateModuleDescriptionQuery
	if s.c.EmbeddingsEnabled && summary != "" {
		query = updateModuleDescriptionWithEmbeddingQuery
	}
	res, err := s.h.Query(ctx, query, map[string]any{
		"name":         name,
		"desc":         desc,
		"long":         long,
		"summary":      summary,
		"isSummarized": isSummarized,
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

const updateTypeDescriptionQuery = `mutation ($name: String!, $desc: String!, $long: String!, $isSummarized: Boolean!) {
		core {
			mcp {
				update_types(
					filter: { name: { eq: $name }}
					data: {
						description: $desc
						long_description: $long
						is_summarized: $isSummarized
					}
				) {
					success
				}
			}
		}
	}`

const updateTypeDescriptionWithEmbeddingQuery = `mutation ($name: String!, $desc: String!, $long: String!, $summary: String!, $isSummarized: Boolean!) {
		core {
			mcp {
				update_types(
					filter: { name: { eq: $name }}
					data: {
						description: $desc
						long_description: $long
						is_summarized: $isSummarized
					}
					summary: $summary
				) {
					success
				}
			}
		}
	}`

func (s *Service) UpdateTypeDescription(ctx context.Context, name, desc, long string, isSummarized bool) error {
	query := updateTypeDescriptionQuery
	if s.c.EmbeddingsEnabled {
		query = updateTypeDescriptionWithEmbeddingQuery
	}
	summary := long
	if summary == "" {
		summary = desc
	}
	if len(summary) > 1000 {
		summary = summary[:1000]
	}

	res, err := s.h.Query(ctx, query, map[string]any{
		"name":         name,
		"desc":         desc,
		"long":         long,
		"summary":      summary,
		"isSummarized": isSummarized,
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

const updateFieldDescriptionQuery = `mutation ($typeName: String!, $fieldName: String!, $desc: String!, $summarized: Boolean) {
		core {
			mcp {
				update_fields(
					filter: { 
						type_name: { eq: $typeName }
						name: { eq: $fieldName }
					}
					data: {
						description: $desc
						is_summarized: $summarized
					}
				) {
					success
				}
			}
		}
	}`

const updateFieldDescriptionWithEmbeddingQuery = `mutation ($typeName: String!, $fieldName: String!, $desc: String!, $summarized: Boolean) {
		core {
			mcp {
				update_fields(
					filter: { 
						type_name: { eq: $typeName }
						name: { eq: $fieldName }
					}
					data: {
						description: $desc
						is_summarized: $summarized
					}
					summary: $desc
				) {
					success
				}
			}
		}
	}`

func (s *Service) UpdateFieldDescription(ctx context.Context, typeName, fieldName, desc string, summarized bool) error {
	query := updateFieldDescriptionQuery
	if s.c.EmbeddingsEnabled {
		query = updateFieldDescriptionWithEmbeddingQuery
	}
	res, err := s.h.Query(ctx, query, map[string]any{
		"typeName":   typeName,
		"fieldName":  fieldName,
		"desc":       desc,
		"summarized": summarized,
	})
	if err != nil {
		return fmt.Errorf("failed to update field description: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update field description: %w", res.Err())
	}
	return nil
}

func (s *Service) UpdateArgumentDescription(ctx context.Context, typeName, fieldName, argName, desc string) error {
	res, err := s.h.Query(ctx, `mutation ($typeName: String!, $fieldName: String!, $argName: String!, $desc: String!) {
		core {
			mcp {
				update_arguments(
					filter: { 
						type_name: { eq: $typeName }
						field_name: { eq: $fieldName }
						name: { eq: $argName }
					}
					data: {
						description: $desc
					}
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"typeName":  typeName,
		"fieldName": fieldName,
		"argName":   argName,
		"desc":      desc,
	})
	if err != nil {
		return fmt.Errorf("failed to update argument description: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update argument description: %w", res.Err())
	}
	return nil
}
