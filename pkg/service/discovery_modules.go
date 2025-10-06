package service

import (
	"context"
	"slices"

	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/mark3labs/mcp-go/mcp"
)

var discoveryModulesTool = mcp.NewTool("discovery-search_modules",
	mcp.WithDescription("Return top-K modules relevant to a natural-language query"),
	mcp.WithInputSchema[schemaSearchModulesInput](),
	mcp.WithOutputSchema[indexer.SearchResult[indexer.ModuleRanked]](),
)

type schemaSearchModulesInput struct {
	Query    string  `json:"query" jsonschema_description:"The natural-language query to search for relevant modules"`
	TopK     int     `json:"top_k" jsonschema_description:"The number of top relevant modules to return" jsonschema:"minimum=1,default=5,maximum=50"`
	MinScore float64 `json:"min_score" jsonschema_description:"Minimum relevance score threshold (between 0 and 1) to filter the results" jsonschema:"minimum=0,maximum=1,default=0.3"`
}

// discoveryModulesHandler handles the "discovery-search_modules" tool request.
// the user role should have read access to the mcp_core_modules table, with proper permissions (row level constraints).
func (s *Service) discoveryModulesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &schemaSearchModulesInput{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	modules, err := s.indexer.SearchModules(ctx, input.Query, input.TopK, 0)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to search modules", err), nil
	}

	if input.MinScore > 0 {
		modules.Items = slices.DeleteFunc(modules.Items, func(m indexer.ModuleRanked) bool {
			return m.Score < input.MinScore
		})
	}

	out := mcp.NewToolResultStructuredOnly(modules)

	return out, nil
}
