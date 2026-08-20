package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"gopkg.in/yaml.v3"
)

// ── types ────────────────────────────────────────────────

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
	Name      string
	Accent    lipgloss.Color
	AccentDim lipgloss.Color
	Icon      lipgloss.Color
}

var colorThemes = []ColorTheme{
	{"Indigo", lipgloss.Color("141"), lipgloss.Color("98"), lipgloss.Color("214")},
	{"Teal", lipgloss.Color("51"), lipgloss.Color("45"), lipgloss.Color("81")},
	{"Rose", lipgloss.Color("205"), lipgloss.Color("132"), lipgloss.Color("211")},
	{"Gold", lipgloss.Color("220"), lipgloss.Color("178"), lipgloss.Color("228")},
	{"Lime", lipgloss.Color("112"), lipgloss.Color("70"), lipgloss.Color("154")},
	{"Sky", lipgloss.Color("75"), lipgloss.Color("66"), lipgloss.Color("117")},
}

type Settings struct {
	Theme    int    `json:"theme"`
	Compact  bool   `json:"compact"`
	DateTime bool   `json:"datetime"`
	Alias    string `json:"alias"`
}

type healthResult struct {
	index  int
	status string
}

type healthBatch struct {
	results []healthResult
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

// ── palette ──────────────────────────────────────────────

var (
	accent    = lipgloss.Color("141")
	accentDim = lipgloss.Color("98")
	muted     = lipgloss.Color("242")
	subtle    = lipgloss.Color("245")
	textHi    = lipgloss.Color("230")
	green     = lipgloss.Color("42")
	red       = lipgloss.Color("196")
	yellow    = lipgloss.Color("214")
)

// ── styles ───────────────────────────────────────────────

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

	selIcon = lipgloss.NewStyle().Foreground(yellow)
	normIcon = lipgloss.NewStyle().Foreground(yellow)

	selName = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
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

	searching  bool
	query      string
	filtered   []int
	searchInput textinput.Model

	settings Settings

	showSettings      bool
	settingsCursor    int
	settingsSnapshot  Settings
	installStatus     string
	editingAlias      bool
	aliasInput        textinput.Model
}

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "homepage", "tui.json")
}

func loadSettings() Settings {
	s := Settings{Theme: 0, DateTime: true, Alias: "hp"}
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return s
	}
	json.Unmarshal(data, &s)
	if s.Theme < 0 || s.Theme >= len(colorThemes) {
		s.Theme = 0
	}
	if s.Alias == "" {
		s.Alias = "hp"
	}
	return s
}

func saveSettings(s Settings) {
	dir := filepath.Dir(settingsPath())
	os.MkdirAll(dir, 0755)
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(settingsPath(), data, 0644)
}

func (s Settings) applyTheme() {
	t := colorThemes[s.Theme]
	accent = t.Accent
	accentDim = t.AccentDim
	normIcon = lipgloss.NewStyle().Foreground(t.Icon)
	selIcon = lipgloss.NewStyle().Foreground(t.Icon)
	selName = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	normName = lipgloss.NewStyle().Foreground(lipgloss.Color("230"))
	textHi = lipgloss.Color("230")
	// rebuild styles that reference accent
	titleText = lipgloss.NewStyle().Foreground(accent).Bold(true)
	grpHdr = lipgloss.NewStyle().Foreground(accentDim).Bold(true)
	helpKeyStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	settingValue = lipgloss.NewStyle().Foreground(accent)
	settingLabel = lipgloss.NewStyle().Foreground(textHi).Bold(true)
	searchPrompt = lipgloss.NewStyle().Foreground(accent).Bold(true).Render("/ ")
}

//go:embed services.yml
var servicesYML []byte

func loadConfig() (Config, error) {
	var cfg Config
	return cfg, yaml.Unmarshal(servicesYML, &cfg)
}

func buildRows(cfg Config, services []Service) []row {
	groupName := make(map[string]string)
	groupIcon := make(map[string]string)
	for _, g := range cfg.Groups {
		groupName[g.ID] = g.Name
		groupIcon[g.ID] = g.Icon
	}

	grouped := make(map[string][]int)
	var order []string
	seen := make(map[string]bool)
	for i, svc := range services {
		g := "Other"
		if len(svc.Groups) > 0 {
			g = svc.Groups[0]
		}
		if !seen[g] {
			seen[g] = true
			order = append(order, g)
		}
		grouped[g] = append(grouped[g], i)
	}

	var rows []row
	for _, gid := range order {
		name := groupName[gid]
		if name == "" {
			name = gid
		}
		rows = append(rows, row{kind: rowGroup, groupID: gid, groupName: name, groupIcon: groupIcon[gid]})
		for _, si := range grouped[gid] {
			rows = append(rows, row{kind: rowService, serviceIdx: si})
		}
	}
	return rows
}

