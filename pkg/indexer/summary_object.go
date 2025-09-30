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

func (s *Service) DataObjectTypesForSummary(ctx context.Context) ([]Type, error) {
	res, err := s.h.Query(ctx, `query ($dtTypes: [String!]) {
		core {
			mcp {
				types(
					filter: {
						hugr_type: {in: $dtTypes}, 
						is_summarized: {eq: false}
					}
					order_by: [
						{field: "module"}
						{field: "name"}
					]
				) {
					name
					module
					hugr_type
					description
				}
			}
		}
	}`, map[string]interface{}{
		"dtTypes": []base.HugrType{base.HugrTypeTable, base.HugrTypeView},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to summarize data objects: %w", err)
	}
	defer res.Close()
	// Process the result
	var dataObjects []Type
	err = res.ScanData("core.mcp.types", &dataObjects)
	if errors.Is(err, types.ErrNoData) {
		log.Println("No data objects to summarize")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode data objects: %w", err)
	}
	return dataObjects, nil
}

func (s *Service) SummarizeDataObject(ctx context.Context, sum *summary.Service, meta *metainfo.SchemaInfo, t Type) error {
	var do *metainfo.DataObjectInfo
	path := t.Name
	if t.Module != "" {
		path = t.Module + "." + t.Name
	}
	switch t.HugrType {
	case base.HugrTypeTable:
		do = meta.Table(path)
	case base.HugrTypeView:
		do = meta.View(path)
	}
	if do == nil {
		log.Printf("Skipping summary for %s: not found", path)
	}
	start := time.Now()
	log.Printf("data object %s: summarization start", path)
	ds, err := sum.SummarizeDataObject(ctx, meta, do)
	if err != nil {
		return fmt.Errorf("failed to summarize data object %s: %w", path, err)
	}
	log.Printf("data object %s: Summary complete in %v", path, time.Since(start))

	// update types
	// 2. Update fields descriptions
	for name, desc := range ds.Fields {
		if err := s.UpdateFieldDescription(ctx, t.Name, name, desc, true); err != nil {
			return fmt.Errorf("failed to update field description: %w", err)
		}
		// 2.1. update aggregation type fields
		if do.AggregationType != "" {
			if err := s.UpdateFieldDescription(ctx, do.AggregationType, name, desc, true); err != nil {
				return fmt.Errorf("failed to update agg field description: %w", err)
			}
		}
		// 2.2. update subaggregation type fields
		if do.SubAggregationType != "" {
			if err := s.UpdateFieldDescription(ctx, do.SubAggregationType, name, desc, true); err != nil {
				return fmt.Errorf("failed to update subagg field description: %w", err)
			}
		}
	}

	// 3. Update extra field description
	for name, desc := range ds.ExtraFields {
		if err := s.UpdateFieldDescription(ctx, t.Name, name, desc, true); err != nil {
			return fmt.Errorf("failed to update extra field description: %w", err)
		}
		// 3.1. update aggregation type extra fields
		if do.AggregationType != "" {
			if err := s.UpdateFieldDescription(ctx, do.AggregationType, name, desc, true); err != nil {
				return fmt.Errorf("failed to update agg extra field description: %w", err)
			}
		}
		// 3.2. update subaggregation type extra fields
		if do.SubAggregationType != "" {
			if err := s.UpdateFieldDescription(ctx, do.SubAggregationType, name, desc, true); err != nil {
				return fmt.Errorf("failed to update subagg extra field description: %w", err)
			}
		}
	}

	// 3. update filter type desc
	if do.FilterType != "" {
		for name, desc := range ds.Filter.Fields {
			if err := s.UpdateFieldDescription(ctx, do.FilterType, name, desc, true); err != nil {
				return fmt.Errorf("failed to update filter field description: %w", err)
			}
		}
		for name, desc := range ds.Filter.References {
			if err := s.UpdateFieldDescription(ctx, do.FilterType, name, desc, true); err != nil {
				return fmt.Errorf("failed to update filter reference description: %w", err)
			}
		}
		err := s.UpdateTypeDescription(ctx, do.FilterType, ds.Filter.Row, "", true)
		if err != nil {
			return fmt.Errorf("failed to update filter type description: %w", err)
		}
	}

	// 4. Update references
	if ds.References != nil {
		for _, ref := range do.References {
			rs, ok := ds.References[ref.Name]
			if !ok {
				log.Printf("data object %s: Reference %s not found", t.Name, ref.Name)
				continue
			}
			// 4.1. Update data query field
			if rs.Select != "" {
				if err := s.UpdateFieldDescription(ctx, t.Name, ref.FieldDataQuery, rs.Select, true); err != nil {
					return fmt.Errorf("failed to update reference field description: %w", err)
				}
			}
			// 4.2. Update agg query field
			if rs.Agg != "" && ref.FieldAggQuery != "" {
				if err := s.UpdateFieldDescription(ctx, t.Name, ref.FieldAggQuery, rs.Agg, true); err != nil {
					return fmt.Errorf("failed to update reference agg field description: %w", err)
				}
			}
			// 4.3. Update bucket agg query field
			if rs.BucketAgg != "" && ref.FieldBucketAggQuery != "" {
				if err := s.UpdateFieldDescription(ctx, t.Name, ref.FieldBucketAggQuery, rs.BucketAgg, true); err != nil {
					return fmt.Errorf("failed to update reference bucket agg field description: %w", err)
				}
			}
			// 4.4. Update aggregation type fields
			if do.AggregationType != "" && rs.Select != "" {
				if err := s.UpdateFieldDescription(ctx, do.AggregationType, ref.FieldDataQuery, rs.Select, true); err != nil {
					return fmt.Errorf("failed to update reference agg field description: %w", err)
				}
				if ref.FieldAggQuery != "" {
					if err := s.UpdateFieldDescription(ctx, do.AggregationType, ref.FieldAggQuery, rs.Agg, true); err != nil {
						return fmt.Errorf("failed to update reference agg field description: %w", err)
					}
				}
			}
			// 4.5. Update aggregation type fields for aggregations
			if do.SubAggregationType != "" && ref.FieldAggQuery != "" && rs.Agg != "" {
				if err := s.UpdateFieldDescription(ctx, do.SubAggregationType, ref.FieldDataQuery, rs.Agg, true); err != nil {
					return fmt.Errorf("failed to update reference subagg field description: %w", err)
				}
			}
		}
	}

	// 5. Update subqueries
	if ds.SubQueries != nil {
		for _, sq := range do.Subqueries {
			sqs, ok := ds.SubQueries[sq.Name]
			if !ok {
				log.Printf("data object %s: Subquery %s not found", t.Name, sq.Name)
				continue
			}
			// 5.1. Update data query field
			if sqs.Select != "" {
				if err := s.UpdateFieldDescription(ctx, t.Name, sq.FieldDataQuery, sqs.Select, true); err != nil {
					return fmt.Errorf("failed to update subquery field description: %w", err)
				}
			}
			// 5.2. Update agg query field
			if sqs.Agg != "" {
				if err := s.UpdateFieldDescription(ctx, t.Name, sq.FieldAggQuery, sqs.Agg, true); err != nil {
					return fmt.Errorf("failed to update subquery field description: %w", err)
				}
			}
			// 5.3. Update bucket agg query field
			if sqs.BucketAgg != "" {
				if err := s.UpdateFieldDescription(ctx, t.Name, sq.FieldBucketAggQuery, sqs.BucketAgg, true); err != nil {
					return fmt.Errorf("failed to update subquery field description: %w", err)
				}
			}
			// 4.4. Update aggregation type fields
			if do.AggregationType != "" && sqs.Select != "" {
				if err := s.UpdateFieldDescription(ctx, do.AggregationType, sq.FieldDataQuery, sqs.Select, true); err != nil {
					return fmt.Errorf("failed to update reference agg field description: %w", err)
				}
				if sq.FieldAggQuery != "" {
					if err := s.UpdateFieldDescription(ctx, do.AggregationType, sq.FieldAggQuery, sqs.Agg, true); err != nil {
						return fmt.Errorf("failed to update reference agg field description: %w", err)
					}
				}
			}
			// 4.5. Update aggregation type fields for aggregations
			if do.SubAggregationType != "" && sq.FieldAggQuery != "" && sqs.Agg != "" {
				if err := s.UpdateFieldDescription(ctx, do.SubAggregationType, sq.FieldDataQuery, sqs.Agg, true); err != nil {
					return fmt.Errorf("failed to update reference subagg field description: %w", err)
				}
			}
		}
	}

	// 6. Update function calls
	if ds.FunctionCalls != nil {
		for _, fc := range do.FunctionCalls {
			desc := ds.FunctionCalls[fc.Name]
			if desc == "" {
				continue
			}
			if err := s.UpdateFieldDescription(ctx, t.Name, fc.FieldName, desc, true); err != nil {
				return fmt.Errorf("failed to update function call field description: %w", err)
			}
			if do.AggregationType != "" {
				if err := s.UpdateFieldDescription(ctx, do.AggregationType, fc.FieldName, desc, true); err != nil {
					return fmt.Errorf("failed to update reference agg field description: %w", err)
				}
			}
		}
	}

	// 7. Update query description
	m, err := s.ModuleByName(ctx, t.Module)
	if err != nil {
		return fmt.Errorf("data object %s: Failed to get module by name: %w", t.Name, err)
	}
	if ds.Queries == nil {
		log.Printf("data object %s: No queries to update", t.Name)
	}
	if ds.Queries != nil {
		var dsPrefix string
		for _, ds := range meta.DataSources {
			if ds.Name == do.DataSource {
				if ds.AsModule && ds.Prefix != "" {
					dsPrefix = ds.Prefix + "_"
				}
			}
		}
		for _, sq := range do.Queries {
			desc, ok := ds.Queries[sq.Name]
			if !ok {
				log.Printf("data object %s: Query %s not found", t.Name, sq.Name)
				continue
			}
			// 7.1. Update module queries
			if err := s.UpdateFieldDescription(ctx, m.QueryRoot, sq.Name, desc, true); err != nil {
				return fmt.Errorf("failed to update module query field description: %w", err)
			}
			if sq.Type == metainfo.QueryTypeSelectOne ||
				sq.Type == metainfo.QueryTypeH3 ||
				sq.Type == metainfo.QueryTypeJQ {
				continue
			}
			// 7.2. Update _join
			if err := s.UpdateFieldDescription(ctx, base.QueryTimeJoinsTypeName, dsPrefix+sq.Name, desc, true); err != nil {
				return fmt.Errorf("data object %s: Failed to update subquery field description: %w", t.Name, err)
			}
			// 7.3. Update _spatial
			if err := s.UpdateFieldDescription(ctx, base.QueryTimeSpatialTypeName, dsPrefix+sq.Name, desc, true); err != nil {
				return fmt.Errorf("data object %s: Failed to update subquery field description: %w", t.Name, err)
			}
			// 7.4. Update _h3_data
			if err := s.UpdateFieldDescription(ctx, base.H3DataQueryTypeName, dsPrefix+sq.Name, desc, true); err != nil {
				return fmt.Errorf("data object %s: Failed to update subquery field description: %w", t.Name, err)
			}
		}
	}

	// 8. Update mutation
	if do.Mutations != nil && ds.Mutations != nil && m.MutationRoot != "" {
		for name, desc := range ds.Mutations {
			if desc == "" {
				continue
			}
			if err := s.UpdateFieldDescription(ctx, m.MutationRoot, name, desc, true); err != nil {
				return fmt.Errorf("data object %s: Failed to update mutation field description: %w", t.Name, err)
			}
		}
	}

	// 9. Update arguments
	if do.Arguments != nil && ds.Arguments.Short != "" {
		for _, arg := range do.Arguments.NestedFields {
			desc, ok := ds.Arguments.Fields[arg.Name]
			if !ok {
				continue
			}
			if err := s.UpdateFieldDescription(ctx, do.Arguments.Type, arg.Name, desc, true); err != nil {
				return fmt.Errorf("data object %s: Failed to update argument field description: %w", t.Name, err)
			}
		}
		err := s.UpdateTypeDescription(ctx, do.Arguments.Type, ds.Arguments.Short, "", true)
		if err != nil {
			return fmt.Errorf("data object %s: Failed to update argument type description: %w", t.Name, err)
		}
	}

	// 1. Update data object description
	if err := s.UpdateTypeDescription(ctx, t.Name, ds.Short, ds.Long, true); err != nil {
		return fmt.Errorf("failed to update type description: %w", err)
	}

	if ds.AggregationTypeShort != "" && do.AggregationType != "" {
		if err := s.UpdateTypeDescription(ctx, do.AggregationType, ds.AggregationTypeShort, ds.AggregationTypeLong, true); err != nil {
			return fmt.Errorf("failed to update agg type description: %w", err)
		}
	}

	if ds.SubAggregationTypeShort != "" && do.SubAggregationType != "" {
		if err := s.UpdateTypeDescription(ctx, do.SubAggregationType, ds.SubAggregationTypeShort, ds.SubAggregationTypeLong, true); err != nil {
			return fmt.Errorf("failed to update subagg type description: %w", err)
		}
	}

	if ds.BucketAggregationTypeShort != "" && do.BucketAggregationType != "" {
		if err := s.UpdateTypeDescription(ctx, do.BucketAggregationType, ds.BucketAggregationTypeShort, ds.BucketAggregationTypeLong, true); err != nil {
			return fmt.Errorf("failed to update bucket agg type description: %w", err)
		}
	}

	log.Println("data object", t.Name, "updated successfully")

	return nil
}
