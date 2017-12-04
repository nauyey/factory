package factory

import (
	"fmt"
	"reflect"
)

const (
	invalidFieldNameErr      = "invalid field name %s to define factory of %s"
	invalidFieldValueTypeErr = "cannot use value (type %v) as type %v of field %s to define factory of %s"
	undefinedTraitErr        = "undefined trait name %s of type %s factory"
)

func newDefaultBlueprint(f *Factory) *blueprint {
	return &blueprint{
		factory:     f,
		filedValues: map[string]interface{}{},
	}
}

func newDefaultBlueprintForCreate(f *Factory) *blueprint {
	bp := newDefaultBlueprint(f)
	bp.table = newTable(f)

	return bp
}

func newDefaultBlueprintForDelete(f *Factory) *blueprint {
	bp := newDefaultBlueprint(f)
	bp.table = newTable(f)

	return bp
}

type factoryOption func(*blueprint) error

// WithTraits defines which traits the new instance will use.
// It can take multiple traits. These traits will be executed one by one.
// So the later one may override the one before.
//
// For example:
//
// The trait "trait1" set Field1 as "value1", and at the same time, trait "trait2" set Field1 as "value2".
// The WithTraits("trait1", "trait2") will finally set Field1 as "value2".
func WithTraits(traits ...string) factoryOption {
	return func(bp *blueprint) error {
		for _, trait := range traits {
			if _, ok := bp.factory.Traits[trait]; !ok {
				return fmt.Errorf(undefinedTraitErr, trait, bp.factory.ModelType.Name())
			}
			bp.traits = append(bp.traits, trait)
		}
		return nil
	}
}

// WithField sets the value of a specific field.
// This way has the highest priority to set the field value.
func WithField(name string, value interface{}) factoryOption {
	return func(bp *blueprint) error {
		modelTypeName := bp.factory.ModelType.Name()
		if ok := fieldExists(bp.factory.ModelType, name); !ok {
			return fmt.Errorf(invalidFieldNameErr, name, modelTypeName)
		}

		field, _ := structFieldByName(bp.factory.ModelType, name)
		if valueType := reflect.TypeOf(value); valueType != field.Type {
			return fmt.Errorf(invalidFieldValueTypeErr, valueType, field.Type, name, modelTypeName)
		}

		bp.filedValues[name] = value
		return nil
	}
}

// Build creates an instance from a factory
// but won't store it into database.
//
// model := &Model{}
//
// err := Build(FactoryModel,
// 	WithTrait("Chinese"),
// 	WithField("Name", "new name"),
// 	WithField("ID", 123),
// ).To(model)
//
func Build(f *Factory, opts ...factoryOption) to {
	bp := newDefaultBlueprint(f)

	for _, opt := range opts {
		opt(bp)
	}

	return &buildTo{
		blueprint: bp,
	}
}

// BuildSlice creates a slice instance from a factory
// but won't store them into database.
//
// modelSlice := []*Model{}
//
// err := Build(FactoryModel,
// 	WithTrait("Chinese"),
// 	WithField("Name", "new name"),
// ).To(&modelSlice)
//
func BuildSlice(f *Factory, count int, opts ...factoryOption) to {
	bp := newDefaultBlueprint(f)

	for _, opt := range opts {
		opt(bp)
	}

	return &buildSliceTo{
		blueprint: bp,
		count:     count,
	}
}

// Create creates an instance from a factory
// and stores it into database.
//
// model := &Model{}
//
// err := Create(FactoryModel,
// 	WithTrait("Chinese"),
// 	WithField("Name", "new name"),
// 	WithField("ID", 123),
// ).To(model)
//
func Create(f *Factory, opts ...factoryOption) to {
	bp := newDefaultBlueprintForCreate(f)

	for _, opt := range opts {
		opt(bp)
	}

	return &createTo{
		blueprint:    bp,
		dbConnection: getDB(),
	}
}

// CreateSlice creates a slice of instance from a factory
// and stores them into database.
//
// modelSlice := []*Model{}
//
// err := CreateSlice(FactoryModel,
// 	WithTrait("Chinese"),
// 	WithField("Name", "new name"),
// ).To(&modelSlice)
//
func CreateSlice(f *Factory, count int, opts ...factoryOption) to {
	bp := newDefaultBlueprintForCreate(f)

	for _, opt := range opts {
		opt(bp)
	}

	return &createSliceTo{
		blueprint:    bp,
		count:        count,
		dbConnection: getDB(),
	}
}

// Delete deletes an instance of a factory model from database.
// Example:
// err := Delete(FactoryModel, Model{})
//
func Delete(f *Factory, instance interface{}) error {
	bp := newDefaultBlueprintForDelete(f)

	return bp.delete(getDB(), instance)
}

// the following code are duplicated with "github.com/nauyey/factory/def"

// TODO: confirm if should handle panic
func fieldExists(typ reflect.Type, name string) bool {
	fields := chainedFieldNameToFieldNames(name)

	for i, field := range fields {
		f, ok := typ.FieldByName(field)
		if !ok {
			return false
		}

		if i == len(fields)-1 {
			break
		}

		// TODO: Optimize me for only type struct or *struct is valid
		typ = f.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	}

	return true
}

// TODO: confirm if should handle panic
func structFieldByName(typ reflect.Type, name string) (*reflect.StructField, bool) {
	var field *reflect.StructField

	fieldNames := chainedFieldNameToFieldNames(name)
	if len(fieldNames) == 0 {
		return nil, false
	}

	for i, fieldName := range fieldNames {
		f, ok := typ.FieldByName(fieldName)
		if !ok {
			return nil, false
		}
		field = &f

		if i == len(fieldNames)-1 {
			break
		}

		typ = f.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	}

	return field, true
}
