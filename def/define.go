package def

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nauyey/factory"
)

// def supports the following features:
// Aliases
// Dynamic Fields
// Dependent Fields
// Sequences
// Traits
// Callbacks
// Associations
// Multilevel Fields

// TODO: generate association tree and check associations circle

const (
	invalidFieldNameErr         = "invalid field name %s to define factory of %s"
	invalidFieldValueTypeErr    = "cannot use value (type %v) as type %v of field %s to define factory of %s"
	nestedAssociationErr        = "association %s error: nested associations isn't allowed"
	nestedTraitErr              = "Trait %s error: nested traits is not allowed"
	callbackInAssociationErr    = "%s is not allowed in Associations"
	duplicateFieldDefinitionErr = "duplicate definition of field %s"
)

func newDefaultFactory(model interface{}, table string) *factory.Factory {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return &factory.Factory{
		ModelType:              modelType,
		Table:                  table,
		FiledValues:            map[string]interface{}{},
		DynamicFieldValues:     map[string]factory.DynamicFieldValue{},
		AssociationFieldValues: map[string]*factory.AssociationFieldValue{},
		Traits:                 map[string]*factory.Factory{},

		CanHaveAssociations: true,
		CanHaveTraits:       true,
		CanHaveCallbacks:    true,
	}
}

func newDefaultFactoryForTrait(f *factory.Factory) *factory.Factory {
	return &factory.Factory{
		ModelType:              f.ModelType,
		FiledValues:            map[string]interface{}{},
		DynamicFieldValues:     map[string]factory.DynamicFieldValue{},
		AssociationFieldValues: map[string]*factory.AssociationFieldValue{},

		CanHaveAssociations: true,
		CanHaveTraits:       false,
		CanHaveCallbacks:    true,
	}
}

func newDefaultFactoryForAssociation(f *factory.Factory) *factory.Factory {
	return &factory.Factory{
		ModelType:          f.ModelType,
		FiledValues:        map[string]interface{}{},
		DynamicFieldValues: map[string]factory.DynamicFieldValue{},

		CanHaveAssociations: false,
		CanHaveTraits:       false,
		CanHaveCallbacks:    false,
	}
}

type definitionOption func(*factory.Factory) error

// Field defines the value of a field in the factory
func Field(name string, value interface{}) definitionOption {
	return func(f *factory.Factory) error {
		if ok := fieldExists(f.ModelType, name); !ok {
			return fmt.Errorf(invalidFieldNameErr, name, f.ModelType.Name())
		}

		field, _ := structFieldByName(f.ModelType, name)
		if valueType := reflect.TypeOf(value); valueType != field.Type {
			return fmt.Errorf(invalidFieldValueTypeErr, valueType, field.Type, name, f.ModelType.Name())
		}

		if ok := definedField(f, name); ok {
			return fmt.Errorf(duplicateFieldDefinitionErr, name)
		}

		f.FiledValues[name] = value
		return nil
	}
}

// SequenceField defines the value of a squence field in the factory.
// Unique values in a specific format (for example, e-mail addresses) can be generated using sequences.
func SequenceField(name string, first int64, value factory.SequenceFieldValue) definitionOption {
	return func(f *factory.Factory) error {
		if ok := fieldExists(f.ModelType, name); !ok {
			return fmt.Errorf(invalidFieldNameErr, name, f.ModelType.Name())
		}

		if ok := definedField(f, name); ok {
			return fmt.Errorf(duplicateFieldDefinitionErr, name)
		}

		f.AddSequenceFiledValue(name, first, value)
		return nil
	}
}

// DynamicField defines the value generator of a dynamic field in the factory.
func DynamicField(name string, value factory.DynamicFieldValue) definitionOption {
	return func(f *factory.Factory) error {
		if ok := fieldExists(f.ModelType, name); !ok {
			return fmt.Errorf(invalidFieldNameErr, name, f.ModelType.Name())
		}

		if ok := definedField(f, name); ok {
			return fmt.Errorf(duplicateFieldDefinitionErr, name)
		}

		f.DynamicFieldValues[name] = value
		return nil
	}
}

