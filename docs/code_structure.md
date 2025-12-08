1️⃣ Folder structure + initial files

Suggested project name: git-scope (you can rename later).

git-scope/
├── cmd/
│   └── git-scope/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── model/
│   │   └── repo.go
│   ├── scan/
│   │   └── scan.go
│   ├── gitstatus/
│   │   └── gitstatus.go
│   ├── cache/
│   │   └── cache.go
│   └── tui/
│       ├── app.go
│       ├── model.go
│       ├── view.go
│       └── update.go
├── configs/
│   └── config.example.yml
├── .gitignore
├── go.mod
└── README.md

2️⃣ Go code scaffolding
go.mod

Update module to your GitHub handle later.

module github.com/yourname/git-scope

go 1.22

require (
    github.com/charmbracelet/bubbletea v0.26.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    gopkg.in/yaml.v3 v3.0.1
)


cmd/git-scope/main.go

Entry point with two commands: scan and tui.

package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/yourname/git-scope/internal/config"
    "github.com/yourname/git-scope/internal/scan"
    "github.com/yourname/git-scope/internal/tui"
)

const version = "0.1.0"

func usage() {
    fmt.Fprintf(os.Stderr, `git-scope %s

Usage:
  git-scope scan           Scan configured roots and print repos (JSON)
  git-scope tui            Launch TUI dashboard
  git-scope help           Show this help

Flags:
`, version)
    flag.PrintDefaults()
}

func main() {
    flag.Usage = usage
    configPath := flag.String("config", defaultConfigPath(), "Path to config file")
    flag.Parse()

    if flag.NArg() < 1 {
        usage()
        os.Exit(1)
    }

    cmd := flag.Arg(0)

    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    switch cmd {
    case "scan":
        repos, err := scan.ScanRoots(cfg.Roots, cfg.Ignore)
        if err != nil {
            log.Fatalf("scan error: %v", err)
        }
        if err := scan.PrintJSON(os.Stdout, repos); err != nil {
            log.Fatalf("print error: %v", err)
        }
    case "tui":
        if err := tui.Run(cfg); err != nil {
            log.Fatalf("tui error: %v", err)
        }
    case "help", "-h", "--help":
        usage()
    default:
        fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
        usage()
        os.Exit(1)
    }
}

func defaultConfigPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return "./config.yml"
    }
    return filepath.Join(home, ".config", "git-scope", "config.yml")
}



internal/config/config.go

Basic YAML config loader.


package config

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Roots  []string `yaml:"roots"`
    Ignore []string `yaml:"ignore"`
    Editor string   `yaml:"editor"` // e.g. "code", "idea", "nvim"
}

func defaultConfig() *Config {
    home, _ := os.UserHomeDir()
    return &Config{
        Roots: []string{
            filepath.Join(home, "code"),
            filepath.Join(home, "projects"),
        },
        Ignore: []string{
            "node_modules",
            ".next",
            "dist",
            "build",
            "target",
            ".venv",
        },
        Editor: "code",
    }
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // If file does not exist, return defaults (no error)
        if os.IsNotExist(err) {
            return defaultConfig(), nil
        }
        return nil, fmt.Errorf("read config: %w", err)
    }
    cfg := defaultConfig()
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return cfg, nil
}

configs/config.example.yml

roots:
  - ~/code
  - ~/projects

ignore:
  - node_modules
  - .next
  - dist
  - build
  - target
  - .venv

editor: code

internal/model/repo.go

Central data model.

package model

import "time"

type RepoStatus struct {
    Branch      string    `json:"branch"`
    Ahead       int       `json:"ahead"`
    Behind      int       `json:"behind"`
    Staged      int       `json:"staged"`
    Unstaged    int       `json:"unstaged"`
    Untracked   int       `json:"untracked"`
    LastCommit  time.Time `json:"last_commit"`
    IsDirty     bool      `json:"is_dirty"`
    ScanError   string    `json:"scan_error,omitempty"`
}

