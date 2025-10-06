package service

import (
	"context"

	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/mark3labs/mcp-go/mcp"
)

var discoveryModuleObjectsTool = mcp.NewTool("discovery-search_module_data_objects",
	mcp.WithDescription("Return top-K data objects relevant to a natural-language query within a specific module and its sub-modules if specified"),
	mcp.WithInputSchema[indexer.SearchDataObjectsRequest](),
	mcp.WithOutputSchema[indexer.SearchResult[indexer.DataObjectSearchItem]](),
)

func (s *Service) discoveryModuleObjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &indexer.SearchDataObjectsRequest{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	dataObjects, err := s.indexer.SearchModuleDataObjects(ctx, input)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to search module data objects", err), nil
	}

	out := mcp.NewToolResultStructuredOnly(dataObjects)

	return out, nil
}
