package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
)

type Hello struct {
	runtime.TypeMeta `json:",inline"`
	v1.ObjectMeta    `json:"metadata,omitempty"`

	Text     string              `json:"text,omitempty"`
	Text2    string              `json:"test,omitempty"`
	Template *v1.PodTemplateSpec `json:"template,omitempty"`
}

type HelloList struct {
	runtime.TypeMeta `json:",inline"`
	v1.ObjectMeta    `json:"metadata,omitempty"`

	Items []Hello `json:"items,omitempty"`
}

type DeleteOptions struct{ v1.DeleteOptions }
type Namespace struct{ v1.Namespace }
type ListOptions struct{ v1.ListOptions }
type Status struct{ v1.Status }
