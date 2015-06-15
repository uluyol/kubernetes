package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
)

// Version is the string that represents the current external default version.
var Version = "v0"

// Codec is the default codec for serializing output that should use
// the latest supported version.  Use this Codec when writing to
// disk, a data store that is not dynamically versioned, or in tests.
// This codec can decode any object that Kubernetes is aware of.
var Codec = runtime.CodecFor(Scheme, Version)

// Scheme is the default instance of runtime.Scheme to which types in the experimental API are already registered.
var Scheme = runtime.NewScheme()

func init() {
	Scheme.AddKnownTypes("",
		&Hello{},
		&HelloList{},

		&api.Namespace{},
		&api.ListOptions{},
		&api.DeleteOptions{},
		&api.Status{},
	)

	Scheme.AddKnownTypes(Version,
		&Hello{},
		&HelloList{},

		&api.Namespace{},
		&api.ListOptions{},
		&api.DeleteOptions{},
		&api.Status{},
	)

	addDefaultingFuncs()
	addConversionFuncs()
	initExperimental()
}

func (*Hello) IsAnAPIObject()     {}
func (*HelloList) IsAnAPIObject() {}
