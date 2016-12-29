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
	"strings"
	"time"
	"strconv"

	"github.com/GeertJohan/go.linenoise"
	"github.com/mgutz/ansi"
)

type XPObj struct {
	StartTS string
	StartXP int
	LastTS string
	LastXP int
	AverageXP int
	TotalXP int
}

var (
	cHost string = "disctemp.starturtle.net"
	cPort int = 4242
	tellSaverMaxLength int= 35
	groupSaverMaxLength int = 35
)

func main() {
	XP := XPObj{
		StartTS: "",
		StartXP: 0,
		LastTS: "",
		LastXP: 0,
		AverageXP: 0,
		TotalXP: 0,
	}
	msgchan := make(chan string)
	goDiscInit()
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", cHost, cPort))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)
	rand.Seed(time.Now().UTC().UnixNano())
	go printMessages(msgchan, conn, &XP)
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
	xpFileExists, _ := exists(os.Getenv("goDiscCfgDir") + "xp.log")
	if xpFileExists == false {
		xpFile, err := os.Create(os.Getenv("goDiscCfgDir") + "xp.log")
		if err != nil {
			wlog("could not create xp.log", err.Error())
		} else {
			wlog("created ", xpFile)
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

func getCurrentTime()string{
    t := time.Now().Local()
    return fmt.Sprintf("%s", t.Format("2006-01-02 15:04:05 +0800"))
}

func saveXp(str string, XP *XPObj) *XPObj {
	res := regComp(str, "Xp: ([0-9]+)")
	if len(res) > 1 {
		resI, err := strconv.Atoi(res[1])
		if err != nil {
			wlog(err)
		}
		if XP.StartXP == 0 {
			XP.StartXP = resI
		}
		if XP.StartTS == "" {
			XP.StartTS = getCurrentTime()
		}
		now, err := time.Parse("2006-01-02 15:04:05 +0800", getCurrentTime())
		if err != nil {
			wlog(err)
		}
		lT, err := time.Parse("2006-01-02 15:04:05 +0800", XP.StartTS)
		if err != nil {
			wlog(err)
		}
		diff := now.Sub(lT)
		m := int(diff.Minutes())
		XP.LastXP = resI
		XP.LastTS = getCurrentTime()
		XP.TotalXP = XP.LastXP - XP.StartXP
		XP.AverageXP = XP.TotalXP
		if m > 1 && XP.TotalXP > 0 {
			XP.AverageXP = ((XP.TotalXP / m) * 60)
		}
		var avS string
		switch {
			case XP.AverageXP > 9999:
				avS = fmt.Sprintf("%dK", (XP.AverageXP / 1000))
			case XP.AverageXP > 99999:
				avS = fmt.Sprintf("%dM", (XP.AverageXP / 1000000))
			default:
				avS = fmt.Sprintf("%d", XP.AverageXP)
		}
		var totS string
		switch {
			case XP.TotalXP > 9999:
				totS = fmt.Sprintf("%dK", (XP.TotalXP / 1000))
			case XP.TotalXP > 99999:
				totS = fmt.Sprintf("%dM", (XP.TotalXP / 1000000))
			default:
				totS = fmt.Sprintf("%d", XP.TotalXP)
		}
		stringToWrite := fmt.Sprintf("%s:\t%s\t\t%s:\t%s", ansi.Color("Average XP / h", "blue+b"), ansi.Color(avS, "yellow+b"), ansi.Color("Total XP", "blue+b"), ansi.Color(totS, "yellow+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir")+"xp.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			wlog(err.Error)
		}

		defer f.Close()

		if _, err = f.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
	}
	return XP
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
		if listRem == false && rmRem == false {
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
	ignoreNpcs := []string{"sailor", "seagull", "barman", "samurai", "tramp", "Mihk-gran-bohp", "engineer", "warrior", "pickpocket", "Khepresh", "smuggler", "citadel", "guard", "hopelite", "lady", "giant", "schoolboy", "farmer", "soldier", "ceremonial", "Kang Wu", "rickshaw driver", "Imperial guard", "Ryattenoki"}
	for _, v := range ignoreNpcs {
		if strings.Contains(str, v) == true {
			return true
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

// printMessages listens on the msgchan and then filters the text. Everything not written to history files should be written to stdout.
func printMessages(msgchan <-chan string, c net.Conn, XP *XPObj) {
	fmt.Printf("\n")
	for msg := range msgchan {
		if len(msg) > 1 {

			// Parse msg to see if it should be written to a file instead of being printed.
			ignoreTaxiPrint := taxiSaver(msg)
			ignoreChatPrint := chatSaver(msg)
			ignoreTellPrint := tellSaver(msg)
			ignoreGroupPrint := groupSaver(msg)
			rememberSaver(msg)
			XP = saveXp(msg, XP)
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
