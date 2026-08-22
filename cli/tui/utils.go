package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"gopkg.in/yaml.v3"
)

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

func loadConfig(path string) (Config, error) {
	var cfg Config
	var data []byte
	if path != "" {
		var err error
		data, err = os.ReadFile(path)
		if err != nil {
			return cfg, fmt.Errorf("reading %s: %w", path, err)
		}
	} else {
		data = servicesYML
	}
	return cfg, yaml.Unmarshal(data, &cfg)
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

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "homepage", "tui.json")
}

func loadSettings() Settings {
	s := Settings{
		Theme:       0,
		BorderStyle: 0,
		IconStyle:   0,
		DateTime:    true,
		Compact:     false,
		Alias:       "hp",
		GroupIcons:  true,
	}
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return s
	}
	json.Unmarshal(data, &s)
	if s.ShowIcons != nil && !*s.ShowIcons && s.IconStyle == 0 {
		s.IconStyle = 2 // Off
	}
	if s.Theme < 0 || s.Theme >= len(colorThemes) {
		s.Theme = 0
	}
	if s.BorderStyle < 0 || s.BorderStyle >= len(borderSets) {
		s.BorderStyle = 0
	}
	if s.IconStyle < 0 || s.IconStyle >= 3 {
		s.IconStyle = 0
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

func validAlias(a string) bool {
	if a == "" {
		return false
	}
	for _, r := range a {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// ponytail: indirection so tests can fake the installed binary location
var exePath = os.Executable

func installCLI(oldAlias, alias string) (string, error) {
	if alias == "" {
		alias = "hp"
	}
	if !validAlias(alias) {
		return "", fmt.Errorf("invalid alias %q (use letters, digits, - or _)", alias)
	}

	bin, err := exePath()
	if err != nil {
		return "", fmt.Errorf("locate binary error: %w", err)
	}
	// ponytail: launched via the installed symlink -> resolve or we'd link to ourselves
	if resolved, err := filepath.EvalSymlinks(bin); err == nil {
		bin = resolved
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
	_ = os.Remove(dst)

	data, err := os.ReadFile(bin)
	if err != nil {
		return "", fmt.Errorf("read binary error: %w", err)
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return "", fmt.Errorf("write binary error: %w", err)
	}

	if oldAlias != "" && oldAlias != alias {
		_ = os.Remove(filepath.Join(binDir, oldAlias))
	}

	pathMsg := ensureLocalBinInPATH()
	return fmt.Sprintf("Installed ~/.local/bin/%s%s", alias, pathMsg), nil
}

// refreshInstalledCLI recopies the installed alias if it differs from the
// running binary, so a rebuilt output/tui updates hp on next launch.
// Survives repo deletion (plain copy, not symlink).
func refreshInstalledCLI(alias string) {
	if alias == "" {
		alias = "hp"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dst := filepath.Join(home, ".local", "bin", alias)
	if _, err := os.Lstat(dst); err != nil {
		return // never installed
	}
	bin, err := exePath()
	if err != nil {
		return
	}
	if resolved, err := filepath.EvalSymlinks(bin); err == nil {
		bin = resolved
	}
	stale := true
	if info, err := os.Stat(dst); err == nil {
		if src, err := os.Stat(bin); err == nil &&
			info.Size() == src.Size() && info.ModTime().Sub(src.ModTime()) < time.Second {
			stale = false
		}
	}
	if !stale {
		return
	}
	// ponytail: naive size+mtime staleness check; checksums if false refreshes bite
	data, err := os.ReadFile(bin)
	if err != nil {
		return
	}
	_ = os.Remove(dst)
	_ = os.WriteFile(dst, data, 0755)
}

func aliasInstalled(alias string) bool {
	if alias == "" {
		alias = "hp"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".local", "bin", alias))
	return err == nil
}
