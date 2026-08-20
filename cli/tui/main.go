package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

//go:generate cp ../../services.yml services.yml

//go:embed services.yml
var servicesYML []byte

// ── palette & styles ─────────────────────────────────────

var (
	accent    = lipgloss.Color("#7aa2f7")
	accentDim = lipgloss.Color("#bb9af7")
	muted     = lipgloss.Color("#565f89")
	textHi    = lipgloss.Color("#c0caf5")
	green     = lipgloss.Color("#9ece6a")
	red       = lipgloss.Color("#f7768e")
	yellow    = lipgloss.Color("#e0af68")
)

var (
	titleText = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	subText = lipgloss.NewStyle().
		Foreground(muted)

	statOk = lipgloss.NewStyle().
		Foreground(green).
		Bold(true)

	statFail = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	grpHdr = lipgloss.NewStyle().
		Foreground(accentDim).
		Bold(true)

	grpCnt = lipgloss.NewStyle().
		Foreground(muted)

	selIcon  = lipgloss.NewStyle().Foreground(yellow)
	normIcon = lipgloss.NewStyle().Foreground(accentDim)

	selName = lipgloss.NewStyle().
			Foreground(textHi).
			Bold(true)

	normName = lipgloss.NewStyle().
			Foreground(textHi)

	urlDim = lipgloss.NewStyle().
		Foreground(muted)

	okDot   = lipgloss.NewStyle().Foreground(green).Bold(true).Render("●")
	failDot = lipgloss.NewStyle().Foreground(red).Bold(true).Render("●")

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(muted)

	searchPrompt = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Render("/ ")

	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	iconPadWidth = 2

	settingLabel = lipgloss.NewStyle().
			Foreground(textHi).
			Bold(true)

	settingValue = lipgloss.NewStyle().
			Foreground(accent)

	settingSel = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)
)

// ── model ────────────────────────────────────────────────

type model struct {
	cfg      Config
	services []Service
	rows     []row
	cursor   int
	health   map[int]string
	tick     int
	checking bool
	width    int
	height   int

	searching   bool
	query       string
	filtered    []int
	searchInput textinput.Model

	groupFilter     string
	collapsedGroups map[string]bool

	settings Settings

	showSettings     bool
	settingsCursor   int
	settingsSnapshot Settings
	installStatus    string
	editingAlias     bool
	aliasInput       textinput.Model

	showHelp bool
}

// ── tea ──────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink, m.startHealth()}
	if m.settings.DateTime {
		cmds = append(cmds, tickCmd())
	}
	return tea.Batch(cmds...)
}

func (m model) startHealth() tea.Cmd {
	return func() tea.Msg {
		dlog("health check started")
		ch := make(chan healthResult, len(m.services))
		for i, svc := range m.services {
			go func(idx int, url string) {
				status := checkURL(url)
				dlog(fmt.Sprintf("  [%d] %s -> %s", idx, url, status))
				ch <- healthResult{index: idx, status: status}
			}(i, svc.URL)
		}
		for range m.services {
			r := <-ch
			sendTeaMsg(r)
		}
		dlog("health check complete")
		return healthDoneMsg{}
	}
}

type healthDoneMsg struct{}

func sendTeaMsg(msg tea.Msg) {
	// ponytail: global program reference, only way to send into tea from a goroutine
	p.Send(msg)
}

var p *tea.Program

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

