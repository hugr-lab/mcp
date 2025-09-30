package indexer

import (
	"errors"
	"testing"

	"github.com/hugr-lab/query-engine/pkg/types"
)

func TestService_LoadDataObject(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	doName := "tf_road_parts"

	// get schema intro
	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	// get type intro for data objects
	ti := schema.TypeByName(doName)
	if ti == nil {
		t.Fatalf("type %s not found in schema", doName)
	}

	mi := meta.Module(ti.Module)
	if mi == nil {
		t.Fatalf("module %s not found in summary", ti.Module)
	}

	doi := mi.DataObject(ti.Name)
	if doi == nil {
		t.Fatalf("data object %s not found in module %s", ti.Name, ti.Module)
	}

	typesMap := map[string]struct{}{doName: {}}
	fieldsMap := map[string]map[string]struct{}{
		doName: {},
	}
	// filter type
	if doi.FilterType != "" {
		typesMap[doi.FilterType] = struct{}{}
		// also add filter fields types
		if fti := schema.TypeByName(doi.FilterType); fti != nil {
			fieldsMap[fti.Name] = map[string]struct{}{}
			for _, ff := range fti.Fields {
				if ff.Type.TypeName() != "" {
					typesMap[ff.Type.TypeName()] = struct{}{}
					fieldsMap[fti.Name][ff.Name] = struct{}{}
				}
			}
		}
	}
	// args type
	if doi.Arguments != nil && doi.Arguments.Type != "" {
		typesMap[doi.Arguments.Type] = struct{}{}
		// also add args fields types
		if ati := schema.TypeByName(doi.Arguments.Type); ati != nil {
			fieldsMap[ati.Name] = map[string]struct{}{}
			for _, af := range ati.Fields {
				if af.Type.TypeName() != "" {
					typesMap[af.Type.TypeName()] = struct{}{}
					fieldsMap[ati.Name][af.Name] = struct{}{}
				}
			}
		}
	}
	// field types
	for _, f := range doi.Columns {
		if f.Type != "" {
			typesMap[f.Type] = struct{}{}
			fieldsMap[doName][f.Name] = struct{}{}
			// add fields if it's an object
			if ft := schema.TypeByName(f.Type); ft != nil {
				fieldsMap[ft.Name] = map[string]struct{}{}
				for _, ff := range ft.Fields {
					if ff.Type.TypeName() != "" {
						typesMap[ff.Type.TypeName()] = struct{}{}
						fieldsMap[ft.Name][ff.Name] = struct{}{}
					}
				}
			}
		}
	}
	// relation types
	for _, r := range doi.References {
		if r.FieldDataQuery != "" {
			typesMap[r.FieldDataType] = struct{}{}
			fieldsMap[doName][r.FieldDataQuery] = struct{}{}
		}
		if r.FieldAggQuery != "" {
			typesMap[r.FieldAggDataType] = struct{}{}
			fieldsMap[doName][r.FieldAggQuery] = struct{}{}
		}
		if r.FieldBucketAggQuery != "" {
			typesMap[r.FieldBucketAggType] = struct{}{}
			fieldsMap[doName][r.FieldBucketAggQuery] = struct{}{}
		}
		// filter type
		rm := meta.Module(r.Module)
		if rm == nil {
			t.Fatalf("module %s not found in summary", r.Module)
		}
		rdo := rm.DataObject(r.DataObject)
		if rdo == nil {
			t.Fatalf("data object %s not found in module %s", r.DataObject, r.Module)
		}
		if rdo.FilterType != "" {
			typesMap[rdo.FilterType] = struct{}{}
		}
	}

	// subquery types
	for _, sq := range doi.Subqueries {
		if sq.FieldDataQuery != "" {
			typesMap[sq.FieldDataType] = struct{}{}
			fieldsMap[doName][sq.FieldDataQuery] = struct{}{}
		}
		if sq.FieldAggQuery != "" {
			typesMap[sq.FieldAggDataType] = struct{}{}
			fieldsMap[doName][sq.FieldAggQuery] = struct{}{}
		}
		if sq.FieldBucketAggQuery != "" {
			typesMap[sq.FieldBucketAggType] = struct{}{}
			fieldsMap[doName][sq.FieldBucketAggQuery] = struct{}{}
		}
		// filter type
		sm := meta.Module(sq.Module)
		if sm == nil {
			t.Fatalf("module %s not found in summary", sq.Module)
		}
		sdo := sm.DataObject(sq.DataObject)
		if sdo == nil {
			t.Fatalf("data object %s not found in module %s", sq.DataObject, sq.Module)
		}
		if sdo.FilterType != "" {
			typesMap[sdo.FilterType] = struct{}{}
		}
	}

	// types to add
	// 1. data object type

	// 1. get data object type if not exists, create it
	// 2. get filter type if not exists, create it
	// 3. get args type if not exists, create it

	// process columns

	// data objects to load
	// 1. Data Object types
	// 2. Fields types
	// 3. Args types
	// 4. Filter types
	// 5. Mutation input types

	// types to add if not exists
	// 1. Relations, including aggregations
	// 2. Subqueries, including aggregations
	// 3. Function calls (fields)
	// 4. Enums
	// 5. Scalars
	// 6. Modules
	// 7. Data sources
	// 8. Data objects

	// 1. Check type for data object is exists
	do, err := s.typeInfo(t.Context(), ti.Name)
	if err != nil && !errors.Is(err, types.ErrNoData) {
		t.Fatalf("failed to get type info: %v", err)
	}

	if errors.Is(err, types.ErrNoData) {
		// Load all data object infos (fields, args, queries, filters)
	}
	_ = do
}