// ── icon padding ─────────────────────────────────────────

func realStringWidth(s string) int {
	clean := strings.ReplaceAll(s, "\ufe0f", "")
	clean = strings.ReplaceAll(clean, "\ufe0e", "")
	return ansi.StringWidth(clean)
}

func padIcon(icon string) string {
	sw := realStringWidth(icon)
	if sw < iconPadWidth {
		icon += strings.Repeat(" ", iconPadWidth-sw)
	}
	return icon
}

// ── search ───────────────────────────────────────────────

func searchMatch(query, target string) bool {
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(target), strings.ToLower(query))
}

func searchMatchAny(query string, svc Service) bool {
	return searchMatch(query, svc.Name) ||
		searchMatch(query, svc.URL) ||
		searchMatch(query, strings.Join(svc.Groups, " "))
}

// ── debug ────────────────────────────────────────────────

const debugLog = "/tmp/tui-debug.log"

func dlog(msg string) {
	f, err := os.OpenFile(debugLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println(msg)
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
		results := make([]healthResult, len(m.services))
		var wg sync.WaitGroup
		for i, svc := range m.services {
			wg.Add(1)
			go func(idx int, url string) {
				defer wg.Done()
				status := checkURL(url)
				dlog(fmt.Sprintf("  [%d] %s -> %s", idx, url, status))
				results[idx] = healthResult{index: idx, status: status}
			}(i, svc.URL)
		}
		wg.Wait()
		dlog("health check complete")
		return healthBatch{results: results}
	}
}

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

// ── update ───────────────────────────────────────────────

var settingItems = []struct{ label, key string }{
	{"Theme", "theme"},
	{"DateTime", "datetime"},
	{"Compact", "compact"},
	{"Install CLI", "install"},
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
			case "up":
				m.settingsCursor--
				if m.settingsCursor < 0 {
					m.settingsCursor = len(settingItems) - 1
				}
			case "down":
				m.settingsCursor++
				if m.settingsCursor >= len(settingItems) {
					m.settingsCursor = 0
				}
			case "enter", "space":
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m.moveUp()
		case "down":
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
			if svc := m.selectedService(); svc != nil {
				openBrowser(svc.URL)
			}
		case ",":
			m.settingsSnapshot = m.settings
			m.showSettings = true
			m.settingsCursor = 0
		}

	case healthBatch:
		for _, r := range msg.results {
			m.health[r.index] = r.status
		}
		m.checking = false
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// ── filter / navigation ──────────────────────────────────

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
	if len(m.filtered) == 0 {
		return
	}
	for _, fi := range m.filtered {
		if fi >= m.cursor && m.rows[fi].kind == rowService {
			m.cursor = fi
			return
		}
	}
	for i := len(m.filtered) - 1; i >= 0; i-- {
		if m.rows[m.filtered[i]].kind == rowService {
			m.cursor = m.filtered[i]
			return
		}
	}
}

func (m model) visibleRows() []int {
	if m.query == "" {
		var vis []int
		for i := range m.rows {
			vis = append(vis, i)
		}
		return vis
	}
	return m.filtered
}

func (m *model) firstVisible() int {
	for i := range m.rows {
		if m.rows[i].kind == rowService && m.isRowVisible(i) {
			return i
		}
	}
	return 0
}

func (m *model) lastVisible() int {
	for i := len(m.rows) - 1; i >= 0; i-- {
		if m.rows[i].kind == rowService && m.isRowVisible(i) {
			return i
		}
	}
	return len(m.rows) - 1
}

func (m *model) moveUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if m.rows[i].kind == rowService && m.isRowVisible(i) {
			m.cursor = i
			return
		}
	}
	for i := len(m.rows) - 1; i > m.cursor; i-- {
		if m.rows[i].kind == rowService && m.isRowVisible(i) {
			m.cursor = i
			return
		}
	}
}

