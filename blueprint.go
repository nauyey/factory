package factory

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

const (
	invalidDeleteInstanceTypeErr = "can't delete type(%s) instance, want type(%s) instance"
)

// blueprint represents the runtime instance of a specific Factory model defined before.
// It is created by the function Create and Build.
// A model struct instance will be generated from it.
type blueprint struct {
	factory     *Factory
	table       *table
	traits      []string
	filedValues map[string]interface{}
}

// create creates a model struct instance and save it into database.
// Callback BeforeCreate will be executed after the model struct instance been created
// and before the instance been saved into database.
// Callback AfterCreate will be execute after the model struct instance been saved into database.
func (bp *blueprint) create(db *sql.DB) (interface{}, error) {
	instance := bp.newDefaultInstance()
	bpFieldValues := makeBlueprintFieldValues(bp)

	if err := createInstanceAssociations(db, instance, bpFieldValues.associationFieldValues()); err != nil {
		return nil, err
	}
	if err := bp.setInstanceFieldValues(instance, bpFieldValues); err != nil {
		return nil, err
	}

	// callbacks
	// execute after build callback
	if err := bp.executeAfterBuildCallbacks(instance); err != nil {
		return nil, err
	}
	// execute before create callback
	if err := bp.executeBeforeCreateCallbacks(instance); err != nil {
		return nil, err
	}

	if err := bp.createInstance(db, instance); err != nil {
		return nil, err
	}

	// callbacks
	// execute after build callback
	if err := bp.executeAfterCreateCallbacks(instance); err != nil {
		return nil, err
	}

	return instance.Addr().Interface(), nil
}

// delete deletes a blueprint created instance from database.
// It uses the primary key related field values of the instance.
func (bp *blueprint) delete(db *sql.DB, instance interface{}) error {
	instanceType := reflect.TypeOf(instance)
	instanceValue := reflect.ValueOf(instance)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
		instanceValue = instanceValue.Elem()
	}
	if instanceType != bp.factory.ModelType {
		return fmt.Errorf(invalidDeleteInstanceTypeErr, instanceType.Name(), bp.factory.ModelType.Name())
	}

	primaryValues := []interface{}{}
	for _, col := range bp.table.columns {
		if col.isPrimaryKey {
			primaryValues = append(primaryValues, instanceValue.Field(col.originalModelIndex).Interface())
		}
	}

	_, err := db.Exec(deleteSQL(bp.table.name, bp.table.getPrimaryKeys()), primaryValues...)
	return err
}

// build creates a model struct instance but won't save into database.
// Callback AfterBuild will be execute after the model struct instance been created.
func (bp *blueprint) build() (interface{}, error) {
	instance := bp.newDefaultInstance()
	bpFieldValues := makeBlueprintFieldValues(bp)

	if err := buildInstanceAssociations(instance, bpFieldValues.associationFieldValues()); err != nil {
		return nil, err
	}

	if err := bp.setInstanceFieldValues(instance, bpFieldValues); err != nil {
		return nil, err
	}

	if err := bp.executeAfterBuildCallbacks(instance); err != nil {
		return nil, err
	}

	return instance.Addr().Interface(), nil
}

func (bp *blueprint) newDefaultInstance() reflect.Value {
	f := bp.factory
	return reflect.New(f.ModelType).Elem()
}

func (bp *blueprint) setInstanceFieldValues(instance reflect.Value, bpFieldValues blueprintFieldValues) error {
	for fieldName, fieldValue := range bpFieldValues.filedValues() {
		setInstanceFieldValue(instance, fieldName, fieldValue)
	}

	for fieldName, sequenceValue := range bpFieldValues.sequenceFieldValues() {
		fieldValue, err := sequenceValue.value()
		if err != nil {
			return err
		}
		setInstanceFieldValue(instance, fieldName, fieldValue)
	}

	for fieldName, dynamicFieldValue := range bpFieldValues.dynamicFieldValues() {
		fieldValue, err := dynamicFieldValue(instance.Addr().Interface())
		if err != nil {
			return err
		}
		setInstanceFieldValue(instance, fieldName, fieldValue)
	}

	return nil
}

func (bp *blueprint) createInstance(db *sql.DB, instance reflect.Value) error {
	var (
		fields                  []string
		insertFields            []string
		values                  []interface{}
		queryFieldValuePointers []interface{}
	)

	tbl := bp.table
	queryFieldValuePointers = make([]interface{}, len(tbl.columns))

	for i, col := range tbl.columns {
		iField := instance.Field(col.originalModelIndex)
		queryFieldValuePointers[i] = iField.Addr().Interface()
		fields = append(fields, col.name)
		insertFields = append(insertFields, col.name)
		values = append(values, iField.Interface())

	}

	// insert
	_, err := insertRow(db, insertSQL(tbl.name, insertFields), values...)
	if err != nil {
		return err
	}

	// query
	primaryColumns := tbl.getPrimaryColumns()
	primaryKeys := make([]string, len(primaryColumns))
	primaryKeyValues := make([]interface{}, len(primaryColumns))

	for i, col := range primaryColumns {
		primaryKeys[i] = col.name
		primaryKeyValues[i] = instance.Field(col.originalModelIndex).Interface()
	}

	err = selectRow(db, selectSQL(tbl.name, fields, primaryKeys), primaryKeyValues, queryFieldValuePointers)
	if err != nil {
		return err
	}

	for i, col := range tbl.columns {
		bp.updateModelInstanceField(instance, col.originalModelIndex, queryFieldValuePointers[i])
	}

	return nil
}