// ── update ───────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.tick++
		if m.checking || m.settings.DateTime {
			cmds = append(cmds, tickCmd())
		}
		return m, tea.Batch(cmds...)

	case healthResult:
		m.health[msg.index] = msg.status
		return m, nil

	case healthDoneMsg:
		m.checking = false
		return m, nil
	}

	if m.showHelp {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "?", "esc", "q", "enter", "space", " ":
				m.showHelp = false
				return m, nil
			}
		}
		return m, nil
	}

	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searching = false
				m.query = ""
				m.searchInput.SetValue("")
				m.rebuildFilter()
				return m, nil
		case "enter":
			m.searching = false
			m.rebuildFilter()
			if len(m.filtered) > 0 {
				m.cursor = m.filtered[0]
				if svc := m.selectedService(); svc != nil {
					openBrowser(svc.URL)
				}
			}
			return m, nil
			}
		}
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
		if v := m.searchInput.Value(); v != m.query {
			m.query = v
			m.rebuildFilter()
			m.snapCursor()
		}
		return m, tea.Batch(cmds...)
	}

	if m.showSettings {
		if m.editingAlias {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "esc":
					m.editingAlias = false
					return m, nil
				case "enter":
					alias := strings.TrimSpace(m.aliasInput.Value())
					if alias == "" {
						alias = "hp"
					}
					m.settings.Alias = alias
					saveSettings(m.settings)
					msgStr, err := installCLI(alias)
					if err != nil {
						m.installStatus = lipgloss.NewStyle().Foreground(red).Render("✗ " + err.Error())
					} else {
						m.installStatus = lipgloss.NewStyle().Foreground(green).Render("✓ " + msgStr)
					}
					m.editingAlias = false
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.aliasInput, cmd = m.aliasInput.Update(msg)
			return m, cmd
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.settings = m.settingsSnapshot
				m.settings.applyTheme()
				m.showSettings = false
				m.installStatus = ""
				m.editingAlias = false
				return m, nil
			case ",":
				saveSettings(m.settings)
				m.showSettings = false
				m.installStatus = ""
				m.editingAlias = false
				return m, nil
			case "up", "k":
				m.settingsCursor--
				if m.settingsCursor < 0 {
					m.settingsCursor = len(settingItems) - 1
				}
			case "down", "j":
				m.settingsCursor++
				if m.settingsCursor >= len(settingItems) {
					m.settingsCursor = 0
				}
			case "enter", "space", " ":
				item := settingItems[m.settingsCursor]
				if item.key == "install" {
					m.editingAlias = true
					m.aliasInput.SetValue(m.settings.Alias)
					m.aliasInput.Focus()
					m.installStatus = ""
					return m, textinput.Blink
				}
				saveSettings(m.settings)
				m.showSettings = false
				m.installStatus = ""
				return m, nil
			case "left", "right":
				item := settingItems[m.settingsCursor]
				switch item.key {
				case "theme":
					if msg.String() == "left" {
						m.settings.Theme--
						if m.settings.Theme < 0 {
							m.settings.Theme = len(colorThemes) - 1
						}
					} else {
						m.settings.Theme++
						if m.settings.Theme >= len(colorThemes) {
							m.settings.Theme = 0
						}
					}
				case "borders":
					if msg.String() == "left" {
						m.settings.BorderStyle--
						if m.settings.BorderStyle < 0 {
							m.settings.BorderStyle = len(borderSets) - 1
						}
					} else {
						m.settings.BorderStyle++
						if m.settings.BorderStyle >= len(borderSets) {
							m.settings.BorderStyle = 0
						}
					}
				case "icons":
					if msg.String() == "left" {
						m.settings.IconStyle--
						if m.settings.IconStyle < 0 {
							m.settings.IconStyle = 2
						}
					} else {
						m.settings.IconStyle++
						if m.settings.IconStyle > 2 {
							m.settings.IconStyle = 0
						}
					}
				case "datetime":
					m.settings.DateTime = !m.settings.DateTime
					if m.settings.DateTime {
						cmds = append(cmds, tickCmd())
					}
				case "compact":
					m.settings.Compact = !m.settings.Compact
				case "install":
					m.editingAlias = true
					m.aliasInput.SetValue(m.settings.Alias)
					m.aliasInput.Focus()
					m.installStatus = ""
					return m, textinput.Blink
				}
				m.settings.applyTheme()
			}
		}
		return m, tea.Batch(cmds...)
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		s := msg.String()
		switch s {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.query != "" {
				m.query = ""
				m.rebuildFilter()
			}
		case "?":
			m.showHelp = true
		case "g":
			m.cycleGroup()
		case "G":
			m.cursor = m.lastVisible()
		case "tab", " ", "space":
			m.toggleCurrentGroup()
		case "up", "k":
			m.moveUp()
		case "down", "j":
			m.moveDown()
		case "/":
			m.searching = true
			m.searchInput.SetValue(m.query)
			m.searchInput.Focus()
			return m, textinput.Blink
		case "h":
			if !m.checking {
				m.checking = true
				m.health = make(map[int]string)
				return m, tea.Batch(m.startHealth(), tickCmd())
			}
		case "enter", "o":
			if m.cursor < len(m.rows) {
				if m.rows[m.cursor].kind == rowGroup {
					m.toggleCurrentGroup()
				} else {
					if svc := m.selectedService(); svc != nil {
						openBrowser(svc.URL)
					}
				}
			}
		case ",":
			m.settingsSnapshot = m.settings
			m.showSettings = true
			m.settingsCursor = 0
		default:
			if len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
				m.jumpToNumber(int(s[0] - '0'))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// ── filter / navigation ──────────────────────────────────

func (m *model) jumpToNumber(num int) {
	count := 0
	for i, r := range m.rows {
		if r.kind == rowService && m.isRowVisible(i) {
			count++
			if count == num {
				m.cursor = i
				return
			}
		}
	}
}

func (m *model) toggleCurrentGroup() {
	if m.cursor >= len(m.rows) {
		return
	}
	r := m.rows[m.cursor]
	gid := ""
	groupRowIdx := -1
	if r.kind == rowGroup {
		gid = r.groupID
		groupRowIdx = m.cursor
	} else {
		svc := m.services[r.serviceIdx]
		if len(svc.Groups) > 0 {
			gid = svc.Groups[0]
		} else {
			gid = "Other"
		}
		for i, gr := range m.rows {
			if gr.kind == rowGroup && gr.groupID == gid {
				groupRowIdx = i
				break
			}
		}
	}
	if gid == "" {
		return
	}
	if m.collapsedGroups == nil {
		m.collapsedGroups = make(map[string]bool)
	}
	m.collapsedGroups[gid] = !m.collapsedGroups[gid]
	if m.collapsedGroups[gid] && groupRowIdx >= 0 {
		m.cursor = groupRowIdx
	} else {
		m.snapCursor()
	}
}

func (m *model) cycleGroup() {
	if len(m.cfg.Groups) == 0 {
		return
	}
	if m.groupFilter == "" {
		m.groupFilter = m.cfg.Groups[0].ID
	} else {
		found := false
		for _, g := range m.cfg.Groups {
			if found {
				m.groupFilter = g.ID
				return
			}
			if g.ID == m.groupFilter {
				found = true
			}
		}
		m.groupFilter = ""
	}
	m.snapCursor()
}

func (m *model) rebuildFilter() {
	m.filtered = nil
	for i, r := range m.rows {
		if r.kind == rowGroup {
			continue
		}
		if searchMatchAny(m.query, m.services[r.serviceIdx]) {
			m.filtered = append(m.filtered, i)
		}
	}
}

func (m *model) snapCursor() {
	if m.cursor < len(m.rows) && m.isRowVisible(m.cursor) {
		return
	}
	for i := m.cursor; i < len(m.rows); i++ {
		if m.isRowVisible(i) {
			m.cursor = i
			return
		}
	}
	for i := m.cursor - 1; i >= 0; i-- {
		if m.isRowVisible(i) {
			m.cursor = i
			return
		}
	}
	m.cursor = 0
}

func (m model) visibleRows() []int {
	if m.query == "" {
		var vis []int
		for i := range m.rows {
			if m.isRowVisible(i) {
				vis = append(vis, i)
			}
		}
		return vis
	}
	return m.filtered
}

func (m *model) firstVisible() int {
	for i := range m.rows {
		if m.isRowVisible(i) {
			return i
		}
	}
	return 0
}

func (m *model) lastVisible() int {
	for i := len(m.rows) - 1; i >= 0; i-- {
		if m.isRowVisible(i) {
			return i
		}
	}
	return len(m.rows) - 1
}

func (m *model) moveUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if m.isRowVisible(i) && (m.rows[i].kind == rowService || (m.rows[i].kind == rowGroup && m.collapsedGroups[m.rows[i].groupID])) {
			m.cursor = i
			return
		}
	}
	for i := len(m.rows) - 1; i > m.cursor; i-- {
		if m.isRowVisible(i) && (m.rows[i].kind == rowService || (m.rows[i].kind == rowGroup && m.collapsedGroups[m.rows[i].groupID])) {
			m.cursor = i
			return
		}
	}
}