func (m *model) moveDown() {
	for i := m.cursor + 1; i < len(m.rows); i++ {
		if m.rows[i].kind == rowService && m.isRowVisible(i) {
			m.cursor = i
			return
		}
	}
	for i := 0; i < m.cursor; i++ {
		if m.rows[i].kind == rowService && m.isRowVisible(i) {
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
	if m.query == "" {
		return true
	}
	for _, fi := range m.filtered {
		if fi == i {
			return true
		}
	}
	return false
}

// ── health ───────────────────────────────────────────────

func checkURL(url string) string {
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := client.Head(url)
	if err != nil {
		resp, err = client.Get(url)
		if err != nil {
			return "✗"
		}
	}
	defer resp.Body.Close()
	return "✓"
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func ensureLocalBinInPATH() string {
	pathEnv := os.Getenv("PATH")
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	binDir := filepath.Join(home, ".local", "bin")

	for _, p := range filepath.SplitList(pathEnv) {
		if p == binDir {
			return ""
		}
	}

	exportLine := "\n# Added by homepage CLI\nexport PATH=\"$HOME/.local/bin:$PATH\"\n"
	var updated []string

	zshrc := filepath.Join(home, ".zshrc")
	if data, err := os.ReadFile(zshrc); err == nil {
		if !strings.Contains(string(data), `export PATH="$HOME/.local/bin:$PATH"`) {
			if f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				f.WriteString(exportLine)
				f.Close()
				updated = append(updated, "~/.zshrc")
			}
		}
	}

	bashrc := filepath.Join(home, ".bashrc")
	if data, err := os.ReadFile(bashrc); err == nil {
		if !strings.Contains(string(data), `export PATH="$HOME/.local/bin:$PATH"`) {
			if f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				f.WriteString(exportLine)
				f.Close()
				updated = append(updated, "~/.bashrc")
			}
		}
	}

	if len(updated) > 0 {
		return fmt.Sprintf(" (Added to %s)", strings.Join(updated, ", "))
	}
	return " (Run: source ~/.zshrc)"
}

func installCLI(alias string) (string, error) {
	if alias == "" {
		alias = "hp"
	}

	bin, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate binary error: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir error: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir error: %w", err)
	}

	dst := filepath.Join(binDir, alias)

	data, err := os.ReadFile(bin)
	if err != nil {
		if fallbackData, err2 := os.ReadFile("output/tui"); err2 == nil {
			data = fallbackData
		} else {
			return "", fmt.Errorf("read binary error: %w", err)
		}
	}

	_ = os.Remove(dst)

	if err := os.WriteFile(dst, data, 0755); err != nil {
		return "", fmt.Errorf("write binary error: %w", err)
	}

	pathMsg := ensureLocalBinInPATH()
	return fmt.Sprintf("Installed ~/.local/bin/%s%s", alias, pathMsg), nil
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
	title := m.cfg.Header
	titleLine := titleText.Render(fmt.Sprintf("  %s ", title))

	var rightTop string
	if m.settings.DateTime {
		now := time.Now()
		rightTop = subText.Render(now.Format("Mon Jan 2  15:04:05"))
	}

	subLine := subText.Render(fmt.Sprintf("  %d services", len(m.services)))
	if m.checking {
		sp := spinnerFrames[m.tick%len(spinnerFrames)]
		subLine += "  " + lipgloss.NewStyle().Foreground(yellow).Render(sp)
	} else if len(m.health) > 0 {
		ok, fail := m.countHealth()
		subLine += "  " + statOk.Render(fmt.Sprintf("%d✓", ok))
		if fail > 0 {
			subLine += " " + statFail.Render(fmt.Sprintf("%d✗", fail))
		}
	}

	w := m.width - 2
	bar := strings.Repeat("═", w)

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

	top := subText.Render("╔" + bar + "╗")
	m1 := subText.Render("║") + mid1 + subText.Render("║")
	m2 := subText.Render("║") + mid2 + subText.Render("║")
	bot := subText.Render("╚" + bar + "╝")

	return top + "\n" + m1 + "\n" + m2 + "\n" + bot
}

func (m model) renderDetail() string {
	svc := m.selectedService()
	if svc == nil || m.cursor >= len(m.rows) {
		return ""
	}

	icon := svc.Icon
	if icon == "" {
		icon = "⚙"
	}

	healthStr := ""
	if s, ok := m.health[m.rows[m.cursor].serviceIdx]; ok {
		if s == "✓" {
			healthStr = okDot + lipgloss.NewStyle().Foreground(green).Render(" online")
		} else {
			healthStr = failDot + lipgloss.NewStyle().Foreground(red).Render(" offline")
		}
	}

	line1 := fmt.Sprintf("  %s  %s  %s", icon,
		lipgloss.NewStyle().Bold(true).Foreground(textHi).Render(svc.Name),
		healthStr,
	)
	line2 := fmt.Sprintf("  %s", lipgloss.NewStyle().Foreground(accent).Render(svc.URL))
	line3 := fmt.Sprintf("  %s%s",
		lipgloss.NewStyle().Foreground(muted).Render("groups: "),
		lipgloss.NewStyle().Foreground(subtle).Render(strings.Join(svc.Groups, ", ")),
	)

	w := m.width - 2
	bar := strings.Repeat("═", w)
	bdr := subText

	padTo := func(s string) string {
		n := w - realStringWidth(s)
		if n > 0 {
			return s + strings.Repeat(" ", n)
		}
		return s
	}

	left := bdr.Render("║")
	right := bdr.Render("║")
	top := bdr.Render("╔" + bar + "╗")
	bot := bdr.Render("╚" + bar + "╝")

	return "\n" + top + "\n" +
		left + padTo(line1) + right + "\n" +
		left + padTo(line2) + right + "\n" +
		left + padTo(line3) + right + "\n" +
		bot
}

