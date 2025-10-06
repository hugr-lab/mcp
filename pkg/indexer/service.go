package indexer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hugr-lab/mcp/pkg/pool"
	hugr "github.com/hugr-lab/query-engine"
	csources "github.com/hugr-lab/query-engine/pkg/catalogs/sources"
	"github.com/hugr-lab/query-engine/pkg/data-sources/sources"
	"github.com/hugr-lab/query-engine/pkg/db"
	"github.com/hugr-lab/query-engine/pkg/types"

	_ "embed"
)

//go:embed schema.graphql
var hschema string

const dbVersion = "0.0.1"
const dataSourceName = "core.mcp"

type Config struct {
	Path       string
	ReadOnly   bool
	VectorSize int

	SummarizeSchema bool
	Summarize       pool.Config

	EmbeddingsEnabled bool
	EmbeddingModel    string

	CacheTTL time.Duration
	ttl      int // cache ttl in seconds
}

// Indexed storage for hugr schema
type Service struct {
	c Config
	h *hugr.Client

	is_init bool
	loaded  bool // types are loaded
}

func New(config Config, h *hugr.Client) *Service {
	if config.CacheTTL != 0 {
		config.ttl = int(config.CacheTTL.Seconds())
	}
	return &Service{
		c: config,
		h: h,
	}
}

func (s *Service) Init(ctx context.Context) error {
	res, err := s.h.Query(ctx, `query ($ds: String!) {
		core {
			data_sources_by_pk (name: $ds) {
				name
				description
			}
		}
	}`, map[string]interface{}{
		"ds": dataSourceName,
	})
	if err != nil {
		return fmt.Errorf("query init: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query init: %w", res.Err())
	}
	var ds types.DataSource
	err = res.ScanData("core.data_sources_by_pk", &ds)
	if errors.Is(err, types.ErrNoData) {
		return s.registerDataSource(ctx)
	}
	if err != nil {
		return err
	}
	// update catalog
	err = s.updateMCPDataSource(ctx)
	if err != nil {
		return err
	}
	// check db version
	err = s.checkDBVersion(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Initialize data sources
func (s *Service) registerDataSource(ctx context.Context) error {
	if strings.HasPrefix(s.c.Path, "s3://") {
		s.c.ReadOnly = true
	}
	var dsType types.DataSourceType
	var dbType db.ScriptDBType
	switch {
	case strings.HasPrefix(s.c.Path, "postgres://"):
		dsType = sources.Postgres
		dbType = db.SDBPostgres
	default:
		dsType = sources.DuckDB
		dbType = db.SDBDuckDB
	}
	schema, err := db.ParseSQLScriptTemplate(dbType, hschema, dbInitParams{
		DBVersion:         dbVersion,
		VectorSize:        s.c.VectorSize,
		EmbeddingsEnabled: s.c.EmbeddingsEnabled,
		EmbeddingModel:    s.c.EmbeddingModel,
	})
	if err != nil {
		return fmt.Errorf("parse GraphQL schema template: %w", err)
	}
	err = s.h.RegisterDataSource(ctx, types.DataSource{
		Name:        dataSourceName,
		Description: "Core MCP Data Source",
		Type:        dsType,
		Path:        s.c.Path,
		Prefix:      "mcp",
		AsModule:    true,
		ReadOnly:    s.c.ReadOnly,
		Sources: []types.CatalogSource{{
			Name:        dataSourceName,
			Description: "MCP internal data source",
			Type:        csources.TextSourceType,
			Path:        schema,
		}},
	})
	if err != nil {
		return fmt.Errorf("register data source: %w", err)
	}
	return s.h.LoadDataSource(ctx, dataSourceName)
}

// Update the catalog source with the latest schema
func (s *Service) updateMCPDataSource(ctx context.Context) error {
	var dbType db.ScriptDBType
	switch {
	case strings.HasPrefix(s.c.Path, "postgres://"):
		dbType = db.SDBPostgres
	default:
		dbType = db.SDBDuckDB
	}
	schema, err := db.ParseSQLScriptTemplate(dbType, hschema, dbInitParams{
		DBVersion:         dbVersion,
		VectorSize:        s.c.VectorSize,
		EmbeddingsEnabled: s.c.EmbeddingsEnabled,
		EmbeddingModel:    s.c.EmbeddingModel,
	})
	if err != nil {
		return fmt.Errorf("parse GraphQL schema template: %w", err)
	}
	res, err := s.h.Query(ctx, `mutation ($name: String!, $data: String!, $path: String!) {
		core {
			update_catalog_sources (filter: { name: { eq: $name }}, data: { path: $data }) {
				success
			}
			update_data_sources (filter: { name: { eq: $name }}, data: { path: $path }) {
				success
			}
		}
	}`, map[string]any{
		"name": dataSourceName,
		"data": schema,
		"path": s.c.Path,
	})
	if err != nil {
		return fmt.Errorf("query update catalog source: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query update catalog source: %w", res.Err())
	}
	return s.h.LoadDataSource(ctx, dataSourceName)
}

var ErrWrongDBVersion = errors.New("wrong db version")

func (s *Service) checkDBVersion(ctx context.Context) error {
	res, err := s.h.Query(ctx, `query mcp {
		core{
			mcp{
			version_info(limit: 1, order_by: [{field: "applied_at"}]){
				version
				applied_at
			}
			}
		}
	}`, nil)
	if err != nil {
		return fmt.Errorf("query db version: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("query db version: %w", res.Err())
	}

	var info []struct {
		Version   string `json:"version"`
		AppliedAt string `json:"applied_at"`
	}
	if err := res.ScanData("core.mcp.version_info", &info); err != nil {
		return fmt.Errorf("scan db version: %w", err)
	}
	if len(info) == 0 {
		return ErrWrongDBVersion
	}
	if dbVersion != info[0].Version {
		return ErrWrongDBVersion
	}

	return nil
}

type ModuleRanked struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	LongDescription string  `json:"long_description"`
	Score           float64 `json:"score"`
}

const modulesRankQuery = `query ($limit: Int!, $offset: Int!) {
		core {
			mcp {
				agg: modules_aggregation(filter: { disabled: { eq: false }}) {
					total: _rows_count
				}
				modules(limit: $limit, offset: $offset, filter: { disabled: { eq: false }}) {
					name
					description
					long_description
				}
			}
		}
	}`

const modulesRankQueryWithEmbedding = `query ($limit: Int!, $offset: Int!, $query: String!, $ttl: Int!) {
		core {
			mcp {
				agg: modules_aggregation(filter: { disabled: { eq: false }}) {
					total: _rows_count
				}
				modules(
					limit: $limit
					offset: $offset
					order_by: [
						{ field: "score" }
					]
					filter: { disabled: { eq: false } }
				) @cache(ttl: $ttl) {
					name
					description
					long_description
					score: _distance_to_query(query: $query)
				}
			}
		}
	}`

func (s *Service) SearchModules(ctx context.Context, query string, limit, offset int) (*SearchResult[ModuleRanked], error) {
	q := modulesRankQuery
	variables := map[string]any{
		"limit":  limit,
		"offset": offset,
	}
	if s.c.EmbeddingsEnabled && query != "" {
		q = modulesRankQueryWithEmbedding
		variables["query"] = query
		variables["ttl"] = s.c.ttl
	}

	res, err := s.h.Query(ctx, q, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query modules: %w", res.Err())
	}
	var out SearchResult[ModuleRanked]
	err = res.ScanData("core.mcp.agg", &out)

	err = res.ScanData("core.mcp.modules", &out.Items)
	if errors.Is(err, types.ErrNoData) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode modules: %w", err)
	}
	if !s.c.EmbeddingsEnabled || query == "" {
		return &out, nil
	}
	// convert distance to score
	for i := range out.Items {
		out.Items[i].Score = 1.0 - out.Items[i].Score
	}
	return &out, nil
}

type DataSourceSearchItem struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	LongDescription string  `json:"long_description"`
	Type            string  `json:"type"`
	ReadOnly        bool    `json:"read_only"`
	AsModule        bool    `json:"as_module"`
	Score           float64 `json:"score"`
}

