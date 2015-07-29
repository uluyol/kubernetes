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

package hello

import (
	"errors"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/rest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/expapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/generic"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util/fielderrors"
)

var (
	NotHelloErr = errors.New("not a hello")
	Strategy    = strategy{api.Scheme, api.SimpleNameGenerator}
)

func Matcher(label labels.Selector, field fields.Selector) generic.Matcher {
	return generic.MatcherFunc(func(obj runtime.Object) (bool, error) {
		if h, ok := obj.(*expapi.Hello); ok {
			fields := labels.Set{"text": h.Text}
			return label.Matches(labels.Set(h.Labels)) && field.Matches(fields), nil
		}
		return false, NotHelloErr
	})
}

var (
	// Ensure that Strategy implements these interfaces.
	_ = rest.RESTCreateStrategy(Strategy)
	_ = rest.RESTUpdateStrategy(Strategy)
)

type strategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

func (strategy) NamespaceScoped() bool                { return true }
func (strategy) PrepareForCreate(_ runtime.Object)    {}
func (strategy) PrepareForUpdate(_, _ runtime.Object) {}
func (strategy) AllowCreateOnUpdate() bool            { return true }
func (strategy) AllowUnconditionalUpdate() bool       { return true }

func (strategy) Validate(ctx api.Context, obj runtime.Object) fielderrors.ValidationErrorList {
	return nil
}

func (strategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) fielderrors.ValidationErrorList {
	return nil
}
