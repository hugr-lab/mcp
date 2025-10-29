package indexer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hugr-lab/query-engine/pkg/compiler"
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
	"github.com/hugr-lab/query-engine/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// PartialLoad loads parts of schema on demand

// LoadDataObject loads/reloads data object by name, including its fields, arguments, inputs, aggregations.
func (s *Service) LoadDataObject(ctx context.Context, name string, patrial bool) error {
	schema, err := s.fetchSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch schema: %w", err)
	}
	meta, err := s.fetchSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch meta summary: %w", err)
	}
	if !patrial {
		err := s.clearDataObjectTypes(ctx, schema, meta, name)
		if err != nil {
			return fmt.Errorf("failed to clear data object types: %w", err)
		}
	}
	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}
	err = fillDataObjectTypesForUpdate(schema, meta, name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		return fmt.Errorf("failed to fill data object types for update: %w", err)
	}
	err = s.loadSchemaPatrial(ctx, schema, meta, patrial, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		return fmt.Errorf("failed to load schema patrial: %w", err)
	}
	return nil
}

func (s *Service) LoadFunction(ctx context.Context, module, name string, patrial bool) error {
	schema, err := s.fetchSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch schema: %w", err)
	}
	meta, err := s.fetchSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch meta summary: %w", err)
	}
	if !patrial {
		err := s.clearFunctionTypes(ctx, schema, meta, module, name)
		if err != nil {
			return fmt.Errorf("failed to clear function types: %w", err)
		}
	}
	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}
	err = fillFunctionTypesForUpdate(schema, meta, module, name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		return fmt.Errorf("failed to fill function types for update: %w", err)
	}
	err = s.loadSchemaPatrial(ctx, schema, meta, patrial, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		return fmt.Errorf("failed to load schema patrial: %w", err)
	}
	return nil
}

func (s *Service) LoadDataSource(ctx context.Context, name string) error {
	return nil
}

func (s *Service) LoadModule(ctx context.Context, name string) error {
	// 1. Fetch schema and summary
	// 2. Get all data objects and functions for the module
	// 2. Clear module, data objects and functions types, fields, arguments if needed
	// 3. Load functions, data objects and modules types, fields, arguments

	return nil
}

func (s *Service) loadSchemaPatrial(ctx context.Context, schema *SchemaIntro, meta *metainfo.SchemaInfo, update bool,
	typesMap map[string]struct{},
	fieldsMap map[string]map[string]struct{},
	modulesMap map[string]struct{},
	dataSourcesMap map[string]struct{},
) error {
	// 3. Add all types from schema
	// add unknown type
	err := s.mergeType(ctx, Type{
		Name:        "Unknown",
		Description: "Unknown type",
		Kind:        "SCALAR",
	}, false)
	if err != nil {
		return fmt.Errorf("failed to add unknown type: %w", err)
	}
	var ff []Field
	var aa []Argument
	am := map[string]struct{}{}
	for _, st := range schema.Types {
		if len(typesMap) != 0 {
			if _, ok := typesMap[st.Name]; !ok {
				continue
			}
		}
		t := Type{
			Name:        st.Name,
			Description: st.Description,
			Kind:        st.Kind,
			HugrType:    st.HugrType,
			Module:      st.Module,
			Catalog:     st.Catalog,
		}
		err := s.mergeType(ctx, t, update)
		if err != nil {
			return fmt.Errorf("failed to add type %q: %w", st.Name, err)
		}
		fields := st.Fields
		if st.Kind == string(ast.InputObject) {
			fields = st.InputFields
		}
		for _, f := range fields {
			if len(fieldsMap) != 0 {
				if _, ok := fieldsMap[st.Name]; !ok {
					continue
				}
				if _, ok := fieldsMap[st.Name][f.Name]; !ok {
					continue
				}
			}
			field := Field{
				Name:        f.Name,
				Description: f.Description,
				TypeName:    st.Name,
				HugrType:    f.HugrType,
				Catalog:     f.Catalog,
				Exclude:     f.Exclude,
				Type:        f.Type.TypeName(),
				IsList:      f.Type.IsList(),
				IsNotNull:   f.Type.IsNotNull(),
			}
			ff = append(ff, field)
			for _, a := range f.Args {
				key := fmt.Sprintf("%s.%s.%s", st.Name, f.Name, a.Name)
				if _, ok := am[key]; ok {
					return fmt.Errorf("duplicate argument %q in hugr schema", key)
				}
				am[key] = struct{}{}
				arg := Argument{
					Name:        a.Name,
					FieldName:   f.Name,
					TypeName:    st.Name,
					Description: a.Description,
					Type:        a.Type.TypeName(),
					IsList:      a.Type.IsList(),
					IsNotNull:   a.Type.IsNotNull(),
				}
				aa = append(aa, arg)
			}
		}
	}

	// 4. fields
	for _, f := range ff {
		// add new fields or update existing if type, description, module, catalog, exclude changed
		err := s.mergeField(ctx, f, update)
		if err != nil {
			return fmt.Errorf("failed to add field %q.%q: %w", f.TypeName, f.Name, err)
		}
	}
	// 5. arguments
	for _, a := range aa {
		// add new arguments or update existing if type, description, required, is_list, is_not_null changed
		err := s.mergeArgument(ctx, a, update)
		if err != nil {
			return fmt.Errorf("failed to add argument %q: %w", a.Name, err)
		}
	}

	// 6. Modules
	for _, m := range meta.Modules() {
		if len(modulesMap) != 0 {
			if _, ok := modulesMap[m.Name]; !ok {
				continue
			}
		}
		err := s.mergeModule(ctx, Module{
			Name:            m.Name,
			Description:     m.Description,
			QueryRoot:       m.QueryType,
			MutationRoot:    m.MutationType,
			FunctionRoot:    m.FunctionType,
			MutFunctionRoot: m.MutationFunctionType,
		}, false)
		if err != nil {
			return fmt.Errorf("failed to add module %q: %w", m.Name, err)
		}
	}

	// 7. Data sources
	for _, ds := range meta.DataSources {
		if len(dataSourcesMap) != 0 {
			if _, ok := dataSourcesMap[ds.Name]; !ok {
				continue
			}
		}
		err := s.mergeDataSource(ctx, DataSource{
			Name:        ds.Name,
			Description: ds.Description,
			Type:        ds.Type,
			Prefix:      ds.Prefix,
			AsModule:    ds.AsModule,
			ReadOnly:    ds.ReadOnly,
		}, update)
		if err != nil {
			return fmt.Errorf("failed to add data source %q: %w", ds.Name, err)
		}
	}

	// 8. Data objects
	for _, do := range meta.DataObjects() {
		if len(typesMap) != 0 {
			if _, ok := typesMap[do.Name]; !ok {
				continue
			}
		}
		m := meta.Module(do.Module)
		if m == nil || m.QueryType == "" {
			return fmt.Errorf("module %q not found for data object %q", do.Module, do.Name)
		}
		object := DataObject{
			Name:           do.Name,
			FilterTypeName: do.FilterType,
		}
		if do.Arguments != nil {
			object.ArgsTypeName = do.Arguments.Type
		}
		for _, q := range do.Queries {
			object.Queries = append(object.Queries, DataObjectQuery{
				Name:      q.Name,
				QueryType: string(q.Type),
				QueryRoot: m.QueryType,
			})
		}

		err := s.addDataObject(ctx, object)
		if err != nil {
			return fmt.Errorf("failed to add data object %q: %w", do.Name, err)
		}
	}

	return nil
}

