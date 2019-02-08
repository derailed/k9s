package config

const defaultView = "po"

// View tracks view configuration options.
type View struct {
	Active string `yaml:"active"`
}

func NewView() *View {
	return &View{Active: defaultView}
}

func (v *View) Validate(ClusterInfo) {
	if len(v.Active) == 0 {
		v.Active = defaultView
	}
}
