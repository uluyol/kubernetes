package experimental

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
)

type Hello struct {
	api.TypeMeta   `json:",inline"`
	api.ObjectMeta `json:"metadata,omitempty"`

	Text string `json"text,omitempty"`
}
