package main

// this package will consistently check /proc/#### directories
// for the environ file. When it finds a new one, it checks to
// see if the SSH_AUTH_SOCK variable is set, and if it is it will
// dump it to the screen. This will allow a user to hijack an SSH
// key as long as that user is still in the system.

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// first and foremost, let's confirm that we're root.
	if !areWeRoot() {
		fmt.Println("You must be root to run this application. Quitting...")
		return
	}
	s, _ := findEnviron("/proc")
	for idx, i := range s {
		fmt.Println("Got:", idx, i)
	}
}

func areWeRoot() bool {
	user, err := user.Current()
	if err != nil {
		log.Fatalln("Could not get current user!")
	}
	thisUID, err := strconv.Atoi(user.Uid)
	if err != nil {
		return false //wha?
	}
	if thisUID == 0 {
		// we are root
		return true
	}
	return false
}

func findEnviron(startPath string) ([]string, error) {
	// This function will return a slice of strings pointing
	// to the location of an environ file. This will be split
	// to a workqueue later.
	if !strings.HasSuffix(startPath, "/") {
		startPath += "/"
	}
	// just making sure we end with a /
	rex, err := regexp.Compile(startPath + `\d+/environ`)

	// create the slice to append to
	var enviroSlice []string

	if err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		fmt.Println("On:", path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if rex.MatchString(path) {
			enviroSlice = append(enviroSlice, path)
		}
		return nil
	}); err != nil {
		log.Println("Could not read path!")
		log.Println(err)
		return enviroSlice, err
	}
	return enviroSlice, err
}
