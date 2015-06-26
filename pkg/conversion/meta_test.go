/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conversion

import (
	"reflect"
	"testing"
)

func TestSimpleMetaFactoryInterpret(t *testing.T) {
	factory := SimpleMetaFactory{}
	tm, err := factory.Interpret([]byte(`{"apiGroup":"0","apiVersion":"1","kind":"object"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tm.APIGroup != "0" || tm.APIVersion != "1" || tm.Kind != "object" {
		t.Errorf("unexpected interpret: %v", tm)
	}

	// no kind or version
	tm, err = factory.Interpret([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (tm != TypeMeta{}) {
		t.Errorf("unexpected interpret: %s %s", tm)
	}

	// unparsable
	tm, err = factory.Interpret([]byte(`{`))
	if err == nil {
		t.Errorf("unexpected non-error")
	}
}

func TestSimpleMetaFactoryUpdate(t *testing.T) {
	factory := SimpleMetaFactory{GroupField: "G", VersionField: "V", KindField: "K"}

	obj := struct {
		G string
		V string
		K string
	}{"0", "1", "2"}

	// must pass a pointer
	if err := factory.Update(TypeMeta{"first", "test", "other"}, obj); err == nil {
		t.Errorf("unexpected non-error")
	}
	if obj.G != "0" || obj.V != "1" || obj.K != "2" {
		t.Errorf("unexpected update: %v", obj)
	}

	// updates
	if err := factory.Update(TypeMeta{"first", "test", "other"}, &obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.G != "first" || obj.V != "test" || obj.K != "other" {
		t.Errorf("unexpected update: %v", obj)
	}
}

func TestSimpleMetaFactoryUpdateStruct(t *testing.T) {
	factory := SimpleMetaFactory{
		BaseFields:   []string{"Test"},
		GroupField:   "G",
		VersionField: "V",
		KindField:    "K",
	}

	type Inner struct {
		G string
		V string
		K string
	}
	obj := struct {
		Test Inner
	}{Test: Inner{"0", "1", "2"}}

	// updates
	if err := factory.Update(TypeMeta{"first", "test", "other"}, &obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Test.G != "first" || obj.Test.V != "test" || obj.Test.K != "other" {
		t.Errorf("unexpected update: %v", obj)
	}
}

func TestMetaValues(t *testing.T) {
	type InternalSimple struct {
		APIGroup   string `json:"apiGroup,omitempty"`
		APIVersion string `json:"apiVersion,omitempty"`
		Kind       string `json:"kind,omitempty"`
		TestString string `json:"testString"`
	}
	type ExternalSimple struct {
		APIGroup   string `json:"apiGroup,omitempty"`
		APIVersion string `json:"apiVersion,omitempty"`
		Kind       string `json:"kind,omitempty"`
		TestString string `json:"testString"`
	}
	s := NewScheme("")
	s.AddKnownTypeWithName("myapi", "", "Simple", &InternalSimple{})
	s.AddKnownTypeWithName("myapi", "externalVersion", "Simple", &ExternalSimple{})

	internalToExternalCalls := 0
	externalToInternalCalls := 0

	// Register functions to verify that scope.Meta() gets set correctly.
	err := s.AddConversionFuncs(
		func(in *InternalSimple, out *ExternalSimple, scope Scope) error {
			t.Logf("internal -> external")
			if e, a := "", scope.Meta().SrcVersion; e != a {
				t.Fatalf("Expected '%v', got '%v'", e, a)
			}
			if e, a := "externalVersion", scope.Meta().DestVersion; e != a {
				t.Fatalf("Expected '%v', got '%v'", e, a)
			}
			scope.Convert(&in.TestString, &out.TestString, 0)
			internalToExternalCalls++
			return nil
		},
		func(in *ExternalSimple, out *InternalSimple, scope Scope) error {
			t.Logf("external -> internal")
			if e, a := "externalVersion", scope.Meta().SrcVersion; e != a {
				t.Errorf("Expected '%v', got '%v'", e, a)
			}
			if e, a := "", scope.Meta().DestVersion; e != a {
				t.Fatalf("Expected '%v', got '%v'", e, a)
			}
			scope.Convert(&in.TestString, &out.TestString, 0)
			externalToInternalCalls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	simple := &InternalSimple{
		TestString: "foo",
	}

	s.Log(t)

	// Test Encode, Decode, and DecodeInto
	data, err := s.EncodeToVersion(simple, "externalVersion")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf(string(data))
	obj2, err := s.Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := obj2.(*InternalSimple); !ok {
		t.Fatalf("Got wrong type")
	}
	if e, a := simple, obj2; !reflect.DeepEqual(e, a) {
		t.Errorf("Expected:\n %#v,\n Got:\n %#v", e, a)
	}

	obj3 := &InternalSimple{}
	if err := s.DecodeInto(data, obj3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e, a := simple, obj3; !reflect.DeepEqual(e, a) {
		t.Errorf("Expected:\n %#v,\n Got:\n %#v", e, a)
	}

	// Test Convert
	external := &ExternalSimple{}
	err = s.Convert(simple, external)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if e, a := simple.TestString, external.TestString; e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}

	// Encode and Convert should each have caused an increment.
	if e, a := 2, internalToExternalCalls; e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
	// Decode and DecodeInto should each have caused an increment.
	if e, a := 2, externalToInternalCalls; e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
}

func TestMetaValuesUnregisteredConvert(t *testing.T) {
	type InternalSimple struct {
		Group      string `json:"apiGroup,omitempty"`
		Version    string `json:"apiVersion,omitempty"`
		Kind       string `json:"kind,omitempty"`
		TestString string `json:"testString"`
	}
	type ExternalSimple struct {
		Group      string `json:"apiGroup,omitempty"`
		Version    string `json:"apiVersion,omitempty"`
		Kind       string `json:"kind,omitempty"`
		TestString string `json:"testString"`
	}
	s := NewScheme("")
	s.InternalVersion = ""
	// We deliberately don't register the types.

	internalToExternalCalls := 0

	// Register functions to verify that scope.Meta() gets set correctly.
	err := s.AddConversionFuncs(
		func(in *InternalSimple, out *ExternalSimple, scope Scope) error {
			if e, a := "unknown", scope.Meta().SrcVersion; e != a {
				t.Fatalf("Expected '%v', got '%v'", e, a)
			}
			if e, a := "unknown", scope.Meta().DestVersion; e != a {
				t.Fatalf("Expected '%v', got '%v'", e, a)
			}
			scope.Convert(&in.TestString, &out.TestString, 0)
			internalToExternalCalls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	simple := &InternalSimple{TestString: "foo"}
	external := &ExternalSimple{}
	err = s.Convert(simple, external)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if e, a := simple.TestString, external.TestString; e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}

	// Verify that our conversion handler got called.
	if e, a := 1, internalToExternalCalls; e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
}

func TestInvalidPtrValueKind(t *testing.T) {
	var simple interface{}
	switch obj := simple.(type) {
	default:
		_, err := EnforcePtr(obj)
		if err == nil {
			t.Errorf("Expected error on invalid kind")
		}
	}
}

func TestEnforceNilPtr(t *testing.T) {
	var nilPtr *struct{}
	_, err := EnforcePtr(nilPtr)
	if err == nil {
		t.Errorf("Expected error on nil pointer")
	}
}
