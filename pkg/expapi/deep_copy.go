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

package expapi

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/conversion"
)

func deepCopy_expapi_Hello(in Hello, out *Hello, c *conversion.Cloner) error {
	out.TypeMeta = in.TypeMeta
	m, err := api.Scheme.DeepCopy(&in.ObjectMeta)
	if err != nil {
		return err
	}
	out.ObjectMeta = *m.(*api.ObjectMeta)
	out.Text = in.Text
	out.Text2 = in.Text2
	t, err := api.Scheme.DeepCopy(in.Template)
	if err != nil {
		return err
	}
	out.Template = t.(*api.PodTemplateSpec)
	return nil
}

func deepCopy_expapi_HelloList(in HelloList, out *HelloList, c *conversion.Cloner) error {
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	out.Items = make([]Hello, len(in.Items))
	for i := range in.Items {
		if err := deepCopy_expapi_Hello(in.Items[i], &out.Items[i], c); err != nil {
			return err
		}
	}
	return nil
}

func addDeepCopyFuncs() {
	err := api.Scheme.AddDeepCopyFuncs(
		deepCopy_expapi_Hello,
		deepCopy_expapi_HelloList,
	)
	if err != nil {
		panic(err)
	}
}
