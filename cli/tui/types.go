package main

import "github.com/charmbracelet/lipgloss"

type Service struct {
	Name   string   `yaml:"name"`
	URL    string   `yaml:"url"`
	Icon   string   `yaml:"icon"`
	Groups []string `yaml:"groups"`
}

type Group struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
}

type Config struct {
	Title    string    `yaml:"title"`
	Header   string    `yaml:"header"`
	Groups   []Group   `yaml:"groups"`
	Services []Service `yaml:"services"`
}

type ColorTheme struct {
	Name         string
	Accent       lipgloss.Color
	AccentDim    lipgloss.Color
	Border       lipgloss.Color
	SelectedFg   lipgloss.Color
	SelectedIcon lipgloss.Color
	Muted        lipgloss.Color
	Ok           lipgloss.Color
	Fail         lipgloss.Color
}

type BorderSet struct {
	Name        string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	Lipgloss    lipgloss.Border
}

type Settings struct {
	Theme       int    `json:"theme"`
	BorderStyle int    `json:"borderStyle"`
	IconStyle   int    `json:"iconStyle"`
	DateTime    bool   `json:"datetime"`
	Compact     bool   `json:"compact"`
	Alias       string `json:"alias"`
	ShowIcons   *bool  `json:"showIcons,omitempty"`
	GroupIcons  bool   `json:"groupIcons"`
}

type healthResult struct {
	index  int
	status string
}

type rowKind int

const (
	rowGroup rowKind = iota
	rowService
)

type row struct {
	kind       rowKind
	groupID    string
	groupName  string
	groupIcon  string
	serviceIdx int
}

var settingItems = []struct{ label, key string }{
	{"Theme", "theme"},
	{"Borders", "borders"},
	{"Icons", "icons"},
	{"Group Icons", "grpicons"},
	{"DateTime", "datetime"},
	{"Compact", "compact"},
	{"Install CLI", "install"},
}
