package indexer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/hugr-lab/mcp/pkg/summary"
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
	"github.com/hugr-lab/query-engine/pkg/types"
)

func (s *Service) SummarizeModule(ctx context.Context, sum *summary.Service, name string) error {
	mm, err := s.moduleForSummary(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get module for summary: %w", err)
	}
	if mm == nil {
		return fmt.Errorf("module %s not found", name)
	}
	input := summary.ModuleInfo{
		Name:         mm.Name,
		Description:  mm.Description,
		Tables:       map[string]string{},
		Views:        map[string]string{},
		Functions:    map[string]string{},
		MutFunctions: map[string]string{},
		SubModules:   map[string]string{},
	}

	dsm := make(map[string]struct{})
	for _, t := range mm.Tables {
		input.Tables[t.Name] = t.Description
		if t.Catalog != "" {
			dsm[t.Catalog] = struct{}{}
		}
	}

	for _, v := range mm.Views {
		input.Views[v.Name] = v.Description
		if v.Catalog != "" {
			dsm[v.Catalog] = struct{}{}
		}
	}

	for _, f := range mm.Functions {
		switch f.Type {
		case "function":
			input.Functions[f.Name] = f.Description
		case "mut_function":
			input.MutFunctions[f.Name] = f.Description
		}
		if f.Catalog != "" {
			dsm[f.Catalog] = struct{}{}
		}
	}

	for _, sm := range mm.SubModules {
		input.SubModules[sm.Name] = sm.Description
	}

	// Data source
	dsNames := []string{mm.Name}
	for n := range dsm {
		dsNames = append(dsNames, n)
	}
	input.DataSources, err = s.dataSourcesByNames(ctx, dsNames...)
	if err != nil {
		return fmt.Errorf("failed to get data sources for module %s: %w", name, err)
	}
	input.DataSources = slices.DeleteFunc(input.DataSources, func(ds metainfo.DataSourceInfo) bool {
		_, ok := dsm[ds.Name]
		return ds.Name == mm.Name && !ds.AsModule && !ok
	})

	start := time.Now()
	log.Printf("module %s: summarization start", name)
	ms, err := sum.SummarizeModule(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to summarize module %s: %w", name, err)
	}
	log.Printf("module %s: summarization completed in %v", name, time.Since(start))

	// update module types
	if mm.QueryRoot != "" && ms.QueryType != "" {
		err = s.UpdateTypeDescription(ctx, mm.QueryRoot, ms.QueryType, "", true)
		if err != nil {
			return fmt.Errorf("failed to update query root type description: %w", err)
		}
	}
	if mm.MutationRoot != "" && ms.MutationType != "" {
		err = s.UpdateTypeDescription(ctx, mm.MutationRoot, ms.MutationType, "", true)
		if err != nil {
			return fmt.Errorf("failed to update mutation root type description: %w", err)
		}
	}
	if mm.FunctionRoot != "" && ms.FunctionType != "" {
		err = s.UpdateTypeDescription(ctx, mm.FunctionRoot, ms.FunctionType, "", true)
		if err != nil {
			return fmt.Errorf("failed to update function root type description: %w", err)
		}
	}
	if mm.MutFunctionRoot != "" && ms.MutFunctionType != "" {
		err = s.UpdateTypeDescription(ctx, mm.MutFunctionRoot, ms.MutFunctionType, "", true)
		if err != nil {
			return fmt.Errorf("failed to update mutation function root type description: %w", err)
		}
	}
	// update parent module fields
	pp := strings.Split(name, ".")
	if len(pp) > 1 {
		parentName := strings.Join(pp[:len(pp)-1], ".")
		parent, err := s.ModuleByName(ctx, parentName)
		if err != nil {
			return fmt.Errorf("failed to get parent module %s: %w", parentName, err)
		}
		if parent.QueryRoot != "" && ms.QueryType != "" {
			err = s.UpdateFieldDescription(ctx, parent.QueryRoot, pp[len(pp)-1], ms.QueryType, true)
			if err != nil {
				return fmt.Errorf("failed to update parent module %s field %s: %w", parentName, pp[len(pp)-1], err)
			}
		}
		if parent.MutationRoot != "" && ms.MutationType != "" {
			err = s.UpdateFieldDescription(ctx, parent.MutationRoot, pp[len(pp)-1], ms.MutationType, true)
			if err != nil {
				return fmt.Errorf("failed to update parent module %s field %s: %w", parentName, pp[len(pp)-1], err)
			}
		}
		if parent.FunctionRoot != "" && ms.FunctionType != "" {
			err = s.UpdateFieldDescription(ctx, parent.FunctionRoot, pp[len(pp)-1], ms.FunctionType, true)
			if err != nil {
				return fmt.Errorf("failed to update parent module %s field %s: %w", parentName, pp[len(pp)-1], err)
			}
		}
		if parent.MutFunctionRoot != "" && ms.MutFunctionType != "" {
			err = s.UpdateFieldDescription(ctx, parent.MutFunctionRoot, pp[len(pp)-1], ms.MutFunctionType, true)
			if err != nil {
				return fmt.Errorf("failed to update parent module %s field %s: %w", parentName, pp[len(pp)-1], err)
			}
		}
	}

	// update module description
	if ms.Short != "" || ms.Long != "" {
		err = s.UpdateModuleDescription(ctx, name, ms.Short, ms.Long, true)
		if err != nil {
			return fmt.Errorf("failed to update module %s description: %w", name, err)
		}
	}

	return nil
}

