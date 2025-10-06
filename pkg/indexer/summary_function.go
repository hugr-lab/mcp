package indexer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hugr-lab/mcp/pkg/summary"
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
	"github.com/hugr-lab/query-engine/pkg/types"
)

func (s *Service) FunctionFieldsForSummary(ctx context.Context) ([]Field, error) {
	res, err := s.h.Query(ctx, `query f($f: String!, $rt: String!) {
		core {
			mcp {
				fields(
					filter: {
						is_summarized: {eq: false}
						hugr_type: {eq: $f}
						root_type: {
							hugr_type: {eq: $rt}
						}
					}
				){
					name
					description
					type_name
					type
					is_indexed
					is_summarized
				}
			}
		}
	}`, map[string]any{
		"f":  base.HugrTypeFieldFunction,
		"rt": base.HugrTypeModule,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query function fields: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to query function fields: %w", res.Err())
	}

	var fields []Field
	err = res.ScanData("core.mcp.fields", &fields)
	if errors.Is(err, types.ErrNoData) {
		log.Println("No functions to summarize")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan function fields: %w", err)
	}
	return fields, nil
}

func (s *Service) SummarizeFunction(ctx context.Context, sum *summary.Service, meta *metainfo.SchemaInfo, f Field) error {
	// 1. Get module by type name
	m, err := s.ModuleByTypeName(ctx, f.TypeName)
	if err != nil {
		return fmt.Errorf("function %s failed to get module by type name: %w", f.Name, err)
	}

	// 2. Get function info
	path := f.Name
	if m.Name != "" {
		path = m.Name + "." + path
	}
	fi := meta.Function(path)
	if fi == nil {
		fi = meta.MutateFunction(path)
		if fi == nil {
			return fmt.Errorf("function %s not found", path)
		}
	}
	// 3. Get Summary
	start := time.Now()
	log.Printf("function %s: summarization start", path)
	fs, err := sum.SummarizeFunction(ctx, meta, fi)
	if err != nil {
		return fmt.Errorf("function %s failed to get summary: %w", path, err)
	}
	log.Printf("function %s: summarization completed in %s", path, time.Since(start))

	// 4. update function field desc
	if err := s.UpdateFieldDescription(ctx, f.TypeName, f.Name, fs.Long, true); err != nil {
		return fmt.Errorf("function %s failed to update field description: %w", path, err)
	}

	// 5. update function parameters
	if fs.Parameters != nil {
		for _, arg := range fi.Arguments {
			desc, ok := fs.Parameters[arg.Name]
			if !ok {
				continue
			}
			if err := s.UpdateArgumentDescription(ctx, f.TypeName, f.Name, arg.Name, desc); err != nil {
				return fmt.Errorf("function %s failed to update parameter description: %w", path, err)
			}
		}
	}

	if len(fs.Returns.Fields) != 0 {
		// 6. Update return type
		if err := s.UpdateTypeDescription(ctx, fi.ReturnType, fs.Returns.Short, fs.Long, true); err != nil {
			return fmt.Errorf("function %s failed to update return type description: %w", path, err)
		}
		// 7. Update return type fields
		for _, rf := range fi.ReturnTypeFields {
			desc, ok := fs.Returns.Fields[rf.Name]
			if !ok {
				continue
			}
			if err := s.UpdateFieldDescription(ctx, fi.ReturnType, rf.Name, desc, true); err != nil {
				return fmt.Errorf("function %s failed to update return field description: %w", path, err)
			}
		}
	}

	return nil
}
