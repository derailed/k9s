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

// StyleListener represents a skin's listener.
type StyleListener interface {
	// StylesChanged notifies listener the skin changed.
	StylesChanged(*Styles)
}

type (
	// Styles tracks K9s styling options.
	Styles struct {
		K9s       Style `yaml:"k9s"`
		listeners []StyleListener
	}

	// Body tracks body styles.
	Body struct {
		FgColor   string `yaml:"fgColor"`
		BgColor   string `yaml:"bgColor"`
		LogoColor string `yaml:"logoColor"`
	}

	// Frame tracks frame styles.
	Frame struct {
		Title  Title  `yaml:"title"`
		Border Border `yaml:"border"`
		Menu   Menu   `yaml:"menu"`
		Crumb  Crumb  `yaml:"crumbs"`
		Status Status `yaml:"status"`
	}

	// Views tracks individual view styles.
	Views struct {
		Yaml Yaml `yaml:"yaml"`
		Log  Log  `yaml:"logs"`
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
		FgColor     string      `yaml:"fgColor"`
		BgColor     string      `yaml:"bgColor"`
		CursorColor string      `yaml:"cursorColor"`
		MarkColor   string      `yaml:"markColor"`
		Header      TableHeader `yaml:"header"`
	}

	// TableHeader tracks table header styles.
	TableHeader struct {
		FgColor     string `yaml:"fgColor"`
		BgColor     string `yaml:"bgColor"`
		SorterColor string `yaml:"sorterColor"`
	}

	// Xray tracks xray styles.
	Xray struct {
		FgColor      string `yaml:"fgColor"`
		BgColor      string `yaml:"bgColor"`
		CursorColor  string `yaml:"cursorColor"`
		GraphicColor string `yaml:"graphicColor"`
		ShowIcons    bool   `yaml:"showIcons"`
	}

	// Menu tracks menu styles.
	Menu struct {
		FgColor     string `yaml:"fgColor"`
		KeyColor    string `yaml:"keyColor"`
		NumKeyColor string `yaml:"numKeyColor"`
	}

	// Style tracks K9s styles.
	Style struct {
		Body  Body  `yaml:"body"`
		Frame Frame `yaml:"frame"`
		Info  Info  `yaml:"info"`
		Table Table `yaml:"table"`
		Xray  Xray  `yaml:"xray"`
		Views Views `yaml:"views"`
	}
)

func newStyle() Style {
	return Style{
		Body:  newBody(),
		Frame: newFrame(),
		Info:  newInfo(),
		Table: newTable(),
		Views: newViews(),
		Xray:  newXray(),
	}
}

func newViews() Views {
	return Views{
		Yaml: newYaml(),
		Log:  newLog(),
	}
}

func newFrame() Frame {
	return Frame{
		Title:  newTitle(),
		Border: newBorder(),
		Menu:   newMenu(),
		Crumb:  newCrumb(),
		Status: newStatus(),
	}
}

func newBody() Body {
	return Body{
		FgColor:   "cadetblue",
		BgColor:   "black",
		LogoColor: "orange",
	}
}

