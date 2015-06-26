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
	"encoding/json"
	"fmt"
	"reflect"
)

// MetaFactory is used to store and retrieve the version and kind
// information for all objects in a scheme.
type MetaFactory interface {
	// Update sets the given version and kind onto the object.
	Update(meta TypeMeta, obj interface{}) error
	// Interpret should return the version and kind of the wire-format of
	// the object.
	Interpret(data []byte) (TypeMeta, error)
}

// DefaultMetaFactory is a default factory for versioning objects in JSON. The object
// in memory and in the default JSON serialization will use the "kind" and "apiVersion"
// fields.
var DefaultMetaFactory = SimpleMetaFactory{GroupField: "APIGroup", VersionField: "APIVersion", KindField: "Kind"}

// SimpleMetaFactory provides default methods for retrieving the type and version of objects
// that are identified with an "apiVersion" and "kind" fields in their JSON
// serialization. It may be parameterized with the names of the fields in memory, or an
// optional list of base structs to search for those fields in memory.
type SimpleMetaFactory struct {
	// The names of the API group, version, and kind fields in memory of the struct
	GroupField   string
	VersionField string
	KindField    string
	// Optional, if set will look in the named inline structs to find the fields to set.
	BaseFields []string
}

// Interpret will return the APIVersion and Kind of the JSON wire-format
// encoding of an object, or an error.
func (SimpleMetaFactory) Interpret(data []byte) (TypeMeta, error) {
	findMeta := TypeMeta{}
	err := json.Unmarshal(data, &findMeta)
	if err != nil {
		return TypeMeta{}, fmt.Errorf("couldn't get version/kind; json parse error: %v", err)
	}
	return findMeta, nil
}

func (f SimpleMetaFactory) Update(tm TypeMeta, obj interface{}) error {
	return UpdateTypeMeta(f.BaseFields,
		f.GroupField, tm.APIGroup,
		f.VersionField, tm.APIVersion,
		f.KindField, tm.Kind,
		obj)
}

// UpdateVersionAndKind uses reflection to find and set the versionField and kindField fields
// on a pointer to a struct to version and kind. Provided as a convenience for others
// implementing MetaFactory. Pass an array to baseFields to check one or more nested structs
// for the named fields. The version field is treated as optional if it is not present in the struct.
func UpdateTypeMeta(baseFields []string, groupField, group, versionField, version, kindField, kind string, obj interface{}) error {
	v, err := EnforcePtr(obj)
	if err != nil {
		return err
	}
	t := v.Type()
	name := t.Name()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, but got %v: %v (%#v)", v.Kind(), name, v.Interface())
	}

	for i := range baseFields {
		base := v.FieldByName(baseFields[i])
		if !base.IsValid() {
			continue
		}
		v = base
	}

	field := v.FieldByName(kindField)
	if !field.IsValid() {
		return fmt.Errorf("couldn't find %v field in %#v", kindField, v.Interface())
	}
	field.SetString(kind)

	if field := v.FieldByName(versionField); field.IsValid() {
		field.SetString(version)
	}

	if field := v.FieldByName(groupField); field.IsValid() {
		field.SetString(group)
	}

	return nil
}

// EnforcePtr ensures that obj is a pointer of some sort. Returns a reflect.Value
// of the dereferenced pointer, ensuring that it is settable/addressable.
// Returns an error if this is not possible.
func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}
