// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"gopkg.in/yaml.v2"
)

// StyleListener represents a skin's listener.
type StyleListener interface {
	// StylesChanged notifies listener the skin changed.
	StylesChanged(*Styles)
}

type (
	// Styles tracks K9s styling options.
	Styles struct {
		K9s       Style `json:"k9s" yaml:"k9s"`
		listeners []StyleListener
	}

	// Style tracks K9s styles.
	Style struct {
		Body   Body   `json:"body" yaml:"body"`
		Prompt Prompt `json:"prompt" yaml:"prompt"`
		Help   Help   `json:"help" yaml:"help"`
		Frame  Frame  `json:"frame" yaml:"frame"`
		Info   Info   `json:"info" yaml:"info"`
		Views  Views  `json:"views" yaml:"views"`
		Dialog Dialog `json:"dialog" yaml:"dialog"`
	}

	// Prompt tracks command styles
	Prompt struct {
		FgColor      Color        `json:"fgColor" yaml:"fgColor"`
		BgColor      Color        `json:"bgColor" yaml:"bgColor"`
		SuggestColor Color        `json:"" yaml:"suggestColor"`
		Border       PromptBorder `json:"" yaml:"border"`
	}

	// PromptBorder tracks the color of the prompt depending on its kind (e.g., command or filter)
	PromptBorder struct {
		CommandColor Color `json:"command" yaml:"command"`
		DefaultColor Color `json:"default" yaml:"default"`
	}

	// Help tracks help styles.
	Help struct {
		FgColor      Color `json:"fgColor" yaml:"fgColor"`
		BgColor      Color `json:"bgColor" yaml:"bgColor"`
		SectionColor Color `json:"sectionColor" yaml:"sectionColor"`
		KeyColor     Color `json:"keyColor" yaml:"keyColor"`
		NumKeyColor  Color `json:"numKeyColor" yaml:"numKeyColor"`
	}

	// Body tracks body styles.
	Body struct {
		FgColor        Color `json:"fgColor" yaml:"fgColor"`
		BgColor        Color `json:"bgColor" yaml:"bgColor"`
		LogoColor      Color `json:"logoColor" yaml:"logoColor"`
		LogoColorMsg   Color `json:"logoColorMsg" yaml:"logoColorMsg"`
		LogoColorInfo  Color `json:"logoColorInfo" yaml:"logoColorInfo"`
		LogoColorWarn  Color `json:"logoColorWarn" yaml:"logoColorWarn"`
		LogoColorError Color `json:"logoColorError" yaml:"logoColorError"`
	}

	// Dialog tracks dialog styles.
	Dialog struct {
		FgColor            Color `json:"fgColor" yaml:"fgColor"`
		BgColor            Color `json:"bgColor" yaml:"bgColor"`
		ButtonFgColor      Color `json:"buttonFgColor" yaml:"buttonFgColor"`
		ButtonBgColor      Color `json:"buttonBgColor" yaml:"buttonBgColor"`
		ButtonFocusFgColor Color `json:"buttonFocusFgColor" yaml:"buttonFocusFgColor"`
		ButtonFocusBgColor Color `json:"buttonFocusBgColor" yaml:"buttonFocusBgColor"`
		LabelFgColor       Color `json:"labelFgColor" yaml:"labelFgColor"`
		FieldFgColor       Color `json:"fieldFgColor" yaml:"fieldFgColor"`
	}

	// Frame tracks frame styles.
	Frame struct {
		Title  Title  `json:"title" yaml:"title"`
		Border Border `json:"border" yaml:"border"`
		Menu   Menu   `json:"menu" yaml:"menu"`
		Crumb  Crumb  `json:"crumbs" yaml:"crumbs"`
		Status Status `json:"status" yaml:"status"`
	}

	// Views tracks individual view styles.
	Views struct {
		Table  Table  `json:"table" yaml:"table"`
		Xray   Xray   `json:"xray" yaml:"xray"`
		Charts Charts `json:"charts" yaml:"charts"`
		Yaml   Yaml   `json:"yaml" yaml:"yaml"`
		Picker Picker `json:"picker" yaml:"picker"`
		Log    Log    `json:"logs" yaml:"logs"`
	}

	// Status tracks resource status styles.
	Status struct {
		NewColor       Color `json:"newColor" yaml:"newColor"`
		ModifyColor    Color `json:"modifyColor" yaml:"modifyColor"`
		AddColor       Color `json:"addColor" yaml:"addColor"`
		PendingColor   Color `json:"pendingColor" yaml:"pendingColor"`
		ErrorColor     Color `json:"errorColor" yaml:"errorColor"`
		HighlightColor Color `json:"highlightColor" yaml:"highlightColor"`
		KillColor      Color `json:"killColor" yaml:"killColor"`
		CompletedColor Color `json:"completedColor" yaml:"completedColor"`
	}

	// Log tracks Log styles.
	Log struct {
		FgColor   Color        `json:"fgColor" yaml:"fgColor"`
		BgColor   Color        `json:"bgColor" yaml:"bgColor"`
		Indicator LogIndicator `json:"indicator" yaml:"indicator"`
	}

	// Picker tracks color when selecting containers
	Picker struct {
		MainColor     Color `json:"mainColor" yaml:"mainColor"`
		FocusColor    Color `json:"focusColor" yaml:"focusColor"`
		ShortcutColor Color `json:"shortcutColor" yaml:"shortcutColor"`
	}

	// LogIndicator tracks log view indicator.
	LogIndicator struct {
		FgColor        Color `json:"fgColor" yaml:"fgColor"`
		BgColor        Color `json:"bgColor" yaml:"bgColor"`
		ToggleOnColor  Color `json:"toggleOnColor" yaml:"toggleOnColor"`
		ToggleOffColor Color `json:"toggleOffColor" yaml:"toggleOffColor"`
	}

	// Yaml tracks yaml styles.
	Yaml struct {
		KeyColor   Color `json:"keyColor" yaml:"keyColor"`
		ValueColor Color `json:"valueColor" yaml:"valueColor"`
		ColonColor Color `json:"colonColor" yaml:"colonColor"`
	}

	// Title tracks title styles.
	Title struct {
		FgColor        Color `json:"fgColor" yaml:"fgColor"`
		BgColor        Color `json:"bgColor" yaml:"bgColor"`
		HighlightColor Color `json:"highlightColor" yaml:"highlightColor"`
		CounterColor   Color `json:"counterColor" yaml:"counterColor"`
		FilterColor    Color `json:"filterColor" yaml:"filterColor"`
	}

	// Info tracks info styles.
	Info struct {
		SectionColor Color `json:"sectionColor" yaml:"sectionColor"`
		FgColor      Color `json:"fgColor" yaml:"fgColor"`
	}

	// Border tracks border styles.
	Border struct {
		FgColor    Color `json:"fgColor" yaml:"fgColor"`
		FocusColor Color `json:"focusColor" yaml:"focusColor"`
	}

	// Crumb tracks crumbs styles.
	Crumb struct {
		FgColor     Color `json:"fgColor" yaml:"fgColor"`
		BgColor     Color `json:"bgColor" yaml:"bgColor"`
		ActiveColor Color `json:"activeColor" yaml:"activeColor"`
	}

	// Table tracks table styles.
	Table struct {
		FgColor       Color       `json:"fgColor" yaml:"fgColor"`
		BgColor       Color       `json:"bgColor" yaml:"bgColor"`
		CursorFgColor Color       `json:"cursorFgColor" yaml:"cursorFgColor"`
		CursorBgColor Color       `json:"cursorBgColor" yaml:"cursorBgColor"`
		MarkColor     Color       `json:"markColor" yaml:"markColor"`
		Header        TableHeader `json:"header" yaml:"header"`
	}

	// TableHeader tracks table header styles.
	TableHeader struct {
		FgColor     Color `json:"fgColor" yaml:"fgColor"`
		BgColor     Color `json:"bgColor" yaml:"bgColor"`
		SorterColor Color `json:"sorterColor" yaml:"sorterColor"`
	}

	// Xray tracks xray styles.
	Xray struct {
		FgColor         Color `json:"fgColor" yaml:"fgColor"`
		BgColor         Color `json:"bgColor" yaml:"bgColor"`
		CursorColor     Color `json:"cursorColor" yaml:"cursorColor"`
		CursorTextColor Color `json:"cursorTextColor" yaml:"cursorTextColor"`
		GraphicColor    Color `json:"graphicColor" yaml:"graphicColor"`
	}

	// Menu tracks menu styles.
	Menu struct {
		FgColor     Color `json:"fgColor" yaml:"fgColor"`
		KeyColor    Color `json:"keyColor" yaml:"keyColor"`
		NumKeyColor Color `json:"numKeyColor" yaml:"numKeyColor"`
	}

	// Charts tracks charts styles.
	Charts struct {
		BgColor            Color             `json:"bgColor" yaml:"bgColor"`
		DialBgColor        Color             `json:"dialBgColor" yaml:"dialBgColor"`
		ChartBgColor       Color             `json:"chartBgColor" yaml:"chartBgColor"`
		DefaultDialColors  Colors            `json:"defaultDialColors" yaml:"defaultDialColors"`
		DefaultChartColors Colors            `json:"defaultChartColors" yaml:"defaultChartColors"`
		ResourceColors     map[string]Colors `json:"resourceColors" yaml:"resourceColors"`
	}
)

