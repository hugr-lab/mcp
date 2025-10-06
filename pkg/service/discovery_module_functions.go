package service

import (
	"context"

	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/mark3labs/mcp-go/mcp"
)

var discoveryModuleFunctionsTool = mcp.NewTool("discovery-search_module_functions",
	mcp.WithDescription("Return top-K module functions relevant to a natural-language query"),
	mcp.WithInputSchema[indexer.SearchFunctionsRequest](),
	mcp.WithOutputSchema[indexer.SearchResult[indexer.FunctionSearchItem]](),
)

func (s *Service) discoveryModuleFunctionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &indexer.SearchFunctionsRequest{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	functions, err := s.indexer.SearchModuleFunctions(ctx, input)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to search module functions", err), nil
	}

	out := mcp.NewToolResultStructuredOnly(functions)

	return out, nil
}
