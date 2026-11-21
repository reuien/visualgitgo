package main

import (
	"flag"
)


func main() {
	var folder, email string
	flag.StringVar(&folder, "add", "", "add a new folder for git repositories")
	flag.StringVar(&email, "email", "youemail.com", "your email to scan")
	flag.Parse()
	if folder != "" {
		scan(folder)
		return
	}
	stats(email)
}