func newStyle() Style {
	return Style{
		Body:   newBody(),
		Prompt: newPrompt(),
		Help:   newHelp(),
		Frame:  newFrame(),
		Info:   newInfo(),
		Views:  newViews(),
		Dialog: newDialog(),
	}
}

func newDialog() Dialog {
	return Dialog{
		FgColor:            "cadetblue",
		BgColor:            "black",
		ButtonBgColor:      "darkslateblue",
		ButtonFgColor:      "black",
		ButtonFocusBgColor: "dodgerblue",
		ButtonFocusFgColor: "black",
		LabelFgColor:       "white",
		FieldFgColor:       "white",
	}
}

func newPrompt() Prompt {
	return Prompt{
		FgColor:      "cadetblue",
		BgColor:      "black",
		SuggestColor: "dodgerblue",
		Border: PromptBorder{
			DefaultColor: "seagreen",
			CommandColor: "aqua",
		},
	}
}

func newCharts() Charts {
	return Charts{
		BgColor:            "black",
		DialBgColor:        "black",
		ChartBgColor:       "black",
		DefaultDialColors:  Colors{Color("palegreen"), Color("orangered")},
		DefaultChartColors: Colors{Color("palegreen"), Color("orangered")},
		ResourceColors: map[string]Colors{
			"cpu": {Color("dodgerblue"), Color("darkslateblue")},
			"mem": {Color("yellow"), Color("goldenrod")},
		},
	}
}

