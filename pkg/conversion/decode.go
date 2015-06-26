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
	"errors"
	"fmt"
)

// Decode converts a JSON string back into a pointer to an api object.
// Deduces the type based upon the fields added by the MetaInsertionFactory
// technique. The object will be converted, if necessary, into the
// s.InternalVersion type before being returned. Decode will not decode
// objects without version set unless InternalVersion is also "".
func (s *Scheme) Decode(data []byte) (interface{}, error) {
	tm, err := s.DataTypeMeta(data)
	if err != nil {
		return nil, err
	}
	if tm.APIVersion == "" && s.InternalVersion != "" {
		return nil, fmt.Errorf("version not set in '%s'", string(data))
	}
	if tm.Kind == "" {
		return nil, fmt.Errorf("kind not set in '%s'", string(data))
	}
	if tm.APIGroup == "" {
		return nil, fmt.Errorf("group not set in '%s'", string(data))
	}
	obj, err := s.NewObject(tm)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return nil, err
	}

	// Version and Kind should be blank in memory.
	if err := s.SetTypeMeta(TypeMeta{}, obj); err != nil {
		return nil, err
	}

	// Convert if needed.
	if s.InternalVersion != tm.APIVersion {
		objOut, err := s.NewObject(TypeMeta{tm.APIGroup, s.InternalVersion, tm.Kind})
		if err != nil {
			return nil, err
		}
		flags, meta := s.generateConvertMeta(tm.APIVersion, s.InternalVersion, obj)
		if err := s.converter.Convert(obj, objOut, flags, meta); err != nil {
			return nil, err
		}
		obj = objOut
	}
	return obj, nil
}

// DecodeInto parses a JSON string and stores it in obj. Returns an error
// if data.Kind is set and doesn't match the type of obj. Obj should be a
// pointer to an api type.
// If obj's version doesn't match that in data, an attempt will be made to convert
// data into obj's version.
func (s *Scheme) DecodeInto(data []byte, obj interface{}) error {
	if len(data) == 0 {
		return errors.New("empty input")
	}
	dataTM, err := s.DataTypeMeta(data)
	if err != nil {
		return err
	}
	objTM, err := s.ObjectTypeMeta(obj)
	if err != nil {
		return err
	}
	if dataTM.Kind == "" {
		// Assume objects with unset Kind fields are being unmarshalled into the
		// correct type.
		dataTM.Kind = objTM.Kind
	}
	if dataTM.APIVersion == "" {
		// Assume objects with unset Version fields are being unmarshalled into the
		// correct type.
		dataTM.APIVersion = objTM.APIVersion
	}
	if dataTM.APIGroup == "" {
		// Assume objects with unset Group fields are being unmarshalled into the
		// correct type.
		dataTM.APIGroup = objTM.APIGroup
	}

	external, err := s.NewObject(dataTM)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, external); err != nil {
		return err
	}
	flags, meta := s.generateConvertMeta(dataTM.APIVersion, objTM.APIVersion, external)
	if err := s.converter.Convert(external, obj, flags, meta); err != nil {
		return err
	}

	// Version and Kind should be blank in memory.
	return s.SetTypeMeta(TypeMeta{}, obj)
}
