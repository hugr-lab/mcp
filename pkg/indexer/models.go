package indexer

import (
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	"github.com/hugr-lab/query-engine/pkg/types"
)

type DataSource struct {
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	LongDescription string        `json:"long_description"`
	Type            string        `json:"type"`
	Prefix          string        `json:"prefix"`
	AsModule        bool          `json:"as_module"`
	ReadOnly        bool          `json:"read_only"`
	Disabled        bool          `json:"disabled"`
	IsSummarized    bool          `json:"is_summarized"`
	Vec             *types.Vector `json:"vec,omitempty"`
}

type Type struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Long         string        `json:"long_description"`
	Kind         string        `json:"kind"`
	HugrType     base.HugrType `json:"hugr_type"`
	Catalog      string        `json:"catalog,omitempty"`
	Module       string        `json:"module"`
	IsSummarized bool          `json:"is_summarized"`
	Vec          *types.Vector `json:"vec,omitempty"`

	Fields []Field `json:"fields,omitempty"`
}

type HugrType string

type Module struct {
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	LongDescription string        `json:"long_description"`
	QueryRoot       string        `json:"query_root,omitempty"`
	MutationRoot    string        `json:"mutation_root,omitempty"`
	FunctionRoot    string        `json:"function_root,omitempty"`
	MutFunctionRoot string        `json:"mut_function_root,omitempty"`
	IsSummarized    bool          `json:"is_summarized"`
	Disabled        bool          `json:"disabled"`
	Vec             *types.Vector `json:"vec,omitempty"`
}

type Field struct {
	Name         string             `json:"name"`
	TypeName     string             `json:"type_name"`
	Description  string             `json:"description"`
	Type         string             `json:"type"`
	HugrType     base.HugrTypeField `json:"hugr_type"`
	Catalog      string             `json:"catalog,omitempty"`
	FieldType    *Type              `json:"field_type"`
	IsIndexed    bool               `json:"is_indexed"`
	IsList       bool               `json:"is_list"`
	IsNotNull    bool               `json:"is_non_null"`
	IsPrimaryKey bool               `json:"is_primary_key"`
	Exclude      bool               `json:"mcp_exclude"`
	IsSummarized bool               `json:"is_summarized"`
	Vec          *types.Vector      `json:"vec,omitempty"`

	Arguments []Argument `json:"arguments,omitempty"`
}

type HugrFieldType string

type Argument struct {
	Name         string `json:"name"`
	FieldName    string `json:"field_name"`
	TypeName     string `json:"type_name"`
	Description  string `json:"description"`
	DefaultValue string `json:"default_value"`
	Type         string `json:"type"`
	IsList       bool   `json:"is_list"`
	IsNotNull    bool   `json:"is_non_null"`
}

type HugrArgumentType string

type DataObject struct {
	Name           string            `json:"name"`
	FilterTypeName string            `json:"filter_type_name,omitempty"`
	ArgsTypeName   string            `json:"args_type_name,omitempty"`
	Queries        []DataObjectQuery `json:"queries,omitempty"`
}

type DataObjectQuery struct {
	Name       string `json:"name"`
	ObjectName string `json:"object_name,omitempty"`
	QueryRoot  string `json:"query_root"`
	QueryType  string `json:"query_type"`
}