// updateModelInstanceField updates value for a model struct instance by field index.
func (bp *blueprint) updateModelInstanceField(instance reflect.Value, index int, value interface{}) {
	instance.Field(index).Set(reflect.ValueOf(value).Elem())
}

func (bp *blueprint) executeAfterBuildCallbacks(modelInstance reflect.Value) error {
	ptrIface := modelInstance.Addr().Interface()

	// execute trait after build callbacks in reverse order of traits
	for i := len(bp.traits) - 1; i >= 0; i-- {
		trait := bp.traits[i]
		traitFactory := bp.factory.Traits[trait]
		if err := executeCallbacks(ptrIface, traitFactory.AfterBuildCallbacks); err != nil {
			return err
		}
	}

	// execute after build callbacks in bp.facotry
	return executeCallbacks(ptrIface, bp.factory.AfterBuildCallbacks)
}

func (bp *blueprint) executeBeforeCreateCallbacks(modelInstance reflect.Value) error {
	ptrIface := modelInstance.Addr().Interface()

	// execute trait before create callbacks in reverse order of traits
	for i := len(bp.traits) - 1; i >= 0; i-- {
		trait := bp.traits[i]
		traitFactory := bp.factory.Traits[trait]
		if err := executeCallbacks(ptrIface, traitFactory.BeforeCreateCallbacks); err != nil {
			return err
		}
	}

	// execute before create callbacks in bp.facotry
	return executeCallbacks(ptrIface, bp.factory.BeforeCreateCallbacks)
}

func (bp *blueprint) executeAfterCreateCallbacks(modelInstance reflect.Value) error {
	ptrIface := modelInstance.Addr().Interface()

	// execute trait after create callbacks in reverse order of traits
	for i := len(bp.traits) - 1; i >= 0; i-- {
		trait := bp.traits[i]
		traitFactory := bp.factory.Traits[trait]
		if err := executeCallbacks(ptrIface, traitFactory.AfterCreateCallbacks); err != nil {
			return err
		}
	}

	// execute after create callbacks in bp.facotry
	return executeCallbacks(ptrIface, bp.factory.AfterCreateCallbacks)
}

func buildInstanceAssociations(instance reflect.Value, associationFieldValues map[string]*AssociationFieldValue) error {
	for fieldName, fieldValue := range associationFieldValues {
		associationBlueprint := newDefaultBlueprintFromAssociationFieldValue(fieldValue)
		associationInterface, err := associationBlueprint.build()
		if err != nil {
			return err
		}
		setInstanceFieldValue(instance, fieldName, associationInterface)

		// set instance reference field value
		value := reflect.ValueOf(associationInterface)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		// FIXME: value.FieldByName(fieldValue.AssociationReferenceField) can't handle mix field name
		setInstanceFieldValue(instance, fieldValue.ReferenceField, value.FieldByName(fieldValue.AssociationReferenceField).Interface())
	}
	return nil
}

// TODO: most of the code is duplicate with buildInstanceAssociations
func createInstanceAssociations(db *sql.DB, instance reflect.Value, associationFieldValues map[string]*AssociationFieldValue) error {
	for fieldName, fieldValue := range associationFieldValues {
		associationBlueprint := newBlueprintFromAssociationFieldValueForCreateAndDelete(fieldValue)
		associationInterface, err := associationBlueprint.create(db)
		if err != nil {
			return err
		}
		setInstanceFieldValue(instance, fieldName, associationInterface)

		value := reflect.ValueOf(associationInterface)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		// FIXME: value.FieldByName(fieldValue.AssociationReferenceField) can't handle mix field name
		setInstanceFieldValue(instance, fieldValue.ReferenceField, value.FieldByName(fieldValue.AssociationReferenceField).Interface())
	}
	return nil
}

type blueprintFieldValues map[string]interface{}

func (bpFieldValues blueprintFieldValues) associationFieldValues() map[string]*AssociationFieldValue {
	fieldValues := map[string]*AssociationFieldValue{}

	for fieldName, fieldValue := range bpFieldValues {
		if value, ok := fieldValue.(*AssociationFieldValue); ok {
			fieldValues[fieldName] = value
		}
	}

	return fieldValues
}

func (bpFieldValues blueprintFieldValues) filedValues() map[string]interface{} {
	fieldValues := map[string]interface{}{}

	for fieldName, fieldValue := range bpFieldValues {
		switch value := fieldValue.(type) {
		default:
			fieldValues[fieldName] = value
		case *sequenceValue, DynamicFieldValue, *AssociationFieldValue:
			continue
		}
	}

	return fieldValues
}

