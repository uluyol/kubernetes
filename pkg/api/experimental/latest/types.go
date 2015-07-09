package latest

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
)

type Hello struct {
	runtime.TypeMeta `json:",inline"`
	api.ObjectMeta    `json:"metadata,omitempty"`

	Text     string              `json:"text,omitempty"`
	Text2    string              `json:"test,omitempty"`
	Template *api.PodTemplateSpec `json:"template,omitempty"`
}

type HelloList struct {
	runtime.TypeMeta `json:",inline"`
	api.ObjectMeta    `json:"metadata,omitempty"`

	Items []Hello `json:"items,omitempty"`
}

func (*Hello) IsAnAPIObject()     {}
func (*HelloList) IsAnAPIObject() {}
