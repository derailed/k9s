package config

// Plugin describes a K9s plugin
type Plugin struct {
	ShortCut string            `yaml:"shortCut"`
	Scopes   []string          `yaml:"scopes"`
	Args     map[string]string `yaml:"args"`
}
