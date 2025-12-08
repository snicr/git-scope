Roadmap

 Real git status parsing

 Configurable roots & ignore rules

 Cache for faster startup

 Jump-to-editor actions

 Sorting & filtering in TUI


---

## 3️⃣ User stories + acceptance criteria

### Story 1 – Discover git repos under configured roots

**As a** developer  
**I want** the tool to find all git repos under my directories  
**So that** I can see everything I’m working on in one place.

**Acceptance criteria:**

- Given a config with `roots: [~/code]`, running `git-scope scan`:
  - Recursively finds all directories with `.git`  
  - Outputs a JSON array where each item has `name`, `path`, and `status` fields  
  - Ignores directories listed in `ignore` (e.g. `node_modules`)

---

### Story 2 – Show basic repo status in TUI

**As a** developer  
**I want** a TUI dashboard listing my repos  
**So that** I can quickly see what’s dirty.

**Acceptance criteria:**

- Running `git-scope tui` opens a full-screen TUI  
- Each row shows:
  - Repo name  
  - Path  
  - Branch  
  - Staged / unstaged / untracked counts (MVP: 0s allowed)  
  - Last commit time (approx)  
- I can scroll with ↑/↓ or j/k  
- Pressing `q` exits cleanly

---

### Story 3 – Configurable root directories & ignore patterns

**Acceptance criteria:**

- If no config file exists, defaults are used and no error is thrown  
- User can create config file with `roots` and `ignore`  
- Updating config changes scan behavior without recompiling

---

### Story 4 – Performance on medium codebases

**As a** developer  
**I want** scans to feel fast  
**So that** I actually use this daily.

**Acceptance criteria:**

- On a machine with at least 50 repos under the configured roots:  
  - `git-scope scan` completes under ~3–5 seconds (target; soft)  
  - TUI startup with cached data under ~1 second (later story)

---

### Story 5 – Open repo in editor from TUI (basic)

**Acceptance criteria:**

- When a repo row is selected and I press `Enter`:
  - The configured editor is spawned (`code <repo-path>` by default)  
- If editor is not found, an error message is shown in TUI (non-crash)

---

### Story 6 – Ignore rules respected

**Acceptance criteria:**

- Any directory whose name matches an `ignore` entry is skipped entirely  
- `.git` inside ignored trees is not treated as repos  
- Adding/removing patterns in config changes behavior on next run

---

## 4️⃣ GitHub backlog (issues list, ready to paste)

You can turn these into GitHub Issues as-is.

---

### Issue 1 – Initialize Go module & repo

**Description:**  
Set up initial Go module, `.gitignore`, and basic `README`.

**Checklist:**
- [ ] Create `go.mod` with module path
- [ ] Add `.gitignore`
- [ ] Add minimal `README.md`
- [ ] Verify `go build ./...` works

---

### Issue 2 – Implement config loading

**Description:**  
Add YAML-based config to define `roots`, `ignore`, and `editor`.

**Checklist:**
- [ ] Create `internal/config/config.go`
- [ ] Implement defaults when config file is missing
- [ ] Implement `Load(path string)` that merges file with defaults
- [ ] Add example config at `configs/config.example.yml`
- [ ] Unit tests for default and custom config

---

### Issue 3 – Implement repo scanner

**Description:**  
Walk configured root directories, discover git repos, and return a list.

**Checklist:**
- [ ] Implement `ScanRoots(roots, ignore []string) ([]model.Repo, error)`
- [ ] Skip directories that match ignore patterns
- [ ] Detect `.git` directories and determine repo path/name
- [ ] Add concurrency (goroutines) for scanning multiple roots
- [ ] Add basic tests with a small fixture directory

---

### Issue 4 – Implement `git status` parser

**Description:**  
For a given repo path, run `git` commands to compute basic status.

**Checklist:**
- [ ] Implement `gitstatus.Status(repoPath string) (model.RepoStatus, error)`
- [ ] Parse branch name and ahead/behind from `git status -sb` or porcelain v2
- [ ] Count staged, unstaged, untracked files (approx is okay for MVP)
- [ ] Get last commit timestamp from `git log -1`
- [ ] Handle errors gracefully (populate `ScanError`)