// Association defines the value of a association field
func Association(name, referenceField, associationReferenceField string, originalFactory *factory.Factory, opts ...definitionOption) definitionOption {
	return func(f *factory.Factory) error {
		if ok := fieldExists(f.ModelType, name); !ok {
			return fmt.Errorf(invalidFieldNameErr, name, f.ModelType.Name())
		}
		// TODO: check ReferenceField
		// TODO: check factory.ModelType with association field type

		if !f.CanHaveAssociations {
			return fmt.Errorf(nestedAssociationErr, name)
		}

		associationFieldValue := &factory.AssociationFieldValue{
			ReferenceField:            referenceField,
			AssociationReferenceField: associationReferenceField,
			OriginalFactory:           originalFactory,
			Factory:                   newDefaultFactoryForAssociation(originalFactory),
		}

		for _, opt := range opts {
			err := opt(associationFieldValue.Factory)
			if err != nil {
				return err
			}
		}

		if ok := definedField(f, name); ok {
			return fmt.Errorf(duplicateFieldDefinitionErr, name)
		}

		f.AssociationFieldValues[name] = associationFieldValue

		return nil
	}
}

// Trait allows you to group fields together and then apply them to any factory.
func Trait(traitName string, opts ...definitionOption) definitionOption {
	return func(f *factory.Factory) error {
		if !f.CanHaveTraits {
			return fmt.Errorf(nestedTraitErr, traitName)
		}

		traitFactory := newDefaultFactoryForTrait(f)

		for _, opt := range opts {
			err := opt(traitFactory)
			if err != nil {
				return err
			}
		}

		f.Traits[traitName] = traitFactory

		return nil
	}
}

// AfterBuild sets callback called after the model struct been build.
// REMIND that AfterBuild callback will be called not only when Build a model struct
// but also when Create a model struct. Because to Create a model struct instance,
// we must build it first.
func AfterBuild(callback factory.Callback) definitionOption {
	return func(f *factory.Factory) error {
		if !f.CanHaveCallbacks {
			return fmt.Errorf(callbackInAssociationErr, "AfterBuild")
		}

		f.AfterBuildCallbacks = append(f.AfterBuildCallbacks, callback)
		return nil
	}
}

// BeforeCreate sets callback called before the model struct been saved.
func BeforeCreate(callback factory.Callback) definitionOption {
	return func(f *factory.Factory) error {
		if !f.CanHaveCallbacks {
			return fmt.Errorf(callbackInAssociationErr, "BeforeCreate")
		}

		f.BeforeCreateCallbacks = append(f.BeforeCreateCallbacks, callback)
		return nil
	}
}

// AfterCreate sets callback called after the model struct been saved.
func AfterCreate(callback factory.Callback) definitionOption {
	return func(f *factory.Factory) error {
		if !f.CanHaveCallbacks {
			return fmt.Errorf(callbackInAssociationErr, "AfterCreate")
		}

		f.AfterCreateCallbacks = append(f.AfterCreateCallbacks, callback)
		return nil
	}
}

// NewFactory defines a factory of a model struct.
// Parameter model is the model struct instance(or struct instance pointer).
// Parameter table represents which database table this model will be saved.
// Usage example:
// Defining factories
//
// type Model struct {
// 		ID int64
// 		Name string
// }
//
// FactoryModel := NewFactory(Model{}, "model_table",
// 	Field("Name", "test name"),
// 	SequenceField("ID", func(n int64) interface{} {
// 		return n
// 	}),
// 	Trait("Chinese",
// 		Field("Country", "China"),
// 	),
// 	BeforeCreate(func(model interface{}) error {
// 		// do something
// 	}),
// 	AfterCreate(func(model interface{}) error {
// 		// do something
// 	}),
// )
//
func NewFactory(model interface{}, table string, opts ...definitionOption) *factory.Factory {
	f := newDefaultFactory(model, table)

	for _, opt := range opts {
		err := opt(f)
		if err != nil {
			panic(err)
		}
	}

	return f
}

type factoryField []string

func fieldNameToFactoryField(name string) factoryField {
	return strings.Split(name, ".")
}

// TODO: confirm if should handle panic
func fieldExists(typ reflect.Type, name string) bool {
	fields := fieldNameToFactoryField(name)

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

	fieldNames := fieldNameToFactoryField(name)
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

func definedField(f *factory.Factory, name string) bool {
	// FiledValues
	if _, ok := f.FiledValues[name]; ok {
		return true
	}
	// SequenceFiledValues
	if _, ok := f.SequenceFiledValues[name]; ok {
		return true
	}
	// DynamicFieldValues
	if _, ok := f.DynamicFieldValues[name]; ok {
		return true
	}
	// AssociationFieldValues
	if _, ok := f.AssociationFieldValues[name]; ok {
		return true
	}

	return false
}
