package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/GeertJohan/go.linenoise"
	"github.com/mgutz/ansi"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (

	// Defining var defaults
	defaultConnectionPort = "4242" // Kingpin wants this as string and not int

	// Defining generic consts
	goDiscVersion = "0.1"

	// Defining more func specific consts
	tellSaverMaxLength = 25
)

var (
	connectHost = kingpin.Arg("host", "The IP/Domain to connect to.").Required().String()
	connectPort = kingpin.Arg("port", "Port to connect to.").Default(defaultConnectionPort).String()
)

func main() {

	// Setting the version for kingpin so it can easily be viewed from the terminal.
	kingpin.Version(goDiscVersion)

	// Parsing kingpin arguments.
	kingpin.Parse()

	// Creating the channel on which the connection will post.
	msgchan := make(chan string)

	// Run our initiation function to check for needed directories.
	goDiscInit()

	// Connecting to our host.
	conn, err := net.Dial("tcp", *connectHost+":"+*connectPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)

	// Setting our Seed (used for building random ints) to use a nano unix timestamp.
	rand.Seed(time.Now().UTC().UnixNano())

	// Starting out routines for printing input we get on our buffer and handling keyboard input.
	go printMessages(msgchan, conn)
	go readKeyboardInput(conn)

	// The main loop fomr handling the telnet stream.
	for {
		str, err := connbuf.ReadString('\n')
		if err != nil {
			break
		}
		str = highLight(str)
		msgchan <- str
	}
}

// exists will search for a path, or file, and return true if it exists.
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// returnRand will build a map of 30 random ints and will, when called, return a random one of these random ints.
func returnRand() int {
	randMap := make(map[int]int)
	for i := 0; i < 30; i++ {
		randMap[i] = random(60, 480)
	}
	return randMap[rand.Intn(len(randMap))]
}

// goDiscInit checks for needed directories and files and will create these if they do not exist.
func goDiscInit() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	goDiscCfgDir := usr.HomeDir + "/.config/godisc"
	os.Setenv("goDiscCfgDir", goDiscCfgDir+"/")
	cfgDirExists, err := exists(goDiscCfgDir)
	if err != nil {
		panic(err.Error())
	}
	if cfgDirExists == false {
		err := os.Mkdir(goDiscCfgDir, 0770)
		if err != nil {
			panic(err.Error())
		}
	}
}

