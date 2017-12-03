package factory

import (
	"database/sql"
	"fmt"
	"reflect"
)

const (
	invalidTargetTypeErr      = "cannot use target (type *%v) as type *%v in func To"
	invalidTargetSliceTypeErr = "cannot use target (type []*%v) as type []*%v in func To"
)

type to interface {
	To(target interface{}) error
}

type buildTo struct {
	blueprint *blueprint
}

func (to *buildTo) To(target interface{}) error {
	if err := checkTargetType(to.blueprint.factory.ModelType, target); err != nil {
		return err
	}

	instanceIface, err := to.blueprint.build()
	if err != nil {
		return err
	}

	setValue(target, instanceIface)
	return nil
}

type buildSliceTo struct {
	blueprint *blueprint
	count     int
}

func (to *buildSliceTo) To(target interface{}) error {
	targetType, targetValue := targetTypeAndValue(target)
	elemType, isPtrElem := elemTypeOf(targetType)

	// check element type of target slice
	if err := checkTargetSliceType(to.blueprint.factory.ModelType, elemType); err != nil {
		return err
	}

	sliceValue := reflect.MakeSlice(targetType, 0, to.count)
	for i := 0; i < to.count; i++ {
		elemIface, err := to.blueprint.build()
		if err != nil {
			return err
		}

		sliceValue = appendSliceValue(sliceValue, isPtrElem, reflect.ValueOf(elemIface))
	}
	targetValue.Set(sliceValue)

	return nil
}

type createTo struct {
	blueprint    *blueprint
	dbConnection *sql.DB
}

func (to *createTo) To(target interface{}) error {
	if err := checkTargetType(to.blueprint.factory.ModelType, target); err != nil {
		return err
	}

	instanceIface, err := to.blueprint.create(to.dbConnection)
	if err != nil {
		return err
	}

	setValue(target, instanceIface)
	return nil
}

type createSliceTo struct {
	blueprint    *blueprint
	count        int
	dbConnection *sql.DB
}

func (to *createSliceTo) To(target interface{}) error {
	targetType, targetValue := targetTypeAndValue(target)
	elemType, isPtrElem := elemTypeOf(targetType)

	// check element type of target slice
	if err := checkTargetSliceType(to.blueprint.factory.ModelType, elemType); err != nil {
		return err
	}

	sliceValue := reflect.MakeSlice(targetType, 0, to.count)
	for i := 0; i < to.count; i++ {
		elemIface, err := to.blueprint.create(to.dbConnection)
		if err != nil {
			return err
		}

		sliceValue = appendSliceValue(sliceValue, isPtrElem, reflect.ValueOf(elemIface))
	}
	targetValue.Set(sliceValue)

	return nil
}

func targetTypeAndValue(target interface{}) (reflect.Type, reflect.Value) {
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
		targetValue = targetValue.Elem()
	}

	return targetType, targetValue
}

func elemTypeOf(targetType reflect.Type) (elemType reflect.Type, isPtrElem bool) {
	elemType = targetType.Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
		isPtrElem = true
	}
	return
}

func checkTargetType(wantType reflect.Type, target interface{}) error {
	typ := reflect.TypeOf(target).Elem()
	if typ != wantType {
		return fmt.Errorf(invalidTargetTypeErr, typ, wantType)
	}
	return nil
}

func checkTargetSliceType(wantType, targetSliceElemType reflect.Type) error {
	if targetSliceElemType != wantType {
		return fmt.Errorf(invalidTargetSliceTypeErr, targetSliceElemType, wantType)
	}
	return nil
}

func setValue(dest interface{}, src interface{}) {
	targetValue := reflect.ValueOf(dest).Elem()
	targetValue.Set(reflect.ValueOf(src).Elem())
}

func appendSliceValue(sliceValue reflect.Value, isPtrElem bool, elemValue reflect.Value) reflect.Value {
	if isPtrElem {
		sliceValue = reflect.Append(sliceValue, elemValue)
	} else {
		sliceValue = reflect.Append(sliceValue, elemValue.Elem())
	}

	return sliceValue
}
