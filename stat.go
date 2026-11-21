package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

const daysInLastSixMonth = 183

func dayStart(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

// collectCommits counts each reachable commit once per repository, including
// commits on other local and remote branches. Dates use the local time zone.
func collectCommits(path, email string, now time.Time, counts map[time.Time]int) error {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return fmt.Errorf("open repository %q: %w", path, err)
	}
	defer repo.Close()
	iter, err := repo.Log(&git.LogOptions{All: true})
	if err != nil {
		return fmt.Errorf("read history of %q: %w", path, err)
	}
	defer iter.Close()
	today := dayStart(now, now.Location())
	start := today.AddDate(0, 0, -(daysInLastSixMonth - 1))
	seen := make(map[plumbing.Hash]bool)
	err = iter.ForEach(func(commit *object.Commit) error {
		if seen[commit.Hash] {
			return nil
		}
		seen[commit.Hash] = true
		date := dayStart(commit.Author.When, today.Location())
		if commit.Author.Email == email && !date.Before(start) && !date.After(today) {
			counts[date]++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk history of %q: %w", path, err)
	}
	return nil
}

func printCell(w io.Writer, val int, today bool) {
	escape := "\033[0;37;30m"
	switch {
	case val >= 10:
		escape = "\033[1;30;42m"
	case val >= 5:
		escape = "\033[1;30;43m"
	case val > 0:
		escape = "\033[1;30;47m"
	}
	if today {
		escape = "\033[1;37;45m"
	}
	if val == 0 {
		fmt.Fprint(w, escape+"  - "+"\033[0m")
	} else if val > 999 {
		fmt.Fprint(w, escape+"999+"+"\033[0m")
	} else {
		fmt.Fprintf(w, escape+"%3d "+"\033[0m", val)
	}
}

func printCommits(w io.Writer, counts map[time.Time]int, now time.Time) {
	today := dayStart(now, now.Location())
	start := today.AddDate(0, 0, -(daysInLastSixMonth - 1))
	firstWeek := start.AddDate(0, 0, -int(start.Weekday()))
	lastWeek := today.AddDate(0, 0, -int(today.Weekday()))
	fmt.Fprint(w, "     ")
	previousMonth := time.Month(0)
	for week := firstWeek; !week.After(lastWeek); week = week.AddDate(0, 0, 7) {
		labelDate := week
		if labelDate.Before(start) {
			labelDate = start
		}
		if labelDate.Month() != previousMonth {
			fmt.Fprintf(w, "%-4s", labelDate.Month().String()[:3])
			previousMonth = labelDate.Month()
		} else {
			fmt.Fprint(w, "    ")
		}
	}
	fmt.Fprintln(w)
	labels := [7]string{"    ", " Mon", "    ", " Wed", "    ", " Fri", "    "}
	for weekday := 0; weekday < 7; weekday++ {
		fmt.Fprintf(w, "%s ", labels[weekday])
		for week := firstWeek; !week.After(lastWeek); week = week.AddDate(0, 0, 7) {
			date := week.AddDate(0, 0, weekday)
			if date.Before(start) || date.After(today) {
				fmt.Fprint(w, "    ")
				continue
			}
			printCell(w, counts[date], date.Equal(today))
		}
		fmt.Fprintln(w)
	}
}

func stats(email string) error {
	repos := parseRepoIntoSlice(getDotFilePath())
	if len(repos) == 0 {
		// With no saved list, inspect the current repository (or an ancestor).
		repos = []string{"."}
	}
	counts := make(map[time.Time]int)
	now := time.Now()
	for _, path := range repos {
		if err := collectCommits(path, email, now, counts); err != nil {
			return err
		}
	}
	printCommits(os.Stdout, counts, now)
	return nil
}
