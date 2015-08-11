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

package bench

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	_ "k8s.io/kubernetes/pkg/api/v1"
	_ "k8s.io/kubernetes/pkg/expapi"
	_ "k8s.io/kubernetes/pkg/expapi/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

type typeVersionLists [][]reflect.Type

func (l typeVersionLists) NewOVL() []objectVersionList {
	var allObjectVersions []objectVersionList
	for _, tv := range l {
		var ov []runtime.Object
		for _, t := range tv {
			ov = append(ov, reflect.New(t).Interface().(runtime.Object))
		}
		allObjectVersions = append(allObjectVersions, objectVersionList(ov))
	}
	return allObjectVersions
}

/*
func initObj(o runtime.Object, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex(i)
		ftype := f.Type()
		switch ftype.Kind() {
		case reflect.Ptr:
			val := reflect.New(ftype.Elem()
			o.Field(i).Set(v)
			if subObj, ok := val.Interface().(runtime.Object); ok {
				initObj(subObj, ftype.Elem())
			}
		case reflect.Struct:
			val := o.Field(i)
			if subObj, ok := val.Interface().(runtime.Object); ok {
				initObj(subObj, ftype.Elem())
			}
		case reflect.Map
*/

type objectVersionList []runtime.Object

func getTypeVersionLists(b *testing.B) typeVersionLists {
	versions := []string{"", "v1"}
	versionMaps := make(map[string]map[string]reflect.Type)
	var versionMapKeyCounts []int
	for _, v := range versions {
		versionMaps[v] = api.Scheme.KnownTypes(v)
		c := 0
		for range versionMaps[v] {
			c++
		}
		versionMapKeyCounts = append(versionMapKeyCounts, c)
	}
	for i, c := range versionMapKeyCounts {
		if versionMapKeyCounts[0] != c {
			b.Fatalf("versions %s and %s have different types", versions[0], versions[i])
		}
	}
	var allTypeVersions [][]reflect.Type
	for k := range versionMaps[versions[0]] {
		var tv []reflect.Type
		for _, v := range versions {
			t, ok := versionMaps[v][k]
			if !ok {
				b.Fatalf("version %s does not contain kind %v", v, k)
			}
			tv = append(tv, t)
		}
		allTypeVersions = append(allTypeVersions, tv)
	}
	return typeVersionLists(allTypeVersions)
}

func getKObjectVersionLists(b *testing.B, k int) [][]objectVersionList {
	tvl := getTypeVersionLists(b)
	ovls := make([][]objectVersionList, k)
	for i := 0; i < k; i++ {
		ovls[i] = tvl.NewOVL()
	}
	return ovls
}

// BenchmarkRegisteredConversions measures conversion time from the first
// version to the second (we assume the first is the internal version). This
// may not run all conversions, but it should be comprehensive enough to
// give us enough to compare generated code.
func BenchmarkRegisteredConversions(b *testing.B) {
	ovls := getKObjectVersionLists(b, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ov := range ovls[i] {
			err := api.Scheme.Convert(ov[0], ov[1])
			if err != nil {
				b.Errorf("Unable to convert %T to %T: %v", ov[0], ov[1], err)
			}
		}
	}
}

// Put results here so they're not optimized out
var deepCopyResult interface{}

// BenchmarkRegisteredDeepCopies measures the time to deep copy all objects.
func BenchmarkRegisteredDeepCopies(b *testing.B) {
	ovls := getKObjectVersionLists(b, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ov := range ovls[i] {
			for _, o := range ov {
				//fmt.Printf("type %T\n", o)
				//fmt.Printf("val %+v\n", o)
				r, err := api.Scheme.DeepCopy(o)
				deepCopyResult = r
				if err != nil {
					b.Errorf("Unable to deep copy %T: %v", o, err)
				}
			}
		}
	}
}