func (m model) renderSettings() string {
	var lines []string
	title := lipgloss.NewStyle().Foreground(accent).Bold(true).Render("  Settings")
	lines = append(lines, title, "")

	for i, item := range settingItems {
		label := settingLabel.Render(item.label)
		var val string
		switch item.key {
		case "theme":
			val = settingValue.Render(colorThemes[m.settings.Theme].Name)
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
		lines = append(lines, cursor+label+"  "+val)
	}

	if m.editingAlias {
		lines = append(lines, "")
		lines = append(lines, subText.Render("  Enter CLI alias name:"))
		lines = append(lines, "  "+m.aliasInput.View())
		lines = append(lines, "")
		lines = append(lines, subText.Render("  enter install  esc cancel"))
	} else {
		if m.installStatus != "" {
			lines = append(lines, "", "  "+m.installStatus)
		}
		lines = append(lines, "")
		lines = append(lines, subText.Render("  ↑↓ navigate  ←→ change"))
		lines = append(lines, subText.Render("  enter select  esc cancel"))
	}

	panel := strings.Join(lines, "\n")
	w := 36
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(muted).
		Padding(0, 1).
		Width(w).
		Render(panel)
}

func (m model) renderHelp() string {
	keys := []struct{ key, desc string }{
		{"↑↓", "nav"},
		{"/", "search"},
		{"enter", "open"},
		{",", "settings"},
		{"h", "health"},
		{"q", "quit"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, helpKeyStyle.Render(k.key)+" "+helpDescStyle.Render(k.desc))
	}
	return strings.Join(parts, "  │  ")
}

func (m model) renderServiceLine(idx int, r row, selected bool, maxName, maxURL int) string {
	svc := m.services[r.serviceIdx]
	icon := svc.Icon
	if icon == "" {
		icon = "⚙"
	}

	status := ""
	if s, ok := m.health[r.serviceIdx]; ok {
		if s == "✓" {
			status = " " + okDot
		} else {
			status = " " + failDot
		}
	} else if m.checking {
		sp := spinnerFrames[m.tick%len(spinnerFrames)]
		status = " " + lipgloss.NewStyle().Foreground(muted).Render(sp)
	}

	prefix := "    "
	iconSt := normIcon
	nameSt := normName
	if selected {
		prefix = "  ▸ "
		iconSt = selIcon
		nameSt = selName
	}
	if m.settings.Compact {
		if selected {
			prefix = " ▸ "
		} else {
			prefix = "   "
		}
	}

	nameW := realStringWidth(svc.Name)
	paddedName := svc.Name + strings.Repeat(" ", maxName-nameW)
	paddedURL := svc.URL + strings.Repeat(" ", maxURL-len(svc.URL))

	return prefix + iconSt.Render(padIcon(icon)+" ") + nameSt.Render(paddedName) + urlDim.Render("  "+paddedURL) + status
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
		if w := len(svc.URL); w > maxURL {
			maxURL = w
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
			icon := r.groupIcon
			if icon == "" {
				icon = "▸"
			}
			cnt := m.groupServiceCount(r.groupID)
			b.WriteString(grpHdr.Render(fmt.Sprintf("  %s %s:", icon, r.groupName)) +
				" " + grpCnt.Render(fmt.Sprintf("(%d)", cnt)))
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

	if m.showSettings {
		settingsView := m.renderSettings()
		settingsH := lipgloss.Height(settingsView)
		settingsW := lipgloss.Width(settingsView)
		startY := (m.height - settingsH) / 2
		startX := (m.width - settingsW) / 2
		if startY < 0 {
			startY = 0
		}
		if startX < 0 {
			startX = 0
		}
		_ = startY
		_ = startX
		b.WriteString("\n" + settingsView)
	}

	return b.String()
}

// ── main ─────────────────────────────────────────────────

func main() {
	os.Remove(debugLog)
	dlog("tui started")

	cfg, err := loadConfig()
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

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