// delete fields for data object type, filter type, aggregation types, module query and mutation types
func (s *Service) clearDataObjectTypes(ctx context.Context, schema *SchemaIntro, meta *metainfo.SchemaInfo, names ...string) error {
	var deleteFieldFilters, deleteArgFilters []map[string]map[string]any
	for _, doName := range names {
		// get type intro for data objects
		ti := schema.TypeByName(doName)
		if ti == nil {
			return fmt.Errorf("type %s not found in schema", doName)
		}

		mi := meta.Module(ti.Module)
		if mi == nil {
			return fmt.Errorf("module %s not found in summary", ti.Module)
		}

		doi := mi.DataObject(ti.Name)
		if doi == nil {
			return fmt.Errorf("data object %s not found in module %s", ti.Name, ti.Module)
		}
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doName}})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": doName}})

		// delete filter type fields
		if doi.FilterType != "" {
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.FilterType}})
			// delete list filter type fields
			listFilterName := strings.TrimSuffix(doi.FilterType, compiler.FilterInputSuffix) + compiler.ListFilterInputSuffix
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": listFilterName}})
		}
		// delete arguments type fields
		if doi.Arguments != nil && doi.Arguments.Type != "" {
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.Arguments.Type}})
		}
		// delete aggregation type fields
		if doi.AggregationType != "" {
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.AggregationType}})
			deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": doi.AggregationType}})
		}
		if doi.SubAggregationType != "" {
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.SubAggregationType}})
			deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": doi.SubAggregationType}})
		}
		if doi.BucketAggregationType != "" {
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.BucketAggregationType}})
			deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": doi.BucketAggregationType}})
		}
		// delete mutation input type fields
		if doi.Mutations != nil {
			if doi.Mutations.InsertDataType != "" {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.Mutations.InsertDataType}})
			}
			if doi.Mutations.UpdateDataType != "" {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": doi.Mutations.UpdateDataType}})
			}
		}
		// module query type fields
		if mi.QueryType != "" {
			for _, q := range doi.Queries {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": q.Name},
					"type_name": {"eq": mi.QueryType},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": q.Name},
					"type_name":  {"eq": mi.QueryType},
				})
			}
		}
		// module mutation type fields
		if mi.MutationType != "" && doi.Mutations != nil {
			if doi.Mutations.InsertMutation != "" {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": doi.Mutations.InsertMutation},
					"type_name": {"eq": mi.MutationType},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": doi.Mutations.InsertMutation},
					"type_name":  {"eq": mi.MutationType},
				})
			}
			// module update mutation type fields
			if doi.Mutations.UpdateMutation != "" {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": doi.Mutations.UpdateMutation},
					"type_name": {"eq": mi.MutationType},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": doi.Mutations.UpdateMutation},
					"type_name":  {"eq": mi.MutationType},
				})
			}
			if doi.Mutations.DeleteMutation != "" {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": doi.Mutations.DeleteMutation},
					"type_name": {"eq": mi.MutationType},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": doi.Mutations.DeleteMutation},
					"type_name":  {"eq": mi.MutationType},
				})
			}
		}

		// join, spatial and h3_data types
		prefix := ""
		ds := meta.DataSource(doi.DataSource)
		if ds != nil && ds.AsModule && ds.Prefix != "" {
			prefix = ds.Prefix + "_"
		}
		for _, q := range doi.Queries {
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
				"name":      {"eq": prefix + q.Name},
				"type_name": {"eq": base.QueryTimeJoinsTypeName},
			})
			deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
				"field_name": {"eq": prefix + q.Name},
				"type_name":  {"eq": base.QueryTimeJoinsTypeName},
			})
			// join aggregations type
			deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
				"name":      {"eq": prefix + q.Name},
				"type_name": {"eq": base.QueryTimeJoinsTypeName + compiler.AggregationSuffix},
			})
			deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
				"field_name": {"eq": prefix + q.Name},
				"type_name":  {"eq": base.QueryTimeJoinsTypeName + compiler.AggregationSuffix},
			})
			if doi.HasGeometry {
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": prefix + q.Name},
					"type_name": {"eq": base.QueryTimeSpatialTypeName},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": prefix + q.Name},
					"type_name":  {"eq": base.QueryTimeSpatialTypeName},
				})
				// spatial aggregations type
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": prefix + q.Name},
					"type_name": {"eq": base.QueryTimeSpatialTypeName + compiler.AggregationSuffix},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": prefix + q.Name},
					"type_name":  {"eq": base.QueryTimeSpatialTypeName + compiler.AggregationSuffix},
				})
				// h3_data type
				deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
					"name":      {"eq": prefix + q.Name},
					"type_name": {"eq": base.H3DataQueryTypeName},
				})
				deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
					"field_name": {"eq": prefix + q.Name},
					"type_name":  {"eq": base.H3DataQueryTypeName},
				})
			}
		}
	}

	return s.deleteFieldsAndArguments(ctx, deleteFieldFilters, deleteArgFilters)
}

