package etcd

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/experimental"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	hello "github.com/GoogleCloudPlatform/kubernetes/pkg/registry/experimental-hello"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/generic"
	etcdgeneric "github.com/GoogleCloudPlatform/kubernetes/pkg/registry/generic/etcd"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
)

// REST implements a RESTStorage for hellos against etcd
type REST struct {
	*etcdgeneric.Etcd
}

func NewStorage(h tools.EtcdHelper) *REST {
	prefix := "/k8s.io/experimental/hello"

	store := &etcdgeneric.Etcd{
		NewFunc:     func() runtime.Object { return &experimental.Hello{} },
		NewListFunc: func() runtime.Object { return &experimental.HelloList{} },
		KeyRootFunc: func(ctx api.Context) string {
			return etcdgeneric.NamespaceKeyRootFunc(ctx, prefix)
		},
		KeyFunc: func(ctx api.Context, id string) (string, error) {
			return etcdgeneric.NamespaceKeyFunc(ctx, prefix, id)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*experimental.Hello).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return hello.Matcher(label, field)
		},
		EndpointName: "hello",

		Helper: h,

		CreateStrategy: hello.Strategy,
		UpdateStrategy: hello.Strategy,
	}

	return &REST{store}
}