type Repo struct {
    Name   string     `json:"name"`
    Path   string     `json:"path"`
    Status RepoStatus `json:"status"`
}

internal/scan/scan.go

Repo discovery + JSON output.

package scan

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "sync"

    "github.com/yourname/git-scope/internal/gitstatus"
    "github.com/yourname/git-scope/internal/model"
)

func ScanRoots(roots, ignore []string) ([]model.Repo, error) {
    ignoreSet := make(map[string]struct{}, len(ignore))
    for _, pattern := range ignore {
        ignoreSet[pattern] = struct{}{}
    }

    var mu sync.Mutex
    var repos []model.Repo
    var wg sync.WaitGroup

    for _, root := range roots {
        root := os.ExpandEnv(root)
        wg.Add(1)
        go func(r string) {
            defer wg.Done()
            filepath.WalkDir(r, func(path string, d os.DirEntry, err error) error {
                if err != nil {
                    return nil
                }

                if d.IsDir() && shouldIgnore(d.Name(), ignoreSet) {
                    return filepath.SkipDir
                }

                if d.IsDir() && d.Name() == ".git" {
                    repoPath := filepath.Dir(path)
                    repoName := filepath.Base(repoPath)

                    status, serr := gitstatus.Status(repoPath)

                    repo := model.Repo{
                        Name: repoName,
                        Path: repoPath,
                        Status: status,
                    }
                    if serr != nil {
                        repo.Status.ScanError = serr.Error()
                    }

                    mu.Lock()
                    repos = append(repos, repo)
                    mu.Unlock()

                    return filepath.SkipDir
                }
                return nil
            })
        }(root)
    }

    wg.Wait()
    return repos, nil
}

func shouldIgnore(name string, ignoreSet map[string]struct{}) bool {
    _, ok := ignoreSet[name]
    if ok {
        return true
    }
    // simple suffix-based ignore
    for pat := range ignoreSet {
        if strings.HasSuffix(name, pat) {
            return true
        }
    }
    return false
}

func PrintJSON(w io.Writer, repos []model.Repo) error {
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err := enc.Encode(repos); err != nil {
        return fmt.Errorf("encode json: %w", err)
    }
    return nil
}

internal/gitstatus/gitstatus.go

Minimal placeholder using git CLI; you’ll fill logic later.

package gitstatus

import (
    "bytes"
    "fmt"
    "os/exec"
    "strconv"
    "strings"
    "time"

    "github.com/yourname/git-scope/internal/model"
)

func Status(repoPath string) (model.RepoStatus, error) {
    // TODO: make more robust, parse porcelain v2 properly
    // For scaffolding, keep it simple.
    status := model.RepoStatus{}

    cmd := exec.Command("git", "status", "--porcelain=v2", "-b")
    cmd.Dir = repoPath
    out, err := cmd.Output()
    if err != nil {
        return status, fmt.Errorf("git status: %w", err)
    }

    // parse branch + ahead/behind
    lines := strings.Split(string(out), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "# branch.head") {
            parts := strings.Fields(line)
            if len(parts) >= 3 {
                status.Branch = parts[2]
            }
        }
        if strings.HasPrefix(line, "# branch.ab") {
            parts := strings.Fields(line)
            if len(parts) >= 3 {
                aheadStr := strings.TrimPrefix(parts[2], "+")
                behindStr := strings.TrimPrefix(parts[3], "-")
                status.Ahead, _ = strconv.Atoi(aheadStr)
                status.Behind, _ = strconv.Atoi(behindStr)
            }
        }
        if len(line) > 0 && line[0] != '#' {
            // crude counters: treat all non-comment as dirty
            if strings.Contains(line, "1 ") || strings.Contains(line, "2 ") {
                status.Unstaged++
            }
            if strings.Contains(line, "? ") {
                status.Untracked++
            }
        }
    }

    status.IsDirty = status.Unstaged > 0 || status.Untracked > 0 || status.Staged > 0

    // last commit time
    t, err := lastCommitTime(repoPath)
    if err == nil {
        status.LastCommit = t
    }

    return status, nil
}

