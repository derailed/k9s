package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"gopkg.in/yaml.v2"
)

var (
	// K9sStylesFile represents K9s skins file location.
	K9sStylesFile = filepath.Join(K9sHome, "skin.yml")
)

type (
	// Styles tracks K9s styling options.
	Styles struct {
		Style *Style `yaml:"k9s"`
	}

	// Style tracks K9s styles.
	Style struct {
		FgColor   string `yaml:"fgColor"`
		BgColor   string `yaml:"bgColor"`
		LogoColor string `yaml:"logoColor"`

		Info   *Info   `yaml:"info"`
		Border *Border `yaml:"border"`
		Menu   *Menu   `yaml:"menu"`
		Crumb  *Crumb  `yaml:"crumb"`
		Table  *Table  `yaml:"table"`
		Status *Status `yaml:"status"`
		Title  *Title  `yaml:"title"`
		Yaml   *Yaml   `yaml:"yaml"`
		Log    *Log    `yaml:"logs"`
	}

	// Status tracks resource status styles.
	Status struct {
		NewColor       string `yaml:"newColor"`
		ModifyColor    string `yaml:"modifyColor"`
		AddColor       string `yaml:"addColor"`
		ErrorColor     string `yaml:"errorColor"`
		HighlightColor string `yaml:"highlightColor"`
		KillColor      string `yaml:"killColor"`
		CompletedColor string `yaml:"completedColor"`
	}

	// Log tracks Log styles.
	Log struct {
		FgColor string `yaml:"fgColor"`
		BgColor string `yaml:"bgColor"`
	}

	// Yaml tracks yaml styles.
	Yaml struct {
		KeyColor   string `yaml:"keyColor"`
		ValueColor string `yaml:"valueColor"`
		ColonColor string `yaml:"colonColor"`
	}

	// Title tracks title styles.
	Title struct {
		FgColor        string `yaml:"fgColor"`
		BgColor        string `yaml:"bgColor"`
		HighlightColor string `yaml:"highlightColor"`
		CounterColor   string `yaml:"counterColor"`
		FilterColor    string `yaml:"filterColor"`
	}

	// Info tracks info styles.
	Info struct {
		SectionColor string `yaml:"sectionColor"`
		FgColor      string `yaml:"fgColor"`
	}

	// Border tracks border styles.
	Border struct {
		FgColor    string `yaml:"fgColor"`
		FocusColor string `yaml:"focusColor"`
	}

	// Crumb tracks crumbs styles.
	Crumb struct {
		FgColor     string `yaml:"fgColor"`
		BgColor     string `yaml:"bgColor"`
		ActiveColor string `yaml:"activeColor"`
	}

	// Table tracks table styles.
	Table struct {
		FgColor     string       `yaml:"fgColor"`
		BgColor     string       `yaml:"bgColor"`
		CursorColor string       `yaml:"cursorColor"`
		Header      *TableHeader `yaml:"header"`
	}

	// TableHeader tracks table header styles.
	TableHeader struct {
		FgColor     string `yaml:"fgColor"`
		BgColor     string `yaml:"bgColor"`
		SorterColor string `yaml:"sorterColor"`
	}

	// Menu tracks menu styles.
	Menu struct {
		FgColor     string `yaml:"fgColor"`
		KeyColor    string `yaml:"keyColor"`
		NumKeyColor string `yaml:"numKeyColor"`
	}
)

func newStyle() *Style {
	return &Style{
		FgColor:   "cadetblue",
		BgColor:   "black",
		LogoColor: "orange",
		Info:      newInfo(),
		Border:    newBorder(),
		Menu:      newMenu(),
		Crumb:     newCrumb(),
		Table:     newTable(),
		Status:    newStatus(),
		Title:     newTitle(),
		Yaml:      newYaml(),
		Log:       newLog(),
	}
}

func newStatus() *Status {
	return &Status{
		NewColor:       "lightskyblue",
		ModifyColor:    "greenyellow",
		AddColor:       "dodgerblue",
		ErrorColor:     "orangered",
		HighlightColor: "aqua",
		KillColor:      "mediumpurple",
		CompletedColor: "gray",
	}
}

