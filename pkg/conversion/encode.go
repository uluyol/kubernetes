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
)

// EncodeToVersion turns the given api object into an appropriate JSON string.
// Obj may be a pointer to a struct, or a struct. If a struct, a copy
// will be made, therefore it's recommended to pass a pointer to a
// struct. The type must have been registered.
//
// Memory/wire format differences:
//  * Having to keep track of the Kind and Version fields makes tests
//    very annoying, so the rule is that they are set only in wire format
//    (json), not when in native (memory) format. This is possible because
//    both pieces of information are implicit in the go typed object.
//     * An exception: note that, if there are embedded API objects of known
//       type, for example, PodList{... Items []Pod ...}, these embedded
//       objects must be of the same version of the object they are embedded
//       within, and their Version and Kind must both be empty.
//     * Note that the exception does not apply to a generic APIObject type
//       which recursively does Encode()/Decode(), and is capable of
//       expressing any API object.
//  * Only versioned objects should be encoded. This means that, if you pass
//    a native object, Encode will convert it to a versioned object. For
//    example, an api.Pod will get converted to a v1beta1.Pod. However, if
//    you pass in an object that's already versioned (v1beta1.Pod), Encode
//    will not modify it.
//
// The purpose of the above complex conversion behavior is to allow us to
// change the memory format yet not break compatibility with any stored
// objects, whether they be in our storage layer (e.g., etcd), or in user's
// config files.
//
func (s *Scheme) EncodeToVersion(obj interface{}, destVersion string) (data []byte, err error) {
	obj = maybeCopy(obj)
	v, _ := EnforcePtr(obj) // maybeCopy guarantees a pointer
	if _, registered := s.typeToVersion[v.Type()]; !registered {
		return nil, fmt.Errorf("type %v is not registered for %q and it will be impossible to Decode it, therefore Encode will refuse to encode it.", v.Type(), destVersion)
	}

	group, version, kind, err := s.ObjectTypeMeta(obj)
	if err != nil {
		return nil, err
	}

	// Perform a conversion if necessary.
	if version != destVersion {
		objOut, err := s.NewObject(group, destVersion, kind)
		if err != nil {
			return nil, fmt.Errorf("%v @group: %s version: %s, kind %s@", err, group, version, kind)
		}
		flags, meta := s.generateConvertMeta(version, destVersion, obj)
		err = s.converter.Convert(obj, objOut, flags, meta)
		if err != nil {
			return nil, err
		}
		obj = objOut
	}

	// ensure the output object name comes from the destination type
	group, version, kind, err = s.ObjectTypeMeta(obj)
	if err != nil {
		return nil, err
	}

	// Version and Kind should be set on the wire.
	err = s.SetTypeMeta(group, destVersion, kind, obj)
	if err != nil {
		return nil, err
	}

	// To add metadata, do some simple surgery on the JSON.
	data, err = json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// Version and Kind should be blank in memory. Reset them, since it's
	// possible that we modified a user object and not a copy above.
	err = s.SetTypeMeta("", "", "", obj)
	if err != nil {
		return nil, err
	}

	return data, nil
}
