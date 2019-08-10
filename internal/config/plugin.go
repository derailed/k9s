package config

// Plugin describes a K9s plugin
type Plugin struct {
	ShortCut    string   `yaml:"shortCut"`
	Scopes      []string `yaml:"scopes"`
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Background  bool     `yaml:"background"`
	Args        []string `yaml:"args"`
}
