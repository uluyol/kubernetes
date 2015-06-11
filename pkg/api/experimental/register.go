package experimental

import (
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
	Scheme.AddKnownTypes("v0",
		&Hello{},
		&HelloList{},
	)

	addDefaultingFuncs()
}

func (*Hello) IsAnAPIObject()     {}
func (*HelloList) IsAnAPIObject() {}