func TestService_updateDataObjects(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	doName := "tf_road_parts"

	// get schema intro
	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}

	err = fillDataObjectTypesForUpdate(schema, meta, doName, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		t.Fatalf("failed to fill types for update: %v", err)
	}

	// update types
	t.Logf("Types to update: %d", len(typesMap))
	t.Logf("Types for fields update: %d", len(fieldsMap))
	fc := 0
	for _, fields := range fieldsMap {
		fc += len(fields)
	}
	t.Logf("Fields for update: %d", fc)
	t.Logf("Modules to update: %d", len(modulesMap))
	t.Logf("Data sources to update: %d", len(dataSourcesMap))

	// loop over types and filter it to update, create fields and arguments to insert if exists

}

func TestService_fillAllDataObjects(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	doo := meta.DataObjects()
	if len(doo) == 0 {
		t.Fatal("no data objects found")
	}
	t.Logf("found %d data objects", len(doo))

	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}

	for _, do := range doo {
		err = fillDataObjectTypesForUpdate(schema, meta, do.Name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
		if err != nil {
			t.Fatalf("failed to fill types for update: %v", err)
		}
	}

	// update types
	t.Logf("Types to update: %d", len(typesMap))
	t.Logf("Types for fields update: %d", len(fieldsMap))
	fc := 0
	for _, fields := range fieldsMap {
		fc += len(fields)
	}
	t.Logf("Fields for update: %d", fc)
	t.Logf("Modules to update: %d", len(modulesMap))
	t.Logf("Data sources to update: %d", len(dataSourcesMap))
}

func TestService_fillFunctionTypesForUpdate(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	funcName := "current_weather"
	moduleName := "owm"

	// get schema intro
	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}

	err = fillFunctionTypesForUpdate(schema, meta, moduleName, funcName, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		t.Fatalf("failed to fill types for update: %v", err)
	}

	// update types
	t.Logf("Types to update: %d", len(typesMap))
	t.Logf("Types for fields update: %d", len(fieldsMap))
	fc := 0
	for _, fields := range fieldsMap {
		fc += len(fields)
	}
	t.Logf("Fields for update: %d", fc)
	t.Logf("Modules to update: %d", len(modulesMap))
	t.Logf("Data sources to update: %d", len(dataSourcesMap))
}

