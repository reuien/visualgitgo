package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCollectCommitsAllBranchesAndDateBounds(t *testing.T) {
	dir := t.TempDir()
	run := func(date, email string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_COMMITTER_NAME=Test", "GIT_AUTHOR_EMAIL="+email, "GIT_COMMITTER_EMAIL="+email, "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("2026-01-01T12:00:00Z", "me@example.com", "init", "-q")
	commit := func(date, email, message string) {
		run(date, email, "commit", "--allow-empty", "-q", "-m", message)
	}
	commit("2026-05-22T12:00:00Z", "me@example.com", "outside window")
	commit("2026-05-23T12:00:00Z", "me@example.com", "first day")
	run("2026-11-21T12:00:00Z", "me@example.com", "branch", "other")
	commit("2026-11-21T12:00:00Z", "me@example.com", "today")
	commit("2026-11-21T12:00:00Z", "another@example.com", "other author")
	commit("2026-11-22T12:00:00Z", "me@example.com", "future")
	run("2026-11-20T12:00:00Z", "me@example.com", "checkout", "-q", "other")
	commit("2026-11-20T12:00:00Z", "me@example.com", "other branch")
	now := time.Date(2026, 11, 21, 13, 0, 0, 0, time.UTC)
	counts := make(map[time.Time]int)
	if err := collectCommits(dir, "me@example.com", now, counts); err != nil {
		t.Fatal(err)
	}
	for _, date := range []time.Time{now, now.AddDate(0, 0, -1), now.AddDate(0, 0, -182)} {
		if got := counts[dayStart(date, time.UTC)]; got != 1 {
			t.Errorf("%s: got %d, want 1", date.Format(time.DateOnly), got)
		}
	}
	if len(counts) != 3 {
		t.Errorf("unexpected dates: %v", counts)
	}
}

func TestPrintCommitsAlignsWeekdaysAndToday(t *testing.T) {
	now := time.Date(2026, 11, 21, 13, 0, 0, 0, time.UTC) // Saturday
	counts := map[time.Time]int{dayStart(now, time.UTC): 2, dayStart(now.AddDate(0, 0, -1), time.UTC): 1}
	var out bytes.Buffer
	printCommits(&out, counts, now)
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("got %d lines, want 8", len(lines))
	}
	if !strings.HasPrefix(lines[6], " Fri ") || !strings.HasSuffix(lines[6], "  1 \033[0m") {
		t.Errorf("Friday row: %q", lines[6])
	}
	if !strings.HasSuffix(lines[7], "\033[1;37;45m  2 \033[0m") {
		t.Errorf("today row: %q", lines[7])
	}
	clean := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(out.String(), "")
	for _, line := range strings.Split(strings.TrimSuffix(clean, "\n"), "\n") {
		if len(line) != len(strings.SplitN(clean, "\n", 2)[0]) {
			t.Errorf("misaligned row: %q", line)
		}
	}
}

func TestScanRecognizesWorktreeGitFile(t *testing.T) {
	dir := t.TempDir()
	worktree := filepath.Join(dir, "worktree")
	if err := os.Mkdir(worktree, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: /tmp/example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	repos := recursiveScanFolder(dir)
	if len(repos) != 1 || repos[0] != worktree {
		t.Fatalf("got %v, want [%s]", repos, worktree)
	}
}
