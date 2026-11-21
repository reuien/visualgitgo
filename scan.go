package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

func recursiveScanFolder(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

func scanGitFolders(folders []string, folder string) []string {
	folder = strings.TrimSuffix(folder, "/")
	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	var path string
	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()
			// file is a FILE*
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				fmt.Println(path)
				folders = append(folders, path)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}
			folders = scanGitFolders(folders, path)
		}
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
		lines = append(lines, scanner.Text())
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
	os.WriteFile(filePath, []byte(content), 0755)
}

func addNewSliceElementsToFile(filePath string, newRepo []string) {
	existedRepo := parseRepoIntoSlice(filePath)
	repos := joinRepos(newRepo, existedRepo)
	storeSliceToFile(repos, filePath)
}

func scan(path string) {
	fmt.Println("Found Flolders:\n\n")
	repositories := recursiveScanFolder(path)
	filePath := getDotFilePath()
	addNewSliceElementsToFile(filePath, repositories)
	fmt.Println("\n\nSuccessfully added \n\n")

}