func fillDataObjectTypesForUpdate(schema *SchemaIntro, meta *metainfo.SchemaInfo, doName string,
	typesMap map[string]struct{},
	fieldsMap map[string]map[string]struct{},
	modulesMap map[string]struct{},
	dataSourcesMap map[string]struct{},
) error {
	// get type intro for data objects
	ti := schema.TypeByName(doName)
	if ti == nil {
		return fmt.Errorf("type %s not found in schema", doName)
	}

	mi := meta.Module(ti.Module)
	if mi == nil {
		return fmt.Errorf("module %s not found in summary", ti.Module)
	}

	doi := mi.DataObject(ti.Name)
	if doi == nil {
		return fmt.Errorf("data object %s not found in module %s", ti.Name, ti.Module)
	}

	doFieldsMap := map[string]struct{}{}
	for _, c := range doi.Columns {
		doFieldsMap[c.Name] = struct{}{}
		for _, e := range c.ExtraFields {
			doFieldsMap[e.Name] = struct{}{}
		}
	}
	typesMap[doName] = struct{}{}
	fieldsMap[doName] = map[string]struct{}{}
	// the data object type fields
	hasSpatial := false
	for _, f := range ti.Fields {
		if f.Name == base.QueryTimeSpatialFieldName &&
			f.Type.TypeName() == base.QueryTimeSpatialTypeName {
			hasSpatial = true
		}
		fieldsMap[doName][f.Name] = struct{}{}
		if _, ok := doFieldsMap[f.Name]; !ok {
			typesMap[f.Type.TypeName()] = struct{}{}
			// field arguments
			for _, a := range f.Args {
				typesMap[a.Type.TypeName()] = struct{}{}
			}
			continue
		}
		addFieldTypeRecursively(schema, f.Type.TypeName(), typesMap, fieldsMap)
		// field arguments
		for _, a := range f.Args {
			addFieldTypeRecursively(schema, a.Type.TypeName(), typesMap, fieldsMap)
		}
	}
	_ = hasSpatial
	if doi.FilterType != "" {
		fi := schema.TypeByName(doi.FilterType)
		if fi == nil {
			return fmt.Errorf("filter type %s not found in schema", doi.FilterType)
		}
		typesMap[doi.FilterType] = struct{}{}
		fieldsMap[doi.FilterType] = map[string]struct{}{}
		for _, f := range fi.InputFields {
			fieldsMap[doi.FilterType][f.Name] = struct{}{}
			if _, ok := doFieldsMap[f.Name]; !ok {
				typesMap[f.Type.TypeName()] = struct{}{}
				continue
			}
			addFieldTypeRecursively(schema, f.Type.TypeName(), typesMap, fieldsMap)
		}
		// add list filter type
		listFilterName := strings.TrimSuffix(doi.FilterType, compiler.FilterInputSuffix) + compiler.ListFilterInputSuffix
		delete(typesMap, listFilterName)
		addFieldTypeRecursively(schema, listFilterName, typesMap, fieldsMap)
	}
	if doi.Arguments != nil && doi.Arguments.Type != "" {
		delete(typesMap, doi.Arguments.Type)
		addFieldTypeRecursively(schema, doi.Arguments.Type, typesMap, fieldsMap)
	}

	// aggregation type fields
	if doi.AggregationType != "" {
		ai := schema.TypeByName(doi.AggregationType)
		if ai == nil {
			return fmt.Errorf("aggregation type %s not found in schema", doi.AggregationType)
		}
		typesMap[doi.AggregationType] = struct{}{}
		fieldsMap[doi.AggregationType] = map[string]struct{}{}
		for _, f := range ai.Fields {
			fieldsMap[doi.AggregationType][f.Name] = struct{}{}
			if _, ok := doFieldsMap[f.Name]; !ok && f.Name != "_rows_count" {
				typesMap[f.Type.TypeName()] = struct{}{}
				// field arguments
				for _, a := range f.Args {
					typesMap[a.Type.TypeName()] = struct{}{}
				}
				continue
			}
			addFieldTypeRecursively(schema, f.Type.TypeName(), typesMap, fieldsMap)
			// field arguments
			for _, a := range f.Args {
				addFieldTypeRecursively(schema, a.Type.TypeName(), typesMap, fieldsMap)
			}
		}
	}
	// sub-aggregation type fields
	if doi.SubAggregationType != "" {
		sai := schema.TypeByName(doi.SubAggregationType)
		if sai == nil {
			return fmt.Errorf("sub-aggregation type %s not found in schema", doi.SubAggregationType)
		}
		typesMap[doi.SubAggregationType] = struct{}{}
		fieldsMap[doi.SubAggregationType] = map[string]struct{}{}
		for _, f := range sai.Fields {
			fieldsMap[doi.SubAggregationType][f.Name] = struct{}{}
			if _, ok := doFieldsMap[f.Name]; !ok && f.Name != "_rows_count" {
				typesMap[f.Type.TypeName()] = struct{}{}
				// field arguments
				for _, a := range f.Args {
					typesMap[a.Type.TypeName()] = struct{}{}
				}
				continue
			}
			addFieldTypeRecursively(schema, f.Type.TypeName(), typesMap, fieldsMap)
			// field arguments
			for _, a := range f.Args {
				addFieldTypeRecursively(schema, a.Type.TypeName(), typesMap, fieldsMap)
			}
		}
	}
	// bucket aggregation type fields
	if doi.BucketAggregationType != "" {
		bai := schema.TypeByName(doi.BucketAggregationType)
		if bai == nil {
			return fmt.Errorf("bucket-aggregation type %s not found in schema", doi.BucketAggregationType)
		}
		typesMap[doi.BucketAggregationType] = struct{}{}
		fieldsMap[doi.BucketAggregationType] = map[string]struct{}{}
		for _, f := range bai.Fields {
			fieldsMap[doi.BucketAggregationType][f.Name] = struct{}{}
		}
	}

	// Module types (module hierarchy)
	// module query type
	if mi.QueryType != "" {
		typesMap[mi.QueryType] = struct{}{}
		if _, ok := fieldsMap[mi.QueryType]; !ok {
			fieldsMap[mi.QueryType] = map[string]struct{}{}
		}
		for _, q := range doi.Queries {
			fieldsMap[mi.QueryType][q.Name] = struct{}{}
			typesMap[q.ReturnedTypeName] = struct{}{}
			// arguments types
			for _, a := range q.Arguments {
				addFieldTypeRecursively(schema, a.Type, typesMap, fieldsMap)
			}
		}
	}
	// module mutation type
	if mi.MutationType != "" && doi.Mutations != nil {
		typesMap[mi.MutationType] = struct{}{}
		if _, ok := fieldsMap[mi.MutationType]; !ok {
			fieldsMap[mi.MutationType] = map[string]struct{}{}
		}
		if doi.Mutations.InsertMutation != "" {
			fieldsMap[mi.MutationType][doi.Mutations.InsertMutation] = struct{}{}
			mdi := schema.TypeByName(doi.Mutations.InsertDataType)
			if mdi != nil {
				typesMap[doi.Mutations.InsertDataType] = struct{}{}
				fieldsMap[mdi.Name] = map[string]struct{}{}
				for _, f := range mdi.InputFields {
					fieldsMap[mdi.Name][f.Name] = struct{}{}
					if _, ok := doFieldsMap[f.Name]; !ok {
						typesMap[f.Type.TypeName()] = struct{}{}
						continue
					}
					addFieldTypeRecursively(schema, f.Type.TypeName(), typesMap, fieldsMap)
				}
			}
		}
		if doi.Mutations.UpdateMutation != "" {
			fieldsMap[mi.MutationType][doi.Mutations.UpdateMutation] = struct{}{}
			addFieldTypeRecursively(schema, doi.Mutations.UpdateDataType, typesMap, fieldsMap)
		}
		if doi.Mutations.DeleteMutation != "" {
			fieldsMap[mi.MutationType][doi.Mutations.DeleteMutation] = struct{}{}
		}
	}
	if mi.FunctionType != "" {
		typesMap[mi.FunctionType] = struct{}{}
	}
	if mi.MutationFunctionType != "" {
		typesMap[mi.MutationFunctionType] = struct{}{}
	}

	// Parent module types
	err := fillModulesTypesRecursively(meta, mi, typesMap, fieldsMap, modulesMap)
	if err != nil {
		return err
	}

	ds := meta.DataSource(doi.DataSource)
	prefix := ""
	if ds != nil {
		dataSourcesMap[ds.Name] = struct{}{}
		if ds.AsModule && ds.Prefix != "" {
			prefix = ds.Prefix + "_"
		}
	}

	// _join, _spatial and _h3_data fields
	typesMap[base.QueryTimeJoinsTypeName] = struct{}{}
	if _, ok := fieldsMap[base.QueryTimeJoinsTypeName]; !ok {
		fieldsMap[base.QueryTimeJoinsTypeName] = map[string]struct{}{}
	}
	typesMap[base.QueryTimeJoinsTypeName+compiler.AggregationSuffix] = struct{}{}
	if _, ok := fieldsMap[base.QueryTimeJoinsTypeName+compiler.AggregationSuffix]; !ok {
		fieldsMap[base.QueryTimeJoinsTypeName+compiler.AggregationSuffix] = map[string]struct{}{}
	}
	if hasSpatial {
		typesMap[base.QueryTimeSpatialTypeName] = struct{}{}
		if _, ok := fieldsMap[base.QueryTimeSpatialTypeName]; !ok {
			fieldsMap[base.QueryTimeSpatialTypeName] = map[string]struct{}{}
		}
		typesMap[base.QueryTimeSpatialTypeName+compiler.AggregationSuffix] = struct{}{}
		if _, ok := fieldsMap[base.QueryTimeSpatialTypeName+compiler.AggregationSuffix]; !ok {
			fieldsMap[base.QueryTimeSpatialTypeName+compiler.AggregationSuffix] = map[string]struct{}{}
		}
		typesMap[base.H3DataQueryTypeName] = struct{}{}
		if _, ok := fieldsMap[base.H3DataQueryTypeName]; !ok {
			fieldsMap[base.H3DataQueryTypeName] = map[string]struct{}{}
		}
	}
	for _, q := range doi.Queries {
		if q.Type == metainfo.QueryTypeSelectOne {
			continue
		}
		fieldsMap[base.QueryTimeJoinsTypeName][prefix+q.Name] = struct{}{}
		fieldsMap[base.QueryTimeJoinsTypeName+compiler.AggregationSuffix][prefix+q.Name] = struct{}{}
		if hasSpatial {
			fieldsMap[base.QueryTimeSpatialTypeName][prefix+q.Name] = struct{}{}
			fieldsMap[base.QueryTimeSpatialTypeName+compiler.AggregationSuffix][prefix+q.Name] = struct{}{}
			fieldsMap[base.H3DataQueryTypeName][prefix+q.Name] = struct{}{}
		}
	}
	return nil
}

