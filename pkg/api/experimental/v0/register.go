package v0

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/experimental"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
)

var Codec = runtime.CodecFor(api.Scheme, experimental.Version)

func init() {
	addKnownTypes()
	addDefaultingFuncs()
	initRESTMapper()
}

func addKnownTypes() {
	api.Scheme.AddKnownTypes(experimental.Group, experimental.Version,
		&Hello{},
		&HelloList{},
		&DeleteOptions{},
		&Namespace{},
		&ListOptions{},
		&Status{},
	)
}

func addDefaultingFuncs() {}

func (*Hello) IsAnAPIObject()         {}
func (*HelloList) IsAnAPIObject()     {}
func (*DeleteOptions) IsAnAPIObject() {}
func (*Namespace) IsAnAPIObject()     {}
func (*ListOptions) IsAnAPIObject()   {}
func (*Status) IsAnAPIObject()        {}