func newStatus() Status {
	return Status{
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
func newLog() Log {
	return Log{
		FgColor: "lightskyblue",
		BgColor: "black",
	}
}

// NewYaml returns a new yaml style.
func newYaml() Yaml {
	return Yaml{
		KeyColor:   "steelblue",
		ColonColor: "white",
		ValueColor: "papayawhip",
	}
}

// NewTitle returns a new title style.
func newTitle() Title {
	return Title{
		FgColor:        "aqua",
		BgColor:        "black",
		HighlightColor: "fuchsia",
		CounterColor:   "papayawhip",
		FilterColor:    "seagreen",
	}
}

// NewInfo returns a new info style.
func newInfo() Info {
	return Info{
		SectionColor: "white",
		FgColor:      "orange",
	}
}

// NewXray returns a new xray style.
func newXray() Xray {
	return Xray{
		FgColor:      "aqua",
		BgColor:      "black",
		CursorColor:  "whitesmoke",
		GraphicColor: "floralwhite",
		ShowIcons:    true,
	}
}

// NewTable returns a new table style.
func newTable() Table {
	return Table{
		FgColor:     "aqua",
		BgColor:     "black",
		CursorColor: "aqua",
		MarkColor:   "palegreen",
		Header:      newTableHeader(),
	}
}

// NewTableHeader returns a new table header style.
func newTableHeader() TableHeader {
	return TableHeader{
		FgColor:     "white",
		BgColor:     "black",
		SorterColor: "aqua",
	}
}

// NewCrumb returns a new crumbs style.
func newCrumb() Crumb {
	return Crumb{
		FgColor:     "black",
		BgColor:     "aqua",
		ActiveColor: "orange",
	}
}

// NewBorder returns a new border style.
func newBorder() Border {
	return Border{
		FgColor:    "dodgerblue",
		FocusColor: "lightskyblue",
	}
}

// NewMenu returns a new menu style.
func newMenu() Menu {
	return Menu{
		FgColor:     "white",
		KeyColor:    "dodgerblue",
		NumKeyColor: "fuchsia",
	}
}

// NewStyles creates a new default config.
func NewStyles() *Styles {
	return &Styles{
		K9s: newStyle(),
	}
}

// FgColor returns the foreground color.
func (s *Styles) FgColor() tcell.Color {
	return AsColor(s.Body().FgColor)
}

// BgColor returns the background color.
func (s *Styles) BgColor() tcell.Color {
	return AsColor(s.Body().BgColor)
}

// AddListener registers a new listener.
func (s *Styles) AddListener(l StyleListener) {
	s.listeners = append(s.listeners, l)
}

// RemoveListener unregister a listener.
func (s *Styles) RemoveListener(l StyleListener) {
	victim := -1
	for i, lis := range s.listeners {
		if lis == l {
			victim = i
			break
		}
	}
	if victim == -1 {
		return
	}
	s.listeners = append(s.listeners[:victim], s.listeners[victim+1:]...)
}

func (s *Styles) fireStylesChanged() {
	for _, list := range s.listeners {
		list.StylesChanged(s)
	}
}

// Body returns body styles.
func (s *Styles) Body() Body {
	return s.K9s.Body
}

// Frame returns frame styles.
func (s *Styles) Frame() Frame {
	return s.K9s.Frame
}

// Crumb returns crumb styles.
func (s *Styles) Crumb() Crumb {
	return s.Frame().Crumb
}

// Title returns title styles.
func (s *Styles) Title() Title {
	return s.Frame().Title
}

// GetTable returns table styles.
func (s *Styles) Table() Table {
	return s.K9s.Table
}

// Xray returns xray styles.
func (s *Styles) Xray() Xray {
	return s.K9s.Xray
}

// Views returns views styles.
func (s *Styles) Views() Views {
	return s.K9s.Views
}

// Load K9s configuration from file
func (s *Styles) Load(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(f, s); err != nil {
		return err
	}
	s.fireStylesChanged()

	return nil
}

// Update apply terminal colors based on styles.
func (s *Styles) Update() {
	tview.Styles.PrimitiveBackgroundColor = s.BgColor()
	tview.Styles.ContrastBackgroundColor = s.BgColor()
	tview.Styles.PrimaryTextColor = s.FgColor()
	tview.Styles.BorderColor = AsColor(s.K9s.Frame.Border.FgColor)
	tview.Styles.FocusColor = AsColor(s.K9s.Frame.Border.FocusColor)
}

// AsColor checks color index, if match return color otherwise pink it is.
func AsColor(c string) tcell.Color {
	if color, ok := tcell.ColorNames[c]; ok {
		return color
	}

	return tcell.GetColor(c)
}