const dataSourcesRankQuery = `query ($limit: Int!, $offset: Int!) {
		core {
			mcp {
				agg: data_sources_aggregation (filter: { disabled: { eq: false }}) {
					total: _rows_count
				}
				data_sources(limit: $limit, offset: $offset, filter: { disabled: { eq: false }}) {
					name
					description
					long_description
					type
					read_only
					as_module
				}
			}
		}
	}`

const dataSourcesRankQueryWithEmbedding = `query ($limit: Int!, $offset: Int!, $query: String!, $ttl: Int!) {
		core {
			mcp {
				agg: data_sources_aggregation (filter: { disabled: { eq: false }}) {
					total: _rows_count
				}
				data_sources(
					limit: $limit
					offset: $offset
					order_by: [
						{ field: "score" }
					]
					filter: { disabled: { eq: false } }
				) @cache(ttl: $ttl) {
					name
					description
					long_description
					type
					read_only
					as_module
					score: _distance_to_query(query: $query)
				}
			}
		}
	}`

type SearchResult[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

func (s *Service) SearchDataSources(ctx context.Context, query string, limit, offset int) (*SearchResult[DataSourceSearchItem], error) {
	q := dataSourcesRankQuery
	variables := map[string]any{
		"limit":  limit,
		"offset": offset,
	}
	if s.c.EmbeddingsEnabled && query != "" {
		q = dataSourcesRankQueryWithEmbedding
		variables["query"] = query
		variables["ttl"] = s.c.ttl
	}

	res, err := s.h.Query(ctx, q, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", res.Err())
	}

	var out SearchResult[DataSourceSearchItem]
	err = res.ScanData("core.mcp.agg", &out)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return nil, fmt.Errorf("failed to decode data sources: %w", err)
	}

	err = res.ScanData("core.mcp.data_sources", &out.Items)
	if errors.Is(err, types.ErrNoData) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode data sources: %w", err)
	}
	if !s.c.EmbeddingsEnabled || query == "" {
		return &out, nil
	}
	// convert distance to score
	for i := range out.Items {
		out.Items[i].Score = 1.0 - out.Items[i].Score
	}
	return &out, nil
}
