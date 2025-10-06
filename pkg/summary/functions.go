package summary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hugr-lab/mcp/pkg/pool"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"

	_ "embed"
)

//go:embed templates/function.tmpl
var functionInputTmpl string

func (s *Service) SummarizeFunction(ctx context.Context, schema *metainfo.SchemaInfo, function *metainfo.FunctionInfo) (*FunctionSummary, error) {
	// use function.tmpl to generate input
	input, err := prepareFunctionInput(ctx, schema, function)
	if err != nil {
		return nil, err
	}

	data := FunctionDescribeTemplateData{
		Name:         fmt.Sprintf("%q", input.Name),
		Description:  fmt.Sprintf("%q", input.Description),
		ReturnType:   fmt.Sprintf("%q", input.ReturnType),
		ReturnsArray: input.ReturnsArray,
	}
	b, err := json.Marshal(input.Parameters)
	if err != nil {
		return nil, err
	}
	data.ParametersJSON = string(b)
	b, err = json.Marshal(input.ReturnTypeFields)
	if err != nil {
		return nil, err
	}
	data.ReturnedFieldsJSON = string(b)
	b, err = json.Marshal(input.DataSourceContext)
	if err != nil {
		return nil, err
	}
	data.DataSourceContextJSON = string(b)
	b, err = json.Marshal(input.ModuleContext)
	if err != nil {
		return nil, err
	}
	data.ModuleContextJSON = string(b)

	c, err := s.pool.Connection(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	resp, err := c.Summarize(ctx, &pool.SummarizationTask{
		SystemPrompt:       systemPrompt,
		UserPromptTemplate: functionInputTmpl,
		Data:               data,
		MaxTokens:          4096,
		Temperature:        0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to summarize function %s: %w", function.Name, err)
	}

	var summary FunctionSummary
	if err := json.Unmarshal([]byte(resp), &summary); err != nil {
		return nil, errors.Join(ErrSummarizationOutputFormat, err)
	}

	return &summary, nil
}

type FunctionSummary struct {
	Short      string                 `json:"short"`
	Long       string                 `json:"long,omitempty"`
	Parameters map[string]string      `json:"parameters,omitempty"`
	Returns    FunctionReturnsSummary `json:"returns,omitempty"`
}

type FunctionReturnsSummary struct {
	Short  string            `json:"short"`
	Fields map[string]string `json:"fields,omitempty"`
}

func prepareFunctionInput(ctx context.Context, schema *metainfo.SchemaInfo, function *metainfo.FunctionInfo) (*FunctionInfoInput, error) {
	input := &FunctionInfoInput{
		Name:             function.Name,
		Description:      function.Description,
		Parameters:       function.Arguments,
		ReturnType:       function.ReturnType,
		ReturnsArray:     function.ReturnsArray,
		ReturnTypeFields: functionReturnedFields(function.ReturnTypeFields),
	}

	ds := schema.DataSource(function.DataSource)
	if ds == nil {
		if !strings.HasPrefix(function.Module, "core") {
			return nil, fmt.Errorf("data source not found: %s", function.DataSource)
		}
		ds = &metainfo.DataSourceInfo{
			Name:        function.DataSource,
			Description: "Core built-in data source for managed hugr",
		}
	}
	input.DataSourceContext = DataSourceContext{
		Name:        ds.Name,
		SummaryText: ds.Description,
	}
	m := schema.Module(function.Module)
	if m == nil {
		return nil, fmt.Errorf("module not found: %s", function.Module)
	}
	input.ModuleContext = ModuleContext{
		Name:     m.Name,
		Overview: m.Description,
	}

	return input, nil
}

func functionReturnedFields(fields []metainfo.FieldInfo) []FieldInfo {
	var out []FieldInfo
	for _, f := range fields {
		field := FieldInfo{
			Name:         f.Name,
			Description:  f.Description,
			Type:         f.Type,
			IsPrimaryKey: f.IsPrimaryKey,
			IsArray:      f.ReturnsArray,
		}
		for _, extra := range f.ExtraFields {
			ef := fmt.Sprintf("%s: %s", extra.Name, extra.Type)
			if extra.Description != "" {
				ef += " (" + extra.Description + ")"
			}
			field.ExtraFields = append(field.ExtraFields, ef)
		}
		if len(f.NestedFields) != 0 {
			field.NestedFields = functionReturnedFields(f.NestedFields)
		}
		out = append(out, field)
	}
	return out
}

type FunctionInfoInput struct {
	Name             string                  `json:"name"`
	Description      string                  `json:"description,omitempty"`
	Parameters       []metainfo.ArgumentInfo `json:"parameters,omitempty"`
	ReturnType       string                  `json:"return_type"`
	ReturnsArray     bool                    `json:"returns_array,omitempty"`
	ReturnTypeFields []FieldInfo             `json:"return_type_fields,omitempty"`

	DataSourceContext DataSourceContext `json:"data_source_context"`
	ModuleContext     ModuleContext     `json:"module_context"`
}

type FunctionDescribeTemplateData struct {
	Name                  string
	Description           string
	ParametersJSON        string
	ReturnType            string
	ReturnsArray          bool
	ReturnedFieldsJSON    string
	DataSourceContextJSON string
	ModuleContextJSON     string
}
