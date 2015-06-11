package experimental

func addDefaultingFuncs() {
	Scheme.AddDefaultingFuncs(
		func(obj *Hello) {
			if obj.Text == "" {
				obj.Text = "should enter some text here. BOGO"
			}
		},
	)
}
