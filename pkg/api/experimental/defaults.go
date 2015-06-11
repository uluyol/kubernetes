package experimental

import (
	"strings"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/golang/glog"
)

func addDefaultingFuncs() {
	Scheme.AddDefaultingFuncs(
		func(obj *Hello) {
			if obj.Text == "" {
				obj.Text = "should enter some text here. BOGO"
			}
		},
	)
}
