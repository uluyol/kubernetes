package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
)

// Scheme is the default instance of runtime.Scheme to which types in the experimental API are already registered.
var Scheme = runtime.NewScheme()

func init() {
	Scheme.AddKnownTypes("",
		&Hello{},
	)
}

func (*Hello) IsAnAPIObject() {}
