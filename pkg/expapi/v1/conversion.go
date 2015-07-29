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

package v1

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/conversion"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/expapi"
)

func addConversionFuncs() {
	err := api.Scheme.AddConversionFuncs(
		convert_expapi_Hello_To_v1_Hello,
		convert_v1_Hello_To_expapi_Hello,
		convert_expapi_HelloList_To_v1_HelloList,
		convert_v1_HelloList_To_expapi_HelloList,
	)
	if err != nil {
		panic(err)
	}
}

func convert_expapi_Hello_To_v1_Hello(in *expapi.Hello, out *Hello, s conversion.Scope) error {
	var outTemplate v1.PodTemplateSpec
	conversionPtrs := []struct {
		in  interface{}
		out interface{}
	}{
		{&in.TypeMeta, &out.TypeMeta}, {&in.ObjectMeta, &out.ObjectMeta}, {in.Template, &outTemplate},
	}
	for _, pair := range conversionPtrs {
		if err := api.Scheme.Convert(pair.in, pair.out); err != nil {
			return err
		}
	}
	out.Template = &outTemplate
	out.Text = in.Text
	out.Text2 = in.Text2
	return nil
}

func convert_v1_Hello_To_expapi_Hello(in *Hello, out *expapi.Hello, s conversion.Scope) error {
	var outTemplate api.PodTemplateSpec
	conversionPtrs := []struct {
		in  interface{}
		out interface{}
	}{
		{&in.TypeMeta, &out.TypeMeta}, {&in.ObjectMeta, &out.ObjectMeta}, {in.Template, &outTemplate},
	}
	for _, pair := range conversionPtrs {
		if err := api.Scheme.Convert(pair.in, pair.out); err != nil {
			return err
		}
	}
	out.Template = &outTemplate
	out.Text = in.Text
	out.Text2 = in.Text2
	return nil
}

func convert_expapi_HelloList_To_v1_HelloList(in *expapi.HelloList, out *HelloList, s conversion.Scope) error {
	if err := api.Scheme.Convert(&in.TypeMeta, &out.TypeMeta); err != nil {
		return err
	}
	if err := api.Scheme.Convert(&in.ListMeta, &out.ListMeta); err != nil {
		return err
	}
	out.Items = make([]Hello, len(in.Items))
	for i := range in.Items {
		if err := convert_expapi_Hello_To_v1_Hello(&in.Items[i], &out.Items[i], s); err != nil {
			return err
		}
	}
	return nil
}

func convert_v1_HelloList_To_expapi_HelloList(in *HelloList, out *expapi.HelloList, s conversion.Scope) error {
	if err := api.Scheme.Convert(&in.TypeMeta, &out.TypeMeta); err != nil {
		return err
	}
	if err := api.Scheme.Convert(&in.ListMeta, &out.ListMeta); err != nil {
		return err
	}
	out.Items = make([]expapi.Hello, len(in.Items))
	for i := range in.Items {
		if err := convert_v1_Hello_To_expapi_Hello(&in.Items[i], &out.Items[i], s); err != nil {
			return err
		}
	}
	return nil
}
