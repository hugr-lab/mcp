package service

import (
	"context"

	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/mark3labs/mcp-go/mcp"
)

var schemaTypeInfoTool = mcp.NewTool("schema-type_info",
	mcp.WithDescription("Return high-level metadata for a GraphQL type (kind, counts, description snippet) to decide whether further introspection is needed."),
	mcp.WithInputSchema[schemaTypeInfoInput](),
	mcp.WithOutputSchema[indexer.TypeInfo](),
)

type schemaTypeInfoInput struct {
	TypeName string `json:"type_name" jsonschema_description:"The name of the GraphQL type to get information about"`
	WithDesc bool   `json:"with_description,omitempty" jsonschema_description:"Whether to include the type description in the response"`
}

func (s *Service) schemaTypeInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &schemaTypeInfoInput{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	typeInfo, err := s.indexer.TypeIntrospection(ctx, input.TypeName)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get type info", err), nil
	}
	if typeInfo == nil {
		return mcp.NewToolResultError("type not found"), nil
	}

	if !input.WithDesc {
		typeInfo.DescriptionSnippet = ""
	}

	out := mcp.NewToolResultStructuredOnly(typeInfo)

	return out, nil
}

var schemaTypeFieldsTool = mcp.NewTool("schema-type_fields",
	mcp.WithDescription("Return fields of a GraphQL type or input, optionally pagination (limit/offset), ranked by relevance to a natural-language query. Rank is applied before pagination."),
	mcp.WithInputSchema[indexer.TypeFieldsRequest](),
	mcp.WithOutputSchema[indexer.SearchResult[indexer.TypeFieldInfo]](),
)

func (s *Service) schemaTypeFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &indexer.TypeFieldsRequest{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	fields, err := s.indexer.TypeFieldsIntrospection(ctx, input)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get type fields", err), nil
	}
	if fields == nil {
		return mcp.NewToolResultError("type not found"), nil
	}

	out := mcp.NewToolResultStructuredOnly(fields)

	return out, nil
}

var schemaEnumValuesTool = mcp.NewTool("schema-enum_values",
	mcp.WithDescription("Return enum values for a GraphQL enum type."),
	mcp.WithInputSchema[schemaEnumValuesInput](),
	mcp.WithOutputSchema[[]indexer.EnumValueInfo](),
)

type schemaEnumValuesInput struct {
	TypeName string `json:"type_name" jsonschema_description:"The name of the GraphQL enum type to get values for"`
}

func (s *Service) schemaEnumValuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	input := &schemaEnumValuesInput{}
	if err := request.BindArguments(input); err != nil {
		return mcp.NewToolResultErrorFromErr("invalid input", err), nil
	}

	values, err := s.indexer.EnumValuesIntrospection(ctx, input.TypeName)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get enum values", err), nil
	}
	if values == nil {
		return mcp.NewToolResultError("type not found"), nil
	}

	out := mcp.NewToolResultStructuredOnly(values)

	return out, nil
}
