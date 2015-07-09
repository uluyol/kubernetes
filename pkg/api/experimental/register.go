package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/experimental/latest"
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
	)

	api.Scheme.AddKnownTypes(Group, Version,
		&Hello{},
		&HelloList{},

		&DeleteOptions{},
		&Namespace{},
		&ListOptions{},
		&Status{},
	)

	initExperimental()
}

func (*Hello) IsAnAPIObject()     {}
func (*HelloList) IsAnAPIObject() {}

func (*DeleteOptions) IsAnAPIObject() {}
func (*Namespace) IsAnAPIObject()     {}
func (*ListOptions) IsAnAPIObject()   {}
func (*Status) IsAnAPIObject()        {}