func lastCommitTime(repoPath string) (time.Time, error) {
    cmd := exec.Command("git", "log", "-1", "--format=%ct")
    cmd.Dir = repoPath
    out, err := cmd.Output()
    if err != nil {
        return time.Time{}, fmt.Errorf("git log: %w", err)
    }
    ts := strings.TrimSpace(string(out))
    sec, err := strconv.ParseInt(ts, 10, 64)
    if err != nil {
        return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
    }
    return time.Unix(sec, 0), nil
}


(You can improve parsing later; this is just scaffolding.)

internal/cache/cache.go

Stub for future caching.

package cache

import "github.com/yourname/git-scope/internal/model"

type Store interface {
    Load() ([]model.Repo, error)
    Save([]model.Repo) error
}

// TODO: implement JSON or SQLite-backed store.

internal/tui/app.go

Glue function to start Bubbletea app.

package tui

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/yourname/git-scope/internal/config"
)

func Run(cfg *config.Config) error {
    m := NewModel(cfg)
    p := bubbletea.NewProgram(m)
    _, err := p.Run()
    return err
}


internal/tui/model.go

Minimal Bubbletea model using in-memory data; you’ll wire scanner later.

package tui

import (
    "time"

    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbletea"
    "github.com/yourname/git-scope/internal/config"
    "github.com/yourname/git-scope/internal/model"
)

type Model struct {
    cfg   *config.Config
    table table.Model
}

func NewModel(cfg *config.Config) Model {
    columns := []table.Column{
        {Title: "Repo", Width: 20},
        {Title: "Path", Width: 40},
        {Title: "Branch", Width: 10},
        {Title: "Stg", Width: 4},
        {Title: "Unst", Width: 4},
        {Title: "Untrk", Width: 5},
        {Title: "Last Commit", Width: 19},
    }

    // placeholder rows; later fill with scan results
    repos := []model.Repo{
        {
            Name: "example-repo",
            Path: "/path/to/example-repo",
            Status: model.RepoStatus{
                Branch:     "main",
                LastCommit: time.Now().Add(-2 * time.Hour),
            },
        },
    }

    rows := make([]table.Row, 0, len(repos))
    for _, r := range repos {
        rows = append(rows, table.Row{
            r.Name,
            r.Path,
            r.Status.Branch,
            "0",
            "0",
            "0",
            r.Status.LastCommit.Format("2006-01-02 15:04"),
        })
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
    )

    return Model{
        cfg:   cfg,
        table: t,
    }
}

func (m Model) Init() tea.Msg {
    return nil
}


internal/tui/update.go
package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Msg) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        }
    }

    m.table, cmd = m.table.Update(msg)
    return m, cmd
}


internal/tui/view.go
package tui

import (
    "strings"

    "github.com/charmbracelet/lipgloss"
)

var (
    titleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
    helpStyle  = lipgloss.NewStyle().Faint(true).MarginTop(1)
)

func (m Model) View() string {
    b := &strings.Builder{}
    b.WriteString(titleStyle.Render("git-scope — all your git repos at a glance"))
    b.WriteString("\n\n")
    b.WriteString(m.table.View())
    b.WriteString("\n")
    b.WriteString(helpStyle.Render("↑/↓ move • q quit (MVP)"))
    return b.String()
}

.gitignore
/dist
/build
/bin
*.log
*.tmp
*.swp
.git-scope-cache.json


README.md (skeleton)
# git-scope

> A fast TUI to see the status of all git repositories on your machine.

## Status

⚠️ Early MVP scaffolding. Not ready for production.

## Vision

- Scan configured directories for git repos
- Show branch, dirty status, and last commit time
- Keyboard-driven navigation
- Jump into your editor from the TUI

## Quick Start

```bash
go install github.com/yourname/git-scope/cmd/git-scope@latest

git-scope scan
git-scope tui


