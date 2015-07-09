package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/experimental/latest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
)

const (
	Group   = "experimental"
	Version = "v0"
)

// Codec is the default codec for serializing output that should use
// the latest supported version.  Use this Codec when writing to
// disk, a data store that is not dynamically versioned, or in tests.
// This codec can decode any object that Kubernetes is aware of.
var Codec = runtime.CodecFor(api.Scheme, Version)

func init() {
	api.Scheme.AddKnownTypes(Group, "",
		&latest.Hello{},
		&latest.HelloList{},

		&api.DeleteOptions{},
		&api.Namespace{},
		&api.ListOptions{},
		&api.Status{},
	)

	api.Scheme.AddKnownTypes(Group, Version,
		&Hello{},
		&HelloList{},

		&v1.DeleteOptions{},
		&v1.Namespace{},
		&v1.ListOptions{},
		&v1.Status{},
	)

	initExperimental()
}

func (*Hello) IsAnAPIObject()     {}
func (*HelloList) IsAnAPIObject() {}
