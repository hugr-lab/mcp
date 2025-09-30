package indexer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hugr-lab/mcp/pkg/summary"
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	"github.com/hugr-lab/query-engine/pkg/data-sources/sources"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
	"github.com/hugr-lab/query-engine/pkg/types"
)

func (s *Service) SummarizeDataSource(ctx context.Context, sum *summary.Service, name string) error {
	ds, err := s.dataSourceForSummary(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get data source for summary: %w", err)
	}
	if ds == nil {
		return errors.New("data source not found")
	}

	summary, err := sum.SummarizeDataSource(ctx, *ds)
	if err != nil {
		return fmt.Errorf("failed to summarize data source %s: %w", name, err)
	}

	// Update data source summary in the database
	err = s.UpdateDataSourceDescription(ctx, name, summary.Short, summary.Long, true)
	if err != nil {
		return fmt.Errorf("failed to update data source description: %w", err)
	}

	return nil
}

func (s *Service) dataSourcesForSummary(ctx context.Context) ([]string, error) {
	res, err := s.h.Query(ctx, `query ds($isSummarized: Boolean!) {
		core {
			mcp {
				data_sources(filter: {is_summarized: {eq: $isSummarized}}){
					name
				}
			}
		}
	}`, map[string]any{
		"isSummarized": false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", res.Err())
	}
	var dss []struct {
		Name string `json:"name"`
	}
	err = res.ScanData("core.mcp.data_sources", &dss)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data sources: %w", err)
	}
	if len(dss) == 0 {
		return nil, errors.New("no data sources to summarize")
	}
	var names []string
	for _, ds := range dss {
		if types.DataSourceType(ds.Name) != sources.Extension {
			names = append(names, ds.Name)
		}
	}
	return names, nil
}

func (s *Service) dataSourceForSummary(ctx context.Context, name string) (*summary.DataSource, error) {
	res, err := s.h.Query(ctx, `query dss($name: String!, $tt: String!, $vt: String!, $fnt: String!) {
		core{
			mcp{
				data_sources_by_pk(name: $name){
					name
					type
					description
					read_only
					as_module
					tables: types_in_catalog(filter: {hugr_type: {eq: $tt}}){
						name
						description
					}
					views: types_in_catalog(filter: {hugr_type: {eq: $vt}}){
						name
						description
					}
					functions: module_intro_in_catalog(
						distinct_on: ["module", "name"]
						filter: {hugr_type: {eq: $fnt}}
					){
						module
						name: field_name
						description: field_description
					}
				}
				modules(
					filter: {
						_or: [
							{ module_intro: { any_of:{ catalog: { eq: $name } } } }
							{ types_in_module: { any_of:{ catalog: { eq: $name } } } }
						]
					}
				){
					name
					description
				}
			}
		}
	}`, map[string]any{
		"name": name,
		"tt":   base.HugrTypeTable,
		"vt":   base.HugrTypeView,
		"fnt":  base.HugrTypeFieldFunction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query data source: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query data source: %w", res.Err())
	}
	var dss summary.DataSource
	err = res.ScanData("core.mcp.data_sources_by_pk", &dss)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data source: %w", err)
	}
	if len(dss.Tables) == 0 && len(dss.Views) == 0 && len(dss.Functions) == 0 {
		return nil, errors.New("no data objects to summarize")
	}

	err = res.ScanData("core.mcp.modules", &dss.Submodules)
	if errors.Is(err, types.ErrNoData) {
		log.Println("No data objects to summarize")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode modules: %w", err)
	}

	return &dss, nil
}

func (s *Service) dataSourcesByNames(ctx context.Context, names ...string) ([]metainfo.DataSourceInfo, error) {
	if len(names) == 0 {
		return nil, nil
	}
	res, err := s.h.Query(ctx, `query ds($names: [String!]!) {
		core {
			mcp {
				data_sources(filter: {name: {in: $names}}) {
					name
					description
					type
					as_module
					read_only
					is_summarized
				}
			}
		}
	}`, map[string]any{
		"names": names,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query data source: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query data source: %w", res.Err())
	}

	var ds []metainfo.DataSourceInfo
	err = res.ScanData("core.mcp.data_sources", &ds)
	if errors.Is(err, types.ErrNoData) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode data source: %w", err)
	}

	return ds, nil
}
