package experimental

import "github.com/GoogleCloudPlatform/kubernetes/pkg/api"

const (
	Group   = "experimental"
	Version = "v0"
)

func init() {
	api.Scheme.AddKnownTypes(Group, "",
		&Hello{},
		&HelloList{},
	)
}

func (*Hello) IsAnAPIObject()     {}
func (*HelloList) IsAnAPIObject() {}
