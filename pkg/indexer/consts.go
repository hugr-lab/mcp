package indexer

const (
	HugrTypeModule       HugrType = "module"
	HugrTypeTable        HugrType = "table"
	HugrTypeView         HugrType = "view"
	HugrTypeFilter       HugrType = "filter"
	HugrTypeAggs         HugrType = "aggs"
	HugrTypeViewArgument HugrType = "view_argument"
	HugrTypeFunction     HugrType = "function"
	HugrTypeSystem       HugrType = "system"
	HugrTypeScalar       HugrType = "scalar"
	HugrTypeScalarFilter HugrType = "scalar_filter"
	HugrTypeScalarAggs   HugrType = "scalar_aggs"
)

const (
	HugrFieldTypeSubModule    HugrFieldType = "submodule"
	HugrFieldTypeField        HugrFieldType = "field"
	HugrFieldTypeExtraField   HugrFieldType = "extra_field"
	HugrFieldTypeReference    HugrFieldType = "references_query"
	HugrFieldTypeJoin         HugrFieldType = "join"
	HugrFieldTypeFunctionCall HugrFieldType = "function_call"
	HugrFieldTypeQueryData    HugrFieldType = "query_data"
	HugrFieldTypeQueryOne     HugrFieldType = "query_one"
	HugrFieldTypeQueryAgg     HugrFieldType = "query_agg"
	HugrFieldTypeQuerySubAgg  HugrFieldType = "query_sub_agg"
	HugrFieldTypeQueryBucket  HugrFieldType = "query_bucket_agg"
	HugrFieldTypeFunction     HugrFieldType = "function"
	HugrFieldTypeMutationFunc HugrFieldType = "mutation_function"
	HugrFieldTypeMutationIns  HugrFieldType = "mutation_insert"
	HugrFieldTypeMutationUpd  HugrFieldType = "mutation_update"
	HugrFieldTypeMutationDel  HugrFieldType = "mutation_delete"
)

const (
	HugrArgumentTypeScalar     HugrArgumentType = "scalar"
	HugrArgumentTypeFilter     HugrArgumentType = "filter"
	HugrArgumentTypeDataInsert HugrArgumentType = "data_insert"
	HugrArgumentTypeDataUpdate HugrArgumentType = "data_update"
	HugrArgumentTypeOrderBy    HugrArgumentType = "order_by"
	HugrArgumentTypeFieldName  HugrArgumentType = "field_name"
)