func (s *Service) clearFunctionTypes(ctx context.Context, schema *SchemaIntro, meta *metainfo.SchemaInfo, module, name string) error {
	mi := meta.Module(module)
	if mi == nil {
		return fmt.Errorf("module %s not found in summary", module)
	}
	isMutation := false
	fi := mi.Function(name)
	if fi == nil {
		fi = mi.MutationFunction(name)
		isMutation = true
		if fi == nil {
			return fmt.Errorf("function %s not found in module %s", name, module)
		}
	}

	var deleteFieldFilters, deleteArgFilters []map[string]map[string]any
	// add filter function return type fields
	if fi.ReturnType != "" {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": fi.ReturnType}})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": fi.ReturnType}})
	}
	// add function arguments types fields
	for _, a := range fi.Arguments {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": a.Type}})
	}
	// add aggregation type fields
	if fi.AggregationType != "" {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": fi.AggregationType}})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": fi.AggregationType}})
	}
	// add sub-aggregation type fields
	if fi.SubAggregationType != "" {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": fi.SubAggregationType}})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": fi.SubAggregationType}})
	}
	// add bucket-aggregation type fields
	if fi.BucketAggregationType != "" {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{"type_name": {"eq": fi.BucketAggregationType}})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{"type_name": {"eq": fi.BucketAggregationType}})
	}
	if isMutation && mi.MutationFunctionType != "" {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
			"name":      {"eq": fi.Name},
			"type_name": {"eq": mi.MutationFunctionType},
		})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
			"field_name": {"eq": fi.Name},
			"type_name":  {"eq": mi.MutationFunctionType},
		})
	}
	if !isMutation && mi.FunctionType != "" {
		deleteFieldFilters = append(deleteFieldFilters, map[string]map[string]any{
			"name":      {"eq": fi.Name},
			"type_name": {"eq": mi.FunctionType},
		})
		deleteArgFilters = append(deleteArgFilters, map[string]map[string]any{
			"field_name": {"eq": fi.Name},
			"type_name":  {"eq": mi.FunctionType},
		})
	}
	return s.deleteFieldsAndArguments(ctx, deleteFieldFilters, deleteArgFilters)
}

