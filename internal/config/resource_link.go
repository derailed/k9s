package config

type CustomResourceLinks map[string]*CustomResourceLink

func NewCustomResourceLinks() CustomResourceLinks {
	return CustomResourceLinks{}
}

// CustomResourceLink tracks K9s CustomResourceLink configuration.
type CustomResourceLink struct {
	// Target represents the target GVR to open when activating a custom resource link.
	Target string `yaml:"target"`
	// LabelSelector defines keys (=target label) and values (=json path) to extract from the current resource.
	LabelSelector map[string]string `yaml:"labelSelector,omitempty"`
	// FieldSelector defines keys (=target field) and values (=json path) to extract from the current resource.
	FieldSelector map[string]string `yaml:"fieldSelector,omitempty"`
}
