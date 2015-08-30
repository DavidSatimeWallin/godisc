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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GeertJohan/go.linenoise"
	"github.com/mgutz/ansi"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultConnectionPort = "4242" // Kingpin wants this as string and not int
	goDiscVersion         = "0.1"
	tellSaverMaxLength    = 35
	groupSaverMaxLength   = 35
)

var (
	app         = kingpin.New("chat", "A command-line chat application.")
	connectHost = app.Arg("host", "The IP/Domain to connect to.").Required().String()
	connectPort = app.Arg("port", "Port to connect to.").Default(defaultConnectionPort).String()
	debug       = app.Flag("debug", "Set to true to see debug information.").Bool()
)

func main() {
	kingpin.Version(goDiscVersion)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	msgchan := make(chan string)
	goDiscInit()
	conn, err := net.Dial("tcp", *connectHost+":"+*connectPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)
	rand.Seed(time.Now().UTC().UnixNano())
	go printMessages(msgchan, conn)
	go readKeyboardInput(conn)
	for {
		str, err := connbuf.ReadString('\n')
		if err != nil {
			break
		}
		str = highLight(str)
		msgchan <- str
	}
}
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
func returnRand() int {
	randMap := make(map[int]int)
	for i := 0; i < 30; i++ {
		randMap[i] = random(60, 480)
	}
	return randMap[rand.Intn(len(randMap))]
}
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
	highLightListExists, _ := exists(os.Getenv("goDiscCfgDir") + "highlight.list")
	if highLightListExists == false {
		hiLiFile, err := os.Create(os.Getenv("goDiscCfgDir") + "highlight.list")
		if err != nil {
			wlog("could not create highlight.list", err.Error())
		} else {
			wlog("created ", hiLiFile)
		}
	}
	aliasListExists, _ := exists(os.Getenv("goDiscCfgDir") + "alias.list")
	if aliasListExists == false {
		aliasFile, err := os.Create(os.Getenv("goDiscCfgDir") + "alias.list")
		if err != nil {
			wlog("could not create alias.list", err.Error())
		} else {
			wlog("created ", aliasFile)
		}
	}
}
func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}
func highLight(str string) string {
	highLightListExists, _ := exists(os.Getenv("goDiscCfgDir") + "highlight.list")
	if highLightListExists == true {
		file, err := os.Open(os.Getenv("goDiscCfgDir") + "highlight.list")
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
func random(min, max int) int {
	return min + rand.Intn(max-min)
}
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
func findAlias(str []string) string {
	aliasListExists, _ := exists(os.Getenv("goDiscCfgDir") + "alias.list")
	if aliasListExists == true {
		file, err := os.Open(os.Getenv("goDiscCfgDir") + "alias.list")
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
func rmRemembers(str string) bool {
	res := regComp(str, "(.rmRem)")
	if len(res) > 1 {
		rememberListExists, _ := exists(os.Getenv("goDiscCfgDir") + "remember.log")
		if rememberListExists == true {
			err := os.Remove(os.Getenv("goDiscCfgDir") + "remember.log")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Removed remember.log\n")
			return true
		}
	}
	return false
}
func listRemembers(str string, c net.Conn) bool {
	res := regComp(str, "(.listRem)")
	if len(res) > 1 {
		rememberListExists, _ := exists(os.Getenv("goDiscCfgDir") + "remember.log")
		if rememberListExists == true {
			file, err := os.Open(os.Getenv("goDiscCfgDir") + "remember.log")
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			var lines []string
			var output string
			for scanner.Scan() {
				lines = append(lines, ansi.Color(scanner.Text(), "cyan+bh"))
			}
			RemoveDuplicates(&lines)
			sort.Strings(lines)
			output = strings.Join(lines, " - ")
			fmt.Printf("%d REMEMBERS: %s", len(lines), output+"\n")
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
			return true
		}
	}
	return false
}
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
		joinText := strings.Join(inputText, " ")
		listRem := listRemembers(joinText, c)
		rmRem := rmRemembers(joinText)
		checkForSleep := checkForSleeper(joinText, c)
		if checkForSleep == false && listRem == false && rmRem == false {
			cmd := findAlias(inputText)
			if cmd == "none" {
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
}
func quit() {
	os.Exit(0)
}
func regComp(str string, reg string) []string {
	r, _ := regexp.Compile(reg)
	res := r.FindStringSubmatch(str)
	return res
}
func clearTellSaver(str string) bool {
	str = strings.Replace(str, "[37m", "", -1)
	str = strings.Replace(str, "[1m", "", -1)
	if len(str) > tellSaverMaxLength {
		return true
	}
	if len(str) > 6 {
		if strings.Contains(str[0:6], "The ") == true {
			return true
		}
		if strings.Contains(str[0:6], "One ") == true {
			return true
		}
	}
	if len(str) > 12 {
		if strings.Contains(str[0:12], "On the") == true {
			return true
		}
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

// taxiSaver handles which strings to write to the talker history log.
func taxiSaver(str string) bool {
	res := regComp(str, "(\\(Taxi\\)) ([a-zA-Z0-9 ]+): ([a-zA-Z0-9 ]+)")
	wlog(fmt.Sprintf("Taxisaver: %+v", res))
	if len(res) > 1 {
		var stringToWrite string
		stringToWrite = fmt.Sprintf("[ %s ] %s : %s", ansi.Color("TAXI", "red+b"), ansi.Color(res[2], "yellow+b"), ansi.Color(res[3], "cyan+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"talkerChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			wlog(err.Error)
		}

		defer f.Close()

		if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
		return true
	} else {
		wlog(fmt.Sprintf("Found %d reg responses to Taxi", len(res)))
	}
	return false
}

// groupSaver handles which strings to write to the tell history log.
func groupSaver(str string) bool {
	res := regComp(str, "\\[(.+)\\]\\s(.+) (.+)")
	if len(res) > 2 && len(res) < groupSaverMaxLength {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] [%s] %s %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "magenta+b"), ansi.Color(res[3], "cyan+b"))
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
	wlog(res)
	wlog(len(res))
	return false
}

// rememberSaver handles which strings to write to the tell remember log.
func rememberSaver(str string) bool {
	res := regComp(str, "identified as \"(.+)\"")
	if len(res) > 1 {
		var stringToWrite string
		stringToWrite = fmt.Sprintf("%s", ansi.Color(res[1], "magenta+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"remember.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			wlog(err.Error)
		}

		defer f.Close()

		if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
		return true
	}
	wlog(res)
	wlog(len(res))
	return false
}

func checkForSleeper(str string, c net.Conn) bool {
	res := regComp(str, "setSleeper\\(([0-9]+), (.+)\\)")
	if len(res) > 1 {
		wlog("Found setSleeper", res[1], res[2])
		go func(dur string, act string, c net.Conn) {
			intDur, convErr := strconv.Atoi(dur)
			if convErr != nil {
				wlog(convErr.Error())
			}
			wlog("Waiting for", dur, "minutes and then running", act)
			time.Sleep(time.Duration(intDur) * time.Minute)
			fmt.Fprintf(c, act+"\n")
		}(res[1], res[2], c)
		return true
	}
	return false
}

// printMessages listens on the msgchan and then filters the text. Everything not written to history files should be written to stdout.
func printMessages(msgchan <-chan string, c net.Conn) {
	fmt.Printf("\n")
	for msg := range msgchan {
		if len(msg) > 1 {
			if *debug == true {
				wlog(ansi.Color(fmt.Sprintf("%s", msg), "white+B:red+h"))
			}

			// Parse msg to see if it should be written to a file instead of being printed.
			ignoreTaxiPrint := taxiSaver(msg)
			ignoreChatPrint := chatSaver(msg)
			ignoreTellPrint := tellSaver(msg)
			ignoreGroupPrint := groupSaver(msg)
			rememberSaver(msg)
			// If the three Print filters above are all false print msg to screen.
			if ignoreTaxiPrint == false && ignoreChatPrint == false && ignoreTellPrint == false && ignoreGroupPrint == false {
				if strings.Contains(msg, "There is a sudden white flash.  Your magical shield has broken.") == true {
					msg = strings.Replace(msg, "There is a sudden white flash.  Your magical shield has broken.", ansi.Color("There is a sudden white flash.  Your magical shield has broken.", "red+bB"), -1)
				}
				fmt.Printf("%s", msg)
			}
		} else {
			fmt.Printf("%s", msg)
		}
	}
}

// wlog is the generic log function which writes and given arguments to a lot file.
func wlog(s ...interface{}) {
	f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err.Error())
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(s)
}