func (s *Service) deleteFieldsAndArguments(ctx context.Context, filterFields, filterArgs []map[string]map[string]any) error {
	res, err := s.h.Query(ctx, `mutation ($fieldFilter: [mcp_fields_filter!], $argFilter: [mcp_arguments_filter!]) {
		core {
			mcp {
				delete_arguments(filter: {_or: $argFilter}) { success }
				delete_fields(filter: {_or: $fieldFilter}) { success }
			}
		}
	}`,
		map[string]any{
			"fieldFilter": filterFields,
			"argFilter":   filterArgs,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete function types: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to delete function types: %w", res.Err())
	}
	return nil
}

func fillFunctionTypesForUpdate(schema *SchemaIntro, meta *metainfo.SchemaInfo, module, name string,
	typesMap map[string]struct{},
	fieldsMap map[string]map[string]struct{},
	modulesMap map[string]struct{},
	dataSourcesMap map[string]struct{},
) error {
	mi := meta.Module(module)
	if mi == nil {
		return fmt.Errorf("module %s not found in summary", module)
	}
	isMutation := false
	fi := mi.Function(name)
	if fi == nil {
		fi = mi.MutationFunction(name)
		isMutation = true
		if fi == nil {
			return fmt.Errorf("function %s not found in module %s", name, module)
		}
	}

	// return types
	delete(typesMap, fi.ReturnType)
	addFieldTypeRecursively(schema, fi.ReturnType, typesMap, fieldsMap)
	// arguments
	for _, a := range fi.Arguments {
		addFieldTypeRecursively(schema, a.Type, typesMap, fieldsMap)
	}
	// aggregation type
	if fi.AggregationType != "" {
		delete(typesMap, fi.AggregationType)
		addFieldTypeRecursively(schema, fi.AggregationType, typesMap, fieldsMap)
	}
	if fi.SubAggregationType != "" {
		delete(typesMap, fi.SubAggregationType)
		addFieldTypeRecursively(schema, fi.SubAggregationType, typesMap, fieldsMap)
	}
	if fi.BucketAggregationType != "" {
		delete(typesMap, fi.BucketAggregationType)
		addFieldTypeRecursively(schema, fi.BucketAggregationType, typesMap, fieldsMap)
	}

	// module type
	if isMutation && mi.MutationFunctionType != "" {
		typesMap[mi.MutationFunctionType] = struct{}{}
		if _, ok := fieldsMap[mi.MutationFunctionType]; !ok {
			fieldsMap[mi.MutationFunctionType] = map[string]struct{}{}
		}
		fieldsMap[mi.MutationFunctionType][fi.Name] = struct{}{}
	}
	if !isMutation && mi.FunctionType != "" {
		typesMap[mi.FunctionType] = struct{}{}
		if _, ok := fieldsMap[mi.FunctionType]; !ok {
			fieldsMap[mi.FunctionType] = map[string]struct{}{}
		}
		fieldsMap[mi.FunctionType][fi.Name] = struct{}{}
	}

	// modules
	err := fillModulesTypesRecursively(meta, mi, typesMap, fieldsMap, modulesMap)
	if err != nil {
		return err
	}

	// data sources
	if fi.DataSource != "" {
		dataSourcesMap[fi.DataSource] = struct{}{}
	}

	return nil
}

func fillModulesTypesRecursively(meta *metainfo.SchemaInfo, mi *metainfo.ModuleInfo,
	typesMap map[string]struct{},
	fieldsMap map[string]map[string]struct{},
	modulesMap map[string]struct{},
) error {
	pp := strings.Split(mi.Name, ".")
	pmi := &meta.RootModule
	modulesMap[""] = struct{}{}
	for i, n := range pp {
		moduleName := strings.Join(pp[:i+1], ".")
		modulesMap[moduleName] = struct{}{}
		spi := meta.Module(moduleName)
		if spi == nil {
			return fmt.Errorf("parent module %s not found in summary", moduleName)
		}
		if pmi.QueryType != "" {
			typesMap[pmi.QueryType] = struct{}{}
			if _, ok := fieldsMap[pmi.QueryType]; !ok {
				fieldsMap[pmi.QueryType] = map[string]struct{}{}
			}
			if spi.QueryType != "" {
				fieldsMap[pmi.QueryType][n] = struct{}{}
			}
			if pmi.FunctionType != "" && pmi.Name == "" {
				fieldsMap[pmi.QueryType]["function"] = struct{}{}
			}
		}
		if pmi.MutationType != "" {
			typesMap[pmi.MutationType] = struct{}{}
			if _, ok := fieldsMap[pmi.MutationType]; !ok {
				fieldsMap[pmi.MutationType] = map[string]struct{}{}
			}
			if spi.MutationType != "" {
				fieldsMap[pmi.MutationType][n] = struct{}{}
			}
			if pmi.MutationFunctionType != "" && pmi.Name == "" {
				fieldsMap[pmi.MutationType]["function"] = struct{}{}
			}
		}
		if pmi.FunctionType != "" {
			typesMap[pmi.FunctionType] = struct{}{}
			if spi.FunctionType != "" {
				if _, ok := fieldsMap[pmi.FunctionType]; !ok {
					fieldsMap[pmi.FunctionType] = map[string]struct{}{n: {}}
				}
				fieldsMap[pmi.FunctionType][n] = struct{}{}
			}
		}
		if pmi.MutationFunctionType != "" {
			typesMap[pmi.MutationFunctionType] = struct{}{}
			if spi.MutationFunctionType != "" {
				if _, ok := fieldsMap[pmi.MutationFunctionType]; !ok {
					fieldsMap[pmi.MutationFunctionType] = map[string]struct{}{}
				}
				fieldsMap[pmi.MutationFunctionType][n] = struct{}{}
			}
		}
		pmi = spi
	}
	return nil
}

func addFieldTypeRecursively(schema *SchemaIntro, typeName string, types map[string]struct{}, fields map[string]map[string]struct{}) error {
	t := schema.TypeByName(typeName)
	if t == nil {
		return fmt.Errorf("type %q not found in schema", typeName)
	}
	if _, ok := types[typeName]; ok {
		return nil
	}
	types[typeName] = struct{}{}
	if _, ok := fields[typeName]; !ok {
		fields[typeName] = map[string]struct{}{}
	}
	for _, f := range t.Fields {
		fields[typeName][f.Name] = struct{}{}
		// add field arguments types their types
		for _, a := range f.Args {
			if _, ok := types[a.Type.TypeName()]; ok {
				continue
			}
			err := addFieldTypeRecursively(schema, a.Type.TypeName(), types, fields)
			if err != nil {
				return err
			}
		}
		tn := f.Type.TypeName()
		if _, ok := types[tn]; ok {
			continue
		}
		err := addFieldTypeRecursively(schema, tn, types, fields)
		if err != nil {
			return err
		}
	}
	for _, inp := range t.InputFields {
		fields[typeName][inp.Name] = struct{}{}
		if _, ok := types[inp.Type.TypeName()]; ok {
			continue
		}
		err := addFieldTypeRecursively(schema, inp.Type.TypeName(), types, fields)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) checkTypeExists(ctx context.Context, name string) (bool, error) {
	res, err := s.h.Query(ctx, `query ($name: String!) {
		core {
			mcp {
				types_by_pk(name: $name) {
					name
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return false, fmt.Errorf("query check type exists: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return false, fmt.Errorf("query check type exists: %w", res.Err())
	}
	var data any
	err = res.ScanData("core.mcp.types_by_pk", &data)
	if errors.Is(err, types.ErrNoData) {
		return false, nil
	}

	return data != nil, nil
}

func (s *Service) mergeType(ctx context.Context, t Type, update bool) error {
	// 1. Check if type exists
	exists, err := s.checkTypeExists(ctx, t.Name)
	if err != nil {
		return fmt.Errorf("failed to check if type %q exists: %w", t.Name, err)
	}
	if exists && !update {
		// skip
		return nil
	}
	if !exists {
		return s.AddType(ctx, t)
	}
	return s.updateType(ctx, t)
}

const updateTypeQuery = `mutation ($name: String!, $input: mcp_types_mut_data!) {
		core {
			mcp {
				update_types(
					filter: { name: { eq: $name }}
					data: $input
				) {
					success
				}
			}
		}
	}`

const updateTypeQueryWithEmbedding = `mutation ($name: String!, $input: mcp_types_mut_data!, $summary: String!) {
		core {
			mcp {
				update_types(
					filter: { name: { eq: $name }}
					data: $input
					summary: $summary
				) {
					success
				}
			}
		}
	}`

func (s *Service) updateType(ctx context.Context, t Type) error {
	summary := t.Long
	if summary == "" {
		summary = t.Description
	}
	vars := map[string]any{
		"name": t.Name,
		"input": map[string]any{
			"kind":             t.Kind,
			"module":           t.Module,
			"catalog":          t.Catalog,
			"description":      t.Description,
			"long_description": t.Long,
			"hugr_type":        t.HugrType,
			"is_summarized":    t.IsSummarized,
		},
		"summary": summary,
	}
	query := updateTypeQuery
	if s.c.EmbeddingsEnabled && t.Long != "" {
		query = updateTypeQueryWithEmbedding
	}
	res, err := s.h.Query(ctx, query, vars)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (s *Service) deleteType(ctx context.Context, name string) error {
	res, err := s.h.Query(ctx, `mutation ($name: String!) {
		core {
			mcp {
				delete_arguments(
					filter: { type_name: { eq: $name } }
				) {
					success
				}
				delete_fields(
					filter: { type_name: { eq: $name } }
				) {
					success
				}
				delete_types(
					filter: { name: { eq: $name } }
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete type: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to delete type: %w", res.Err())
	}
	return nil
}

func (s *Service) mergeField(ctx context.Context, f Field, update bool) error {
	// 1. Check if field exists
	exists, err := s.checkFieldExists(ctx, f.TypeName, f.Name)
	if err != nil {
		return fmt.Errorf("failed to check if field %q.%q exists: %w", f.TypeName, f.Name, err)
	}
	if exists && !update {
		// skip
		return nil
	}
	if !exists {
		return s.AddField(ctx, f)
	}
	return s.updateTypeField(ctx, f)
}

func (s *Service) checkFieldExists(ctx context.Context, typeName, fieldName string) (bool, error) {
	res, err := s.h.Query(ctx, `query ($typeName: String!, $fieldName: String!) {
		core {
			mcp {
				fields_by_pk(type_name: $typeName, name: $fieldName) {
					name
				}
			}
		}
	}`, map[string]any{
		"typeName":  typeName,
		"fieldName": fieldName,
	})
	if err != nil {
		return false, fmt.Errorf("query check field exists: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return false, fmt.Errorf("query check field exists: %w", res.Err())
	}
	var data any
	err = res.ScanData("core.mcp.fields_by_pk", &data)
	if errors.Is(err, types.ErrNoData) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query check field exists: %w", err)
	}
	return data != nil, nil
}

const updateTypeFieldMutation = `mutation ($name: String!, $type_name: String!, $input: mcp_fields_mut_data!) {
	core {
		mcp {
			update_fields(
				filter: { name: { eq: $name }, type_name: { eq: $type_name } }
				data: $input
			) {
				success
			}
		}
	}
}`

const updateTypeFieldMutationWithEmbedding = `mutation ($name: String!, $type_name: String!, $input: mcp_fields_mut_data!, $summary: String!) {
	core {
		mcp {
			update_fields(
				filter: { name: { eq: $name }, type_name: { eq: $type_name } }
				data: $input
				summary: $summary
			) {
				success
			}
		}
	}
}`

func (s *Service) updateTypeField(ctx context.Context, f Field) error {
	summary := f.Description
	if summary == "" {
		summary = f.Name
	}
	vars := map[string]any{
		"name":      f.Name,
		"type_name": f.TypeName,
		"input": map[string]any{
			"description":   f.Description,
			"type":          f.Type,
			"hugr_type":     f.HugrType,
			"catalog":       f.Catalog,
			"is_list":       f.IsList,
			"is_non_null":   f.IsNotNull,
			"mcp_exclude":   f.Exclude,
			"is_indexed":    f.IsIndexed,
			"is_summarized": f.IsSummarized,
		},
		"summary": summary,
	}
	query := updateTypeFieldMutation
	if s.c.EmbeddingsEnabled && f.Description != "" {
		query = updateTypeFieldMutationWithEmbedding
	}
	res, err := s.h.Query(ctx, query, vars)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (s *Service) deleteTypeField(ctx context.Context, typeName, fieldName string) error {
	res, err := s.h.Query(ctx, `mutation ($typeName: String!, $fieldName: String!) {
		core {
			mcp {
				delete_fields(
					filter: { 
						type_name: { eq: $typeName }
						name: { eq: $fieldName }
					}
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"typeName":  typeName,
		"fieldName": fieldName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete field: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to delete field: %w", res.Err())
	}
	return nil
}

func (s *Service) mergeArgument(ctx context.Context, arg Argument, update bool) error {
	// 1. Check if argument exists
	exists, err := s.checkArgumentExists(ctx, arg.TypeName, arg.FieldName, arg.Name)
	if err != nil {
		return fmt.Errorf("failed to check if argument %q.%q(%q) exists: %w", arg.TypeName, arg.FieldName, arg.Name, err)
	}
	if exists && !update {
		// skip
		return nil
	}
	if !exists {
		return s.AddArgument(ctx, arg)
	}
	return s.updateArgument(ctx, arg)
}

func (s *Service) checkArgumentExists(ctx context.Context, typeName, fieldName, argName string) (bool, error) {
	res, err := s.h.Query(ctx, `query ($typeName: String!, $fieldName: String!, $argName: String!) {
		core {
			mcp {
				arguments_by_pk(type_name: $typeName, field_name: $fieldName, name: $argName) {
					name
				}
			}
		}
	}`, map[string]any{
		"typeName":  typeName,
		"fieldName": fieldName,
		"argName":   argName,
	})
	if err != nil {
		return false, fmt.Errorf("query check argument exists: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return false, fmt.Errorf("query check argument exists: %w", res.Err())
	}
	var data any
	err = res.ScanData("core.mcp.arguments_by_pk", &data)
	if errors.Is(err, types.ErrNoData) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query check argument exists: %w", err)
	}
	return data != nil, nil
}

func (s *Service) updateArgument(ctx context.Context, arg Argument) error {
	res, err := s.h.Query(ctx, `mutation ($typeName: String!, $fieldName: String!, $argName: String!, $input: mcp_arguments_mut_data!) {
		core {
			mcp {
				update_arguments(
					filter: { 
						type_name: { eq: $typeName }
						field_name: { eq: $fieldName }
						name: { eq: $argName }
					}
					data: $input
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"typeName":  arg.TypeName,
		"fieldName": arg.FieldName,
		"argName":   arg.Name,
		"input": map[string]any{
			"description": arg.Description,
			"type":        arg.Type,
			"is_list":     arg.IsList,
			"is_non_null": arg.IsNotNull,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update argument: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update argument: %w", res.Err())
	}
	return nil
}

func (s *Service) deleteArgument(ctx context.Context, typeName, fieldName, argName string) error {
	res, err := s.h.Query(ctx, `mutation ($typeName: String!, $fieldName: String!, $argName: String!) {
		core {
			mcp {
				delete_arguments(
					filter: { 
						type_name: { eq: $typeName }
						field_name: { eq: $fieldName }
						name: { eq: $argName }
					}
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"typeName":  typeName,
		"fieldName": fieldName,
		"argName":   argName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete argument: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to delete argument: %w", res.Err())
	}
	return nil
}

func (s *Service) mergeDataSource(ctx context.Context, source DataSource, update bool) error {
	// 1. Check if data source exists
	exists, err := s.checkDataSourceExists(ctx, source.Name)
	if err != nil {
		return fmt.Errorf("failed to check if data source %q exists: %w", source.Name, err)
	}
	if exists && !update {
		// skip
		return nil
	}
	if !exists {
		return s.AddDataSource(ctx, source)
	}
	return s.updateDataSource(ctx, source)
}

func (s *Service) checkDataSourceExists(ctx context.Context, name string) (bool, error) {
	res, err := s.h.Query(ctx, `query ($name: String!) {
		core {
			mcp {
				data_sources_by_pk(name: $name) {
					name
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return false, fmt.Errorf("query check data source exists: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return false, fmt.Errorf("query check data source exists: %w", res.Err())
	}
	var data any
	err = res.ScanData("core.mcp.data_sources_by_pk", &data)
	if errors.Is(err, types.ErrNoData) {
		return false, nil
	}

	return data != nil, nil
}

const updateDataSourceMutation = `mutation ($name: String!, $data: mcp_data_sources_mut_data!) {
		core {
			mcp {
				update_data_sources(
					filter: { name: { eq: $name }}
					data: $data
				) {
					success
				}
			}
		}
	}`

const updateDataSourceMutationWithEmbedding = `mutation ($name: String!, $data: mcp_data_sources_mut_data!, $summary: String!) {
		core {
			mcp {
				update_data_sources(
					filter: { name: { eq: $name }}
					data: $data
					summary: $summary
				) {
					success
				}
			}
		}
	}`

func (s *Service) updateDataSource(ctx context.Context, source DataSource) error {
	summary := source.LongDescription
	if summary == "" {
		summary = source.Description
	}

	vars := map[string]any{
		"name": source.Name,
		"data": map[string]any{
			"description":      source.Description,
			"long_description": source.LongDescription,
			"type":             source.Type,
			"prefix":           source.Prefix,
			"as_module":        source.AsModule,
			"read_only":        source.ReadOnly,
			"is_summarized":    source.IsSummarized,
		},
		"summary": summary,
	}
	query := updateDataSourceMutation
	if s.c.EmbeddingsEnabled && source.Description != "" {
		query = updateDataSourceMutationWithEmbedding
	}
	res, err := s.h.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to update data source: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update data source: %w", res.Err())
	}
	return nil
}

func (s *Service) deleteDataSource(ctx context.Context, name string) error {
	res, err := s.h.Query(ctx, `mutation ($name: String!) {
		core {
			mcp {
				delete_data_sources(
					filter: { name: { eq: $name }}
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to delete data source: %w", res.Err())
	}
	return nil
}

func (s *Service) mergeModule(ctx context.Context, module Module, update bool) error {
	// 1. Check if module exists
	exists, err := s.checkModuleExists(ctx, module.Name)
	if err != nil {
		return fmt.Errorf("failed to check if module %q exists: %w", module.Name, err)
	}
	if exists && !update {
		// skip
		return nil
	}
	if !exists {
		return s.AddModule(ctx, module)
	}
	return s.updateModule(ctx, module)
}

func (s *Service) checkModuleExists(ctx context.Context, name string) (bool, error) {
	res, err := s.h.Query(ctx, `query ($name: String!) {
		core {
			mcp {
				modules_by_pk(name: $name) {
					name
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return false, fmt.Errorf("query check module exists: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return false, fmt.Errorf("query check module exists: %w", res.Err())
	}
	var data any
	err = res.ScanData("core.mcp.modules_by_pk", &data)
	if errors.Is(err, types.ErrNoData) {
		return false, nil
	}

	return data != nil, nil
}

const updateModuleMutation = `mutation ($name: String!, $data: mcp_modules_mut_data!) {
		core {
			mcp {
				update_modules(
					filter: { name: { eq: $name }}
					data: $data
				) {
					success
				}
			}
		}
	}`

const updateModuleMutationWithEmbedding = `mutation ($name: String!, $data: mcp_modules_mut_data!, $summary: String!) {
		core {
			mcp {
				update_modules(
					filter: { name: { eq: $name }}
					data: $data
					summary: $summary
				) {
					success
				}
			}
		}
	}`

func (s *Service) updateModule(ctx context.Context, module Module) error {
	summary := module.LongDescription
	if summary == "" {
		summary = module.Description
	}
	vars := map[string]any{
		"name": module.Name,
		"data": map[string]any{
			"description":       module.Description,
			"long_description":  module.LongDescription,
			"query_root":        module.QueryRoot,
			"mutation_root":     module.MutationRoot,
			"function_root":     module.FunctionRoot,
			"mut_function_root": module.MutFunctionRoot,
			"is_summarized":     module.IsSummarized,
		},
		"summary": summary,
	}
	query := updateModuleMutation
	if s.c.EmbeddingsEnabled && module.Description != "" {
		query = updateModuleMutationWithEmbedding
	}
	res, err := s.h.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update module: %w", res.Err())
	}
	return nil
}

func (s *Service) deleteModule(ctx context.Context, name string) error {
	res, err := s.h.Query(ctx, `mutation ($name: String!) {
		core {
			mcp {
				delete_modules(
					filter: { name: { eq: $name }}
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"name": name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to delete module: %w", res.Err())
	}
	return nil
}
