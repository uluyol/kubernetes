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

package api

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/meta"
)

var restMapperMap = map[string]meta.RESTMapper{}

func RegisterRESTMapper(group string, mapper meta.RESTMapper) {
	restMapperMap[group] = mapper
}

func GetRESTMapper(group string) (meta.RESTMapper, error) {
	if mapper, ok := restMapperMap[group]; ok {
		return mapper, nil
	}
	return nil, notRegisteredErr{group}
}

type notRegisteredErr struct {
	group string
}

func (e notRegisteredErr) Error() string {
	return fmt.Sprintf("%q does not have a registered RESTMapper", e.group)
}

func IsNotRegisteredErr(e error) bool {
	_, ok := e.(notRegisteredErr)
	return ok
}