func (m *model) moveDown() {
	for i := m.cursor + 1; i < len(m.rows); i++ {
		if m.isRowVisible(i) && (m.rows[i].kind == rowService || (m.rows[i].kind == rowGroup && m.collapsedGroups[m.rows[i].groupID])) {
			m.cursor = i
			return
		}
	}
	for i := 0; i < m.cursor; i++ {
		if m.isRowVisible(i) && (m.rows[i].kind == rowService || (m.rows[i].kind == rowGroup && m.collapsedGroups[m.rows[i].groupID])) {
			m.cursor = i
			return
		}
	}
}

func (m model) selectedService() *Service {
	if m.cursor < len(m.rows) && m.rows[m.cursor].kind == rowService {
		svc := m.services[m.rows[m.cursor].serviceIdx]
		return &svc
	}
	return nil
}

func (m model) isRowVisible(i int) bool {
	r := m.rows[i]
	if m.groupFilter != "" {
		if r.kind == rowGroup {
			return false
		}
		svc := m.services[r.serviceIdx]
		match := false
		for _, g := range svc.Groups {
			if g == m.groupFilter {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if m.query != "" {
		if r.kind == rowGroup {
			for _, fi := range m.filtered {
				if m.rows[fi].kind == rowService {
					svc := m.services[m.rows[fi].serviceIdx]
					for _, g := range svc.Groups {
						if g == r.groupID {
							return true
						}
					}
				}
			}
			return false
		}
		for _, fi := range m.filtered {
			if fi == i {
				return true
			}
		}
		return false
	}
	if r.kind == rowService {
		gid := "Other"
		svc := m.services[r.serviceIdx]
		if len(svc.Groups) > 0 {
			gid = svc.Groups[0]
		}
		if m.collapsedGroups != nil && m.collapsedGroups[gid] {
			return false
		}
	}
	return true
}

// ── icon helper ──────────────────────────────────────────

func (m model) getServiceIcon(svc Service) string {
	switch m.settings.IconStyle {
	case 2: // Off
		return ""
	case 1: // ASCII
		return "[•]"
	default: // 0 = Emoji
		if svc.Icon != "" {
			return svc.Icon
		}
		return "⚙"
	}
}

// ── view ─────────────────────────────────────────────────

func (m model) countHealth() (ok, fail int) {
	for _, s := range m.health {
		if s == "✓" {
			ok++
		} else {
			fail++
		}
	}
	return
}

func (m model) groupServiceCount(gid string) int {
	count := 0
	for _, r := range m.rows {
		if r.kind == rowGroup {
			continue
		}
		svc := m.services[r.serviceIdx]
		g := "Other"
		if len(svc.Groups) > 0 {
			g = svc.Groups[0]
		}
		if g == gid {
			count++
		}
	}
	return count
}

func (m model) renderHeader() string {
	t := colorThemes[m.settings.Theme]
	bdr := borderSets[m.settings.BorderStyle]
	bdrStyle := lipgloss.NewStyle().Foreground(t.Border)

	title := m.cfg.Header
	titleLine := titleText.Render(fmt.Sprintf("  %s ", title))

	var rightTop string
	if m.settings.DateTime {
		now := time.Now()
		rightTop = subText.Render(now.Format("Mon Jan 2  15:04:05"))
	}

	subLine := subText.Render(fmt.Sprintf("  %d services", len(m.services)))
	if m.groupFilter != "" {
		for _, g := range m.cfg.Groups {
			if g.ID == m.groupFilter {
				subLine += lipgloss.NewStyle().Foreground(t.AccentDim).Render(fmt.Sprintf("  [%s]", g.Name))
				break
			}
		}
	}
	if m.checking {
		sp := spinnerFrames[m.tick%len(spinnerFrames)]
		subLine += "  " + lipgloss.NewStyle().Foreground(t.SelectedIcon).Render(sp)
	} else if len(m.health) > 0 {
		ok, fail := m.countHealth()
		subLine += "  " + statOk.Render(fmt.Sprintf("%d✓", ok))
		if fail > 0 {
			subLine += " " + statFail.Render(fmt.Sprintf("%d✗", fail))
		}
	}

	w := m.width - 2
	bar := strings.Repeat(bdr.Horizontal, w)

	titleW := realStringWidth(titleLine)
	rightTopW := realStringWidth(rightTop)
	gap := w - titleW - rightTopW
	if gap < 1 {
		gap = 1
	}
	mid1 := titleLine + strings.Repeat(" ", gap) + rightTop

	subW := realStringWidth(subLine)
	padRight := w - subW
	if padRight < 0 {
		padRight = 0
	}
	mid2 := subLine + strings.Repeat(" ", padRight)

	top := bdrStyle.Render(bdr.TopLeft + bar + bdr.TopRight)

	padMid := func(s string) string {
		sw := realStringWidth(s)
		if sw < w {
			return s + strings.Repeat(" ", w-sw)
		}
		if sw > w {
			s = ansi.Truncate(s, w, "")
			sw = realStringWidth(s)
			if sw < w {
				s += strings.Repeat(" ", w-sw)
			}
			return s
		}
		return s
	}

	m1 := bdrStyle.Render(bdr.Vertical) + padMid(mid1) + bdrStyle.Render(bdr.Vertical)
	m2 := bdrStyle.Render(bdr.Vertical) + padMid(mid2) + bdrStyle.Render(bdr.Vertical)
	bot := bdrStyle.Render(bdr.BottomLeft + bar + bdr.BottomRight)

	return top + "\n" + m1 + "\n" + m2 + "\n" + bot
}

func (m model) renderDetail() string {
	if m.cursor >= len(m.rows) {
		return ""
	}
	t := colorThemes[m.settings.Theme]
	bdr := borderSets[m.settings.BorderStyle]
	bdrStyle := lipgloss.NewStyle().Foreground(t.Border)

	var line1, line2, line3 string

	if m.rows[m.cursor].kind == rowGroup {
		r := m.rows[m.cursor]
		icon := r.groupIcon
		if icon == "" {
			icon = "▸"
		}
		isCollapsed := m.collapsedGroups != nil && m.collapsedGroups[r.groupID]
		state := "expanded"
		if isCollapsed {
			state = "folded"
		}
		cnt := m.groupServiceCount(r.groupID)
		var svcNames []string
		for _, s := range m.services {
			for _, g := range s.Groups {
				if g == r.groupID {
					svcNames = append(svcNames, s.Name)
					break
				}
			}
		}

		line1 = fmt.Sprintf("  %s  %s  %s",
			icon,
			lipgloss.NewStyle().Bold(true).Foreground(t.SelectedFg).Render(r.groupName),
			lipgloss.NewStyle().Foreground(t.AccentDim).Render(fmt.Sprintf("(%d services, %s)", cnt, state)),
		)
		line2 = fmt.Sprintf("  %s", lipgloss.NewStyle().Foreground(t.Accent).Render("press Enter / Space / Tab to toggle fold"))
		line3 = fmt.Sprintf("  %s%s",
			lipgloss.NewStyle().Foreground(t.Muted).Render("services: "),
			lipgloss.NewStyle().Foreground(t.AccentDim).Render(strings.Join(svcNames, ", ")),
		)
	} else {
		svc := m.services[m.rows[m.cursor].serviceIdx]
		icon := m.getServiceIcon(svc)

		healthStr := ""
		if s, ok := m.health[m.rows[m.cursor].serviceIdx]; ok {
			if s == "✓" {
				healthStr = okDot + lipgloss.NewStyle().Foreground(green).Render(" online")
			} else {
				healthStr = failDot + lipgloss.NewStyle().Foreground(red).Render(" offline")
			}
		}

		iconStr := ""
		if icon != "" {
			iconStr = icon + "  "
		}

		line1 = fmt.Sprintf("  %s%s  %s",
			iconStr,
			lipgloss.NewStyle().Bold(true).Foreground(t.SelectedFg).Render(svc.Name),
			healthStr,
		)
		line2 = fmt.Sprintf("  %s", lipgloss.NewStyle().Foreground(t.Accent).Render(svc.URL))
		line3 = fmt.Sprintf("  %s%s",
			lipgloss.NewStyle().Foreground(t.Muted).Render("groups: "),
			lipgloss.NewStyle().Foreground(t.AccentDim).Render(strings.Join(svc.Groups, ", ")),
		)
	}

	w := m.width - 2
	bar := strings.Repeat(bdr.Horizontal, w)

	padTo := func(s string) string {
		sw := realStringWidth(s)
		if sw < w {
			return s + strings.Repeat(" ", w-sw)
		}
		if sw > w {
			s = ansi.Truncate(s, w-1, "") + "…"
			sw = realStringWidth(s)
			if sw < w {
				s += strings.Repeat(" ", w-sw)
			}
			return s
		}
		return s
	}

	left := bdrStyle.Render(bdr.Vertical)
	right := bdrStyle.Render(bdr.Vertical)
	top := bdrStyle.Render(bdr.TopLeft + bar + bdr.TopRight)
	bot := bdrStyle.Render(bdr.BottomLeft + bar + bdr.BottomRight)

	return "\n" + top + "\n" +
		left + padTo(line1) + right + "\n" +
		left + padTo(line2) + right + "\n" +
		left + padTo(line3) + right + "\n" +
		bot
}

func (m model) renderSettings() string {
	t := colorThemes[m.settings.Theme]
	bdr := borderSets[m.settings.BorderStyle]

	var lines []string
	title := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("  Settings")
	lines = append(lines, title, "")

	iconModes := []string{"Emoji", "ASCII", "Off"}

	maxLabelLen := 0
	for _, item := range settingItems {
		if len(item.label) > maxLabelLen {
			maxLabelLen = len(item.label)
		}
	}

	for i, item := range settingItems {
		paddedLabel := item.label + strings.Repeat(" ", maxLabelLen-len(item.label))
		label := settingLabel.Render(paddedLabel)
		var val string
		switch item.key {
		case "theme":
			val = settingValue.Render(colorThemes[m.settings.Theme].Name)
		case "borders":
			val = settingValue.Render(borderSets[m.settings.BorderStyle].Name)
		case "icons":
			val = settingValue.Render(iconModes[m.settings.IconStyle])
		case "datetime":
			if m.settings.DateTime {
				val = settingValue.Render("On")
			} else {
				val = settingValue.Render("Off")
			}
		case "compact":
			if m.settings.Compact {
				val = settingValue.Render("On")
			} else {
				val = settingValue.Render("Off")
			}
		case "install":
			val = settingValue.Render("→ [" + m.settings.Alias + "]")
		}

		cursor := "  "
		if i == m.settingsCursor {
			cursor = settingSel.Render("▸ ")
		}
		lines = append(lines, cursor+label+"   "+val)
	}

	if m.editingAlias {
		lines = append(lines, "")
		lines = append(lines, subText.Render("  Enter CLI alias name:"))
		lines = append(lines, "  "+m.aliasInput.View())
		lines = append(lines, "")
		lines = append(lines, subText.Render("  enter install • esc cancel"))
	} else {
		if m.installStatus != "" {
			lines = append(lines, "", "  "+m.installStatus)
		}
		lines = append(lines, "")
		lines = append(lines, subText.Render("  ↑↓/←→ edit • enter save • esc cancel"))
	}

	panel := strings.Join(lines, "\n")
	w := 44
	return lipgloss.NewStyle().
		Border(bdr.Lipgloss).
		BorderForeground(t.Border).
		Padding(0, 1).
		Width(w).
		Render(panel)
}

func (m model) renderHelpModal() string {
	t := colorThemes[m.settings.Theme]
	bdr := borderSets[m.settings.BorderStyle]

	var lines []string
	title := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("  Keyboard Shortcuts")
	lines = append(lines, title, "")

	shortcuts := []struct{ key, desc string }{
		{"↑↓ / j k", "Navigate services"},
		{"1 - 9", "Quick jump to service"},
		{"Enter / o", "Open in browser"},
		{"Tab / space", "Toggle group fold"},
		{"g / G", "Cycle group / Jump bottom"},
		{"/", "Search / filter"},
		{"h", "Re-check health"},
		{",", "Settings panel"},
		{"q", "Quit"},
	}

	for _, sc := range shortcuts {
		k := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(fmt.Sprintf("  %-13s", sc.key))
		d := lipgloss.NewStyle().Foreground(t.SelectedFg).Render(sc.desc)
		lines = append(lines, k+" "+d)
	}

	lines = append(lines, "")
	lines = append(lines, subText.Render("  press ? or esc to close"))

	panel := strings.Join(lines, "\n")
	w := 44
	return lipgloss.NewStyle().
		Border(bdr.Lipgloss).
		BorderForeground(t.Border).
		Padding(0, 1).
		Width(w).
		Render(panel)
}

func (m model) renderHelp() string {
	items := []struct{ key, desc string }{
		{"↑↓", "move"},
		{"/", "search"},
		{",", "settings"},
		{"?", "help"},
	}
	var parts []string
	for _, it := range items {
		parts = append(parts, helpKeyStyle.Render(it.key)+" "+helpDescStyle.Render(it.desc))
	}
	return strings.Join(parts, "   ")
}

func (m model) renderServiceLine(idx int, r row, selected bool, maxName, maxURL int) string {
	svc := m.services[r.serviceIdx]
	t := colorThemes[m.settings.Theme]
	icon := m.getServiceIcon(svc)

	status := ""
	if s, ok := m.health[r.serviceIdx]; ok {
		if s == "✓" {
			status = " " + okDot
		} else {
			status = " " + failDot
		}
	} else if m.checking {
		sp := spinnerFrames[m.tick%len(spinnerFrames)]
		status = " " + lipgloss.NewStyle().Foreground(t.SelectedIcon).Render(sp)
	}

	prefix := "    "
	iconSt := normIcon
	nameSt := normName
	if selected {
		prefix = lipgloss.NewStyle().Foreground(t.Accent).Render("  ▸ ")
		iconSt = selIcon
		nameSt = selName
	}
	if m.settings.Compact {
		if selected {
			prefix = lipgloss.NewStyle().Foreground(t.Accent).Render(" ▸ ")
		} else {
			prefix = "   "
		}
	}

	nameW := realStringWidth(svc.Name)
	paddedName := svc.Name + strings.Repeat(" ", maxName-nameW)

	iconPart := ""
	if m.settings.IconStyle != 2 { // not Off
		if icon != "" {
			iconPart = iconSt.Render(padIcon(icon) + " ")
		}
	}

	if m.settings.Compact {
		return prefix + iconPart + nameSt.Render(paddedName)
	}

	paddedURL := svc.URL + strings.Repeat(" ", maxURL-len(svc.URL))
	return prefix + iconPart + nameSt.Render(paddedName) + urlDim.Render("  "+paddedURL) + status
}

func overlayLine(baseLine, overlayLine string, startX, overlayW int) string {
	if startX < 0 {
		startX = 0
	}

	left := ansi.Truncate(baseLine, startX, "")
	leftW := ansi.StringWidth(left)
	if leftW < startX {
		left += strings.Repeat(" ", startX-leftW)
	}

	right := ""
	baseW := ansi.StringWidth(baseLine)
	if baseW > startX+overlayW {
		right = ansi.TruncateLeft(baseLine, startX+overlayW, "")
	}

	return left + overlayLine + right
}

func (m model) overlayModal(baseView, modalView string) string {
	baseLines := strings.Split(baseView, "\n")
	overlayLines := strings.Split(modalView, "\n")

	overlayH := len(overlayLines)
	overlayW := 0
	for _, l := range overlayLines {
		if w := realStringWidth(l); w > overlayW {
			overlayW = w
		}
	}

	startY := (m.height - overlayH) / 2
	startX := (m.width - overlayW) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	for len(baseLines) < startY+overlayH {
		baseLines = append(baseLines, "")
	}

	for i, ol := range overlayLines {
		y := startY + i
		if y < len(baseLines) {
			baseLines[y] = overlayLine(baseLines[y], ol, startX, overlayW)
		}
	}

	return strings.Join(baseLines, "\n")
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()

	search := ""
	if m.searching {
		search = searchPrompt + m.searchInput.View()
	}

	detail := m.renderDetail()
	help := m.renderHelp()

	fixedH := 4 + 1 // header (top+mid1+mid2+bot) + newline
	if search != "" {
		fixedH += 1 + 1
	}
	detailH := lipgloss.Height(detail)
	fixedH += detailH
	if detailH > 0 {
		fixedH++
	}

	avail := m.height - fixedH

	if avail < 1 {
		avail = 1
	}

	var vis []int
	for i := range m.rows {
		if m.isRowVisible(i) {
			vis = append(vis, i)
		}
	}
	totalVis := len(vis)

	cursorPos := 0
	for pi, vi := range vis {
		if vi == m.cursor {
			cursorPos = pi
			break
		}
	}

	scrollStart := 0
	if totalVis > avail {
		if cursorPos >= avail-1 {
			scrollStart = cursorPos - avail + 2
		}
		if scrollStart+avail > totalVis {
			scrollStart = totalVis - avail
		}
	}
	scrollEnd := scrollStart + avail
	if scrollEnd > totalVis {
		scrollEnd = totalVis
	}

	dlog(fmt.Sprintf("View: termH=%d fixedH=%d detailH=%d avail=%d vis=%d/%d scroll=%d-%d cursor=%d",
		m.height, fixedH, detailH, avail, totalVis, len(m.rows), scrollStart, scrollEnd, cursorPos))

	maxName, maxURL := 0, 0
	for _, svc := range m.services {
		if w := realStringWidth(svc.Name); w > maxName {
			maxName = w
		}
		if !m.settings.Compact {
			if w := len(svc.URL); w > maxURL {
				maxURL = w
			}
		}
	}

	var b strings.Builder

	b.WriteString(header)
	b.WriteString("\n")

	if search != "" {
		b.WriteString(search)
		b.WriteString("\n")
	}

	for pi := scrollStart; pi < scrollEnd; pi++ {
		i := vis[pi]
		r := m.rows[i]

		if r.kind == rowGroup {
			t := colorThemes[m.settings.Theme]
			icon := r.groupIcon
			if icon == "" {
				icon = "▸"
			}
			isCollapsed := m.collapsedGroups != nil && m.collapsedGroups[r.groupID]
			foldSymbol := "▾"
			if isCollapsed {
				foldSymbol = "▸"
			}
			cnt := m.groupServiceCount(r.groupID)
			suffix := ""
			if isCollapsed {
				suffix = " " + subText.Render("[folded]")
			}

			prefix := "  "
			headerStyle := grpHdr
			if i == m.cursor {
				prefix = lipgloss.NewStyle().Foreground(t.Accent).Render("▸ ")
				headerStyle = lipgloss.NewStyle().Foreground(t.SelectedFg).Bold(true)
			}

			b.WriteString(prefix + headerStyle.Render(fmt.Sprintf("%s %s %s:", foldSymbol, icon, r.groupName)) +
				" " + grpCnt.Render(fmt.Sprintf("(%d)", cnt)) + suffix)
			b.WriteString("\n")
			continue
		}

		b.WriteString(m.renderServiceLine(i, r, i == m.cursor, maxName, maxURL))
		b.WriteString("\n")
	}

	if detailH > 0 {
		b.WriteString(detail)
		b.WriteString("\n")
	}
	b.WriteString(help)

	mainView := b.String()
	if m.showSettings {
		return m.overlayModal(mainView, m.renderSettings())
	}
	if m.showHelp {
		return m.overlayModal(mainView, m.renderHelpModal())
	}
	return mainView
}

// ── main ─────────────────────────────────────────────────

func main() {
	os.Remove(debugLog)
	dlog("tui started")

	var cfgPath string
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading services.yml: %v\n", err)
		os.Exit(1)
	}

	si := textinput.New()
	si.Placeholder = "filter services..."
	si.Prompt = ""
	si.CharLimit = 64

	ai := textinput.New()
	ai.Placeholder = "hp"
	ai.Prompt = "Alias: "
	ai.CharLimit = 32

	settings := loadSettings()
	settings.applyTheme()

	m := model{
		cfg:         cfg,
		services:    cfg.Services,
		rows:        buildRows(cfg, cfg.Services),
		health:      make(map[int]string),
		checking:    true,
		searchInput: si,
		aliasInput:  ai,
		settings:    settings,
	}
	for i := range m.rows {
		if m.rows[i].kind == rowService {
			m.cursor = i
			break
		}
	}

	p = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
