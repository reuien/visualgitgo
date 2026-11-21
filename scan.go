package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func recursiveScanFolder(folder string) []string {
	folder, err := filepath.Abs(folder)
	if err != nil {
		log.Fatal(err)
	}
	var folders []string
	err = filepath.WalkDir(folder, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			if entry.IsDir() || entry.Type().IsRegular() {
				repo := filepath.Dir(path)
				fmt.Println(repo)
				folders = append(folders, repo)
			}
			if entry.IsDir() {
				return filepath.SkipDir
			}
		}
		if entry.IsDir() && (entry.Name() == "vendor" || entry.Name() == "node_modules") {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return folders
}

func getDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gogitlocalstats"
	return dotFile
}

func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		// if file is not existed then create one
		if os.IsNotExist(err) {
			f, err = os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			// other fault
			panic(err)
		}
	}
	return f
}

func parseRepoIntoSlice(filePath string) []string {
	f := openFile(filePath)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	return lines
}

// put the elements of the new slice into the old slices
// if the element not already there
func joinRepos(new []string, old []string) []string {
	for _, i := range new {
		if !sliceContains(old, i) {
			old = append(old, i)
		}
	}
	return old
}

// judge if a element in the slice
func sliceContains(existing []string, elem string) bool {
	for _, v := range existing {
		if v == elem {
			return true
		}
	}
	return false
}

func storeSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		log.Fatal(err)
	}
}

func addNewSliceElementsToFile(filePath string, newRepo []string) {
	existedRepo := parseRepoIntoSlice(filePath)
	repos := joinRepos(newRepo, existedRepo)
	storeSliceToFile(repos, filePath)
}

func scan(path string) {
	fmt.Println("Found repositories:")
	repositories := recursiveScanFolder(path)
	filePath := getDotFilePath()
	addNewSliceElementsToFile(filePath, repositories)
	fmt.Println("Successfully added.")

}
