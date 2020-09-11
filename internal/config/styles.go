package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"gopkg.in/yaml.v2"
)

// K9sStylesFile represents K9s skins file location.
var K9sStylesFile = filepath.Join(K9sHome(), "skin.yml")

// StyleListener represents a skin's listener.
type StyleListener interface {
	// StylesChanged notifies listener the skin changed.
	StylesChanged(*Styles)
}

type (
	// Color represents a color.
	Color string

	// Colors tracks multiple colors.
	Colors []Color

	// Styles tracks K9s styling options.
	Styles struct {
		K9s       Style `yaml:"k9s"`
		listeners []StyleListener
	}

	// Style tracks K9s styles.
	Style struct {
		Body   Body   `yaml:"body"`
		Frame  Frame  `yaml:"frame"`
		Info   Info   `yaml:"info"`
		Views  Views  `yaml:"views"`
		Dialog Dialog `yaml:"dialog"`
	}

	// Body tracks body styles.
	Body struct {
		FgColor   Color `yaml:"fgColor"`
		BgColor   Color `yaml:"bgColor"`
		LogoColor Color `yaml:"logoColor"`
	}

	// Dialog tracks dialog styles.
	Dialog struct {
		FgColor            Color `yaml:"fgColor"`
		BgColor            Color `yaml:"bgColor"`
		ButtonFgColor      Color `yaml:"buttonFgColor"`
		ButtonBgColor      Color `yaml:"buttonBgColor"`
		ButtonFocusFgColor Color `yaml:"buttonFocusFgColor"`
		ButtonFocusBgColor Color `yaml:"buttonFocusBgColor"`
		LabelFgColor       Color `yaml:"labelFgColor"`
		FieldFgColor       Color `yaml:"fieldFgColor"`
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
		Table  Table  `yaml:"table"`
		Xray   Xray   `yaml:"xray"`
		Charts Charts `yaml:"charts"`
		Yaml   Yaml   `yaml:"yaml"`
		Log    Log    `yaml:"logs"`
	}

	// Status tracks resource status styles.
	Status struct {
		NewColor       Color `yaml:"newColor"`
		ModifyColor    Color `yaml:"modifyColor"`
		AddColor       Color `yaml:"addColor"`
		PendingColor   Color `yaml:"pendingColor"`
		ErrorColor     Color `yaml:"errorColor"`
		HighlightColor Color `yaml:"highlightColor"`
		KillColor      Color `yaml:"killColor"`
		CompletedColor Color `yaml:"completedColor"`
	}

	// Log tracks Log styles.
	Log struct {
		FgColor   Color        `yaml:"fgColor"`
		BgColor   Color        `yaml:"bgColor"`
		Indicator LogIndicator `yaml:"indicator"`
	}

	// LogIndicator tracks log view indicator.
	LogIndicator struct {
		FgColor Color `yaml:"fgColor"`
		BgColor Color `yaml:"bgColor"`
	}

	// Yaml tracks yaml styles.
	Yaml struct {
		KeyColor   Color `yaml:"keyColor"`
		ValueColor Color `yaml:"valueColor"`
		ColonColor Color `yaml:"colonColor"`
	}

	// Title tracks title styles.
	Title struct {
		FgColor        Color `yaml:"fgColor"`
		BgColor        Color `yaml:"bgColor"`
		HighlightColor Color `yaml:"highlightColor"`
		CounterColor   Color `yaml:"counterColor"`
		FilterColor    Color `yaml:"filterColor"`
	}

	// Info tracks info styles.
	Info struct {
		SectionColor Color `yaml:"sectionColor"`
		FgColor      Color `yaml:"fgColor"`
	}

	// Border tracks border styles.
	Border struct {
		FgColor    Color `yaml:"fgColor"`
		FocusColor Color `yaml:"focusColor"`
	}

	// Crumb tracks crumbs styles.
	Crumb struct {
		FgColor     Color `yaml:"fgColor"`
		BgColor     Color `yaml:"bgColor"`
		ActiveColor Color `yaml:"activeColor"`
	}

	// Table tracks table styles.
	Table struct {
		FgColor       Color       `yaml:"fgColor"`
		BgColor       Color       `yaml:"bgColor"`
		CursorFgColor Color       `yaml:"cursorFgColor"`
		CursorBgColor Color       `yaml:"cursorBgColor"`
		MarkColor     Color       `yaml:"markColor"`
		Header        TableHeader `yaml:"header"`
	}

	// TableHeader tracks table header styles.
	TableHeader struct {
		FgColor     Color `yaml:"fgColor"`
		BgColor     Color `yaml:"bgColor"`
		SorterColor Color `yaml:"sorterColor"`
	}

	// Xray tracks xray styles.
	Xray struct {
		FgColor         Color `yaml:"fgColor"`
		BgColor         Color `yaml:"bgColor"`
		CursorColor     Color `yaml:"cursorColor"`
		CursorTextColor Color `yaml:"cursorTextColor"`
		GraphicColor    Color `yaml:"graphicColor"`
	}

	// Menu tracks menu styles.
	Menu struct {
		FgColor     Color `yaml:"fgColor"`
		KeyColor    Color `yaml:"keyColor"`
		NumKeyColor Color `yaml:"numKeyColor"`
	}

	// Charts tracks charts styles.
	Charts struct {
		BgColor            Color             `yaml:"bgColor"`
		DialBgColor        Color             `yaml:"dialBgColor"`
		ChartBgColor       Color             `yaml:"chartBgColor"`
		DefaultDialColors  Colors            `yaml:"defaultDialColors"`
		DefaultChartColors Colors            `yaml:"defaultChartColors"`
		ResourceColors     map[string]Colors `yaml:"resourceColors"`
	}
)