// NewLog returns a new log style.
func newLog() *Log {
	return &Log{
		FgColor: "lightskyblue",
		BgColor: "black",
	}
}

// NewYaml returns a new yaml style.
func newYaml() *Yaml {
	return &Yaml{
		KeyColor:   "steelblue",
		ColonColor: "white",
		ValueColor: "papayawhip",
	}
}

// NewTitle returns a new title style.
func newTitle() *Title {
	return &Title{
		FgColor:        "aqua",
		BgColor:        "black",
		HighlightColor: "fuchsia",
		CounterColor:   "papayawhip",
		FilterColor:    "steelblue",
	}
}

// NewInfo returns a new info style.
func newInfo() *Info {
	return &Info{
		SectionColor: "white",
		FgColor:      "orange",
	}
}

// NewTable returns a new table style.
func newTable() *Table {
	return &Table{
		FgColor:     "aqua",
		BgColor:     "black",
		CursorColor: "aqua",
		Header:      newTableHeader(),
	}
}

// NewTableHeader returns a new table header style.
func newTableHeader() *TableHeader {
	return &TableHeader{
		FgColor:     "white",
		BgColor:     "black",
		SorterColor: "aqua",
	}
}

// NewCrumb returns a new crumbs style.
func newCrumb() *Crumb {
	return &Crumb{
		FgColor:     "black",
		BgColor:     "aqua",
		ActiveColor: "orange",
	}
}

// NewBorder returns a new border style.
func newBorder() *Border {
	return &Border{
		FgColor:    "dodgerblue",
		FocusColor: "lightskyblue",
	}
}

// NewMenu returns a new menu style.
func newMenu() *Menu {
	return &Menu{
		FgColor:     "white",
		KeyColor:    "dodgerblue",
		NumKeyColor: "fuchsia",
	}
}

// NewStyles creates a new default config.
func NewStyles() (*Styles, error) {
	s := &Styles{Style: newStyle()}
	err := s.load(K9sStylesFile)

	return s, err
}

// Ensure default styles are applied in not in stylesheet.
func (s *Styles) ensure() {
	if s.Style == nil {
		s.Style = newStyle()
	}

	if s.Style.Info == nil {
		s.Style.Info = newInfo()
	}

	if s.Style.Border == nil {
		s.Style.Border = newBorder()
	}

	if s.Style.Table == nil {
		s.Style.Table = newTable()
	}

	if s.Style.Menu == nil {
		s.Style.Menu = newMenu()
	}

	if s.Style.Crumb == nil {
		s.Style.Crumb = newCrumb()
	}

	if s.Style.Status == nil {
		s.Style.Status = newStatus()
	}

	if s.Style.Title == nil {
		s.Style.Title = newTitle()
	}

	if s.Style.Yaml == nil {
		s.Style.Yaml = newYaml()
	}

	if s.Style.Log == nil {
		s.Style.Log = newLog()
	}
}

// FgColor returns the foreground color.
func (s *Styles) FgColor() tcell.Color {
	return AsColor(s.Style.FgColor)
}

// BgColor returns the background color.
func (s *Styles) BgColor() tcell.Color {
	return AsColor(s.Style.BgColor)
}

// Load K9s configuration from file
func (s *Styles) load(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg Styles
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return err
	}

	if cfg.Style != nil {
		s.Style = cfg.Style
	}
	s.ensure()

	return nil
}

// Update apply terminal colors based on styles.
func (s *Styles) Update() {
	tview.Styles.PrimitiveBackgroundColor = AsColor(s.Style.BgColor)
	tview.Styles.ContrastBackgroundColor = AsColor(s.Style.BgColor)
	tview.Styles.PrimaryTextColor = AsColor(s.Style.FgColor)
	tview.Styles.BorderColor = AsColor(s.Style.Border.FgColor)
	tview.Styles.FocusColor = AsColor(s.Style.Border.FocusColor)
}

// AsColor checks color index, if match return color otherwise pink it is.
func AsColor(c string) tcell.Color {
	if color, ok := tcell.ColorNames[c]; ok {
		return color
	}

	return tcell.ColorPink
}
