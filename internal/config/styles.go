// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"gopkg.in/yaml.v3"
)

// StyleListener represents a skin's listener.
type StyleListener interface {
	// StylesChanged notifies listener the skin changed.
	StylesChanged(*Styles)
}

type TextStyle string

const (
	TextStyleNormal TextStyle = "normal"
	TextStyleBold   TextStyle = "bold"
	TextStyleDim    TextStyle = "dim"
)

func (ts TextStyle) ToShortString() string {
	switch ts {
	case TextStyleNormal:
		return "-"
	case TextStyleBold:
		return "b"
	case TextStyleDim:
		return "d"
	default:
		return "d"
	}
}

type (
	// Styles tracks K9s styling options.
	Styles struct {
		Skin  StyleConfig
		Emoji EmojiConfig

		listeners []StyleListener
	}

	StyleConfig struct {
		K9s Style `json:"k9s" yaml:"k9s"`
	}

	// Style tracks K9s styles.
	Style struct {
		Body   Body       `json:"body" yaml:"body"`
		Prompt Prompt     `json:"prompt" yaml:"prompt"`
		Help   Help       `json:"help" yaml:"help"`
		Frame  Frame      `json:"frame" yaml:"frame"`
		Info   Info       `json:"info" yaml:"info"`
		Views  Views      `json:"views" yaml:"views"`
		Dialog Dialog     `json:"dialog" yaml:"dialog"`
		Emoji  EmojiStyle `json:"emoji" yaml:"emoji"`
	}

	EmojiStyle struct {
		Palette string `json:"palette" yaml:"palette"`
	}

	EmojiConfig struct {
		K9s     EmojiPalette `json:"k9s" yaml:"k9s"`
		NoIcons bool
	}

	EmojiPalette struct {
		System SystemEmoji `json:"system" yaml:"system"`
		Prompt PromptEmoji `json:"prompt" yaml:"prompt"`
		Status StatusEmoji `json:"status" yaml:"status"`
		Xray   XrayEmoji   `json:"xray" yaml:"xray"`
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
		FgColor     Color     `json:"fgColor" yaml:"fgColor"`
		FgStyle     TextStyle `json:"fgStyle" yaml:"fgStyle"`
		KeyColor    Color     `json:"keyColor" yaml:"keyColor"`
		NumKeyColor Color     `json:"numKeyColor" yaml:"numKeyColor"`
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

	SystemEmoji struct {
		LogStreamCancel string `json:"log_stream_cancelled" yaml:"log_stream_cancelled"`
		NewVersion      string `json:"new_version" yaml:"new_version"`
		Default         string `json:"default" yaml:"default"`
		LockedIC        string `json:"locked" yaml:"locked"`
		UnlockedIC      string `json:"unlocked" yaml:"unlocked"`
	}

	PromptEmoji struct {
		Query  string `json:"query" yaml:"query"`
		Filter string `json:"filter" yaml:"filter"`
	}

	StatusEmoji struct {
		Info  string `json:"info" yaml:"info"`
		Warn  string `json:"warn" yaml:"warn"`
		Error string `json:"error" yaml:"error"`
	}

	XrayEmoji struct {
		Namespaces               string `json:"namespaces" yaml:"namespaces"`
		DefaultGVR               string `json:"default_gvr" yaml:"default_gvr"`
		Nodes                    string `json:"nodes" yaml:"nodes"`
		Pods                     string `json:"pods" yaml:"pods"`
		Services                 string `json:"services" yaml:"services"`
		ServiceAccounts          string `json:"service_accounts" yaml:"service_accounts"`
		PersistentVolumes        string `json:"persistent_volumes" yaml:"persistent_volumes"`
		PersistentVolumeClaims   string `json:"persistent_volume_claims" yaml:"persistent_volume_claims"`
		Secrets                  string `json:"secrets" yaml:"secrets"`
		HorizontalPodAutoscalers string `json:"horizontal_pod_autoscalers" yaml:"horizontal_pod_autoscalers"`
		ConfigMaps               string `json:"config_maps" yaml:"config_maps"`
		Deployments              string `json:"deployments" yaml:"deployments"`
		StatefulSets             string `json:"stateful_sets" yaml:"stateful_sets"`
		DaemonSets               string `json:"daemon_sets" yaml:"daemon_sets"`
		ReplicaSets              string `json:"replica_sets" yaml:"replica_sets"`
		ClusterRoles             string `json:"cluster_roles" yaml:"cluster_roles"`
		Roles                    string `json:"roles" yaml:"roles"`
		NetworkPolicies          string `json:"network_policies" yaml:"network_policies"`
		PodDisruptionBudgets     string `json:"pod_disruption_budgets" yaml:"pod_disruption_budgets"`
		PodSecurityPolicies      string `json:"pod_security_policies" yaml:"pod_security_policies"`
		Containers               string `json:"containers" yaml:"containers"`
		Report                   string `json:"report" yaml:"report"`
		Issue0                   string `json:"issue_0" yaml:"issue_0"`
		Issue1                   string `json:"issue_1" yaml:"issue_1"`
		Issue2                   string `json:"issue_2" yaml:"issue_2"`
		Issue3                   string `json:"issue_3" yaml:"issue_3"`
	}
)

func newSkin() StyleConfig {
	return StyleConfig{
		K9s: Style{
			Body:   newBody(),
			Prompt: newPrompt(),
			Help:   newHelp(),
			Frame:  newFrame(),
			Info:   newInfo(),
			Views:  newViews(),
			Dialog: newDialog(),
		},
	}
}

func newEmoji() EmojiConfig {
	return EmojiConfig{
		K9s: EmojiPalette{
			System: newSystemEmoji(),
			Prompt: newPromptEmoji(),
			Status: newStatusEmoji(),
			Xray:   newXrayEmoji(),
		},
	}
}

func newSystemEmoji() SystemEmoji {
	return SystemEmoji{
		LogStreamCancel: "🏁",
		NewVersion:      "⚡️",
		Default:         "📎",
		LockedIC:        "🔒",
		UnlockedIC:      "✍️",
	}
}

func newPromptEmoji() PromptEmoji {
	return PromptEmoji{
		Query:  "🐶",
		Filter: "🐩",
	}
}

func newStatusEmoji() StatusEmoji {
	return StatusEmoji{
		Info:  "😎",
		Warn:  "😗",
		Error: "😡",
	}
}

func newXrayEmoji() XrayEmoji {
	return XrayEmoji{
		Namespaces:               "🗂",
		DefaultGVR:               "📎",
		Nodes:                    "🖥",
		Pods:                     "🚛",
		Services:                 "💁‍♀️",
		ServiceAccounts:          "💳",
		PersistentVolumes:        "📚",
		PersistentVolumeClaims:   "🎟",
		Secrets:                  "🔒",
		HorizontalPodAutoscalers: "♎️",
		ConfigMaps:               "🗺",
		Deployments:              "🪂",
		StatefulSets:             "🎎",
		DaemonSets:               "😈",
		ReplicaSets:              "👯‍♂️",
		ClusterRoles:             "👩‍",
		Roles:                    "👨🏻‍",
		NetworkPolicies:          "📕",
		PodDisruptionBudgets:     "🏷",
		PodSecurityPolicies:      "👮‍♂️",
		Containers:               "🐳",
		Report:                   "🧼",
		Issue0:                   "👍",
		Issue1:                   "🔊",
		Issue2:                   "☣️",
		Issue3:                   "🧨",
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
		DefaultDialColors:  Colors{"palegreen", "orangered"},
		DefaultChartColors: Colors{"palegreen", "orangered"},
		ResourceColors: map[string]Colors{
			"cpu": {"dodgerblue", "darkslateblue"},
			"mem": {"yellow", "goldenrod"},
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
func NewStyles(noIcons bool) *Styles {
	var s StyleConfig
	if err := yaml.Unmarshal(stockSkinTpl, &s); err != nil {
		s = newSkin()
	}

	var e EmojiConfig
	if err := yaml.Unmarshal(emojiTpl, &e); err != nil {
		e = newEmoji()
	}

	e.NoIcons = noIcons

	return &Styles{
		Skin:  s,
		Emoji: e,
	}
}

// Reset resets styles.
func (s *Styles) Reset() {
	if err := yaml.Unmarshal(stockSkinTpl, s); err != nil {
		s.Skin = newSkin()
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
	return s.Skin.K9s.Body
}

// Prompt returns prompt styles.
func (s *Styles) Prompt() Prompt {
	return s.Skin.K9s.Prompt
}

// Frame returns frame styles.
func (s *Styles) Frame() Frame {
	return s.Skin.K9s.Frame
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
	return s.Skin.K9s.Views.Charts
}

// Dialog returns dialog styles.
func (s *Styles) Dialog() Dialog {
	return s.Skin.K9s.Dialog
}

// Table returns table styles.
func (s *Styles) Table() Table {
	return s.Skin.K9s.Views.Table
}

// Xray returns xray styles.
func (s *Styles) Xray() Xray {
	return s.Skin.K9s.Views.Xray
}

// Views returns views styles.
func (s *Styles) Views() Views {
	return s.Skin.K9s.Views
}

// EmojiFor returns an emoji for the given key.
// Examples: "prompt.filter", "status.warn", "xray.pod".
func (e *EmojiConfig) EmojiFor(key string) string {
	if e.NoIcons {
		return ""
	}

	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return e.K9s.System.Default
	}

	category, subKey := parts[0], parts[1]
	switch category {
	case "system":
		switch subKey {
		case "log_stream_cancelled":
			return e.K9s.System.LogStreamCancel
		case "new_version":
			return e.K9s.System.NewVersion
		case "locked":
			return e.K9s.System.LockedIC
		case "unlocked":
			return e.K9s.System.UnlockedIC
		}
	case "prompt":
		switch subKey {
		case "query":
			return e.K9s.Prompt.Query
		case "filter":
			return e.K9s.Prompt.Filter
		}
	case "status":
		switch subKey {
		case "info":
			return e.K9s.Status.Info
		case "warn":
			return e.K9s.Status.Warn
		case "error":
			return e.K9s.Status.Error
		}
	case "xray":
		switch subKey {
		case "namespaces":
			return e.K9s.Xray.Namespaces
		case "default_gvr":
			return e.K9s.Xray.DefaultGVR
		case "nodes":
			return e.K9s.Xray.Nodes
		case "pods":
			return e.K9s.Xray.Pods
		case "services":
			return e.K9s.Xray.Services
		case "service_accounts":
			return e.K9s.Xray.ServiceAccounts
		case "persistent_volumes":
			return e.K9s.Xray.PersistentVolumes
		case "persistent_volume_claims":
			return e.K9s.Xray.PersistentVolumeClaims
		case "secrets":
			return e.K9s.Xray.Secrets
		case "horizontal_pod_autoscalers":
			return e.K9s.Xray.HorizontalPodAutoscalers
		case "config_maps":
			return e.K9s.Xray.ConfigMaps
		case "deployments":
			return e.K9s.Xray.Deployments
		case "stateful_sets":
			return e.K9s.Xray.StatefulSets
		case "daemon_sets":
			return e.K9s.Xray.DaemonSets
		case "replica_sets":
			return e.K9s.Xray.ReplicaSets
		case "cluster_roles":
			return e.K9s.Xray.ClusterRoles
		case "roles":
			return e.K9s.Xray.Roles
		case "network_policies":
			return e.K9s.Xray.NetworkPolicies
		case "pod_disruption_budgets":
			return e.K9s.Xray.PodDisruptionBudgets
		case "pod_security_policies":
			return e.K9s.Xray.PodSecurityPolicies
		case "containers":
			return e.K9s.Xray.Containers
		case "report":
			return e.K9s.Xray.Report
		case "issue_0":
			return e.K9s.Xray.Issue0
		case "issue_1":
			return e.K9s.Xray.Issue1
		case "issue_2":
			return e.K9s.Xray.Issue2
		case "issue_3":
			return e.K9s.Xray.Issue3
		}
	}

	return e.K9s.System.Default
}

// LoadSkin K9s configuration from file.
func (s *Styles) LoadSkin(path string) error {
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.SkinSchema, bb); err != nil {
		return err
	}
	if err := yaml.Unmarshal(bb, &s.Skin); err != nil {
		return err
	}

	return nil
}

// LoadEmoji emoji configuration from file.
func (s *Styles) LoadEmoji(path string) error {
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.EmojiSchema, bb); err != nil {
		return err
	}
	if err := yaml.Unmarshal(bb, &s.Emoji); err != nil {
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
	tview.Styles.BorderColor = s.Skin.K9s.Frame.Border.FgColor.Color()
	tview.Styles.FocusColor = s.Skin.K9s.Frame.Border.FocusColor.Color()
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
