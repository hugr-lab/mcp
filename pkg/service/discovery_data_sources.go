package service

import (
	"context"
	"slices"

	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/mark3labs/mcp-go/mcp"
)

var discoveryDataSourcesTool = mcp.NewTool("discovery-search_data_sources",
	mcp.WithDescription("Return top-K data sources relevant to a natural-language query"),
	mcp.WithInputSchema[schemaSearchDataSourcesInput](),
	mcp.WithOutputSchema[indexer.SearchResult[indexer.DataSourceSearchItem]](),
)

type schemaSearchDataSourcesInput struct {
	Query    string  `json:"query" jsonschema_description:"The natural-language query to search for relevant data sources"`
	TopK     int     `json:"top_k" jsonschema_description:"The number of top relevant data sources to return" jsonschema:"minimum=1,default=5,maximum=50"`
	MinScore float64 `json:"min_score" jsonschema_description:"Minimum relevance score threshold (between 0 and 1) to filter the results" jsonschema:"minimum=0,maximum=1,default=0.3"`
}

// discoveryDataSourcesHandler handles the "discovery-search_data_sources" tool request.
// the user role should have read access to the mcp_core_data_sources table, with proper permissions (row level constraints).
func (s *Service) discoveryDataSourcesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &schemaSearchDataSourcesInput{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	dataSources, err := s.indexer.SearchDataSources(ctx, input.Query, input.TopK, 0)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to search data sources", err), nil
	}

	if input.MinScore > 0 {
		dataSources.Items = slices.DeleteFunc(dataSources.Items, func(ds indexer.DataSourceSearchItem) bool {
			return ds.Score < input.MinScore
		})
	}

	out := mcp.NewToolResultStructuredOnly(dataSources)

	return out, nil
}