---

### Issue 5 – Implement `git-scope scan` command

**Description:**  
Wire config + scanner + JSON output into CLI.

**Checklist:**
- [ ] Parse command args: `git-scope scan`
- [ ] Load config
- [ ] Call `ScanRoots`
- [ ] Print JSON to stdout
- [ ] Add simple error handling and exit codes

---

### Issue 6 – Implement TUI skeleton

**Description:**  
Create a Bubbletea program with a table listing repos.

**Checklist:**
- [ ] Create `internal/tui` package with `Model`, `View`, `Update`
- [ ] Use `bubbles/table` for the list
- [ ] Hardcode 1–2 demo rows initially
- [ ] Support ↑/↓ or j/k navigation
- [ ] Support `q` to quit

---

### Issue 7 – Connect scanner to TUI

**Description:**  
Use actual scan results to populate the TUI table.

**Checklist:**
- [ ] On TUI start, run `ScanRoots` using config
- [ ] Convert repo list to table rows
- [ ] Show loading state while scanning (even just “Scanning…”)
- [ ] Handle error state (e.g., cannot scan)

---

### Issue 8 – Implement open-in-editor action

**Description:**  
From TUI, pressing `Enter` should open the selected repo in configured editor.

**Checklist:**
- [ ] Read `editor` from config
- [ ] On `Enter`, spawn `editor <repoPath>`
- [ ] Show a small status line message (“Opening in code…”)
- [ ] Handle missing editor binary gracefully

---

### Issue 9 – Basic tests & CI

**Description:**  
Add tests for config and scanner; set up GitHub Actions.

**Checklist:**
- [ ] Tests for config defaults and overrides
- [ ] Tests for scanner on test directories
- [ ] GitHub Actions workflow: `go test ./...` on push/PR
- [ ] Badge in README

---

### Issue 10 – MVP polish & first release

**Description:**  
Prepare for first public tag v0.1.0.

**Checklist:**
- [ ] Clean up logs and debug prints
- [ ] Improve README with installation & usage
- [ ] Add asciinema or GIF of TUI
- [ ] Tag release `v0.1.0`

---

## 5️⃣ TUI screens wireframes (ASCII)

### Main dashboard

```text
+--------------------------------------------------------------------------------------+
| git-scope — all your git repos at a glance                               [q] quit   |
+--------------------------------------------------------------------------------------+
| Repo           │ Path                               │ Branch │ Stg │ Unst │ Untrk │ Last Commit      |
+----------------+------------------------------------+--------+-----+------+-------+------------------+
| api-service    │ ~/code/api-service                 │ main   │  2  │   1  │   0   │ 2025-12-08 10:12 |
| web-dashboard  │ ~/code/web-dashboard               │ dev    │  0  │   5  │   1   │ 2025-12-05 09:03 |
| cli-tools      │ ~/projects/cli-tools               │ main   │  0  │   0  │   0   │ 2025-12-01 20:45 |
| infra          │ ~/work/infra                       │ prod   │  0  │   3  │   0   │ 2025-11-30 18:21 |
+--------------------------------------------------------------------------------------+
  ↑/↓ move • Enter open in editor • / filter (future) • s sort (future) • q quit


Loading state
+--------------------------------------------------------------------------------------+
| git-scope — all your git repos at a glance                                           |
+--------------------------------------------------------------------------------------+
| Scanning roots:                                                                      |
|   - ~/code                                                                           |
|   - ~/projects                                                                       |
|                                                                                      |
| Please wait...                                                                       |
+--------------------------------------------------------------------------------------+
  q quit

Error state
+--------------------------------------------------------------------------------------+
| git-scope — all your git repos at a glance                                           |
+--------------------------------------------------------------------------------------+
| Error: failed to scan repos                                                          |
| Details: permission denied under /some/path                                          |
|                                                                                      |
| Try adjusting your config or running with different permissions.                     |
+--------------------------------------------------------------------------------------+
  q quit