func TestService_fillAllFunctionTypesForUpdate(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	funcs := meta.Functions()
	t.Logf("found %d functions", len(funcs))
	mutFunctions := meta.MutationFunctions()
	t.Logf("found %d mutation functions", len(mutFunctions))

	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}

	for _, f := range funcs {
		err = fillFunctionTypesForUpdate(schema, meta, f.Module, f.Name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
		if err != nil {
			t.Fatalf("failed to fill types for update: %v", err)
		}
	}
	for _, f := range mutFunctions {
		err = fillFunctionTypesForUpdate(schema, meta, f.Module, f.Name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
		if err != nil {
			t.Fatalf("failed to fill types for update: %v", err)
		}
	}

	// update types
	t.Logf("Types to update: %d", len(typesMap))
	t.Logf("Types for fields update: %d", len(fieldsMap))
	fc := 0
	for _, fields := range fieldsMap {
		fc += len(fields)
	}
	t.Logf("Fields for update: %d", fc)
	t.Logf("Modules to update: %d", len(modulesMap))
	t.Logf("Data sources to update: %d", len(dataSourcesMap))
}

func TestService_fillAllDataObjectsAndFunctionsForUpdate(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}
	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}

	// fill data objects
	doo := meta.DataObjects()
	t.Logf("found %d data objects", len(doo))
	for _, do := range doo {
		err = fillDataObjectTypesForUpdate(schema, meta, do.Name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
		if err != nil {
			t.Fatalf("failed to fill types for update: %v", err)
		}
	}
	// fill functions
	funcs := meta.Functions()
	t.Logf("found %d functions", len(funcs))
	mutFunctions := meta.MutationFunctions()
	t.Logf("found %d mutation functions", len(mutFunctions))
	for _, f := range funcs {
		err = fillFunctionTypesForUpdate(schema, meta, f.Module, f.Name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
		if err != nil {
			t.Fatalf("failed to fill types for update: %v", err)
		}
	}
	for _, f := range mutFunctions {
		err = fillFunctionTypesForUpdate(schema, meta, f.Module, f.Name, typesMap, fieldsMap, modulesMap, dataSourcesMap)
		if err != nil {
			t.Fatalf("failed to fill types for update: %v", err)
		}
	}

	// update types
	t.Logf("Types to update: %d", len(typesMap))
	t.Logf("Types for fields update: %d", len(fieldsMap))
	fc := 0
	for _, fields := range fieldsMap {
		fc += len(fields)
	}
	t.Logf("Fields for update: %d", fc)
	t.Logf("Modules to update: %d", len(modulesMap))
	t.Logf("Data sources to update: %d", len(dataSourcesMap))
	// types that loaded but not field updated
	for tt := range typesMap {
		if _, ok := fieldsMap[tt]; !ok {
			t.Logf("Type %s has no fields to update", tt)
		}
	}
}

func TestService_checkTypeExsits(t *testing.T) {
	s := New(testConfig, testHugr)

	exists, err := s.checkTypeExists(t.Context(), "tf_road_parts")
	if err != nil {
		t.Fatalf("failed to check type exists: %v", err)
	}
	if !exists {
		t.Fatalf("type tf_road_parts does not exist")
	}
	t.Logf("type tf_road_parts exists: %v", exists)

	exists, err = s.checkTypeExists(t.Context(), "____non_existing_type")
	if err != nil {
		t.Fatalf("failed to check type exists: %v", err)
	}
	if exists {
		t.Fatalf("type ____non_existing_type should not exist")
	}
	t.Logf("type ____non_existing_type exists: %v", exists)
}

func TestService_clearDataObjectFields(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}
	doName := "tf2_attributes"
	err = s.clearDataObjectTypes(t.Context(), schema, meta, doName)
	if err != nil {
		t.Fatalf("failed to clear data object types: %v", err)
	}
}

func TestService_loadPatrial(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
	doName := "tf2_road_parts"

	// get schema intro
	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	typesMap := map[string]struct{}{}
	fieldsMap := map[string]map[string]struct{}{}
	modulesMap := map[string]struct{}{}
	dataSourcesMap := map[string]struct{}{}

	err = fillDataObjectTypesForUpdate(schema, meta, doName, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		t.Fatalf("failed to fill types for update: %v", err)
	}

	// update types
	t.Logf("Types to update: %d", len(typesMap))
	t.Logf("Types for fields update: %d", len(fieldsMap))
	fc := 0
	for _, fields := range fieldsMap {
		fc += len(fields)
	}
	t.Logf("Fields for update: %d", fc)
	t.Logf("Modules to update: %d", len(modulesMap))
	t.Logf("Data sources to update: %d", len(dataSourcesMap))
	err = s.loadSchemaPatrial(t.Context(), schema, meta, false, typesMap, fieldsMap, modulesMap, dataSourcesMap)
	if err != nil {
		t.Fatalf("failed to load schema partial: %v", err)
	}
}
