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

package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/resource"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/conversion"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
)

// AUTO-GENERATED FUNCTIONS START HERE
func init() {
	err := Scheme.AddConversionFuncs(
		func(in *util.Time, out *util.Time, s conversion.Scope) error {
			// Cannot deep copy these, because time.Time has unexported fields.
			*out = *in
			return nil
		},
		func(in *string, out *labels.Selector, s conversion.Scope) error {
			selector, err := labels.Parse(*in)
			if err != nil {
				return err
			}
			*out = selector
			return nil
		},
		func(in *string, out *fields.Selector, s conversion.Scope) error {
			selector, err := fields.ParseSelector(*in)
			if err != nil {
				return err
			}
			*out = selector
			return nil
		},
		func(in *labels.Selector, out *string, s conversion.Scope) error {
			if *in == nil {
				return nil
			}
			*out = (*in).String()
			return nil
		},
		func(in *fields.Selector, out *string, s conversion.Scope) error {
			if *in == nil {
				return nil
			}
			*out = (*in).String()
			return nil
		},
		func(in *resource.Quantity, out *resource.Quantity, s conversion.Scope) error {
			// Cannot deep copy these, because inf.Dec has unexported fields.
			*out = *in.Copy()
			return nil
		},
	)
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
}

// AUTO-GENERATED FUNCTIONS END HERE