const (
	// DefaultColor represents  a default color.
	DefaultColor Color = "default"

	// TransparentColor represents the terminal bg color.
	TransparentColor Color = "-"
)

// NewColor returns a new color.
func NewColor(c string) Color {
	return Color(c)
}

// String returns color as string.
func (c Color) String() string {
	return string(c)
}

// Color returns a view color.
func (c Color) Color() tcell.Color {
	if c == DefaultColor {
		return tcell.ColorDefault
	}
	if color, ok := tcell.ColorNames[c.String()]; ok {
		return color
	}
	return tcell.GetColor(c.String())
}

// Colors converts series string colors to colors.
func (c Colors) Colors() []tcell.Color {
	cc := make([]tcell.Color, 0, len(c))
	for _, color := range c {
		cc = append(cc, color.Color())
	}
	return cc
}

func newStyle() Style {
	return Style{
		Body:   newBody(),
		Frame:  newFrame(),
		Info:   newInfo(),
		Views:  newViews(),
		Dialog: newDialog(),
	}
}

func newDialog() Dialog {
	return Dialog{
		FgColor:            "cadetBlue",
		BgColor:            "black",
		ButtonBgColor:      "darkslateblue",
		ButtonFgColor:      "black",
		ButtonFocusBgColor: "dodgerblue",
		ButtonFocusFgColor: "black",
		LabelFgColor:       "white",
		FieldFgColor:       "white",
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
		PendingColor:   "darkorange",
		ErrorColor:     "orangered",
		HighlightColor: "aqua",
		KillColor:      "mediumpurple",
		CompletedColor: "lightslategray",
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
		FgColor: "dodgerblue",
		BgColor: "black",
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
	return &Styles{
		K9s: newStyle(),
	}
}

// Reset resets styles.
func (s *Styles) Reset() {
	s.K9s = newStyle()
}

// DefaultSkin loads the default skin
func (s *Styles) DefaultSkin() {
	s.K9s = newStyle()
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
	tview.Styles.MoreContrastBackgroundColor = s.BgColor()
	tview.Styles.PrimaryTextColor = s.FgColor()
	tview.Styles.BorderColor = s.K9s.Frame.Border.FgColor.Color()
	tview.Styles.FocusColor = s.K9s.Frame.Border.FocusColor.Color()
	s.fireStylesChanged()
}