// highLight parses through the buffer and replaces words, defined in the highlight.list, with ansi color.
func highLight(str string) string {
	highLightListExists, _ := exists(os.Getenv("goDiscCfgDir") + "highlight.list")
	if highLightListExists == true {
		file, err := os.Open("highlight.list")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.Contains(str, scanner.Text()) {
				str = strings.Replace(str, scanner.Text(), ansi.Color(scanner.Text(), "red+b"), -1)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	} else {
		wlog("Could not find", os.Getenv("goDiscCfgDir")+"highlight.list")
	}
	return str
}

// random gives back a random int based on the Seed defined earlier.
func random(min, max int) int {
	return min + rand.Intn(max-min)
}

// mapSearch walks through a map to see if a value exists.
func mapSearch(s string, a map[int]string) (exists bool) {
	exists = false
	for _, v := range a {
		if v == s {
			exists = true
			return
		}
	}
	return
}

// cleanNpcString simply removes some unwanted characters from npc names.
func cleanNpcString(s string) string {
	foundNpcs := strings.Replace(s, "*", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "+", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "$", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "-", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "/", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "\\", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "!", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "|", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "_", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "^", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "~", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "\"", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "'", "", -1)
	foundNpcs = strings.Replace(foundNpcs, ":", "", -1)
	foundNpcs = strings.Replace(foundNpcs, ";", "", -1)
	foundNpcs = strings.Replace(foundNpcs, ".", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "=", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "}", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "{", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "[", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "]", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "(", "", -1)
	foundNpcs = strings.Replace(foundNpcs, ")", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "@", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "#", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "£", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "¤", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "%", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "&", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "<", "", -1)
	foundNpcs = strings.Replace(foundNpcs, ">", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "010m", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "3949m", "", -1)
	foundNpcs = strings.Replace(foundNpcs, "31m", "", -1)
	foundNpcs = strings.Replace(foundNpcs, " and ", ", ", -1)
	foundNpcs = strings.TrimSpace(foundNpcs)
	return foundNpcs
}

// findAlias reads the alias.list file to see if any keyboard input should be replaced with some pre-defined aliases or not.
func findAlias(str []string) string {
	aliasListExists, _ := exists(os.Getenv("goDiscCfgDir") + "alias.list")
	if aliasListExists == true {
		file, err := os.Open("alias.list")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			splitScan := strings.Split(scanner.Text(), "->")
			if len(splitScan) > 1 {
				for _, v := range str {
					if splitScan[0] == v {
						return splitScan[1]
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	return "none"
}

// readKeyboardInput is a go routine that reads the keyboard input and sends it to the connected buffer when enter is hit.
func readKeyboardInput(c net.Conn) {
	for {
		str, err := linenoise.Line("")
		wlog(str)
		if err != nil {
			if err == linenoise.KillSignalError {
				quit()
			}
			fmt.Printf("Unexpected error: %s\n", err)
			quit()
		}

		inputText := strings.Fields(str)
		cmd := findAlias(inputText)
		if cmd == "none" {
			joinText := strings.Join(inputText, " ")
			if strings.Contains(joinText, "|") {
				splitText := strings.Split(joinText, "|")
				for _, sv := range splitText {
					fmt.Fprintf(c, sv+"\n")
				}
				linenoise.AddHistory(joinText)
			} else {
				fmt.Fprintf(c, joinText+"\n")
				linenoise.AddHistory(joinText)
			}
		} else {
			joinText := strings.Join(inputText, " ")
			wlog("[ CMD ]:", joinText)
			if strings.Contains(cmd, "|") {
				splitText := strings.Split(cmd, "|")
				for _, v := range splitText {
					wlog("Split cmd", v)
					fmt.Fprintf(c, v+"\n")
					linenoise.AddHistory(joinText)
				}
			} else {
				fmt.Fprintf(c, cmd+"\n")
				linenoise.AddHistory(joinText)
			}
		}

	}
}

// quit is used by linenoise as its KillSignalError.
func quit() {
	os.Exit(0)
}

// regComp compiles a given regex pattern and does a FindStringSubmatch on it.
func regComp(str string, reg string) []string {
	r, _ := regexp.Compile(reg)
	res := r.FindStringSubmatch(str)
	return res
}

// clearTellSaver goes through the strings of the buffer to check for certain strings that should not be written to the tell history file.
func clearTellSaver(str string) bool {

	str = strings.Replace(str, "[37m", "", -1)
	str = strings.Replace(str, "[1m", "", -1)

	if len(str) > tellSaverMaxLength {
		return true
	}
	if strings.Contains(str[0:6], "The ") == true {
		return true
	}
	if strings.Contains(str[0:6], "One ") == true {
		return true
	}
	if strings.Contains(str[0:12], "On the") == true {
		return true
	}
	ignoreNpcs := make(map[string]bool)
	ignoreNpcs["sailor"] = true
	ignoreNpcs["seagull"] = true
	ignoreNpcs["barman"] = true
	ignoreNpcs["samurai"] = true
	ignoreNpcs["strongarm"] = true
	ignoreNpcs["tramp"] = true
	ignoreNpcs["Mihk-gran-bohp"] = true
	ignoreNpcs["engineer"] = true
	ignoreNpcs["warrior"] = true
	ignoreNpcs["pickpocket"] = true
	ignoreNpcs["Khepresh"] = true
	ignoreNpcs["smuggler"] = true
	ignoreNpcs["citadel"] = true
	ignoreNpcs["guard"] = true
	ignoreNpcs["hopelite"] = true
	ignoreNpcs["giant"] = true
	ignoreNpcs["schoolboy"] = true
	ignoreNpcs["farmer"] = true
	ignoreNpcs["soldier"] = true
	ignoreNpcs["ceremonial"] = true
	ignoreNpcs["Kang Wu"] = true
	ignoreNpcs["rickshaw driver"] = true
	ignoreNpcs["Imperial guard"] = true
	ignoreNpcs["Ryattenoki"] = true

	for k, v := range ignoreNpcs {
		switch v {
		case true:
			if strings.Contains(str, k) == true {
				return true
			}
		case false:
			continue
		}
	}
	return false
}

// tellSaver actually handles what strings to write to the tell history log.
func tellSaver(str string) bool {
	res := regComp(str, "(You tell|You ask|You exclaim|You shout|You yell) (.+):(.+)")
	if len(res) > 1 {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] (%s) %s : %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "yellow+b"), ansi.Color(res[3], "green+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			wlog(err.Error)
		}

		defer f.Close()

		if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
		return true
	}
	res2 := regComp(str, "(.+) (tell|ask|exclaim|tells|asks|exclaims) (.+):(.+)")
	if len(res2) > 1 {
		if clearTellSaver(res2[1]) == false {
			var stringToWrite string
			t := time.Now()
			stringToWrite = fmt.Sprintf("[ %d:%d:%d ] (%s) %s : %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(strings.Replace(strings.TrimSpace(res2[1]), ">", "", -1), "blue+b"), ansi.Color(res2[3], "yellow+b"), ansi.Color(res2[4], "green+b"))
			f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				wlog(err.Error)
			}

			defer f.Close()

			if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
				wlog(err.Error)
			}
			return true
		}
	}
	return false
}

// chatSaver handles which strings to write to the talker history log.
func chatSaver(str string) bool {
	res := regComp(str, "\\((\\D+)\\) (.+)wisps(.+)")
	if len(res) > 1 {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] (%s) %s %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "yellow+b"), ansi.Color(res[3], "green+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"talkerChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			wlog(err.Error)
		}

		defer f.Close()

		if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
		return true
	}
	return false
}

// groupSaver handles which strings to write to the tell history log.
func groupSaver(str string) bool {
	res3 := regComp(str, "\\[(\\D+)\\]\\s(.+) (.+)")
	if len(res3) > 2 {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] [%s] %s %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res3[1], "blue+b"), ansi.Color(res3[2], "magenta+b"), ansi.Color(res3[3], "cyan+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			wlog(err.Error)
		}

		defer f.Close()

		if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
		return true
	} else {
		wlog(res3)
		wlog(len(res3))
	}
	return false
}

// printMessages listens on the msgchan and then filters the text. Everything not written to history files should be written to stdout.
func printMessages(msgchan <-chan string, c net.Conn) {
	fmt.Printf("\n")
	for msg := range msgchan {
		ignoreChatPrint := chatSaver(msg)
		ignoreTellPrint := tellSaver(msg)
		ignoreGroupPrint := groupSaver(msg)
		if ignoreChatPrint == false && ignoreTellPrint == false && ignoreGroupPrint == false {
			if strings.Contains(msg, "There is a sudden white flash.  Your magical shield has broken.") == true {
				msg = strings.Replace(msg, "There is a sudden white flash.  Your magical shield has broken.", ansi.Color("There is a sudden white flash.  Your magical shield has broken.", "red+bB"), -1)
			}
			fmt.Printf("%s", msg)
		}
	}
}

// wlog is the generic log function which writes and given arguments to a lot file.
func wlog(s ...interface{}) {
	f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"godisc.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err.Error())
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(s)
}
