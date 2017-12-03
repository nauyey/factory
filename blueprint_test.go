package factory

import "testing"
import "reflect"

func TestSetInstanceFieldValue(t *testing.T) {
	type test2 struct {
		A string
		b int
	}

	type test struct {
		A string
		b int
		C *string
		D test2
		E *test2
	}

	// test field
	tt := &test{}
	v := reflect.ValueOf(tt)

	setInstanceFieldValue(v, "A", "aaa")
	if tt.A != "aaa" {
		t.Errorf("setInstanceFieldValue failed")
	}

	// test ptr field
	tt = &test{}
	v = reflect.ValueOf(tt)
	c := "ccc"

	setInstanceFieldValue(v, "C", &c)
	if *tt.C != "ccc" {
		t.Errorf("setInstanceFieldValue failed")
	}

	// test struct field

	tt = &test{}
	v = reflect.ValueOf(tt)
	d := test2{
		A: "test2-AAA",
		b: 1,
	}

	setInstanceFieldValue(v, "D", d)
	if tt.D.A != "test2-AAA" {
		t.Errorf("setInstanceFieldValue failed")
	}
	if tt.D.b != 1 {
		t.Errorf("setInstanceFieldValue failed")
	}

	// test ptr struct field

	tt = &test{}
	v = reflect.ValueOf(tt)
	e := &test2{
		A: "ptr test2-AAA",
		b: 2,
	}

	setInstanceFieldValue(v, "E", e)
	if tt.E.A != "ptr test2-AAA" {
		t.Errorf("setInstanceFieldValue failed")
	}
	if tt.E.b != 2 {
		t.Errorf("setInstanceFieldValue failed")
	}

	// test sub field of struct field

	tt = &test{}
	v = reflect.ValueOf(tt)

	setInstanceFieldValue(v, "D.A", "D.A-AAA")
	if tt.D.A != "D.A-AAA" {
		t.Errorf("setInstanceFieldValue failed")
	}
	if tt.D.b != 0 {
		t.Errorf("setInstanceFieldValue failed")
	}

	// test ptr struct field

	tt = &test{}
	v = reflect.ValueOf(tt)

	setInstanceFieldValue(v, "E.A", "ptr E.A-AAA")
	if tt.E.A != "ptr E.A-AAA" {
		t.Errorf("setInstanceFieldValue failed")
	}
	if tt.E.b != 0 {
		t.Errorf("setInstanceFieldValue failed")
	}
}
