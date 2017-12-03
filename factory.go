package factory

import (
	"reflect"
)

// Factory represents a factory defined by some model struct
type Factory struct {
	ModelType              reflect.Type
	Table                  string
	FiledValues            map[string]interface{}
	SequenceFiledValues    map[string]*sequenceValue
	DynamicFieldValues     map[string]DynamicFieldValue
	AssociationFieldValues map[string]*AssociationFieldValue
	Traits                 map[string]*Factory
	AfterBuildCallbacks    []Callback
	BeforeCreateCallbacks  []Callback
	AfterCreateCallbacks   []Callback

	CanHaveAssociations bool
	CanHaveTraits       bool
	CanHaveCallbacks    bool
}

// AddSequenceFiledValue adds sequence field value to factory by field name
func (f *Factory) AddSequenceFiledValue(name string, first int64, value SequenceFieldValue) {
	if f.SequenceFiledValues == nil {
		f.SequenceFiledValues = map[string]*sequenceValue{}
	}
	f.SequenceFiledValues[name] = newSequenceValue(first, value)
}

// sequenceValue defines the value of a sequence field.
type sequenceValue struct {
	valueGenerateFunc SequenceFieldValue
	sequence          *sequence
}

// value calculates the value of current sequenceValue
func (seqValue *sequenceValue) value() (interface{}, error) {
	return seqValue.valueGenerateFunc(seqValue.sequence.next())
}

// newSequenceValue create a new SequenceValue instance
func newSequenceValue(first int64, value SequenceFieldValue) *sequenceValue {
	return &sequenceValue{
		valueGenerateFunc: value,
		sequence: &sequence{
			first: first,
		},
	}
}

// SequenceFieldValue defines the value generator type of sequence field.
// It's return result will be set as the value of the sequence field dynamicly.
type SequenceFieldValue func(n int64) (interface{}, error)

// DynamicFieldValue defines the value generator type of a field.
// It's return result will be set as the value of the field dynamicly.
type DynamicFieldValue func(model interface{}) (interface{}, error)

// AssociationFieldValue represents a struct which contains data to generate value of a association field.
type AssociationFieldValue struct {
	ReferenceField            string
	AssociationReferenceField string
	OriginalFactory           *Factory
	Factory                   *Factory
}

// Callback defines the callback function type
type Callback func(model interface{}) error
