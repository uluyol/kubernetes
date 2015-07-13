package v0

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/experimental"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/meta"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
)

// accessor is the shared static metadata accessor for the API.
var accessor = meta.NewAccessor()

// SelfLinker can set or get the SelfLink field of all API types.
// TODO: when versioning changes, make this part of each API definition.
// TODO(lavalamp): Combine SelfLinker & ResourceVersioner interfaces, force all uses
// to go through the InterfacesFor method below.
var SelfLinker = runtime.SelfLinker(accessor)

// RESTMapper provides the default mapping between REST paths and the objects declared in Scheme and all known
// Kubernetes versions.
var RESTMapper meta.RESTMapper

func initRESTMapper() {
	versions := []string{experimental.Version}

	mapper := meta.NewDefaultRESTMapper(
		versions,
		func(version string) (*meta.VersionInterfaces, bool) {
			if version == experimental.Version {
				return &meta.VersionInterfaces{
					Codec:            Codec,
					ObjectConvertor:  api.Scheme,
					MetadataAccessor: accessor,
				}, true
			}
			return nil, false
		},
	)

	// the list of kinds that are scoped at the root of the api hierarchy
	// if a kind is not enumerated here, it is assumed to have a namespace scope
	kindToRootScope := map[string]bool{}

	// these kinds should be excluded from the list of resources
	ignoredKinds := util.NewStringSet()

	// enumerate all supported versions, get the kinds, and register with the mapper how to address our resources.
	fmt.Println("HERE")
	for kind := range api.Scheme.KnownTypes(experimental.Group, experimental.Version) {
		fmt.Println("GOT kind", kind)
		if ignoredKinds.Has(kind) {
			continue
		}
		scope := meta.RESTScopeNamespace
		if kindToRootScope[kind] {
			scope = meta.RESTScopeRoot
		}
		fmt.Println("STIL GOT kind", kind)
		mapper.Add(scope, kind, experimental.Version, false)
	}
	RESTMapper = mapper
	fmt.Println("RESTMAPPER;", RESTMapper)
	api.RegisterRESTMapper(experimental.Group, RESTMapper)
}