func newViews() Views {
	return Views{
		Table:  newTable(),
		Xray:   newXray(),
		Charts: newCharts(),
		Yaml:   newYaml(),
		Picker: newPicker(),
		Log:    newLog(),
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

func newHelp() Help {
	return Help{
		FgColor:      "cadetblue",
		BgColor:      "black",
		SectionColor: "green",
		KeyColor:     "dodgerblue",
		NumKeyColor:  "fuchsia",
	}
}

func newBody() Body {
	return Body{
		FgColor:        "cadetblue",
		BgColor:        "black",
		LogoColor:      "orange",
		LogoColorMsg:   "white",
		LogoColorInfo:  "green",
		LogoColorWarn:  "mediumvioletred",
		LogoColorError: "red",
	}
}

func newStatus() Status {
	return Status{
		NewColor:       "lightskyblue",
		ModifyColor:    "greenyellow",
		AddColor:       "dodgerblue",
		PendingColor:   "darkorange",
		ErrorColor:     "orangered",
		HighlightColor: "aqua",
		KillColor:      "mediumpurple",
		CompletedColor: "lightslategray",
	}
}

func newPicker() Picker {
	return Picker{
		MainColor:     "white",
		FocusColor:    "aqua",
		ShortcutColor: "aqua",
	}
}

func newLog() Log {
	return Log{
		FgColor:   "lightskyblue",
		BgColor:   "black",
		Indicator: newLogIndicator(),
	}
}

func newLogIndicator() LogIndicator {
	return LogIndicator{
		FgColor:        "dodgerblue",
		BgColor:        "black",
		ToggleOnColor:  "limegreen",
		ToggleOffColor: "gray",
	}
}

func newYaml() Yaml {
	return Yaml{
		KeyColor:   "steelblue",
		ColonColor: "white",
		ValueColor: "papayawhip",
	}
}

func newTitle() Title {
	return Title{
		FgColor:        "aqua",
		BgColor:        "black",
		HighlightColor: "fuchsia",
		CounterColor:   "papayawhip",
		FilterColor:    "seagreen",
	}
}

func newInfo() Info {
	return Info{
		SectionColor: "white",
		FgColor:      "orange",
	}
}

func newXray() Xray {
	return Xray{
		FgColor:         "aqua",
		BgColor:         "black",
		CursorColor:     "dodgerblue",
		CursorTextColor: "black",
		GraphicColor:    "cadetblue",
	}
}

func newTable() Table {
	return Table{
		FgColor:       "aqua",
		BgColor:       "black",
		CursorFgColor: "black",
		CursorBgColor: "aqua",
		MarkColor:     "palegreen",
		Header:        newTableHeader(),
	}
}

func newTableHeader() TableHeader {
	return TableHeader{
		FgColor:     "white",
		BgColor:     "black",
		SorterColor: "aqua",
	}
}

func newCrumb() Crumb {
	return Crumb{
		FgColor:     "black",
		BgColor:     "aqua",
		ActiveColor: "orange",
	}
}

func newBorder() Border {
	return Border{
		FgColor:    "dodgerblue",
		FocusColor: "lightskyblue",
	}
}

func newMenu() Menu {
	return Menu{
		FgColor:     "white",
		KeyColor:    "dodgerblue",
		NumKeyColor: "fuchsia",
	}
}

// NewStyles creates a new default config.
func NewStyles() *Styles {
	var s Styles
	if err := yaml.Unmarshal(stockSkinTpl, &s); err == nil {
		return &s
	}

	return &Styles{
		K9s: newStyle(),
	}
}

// Reset resets styles.
func (s *Styles) Reset() {
	if err := yaml.Unmarshal(stockSkinTpl, s); err != nil {
		s.K9s = newStyle()
	}
}

// FgColor returns the foreground color.
func (s *Styles) FgColor() tcell.Color {
	return s.Body().FgColor.Color()
}

// BgColor returns the background color.
func (s *Styles) BgColor() tcell.Color {
	return s.Body().BgColor.Color()
}

// AddListener registers a new listener.
func (s *Styles) AddListener(l StyleListener) {
	s.listeners = append(s.listeners, l)
}

// RemoveListener removes a listener.
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

// Prompt returns prompt styles.
func (s *Styles) Prompt() Prompt {
	return s.K9s.Prompt
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

// Charts returns charts styles.
func (s *Styles) Charts() Charts {
	return s.K9s.Views.Charts
}

// Dialog returns dialog styles.
func (s *Styles) Dialog() Dialog {
	return s.K9s.Dialog
}

// Table returns table styles.
func (s *Styles) Table() Table {
	return s.K9s.Views.Table
}

// Xray returns xray styles.
func (s *Styles) Xray() Xray {
	return s.K9s.Views.Xray
}

// Views returns views styles.
func (s *Styles) Views() Views {
	return s.K9s.Views
}

// Load K9s configuration from file.
func (s *Styles) Load(path string) error {
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.SkinSchema, bb); err != nil {
		return err
	}
	if err := yaml.Unmarshal(bb, s); err != nil {
		return err
	}

	return nil
}

// Update apply terminal colors based on styles.
func (s *Styles) Update() {
	tview.Styles.PrimitiveBackgroundColor = s.BgColor()
	tview.Styles.ContrastBackgroundColor = s.BgColor()
	tview.Styles.MoreContrastBackgroundColor = s.BgColor()
	tview.Styles.PrimaryTextColor = s.FgColor()
	tview.Styles.BorderColor = s.K9s.Frame.Border.FgColor.Color()
	tview.Styles.FocusColor = s.K9s.Frame.Border.FocusColor.Color()
	tview.Styles.TitleColor = s.FgColor()
	tview.Styles.GraphicsColor = s.FgColor()
	tview.Styles.SecondaryTextColor = s.FgColor()
	tview.Styles.TertiaryTextColor = s.FgColor()
	tview.Styles.InverseTextColor = s.FgColor()
	tview.Styles.ContrastSecondaryTextColor = s.FgColor()

	s.fireStylesChanged()
}

// Dump for debug.
func (s *Styles) Dump() {
	bb, _ := yaml.Marshal(s)
	fmt.Println(string(bb))
}
