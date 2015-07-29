/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

/*
This file (together with pkg/expapi/v1/types.go) contain the experimental
types in kubernetes. These API objects are experimental, meaning that the
APIs may be broken at any time by the kubernetes team.

DISCLAIMER: The implementation of the experimental API group itself is
a temporary one meant as a stopgap solution until kubernetes has proper
support for multiple API groups. The transition may require changes
beyond registration differences. In other words, experimental API group
support is experimental.
*/

package expapi

import "github.com/GoogleCloudPlatform/kubernetes/pkg/api"

type Hello struct {
	api.TypeMeta   `json:",inline"`
	api.ObjectMeta `json:"metadata,omitempty"`

	Text     string               `json:"text,omitempty"`
	Text2    string               `json:"test,omitempty"`
	Template *api.PodTemplateSpec `json:"template,omitempty"`
}

type HelloList struct {
	api.TypeMeta `json:",inline"`
	api.ListMeta `json:"metadata,omitempty"`

	Items []Hello `json:"items,omitempty"`
}
