package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hugr-lab/mcp/pkg/auth"
	"github.com/hugr-lab/mcp/pkg/indexer"
	hugr "github.com/hugr-lab/query-engine"
	"github.com/mark3labs/mcp-go/server"
)

const (
	mcpServerName    = "hugr-mcp"
	mcpServerVersion = "v0.0.1"
)

type Config struct {
	URL          string
	Secret       string
	SecretHeader string
	TTL          time.Duration
	ttl          int // in seconds, for data queries

	Indexer indexer.Config
}

func (c *Config) hugrOptions() (opts []hugr.Option) {
	if c.Secret == "" && c.SecretHeader == "" {
		return
	}
	return append(opts, hugr.WithTransport(
		auth.New(auth.WithSecretHeaderName(c.SecretHeader))),
		hugr.WithApiKey(c.Secret),
	)
}

type Service struct {
	cfg Config

	hugr    *hugr.Client
	mcp     *server.MCPServer
	s       *server.StreamableHTTPServer
	indexer *indexer.Service
}

func New(cfg Config) *Service {
	hugr := hugr.NewClient(cfg.URL,
		cfg.hugrOptions()...,
	)
	indexer := indexer.New(cfg.Indexer, hugr)

	if cfg.TTL <= 0 {
		cfg.TTL = 60 * time.Second
	}
	cfg.ttl = int(cfg.TTL.Seconds())
	if cfg.ttl == 0 {
		cfg.ttl = 60
	}

	mcp := server.NewMCPServer(
		mcpServerName,
		mcpServerVersion,
		server.WithRecovery(),
		server.WithResourceRecovery(),
		server.WithToolCapabilities(false),
	)

	s := server.NewStreamableHTTPServer(mcp, server.WithStateLess(true))

	return &Service{cfg: cfg, hugr: hugr, mcp: mcp, s: s, indexer: indexer}
}

func (s *Service) Init(ctx context.Context) error {
	// Initialize indexer
	if err := s.indexer.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize indexer: %w", err)
	}

	//s.mcp.AddTool(testTool, s.testToolHandler)
	s.mcp.AddTool(discoveryModulesTool, s.discoveryModulesHandler)
	s.mcp.AddTool(discoveryDataSourcesTool, s.discoveryDataSourcesHandler)
	s.mcp.AddTool(discoveryModuleObjectsTool, s.discoveryModuleObjectsHandler)
	s.mcp.AddTool(discoveryModuleFunctionsTool, s.discoveryModuleFunctionsHandler)
	s.mcp.AddTool(discoveryDataObjectFieldValuesTool, s.discoveryDataObjectFieldValuesHandler)
	s.mcp.AddTool(schemaTypeInfoTool, s.schemaTypeInfoHandler)
	s.mcp.AddTool(schemaTypeFieldsTool, s.schemaTypeFieldsHandler)
	s.mcp.AddTool(schemaEnumValuesTool, s.schemaEnumValuesHandler)
	s.mcp.AddTool(dataInlineGraphQLResultTool, s.dataInlineGraphQLResultHandler)

	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("MCP request: %s %s", r.Method, r.URL.Path)
	s.s.ServeHTTP(w, r)
}
