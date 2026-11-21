package main

import (
	"flag"
	"log"
)


func main() {
	var folder, email string
	flag.StringVar(&folder, "add", "", "add a new folder for git repositories")
	flag.StringVar(&email, "email", "179098538+reuien@users.noreply.github.com", "your email to scan")
	flag.Parse()
	if folder != "" {
		scan(folder)
		return
	}
	if err := stats(email); err != nil {
		log.Fatal(err)
	}
}
