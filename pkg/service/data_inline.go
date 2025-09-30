package service

import (
	"context"

	hugr "github.com/hugr-lab/query-engine"
	"github.com/hugr-lab/query-engine/pkg/types"
	"github.com/mark3labs/mcp-go/mcp"
)

var dataInlineGraphQLResultTool = mcp.NewTool("data-inline_graphql_result",
	mcp.WithDescription("Execute a GraphQL query (optionally apply a jq transform) and inline a small JSON result directly in the response. Useful for dynamic data fetching within a single request. Limited by size"),
	mcp.WithInputSchema[simpleGraphQLRequest](),
	mcp.WithOutputSchema[map[string]any](),
)

type simpleGraphQLRequest struct {
	OperationName string         `json:"operation_name,omitempty" jsonschema_description:"Optional name of the GraphQL operation to execute if the query contains multiple operations."`
	Query         string         `json:"query" jsonschema_description:"The GraphQL query to execute. Should be a read-only query (no mutations). The query should not contain any sensitive information, as it may be logged or cached."`
	Variables     map[string]any `json:"variables,omitempty" jsonschema_description:"Optional variables to pass to the GraphQL query. The variables should be a JSON object that represents the GraphQL variables input for the query."`
	JQTransform   string         `json:"jq_transform,omitempty" jsonschema_description:"Optional jq transform to apply to the JSON result of the GraphQL query. The transform should be a valid jq expression that transforms the JSON result into a smaller JSON object. For example, to extract a list of names from a list of users, you might use: .data.users | map(.name)" jsonschema:"default="`
	MaxResultSize int            `json:"max_result_size,omitempty" jsonschema_description:"The maximum size (in bytes) of the JSON result after applying the jq transform. If the result exceeds this size, an error will be returned. This is to prevent inlining excessively large results. Default is 1000 bytes." jsonschema:"minimum=100,maximum=5000,default=1000"`
}

type simpleGraphQLResponse struct {
	IsTruncated bool `json:"is_truncated" jsonschema_description:"Whether the result was truncated due to exceeding the maximum size"`
	Size        int  `json:"original_size" jsonschema_description:"The size (in bytes) of the original JSON result before truncation"`
	Response    any  `json:"data,omitempty" jsonschema_description:"The JSON result of the GraphQL query after applying the jq transform, if any. May be omitted as not full JSON result if it was truncated."`
}

func (s *Service) dataInlineGraphQLResultHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &simpleGraphQLRequest{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}
	if input.Query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}
	if input.MaxResultSize <= 0 {
		input.MaxResultSize = 2 * 1024 * 1024 * 1024
	}

	res, err := s.hugr.QueryJSON(ctx, hugr.JQRequest{
		JQ: input.JQTransform,
		Query: types.Request{
			OperationName: input.OperationName,
			Query:         input.Query,
			Variables:     input.Variables,
		},
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to execute query", err), nil
	}
	if res == nil {
		return mcp.NewToolResultStructuredOnly(simpleGraphQLResponse{
			IsTruncated: false,
			Size:        0,
			Response:    nil,
		}), nil
	}
	out := simpleGraphQLResponse{
		IsTruncated: false,
		Size:        len(*res),
	}

	if input.MaxResultSize > 0 && len(*res) > input.MaxResultSize {
		out.IsTruncated = true
		out.Response = string((*res)[:input.MaxResultSize])
	} else {
		out.Response = res
	}

	return mcp.NewToolResultStructuredOnly(out), nil
}