func (s *Service) modulesForSummary(ctx context.Context) ([]string, error) {
	res, err := s.h.Query(ctx, `query mm($isSummarized: Boolean!) {
		core {
			mcp {
				modules(
					filter: {is_summarized: {eq: $isSummarized}}
					order_by: [{field: "name"}]
				){
					name
				}
			}
		}
	}`, map[string]any{
		"isSummarized": false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query modules: %w", res.Err())
	}

	var modules []moduleForSummary
	err = res.ScanData("core.mcp.modules", &modules)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		return nil, fmt.Errorf("failed to decode modules: %w", err)
	}
	if len(modules) == 0 {
		return nil, errors.New("no modules to summarize")
	}
	var names []string
	for _, m := range modules {
		names = append(names, m.Name)
	}

	// sort modules by name to order from submodule to parent module (descending)
	sort.Slice(names, func(i, j int) bool {
		return names[i] > names[j]
	})

	return names, nil
}

func (s *Service) moduleForSummary(ctx context.Context, name string) (*moduleForSummary, error) {
	res, err := s.h.Query(ctx, `query mm($name: String!, $tt: String!, $vt: String!, $smt: String!, $fnt: String!) {
		core {
			mcp {
				modules_by_pk(name: $name) {
					name
					description
					long_description
					query_root
					mutation_root
					function_root
					mut_function_root
					tables: types_in_module(filter: {hugr_type: {eq: $tt}}){
						name
						description
						long_description
						hugr_type
						catalog
					}
					views: types_in_module(filter: {hugr_type: {eq: $vt}}){
						name
						description
						long_description
						hugr_type
						catalog
					}
					sub: module_intro(
						distinct_on: ["module","name"]
						filter:{
							hugr_type: {eq: $smt}
					}
					){
						module
						name: field_name
						description: field_description
						hugr_type
					}
					functions: module_intro(
						distinct_on: ["module","name"]
						filter:{
							hugr_type: {eq: $fnt}
						}
					){
						module
						name: field_name
						type_name
						type_type
						description: field_description
						hugr_type
						catalog
					}
				}
			}
		}
	}`, map[string]any{
		"name": name,
		"tt":   base.HugrTypeTable,
		"vt":   base.HugrTypeView,
		"smt":  base.HugrTypeFieldSubmodule,
		"fnt":  base.HugrTypeFieldFunction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query modules: %w", res.Err())
	}

	var module moduleForSummary
	err = res.ScanData("core.mcp.modules_by_pk", &module)
	if errors.Is(err, types.ErrNoData) {
		log.Println("No modules to summarize")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode module: %w", err)
	}

	return &module, nil
}

type moduleForSummary struct {
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	QueryRoot       string                       `json:"query_root"`
	MutationRoot    string                       `json:"mutation_root"`
	FunctionRoot    string                       `json:"function_root"`
	MutFunctionRoot string                       `json:"mut_function_root"`
	Tables          []moduleDataObjectForSummary `json:"tables"`
	Views           []moduleDataObjectForSummary `json:"views"`
	Functions       []moduleFunctionForSummary   `json:"functions"`
	SubModules      []moduleSubmoduleForSummary  `json:"sub"`
}

type moduleDataObjectForSummary struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	LongDescription string `json:"long_description"`
	HugrType        string `json:"hugr_type"`
	Catalog         string `json:"catalog"`
}

type moduleFunctionForSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type_type"`
	Catalog     string `json:"catalog"`
}

type moduleSubmoduleForSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Service) ModuleByTypeName(ctx context.Context, name string) (*Module, error) {
	res, err := s.h.Query(ctx, `query ($typeName: String!) {
		core {
			mcp {
				modules(filter: { _or: [
					{ query_root: { eq: $typeName } }
					{ mutation_root: { eq: $typeName } }
					{ function_root: { eq: $typeName } }
					{ mut_function_root: { eq: $typeName } }
				]}) {
					name
					query_root
					mutation_root
					function_root
					mut_function_root
					is_summarized
				}
			}
		}
	}`, map[string]any{
		"typeName": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get module by type name: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get module by type name: %w", res.Err())
	}
	var mm []Module
	if err := res.ScanData("core.mcp.modules", &mm); err != nil {
		return nil, fmt.Errorf("failed to decode module: %w", err)
	}
	if len(mm) == 0 {
		return nil, types.ErrNoData
	}
	return &mm[0], nil
}

func (s *Service) ModuleByName(ctx context.Context, name string) (*Module, error) {
	res, err := s.h.Query(ctx, `query ($name: String!) {
		core {
			mcp {
				modules_by_pk(name: $name) {
					name
					query_root
					mutation_root
					function_root
					mut_function_root
					is_summarized
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get module by name: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get module by name: %w", res.Err())
	}
	var module Module
	if err := res.ScanData("core.mcp.modules_by_pk", &module); err != nil {
		return nil, fmt.Errorf("failed to decode module: %w", err)
	}
	return &module, nil
}