func (bpFieldValues blueprintFieldValues) sequenceFieldValues() map[string]*sequenceValue {
	fieldValues := map[string]*sequenceValue{}

	for fieldName, fieldValue := range bpFieldValues {
		if value, ok := fieldValue.(*sequenceValue); ok {
			fieldValues[fieldName] = value
		}
	}

	return fieldValues
}

func (bpFieldValues blueprintFieldValues) dynamicFieldValues() map[string]DynamicFieldValue {
	fieldValues := map[string]DynamicFieldValue{}

	for fieldName, fieldValue := range bpFieldValues {
		if value, ok := fieldValue.(DynamicFieldValue); ok {
			fieldValues[fieldName] = value
		}
	}

	return fieldValues
}

// makeBlueprintFieldValues create a new field value map of the factory model instance.
// It chooses value for a model struct instance field as following:
// 1. apply the Factory SequenceFiledValues
// 2. apply the Factory FiledValues
// 3. apply the Factory AssociationFieldValue
// 4. apply the Factory DynamicFieldValues
// 5. apply the Factory Traits
// 6. apply the blueprint filedValues
func makeBlueprintFieldValues(bp *blueprint) blueprintFieldValues {
	bpFieldValues := blueprintFieldValues{}
	setBlueprintFieldValuesInBlueprint(bp, bpFieldValues)

	return bpFieldValues
}

func setBlueprintFieldValuesInBlueprint(bp *blueprint, bpFieldValues blueprintFieldValues) {
	// set field values in the Factory
	setBlueprintFieldValuesInFactory(bp.factory, bpFieldValues)

	// set field values in the Factory Traits
	setBlueprintFieldValuesInFactoryTraits(bp.factory, bp.traits, bpFieldValues)

	// set field values in the blueprint filedValues
	for fieldName, fieldValue := range bp.filedValues {
		bpFieldValues[fieldName] = fieldValue
	}
}

func setBlueprintFieldValuesInFactory(f *Factory, bpFieldValues blueprintFieldValues) {
	// set field values in SequenceFiledValues
	for fieldName, sequenceValue := range f.SequenceFiledValues {
		bpFieldValues[fieldName] = sequenceValue
	}

	// set field values in FiledValues
	for fieldName, fieldValue := range f.FiledValues {
		bpFieldValues[fieldName] = fieldValue
	}

	// set filed values in AssociationFieldValue
	for fieldName, associationFieldValue := range f.AssociationFieldValues {
		bpFieldValues[fieldName] = associationFieldValue
	}

	// set field values in DynamicFieldValues
	for fieldName, dynamicFieldValue := range f.DynamicFieldValues {
		bpFieldValues[fieldName] = dynamicFieldValue
	}
}

func setBlueprintFieldValuesInFactoryTraits(f *Factory, traits []string, bpFieldValues blueprintFieldValues) {
	for _, trait := range traits {
		traitFactory := f.Traits[trait]
		setBlueprintFieldValuesInFactory(traitFactory, bpFieldValues)
	}
}

func newDefaultBlueprintFromAssociationFieldValue(fieldValue *AssociationFieldValue) *blueprint {
	return &blueprint{
		factory:     fieldValue.OriginalFactory,
		filedValues: fieldValue.Factory.FiledValues,
	}
}

func newBlueprintFromAssociationFieldValueForCreateAndDelete(fieldValue *AssociationFieldValue) *blueprint {
	bp := newDefaultBlueprintFromAssociationFieldValue(fieldValue)
	bp.table = newTable(fieldValue.OriginalFactory)

	return bp
}

func executeCallbacks(modelInstancePtrIface interface{}, callbacks []Callback) error {
	for _, callback := range callbacks {
		err := callback(modelInstancePtrIface)
		if err != nil {
			return err
		}
	}
	return nil
}

func chainedFieldNameToFieldNames(name string) []string {
	return strings.Split(name, ".")
}

func setInstanceFieldValue(instance reflect.Value, fieldName string, fieldValue interface{}) {
	var field reflect.Value
	var structValue = instance
	fieldNames := chainedFieldNameToFieldNames(fieldName)

	for i, name := range fieldNames {
		if structValue.Kind() == reflect.Ptr {
			structValue = structValue.Elem()
		}
		field = structValue.FieldByName(name)

		if i == len(fieldNames)-1 {
			break
		}

		switch field.Kind() {
		case reflect.Struct:
			structValue = field
		case reflect.Ptr:
			typ := field.Type()
			if typ.Elem().Kind() == reflect.Struct {
				if reflect.DeepEqual(field.Interface(), reflect.Zero(typ).Interface()) {
					field.Set(reflect.New(typ.Elem()).Elem().Addr())
				}
				field = field.Elem()
				structValue = field
			}
		}
	}

	field.Set(reflect.ValueOf(fieldValue))
}
