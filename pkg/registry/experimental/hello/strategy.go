package hello

import (
	"errors"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/experimental"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/rest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/generic"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util/fielderrors"
)

var (
	NotHelloErr = errors.New("not a hello")
	Strategy    = strategy{experimental.Scheme, api.SimpleNameGenerator}
)

func Matcher(label labels.Selector, field fields.Selector) generic.Matcher {
	return generic.MatcherFunc(func(obj runtime.Object) (bool, error) {
		if h, ok := obj.(*experimental.Hello); ok {
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
func (strategy) AllowCreateOnUpdate() bool            { return false }

func (strategy) Validate(ctx api.Context, obj runtime.Object) fielderrors.ValidationErrorList {
	return nil
}

func (strategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) fielderrors.ValidationErrorList {
	return nil
}
