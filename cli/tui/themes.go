package main

import "github.com/charmbracelet/lipgloss"

var colorThemes = []ColorTheme{
	{
		Name:         "Tokyo Night",
		Accent:       lipgloss.Color("#7aa2f7"),
		AccentDim:    lipgloss.Color("#bb9af7"),
		Border:       lipgloss.Color("#3b4261"),
		SelectedFg:   lipgloss.Color("#c0caf5"),
		SelectedIcon: lipgloss.Color("#e0af68"),
		Muted:        lipgloss.Color("#565f89"),
		Ok:           lipgloss.Color("#9ece6a"),
		Fail:         lipgloss.Color("#f7768e"),
	},
	{
		Name:         "Catppuccin",
		Accent:       lipgloss.Color("#cba6f7"),
		AccentDim:    lipgloss.Color("#b4befe"),
		Border:       lipgloss.Color("#585b70"),
		SelectedFg:   lipgloss.Color("#cdd6f4"),
		SelectedIcon: lipgloss.Color("#f9e2af"),
		Muted:        lipgloss.Color("#6c7086"),
		Ok:           lipgloss.Color("#a6e3a1"),
		Fail:         lipgloss.Color("#f38ba8"),
	},
	{
		Name:         "Nord",
		Accent:       lipgloss.Color("#88c0d0"),
		AccentDim:    lipgloss.Color("#81a1c1"),
		Border:       lipgloss.Color("#4c566a"),
		SelectedFg:   lipgloss.Color("#eceff4"),
		SelectedIcon: lipgloss.Color("#ebcb8b"),
		Muted:        lipgloss.Color("#616e88"),
		Ok:           lipgloss.Color("#a3be8c"),
		Fail:         lipgloss.Color("#bf616a"),
	},
	{
		Name:         "Dracula",
		Accent:       lipgloss.Color("#bd93f9"),
		AccentDim:    lipgloss.Color("#ff79c6"),
		Border:       lipgloss.Color("#6272a4"),
		SelectedFg:   lipgloss.Color("#f8f8f2"),
		SelectedIcon: lipgloss.Color("#f1fa8c"),
		Muted:        lipgloss.Color("#6c7393"),
		Ok:           lipgloss.Color("#50fa7b"),
		Fail:         lipgloss.Color("#ff5555"),
	},
	{
		Name:         "Gruvbox",
		Accent:       lipgloss.Color("#fe8019"),
		AccentDim:    lipgloss.Color("#fabd2f"),
		Border:       lipgloss.Color("#665c54"),
		SelectedFg:   lipgloss.Color("#ebdbb2"),
		SelectedIcon: lipgloss.Color("#d79921"),
		Muted:        lipgloss.Color("#7c6f64"),
		Ok:           lipgloss.Color("#b8bb26"),
		Fail:         lipgloss.Color("#fb4934"),
	},
	{
		Name:         "Cyberpunk",
		Accent:       lipgloss.Color("#00f0ff"),
		AccentDim:    lipgloss.Color("#ff0055"),
		Border:       lipgloss.Color("#711c91"),
		SelectedFg:   lipgloss.Color("#ffffff"),
		SelectedIcon: lipgloss.Color("#ffe600"),
		Muted:        lipgloss.Color("#805090"),
		Ok:           lipgloss.Color("#00ff66"),
		Fail:         lipgloss.Color("#ff003c"),
	},
	{
		Name:         "Monochrome",
		Accent:       lipgloss.Color("#ffffff"),
		AccentDim:    lipgloss.Color("#cccccc"),
		Border:       lipgloss.Color("#444444"),
		SelectedFg:   lipgloss.Color("#ffffff"),
		SelectedIcon: lipgloss.Color("#ffffff"),
		Muted:        lipgloss.Color("#777777"),
		Ok:           lipgloss.Color("#99ff99"),
		Fail:         lipgloss.Color("#ff9999"),
	},
}

var borderSets = []BorderSet{
	{
		Name:        "Double",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
		Lipgloss:    lipgloss.DoubleBorder(),
	},
	{
		Name:        "Rounded",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
		Lipgloss:    lipgloss.RoundedBorder(),
	},
	{
		Name:        "Single",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		Lipgloss:    lipgloss.NormalBorder(),
	},
	{
		Name:        "Thick",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
		Horizontal:  "━",
		Vertical:    "┃",
		Lipgloss:    lipgloss.ThickBorder(),
	},
}

func (s Settings) applyTheme() {
	if s.Theme < 0 || s.Theme >= len(colorThemes) {
		s.Theme = 0
	}
	t := colorThemes[s.Theme]
	accent = t.Accent
	accentDim = t.AccentDim
	muted = t.Muted
	textHi = t.SelectedFg
	green = t.Ok
	red = t.Fail
	yellow = t.SelectedIcon

	normIcon = lipgloss.NewStyle().Foreground(t.AccentDim)
	selIcon = lipgloss.NewStyle().Foreground(t.SelectedIcon)
	selName = lipgloss.NewStyle().Foreground(t.SelectedFg).Bold(true)
	normName = lipgloss.NewStyle().Foreground(t.SelectedFg)

	titleText = lipgloss.NewStyle().Foreground(accent).Bold(true)
	grpHdr = lipgloss.NewStyle().Foreground(accentDim).Bold(true)
	grpCnt = lipgloss.NewStyle().Foreground(muted)
	subText = lipgloss.NewStyle().Foreground(muted)
	urlDim = lipgloss.NewStyle().Foreground(muted)

	statOk = lipgloss.NewStyle().Foreground(green).Bold(true)
	statFail = lipgloss.NewStyle().Foreground(red).Bold(true)
	okDot = lipgloss.NewStyle().Foreground(green).Bold(true).Render("●")
	failDot = lipgloss.NewStyle().Foreground(red).Bold(true).Render("●")

	helpKeyStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	helpDescStyle = lipgloss.NewStyle().Foreground(muted)

	settingValue = lipgloss.NewStyle().Foreground(accent)
	settingLabel = lipgloss.NewStyle().Foreground(textHi).Bold(true)
	settingSel = lipgloss.NewStyle().Foreground(t.SelectedIcon).Bold(true)
	searchPrompt = lipgloss.NewStyle().Foreground(accent).Bold(true).Render("/ ")
}
