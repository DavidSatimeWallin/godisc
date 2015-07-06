package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/user"
	"strings"
	"time"
	"regexp"

	"github.com/mgutz/ansi"
	"github.com/GeertJohan/go.linenoise"
	"github.com/pmylund/go-cache"
)

var (
	triAct     map[string]string
	hiLi     map[string]bool
)

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func returnRand() (int) {
	randMap := make(map[int]int)
	for i := 0; i < 30; i++ {
		randMap[i] = random(60, 480)
	}
	return randMap[rand.Intn(len(randMap))]
}

func goDiscInit() {
	usr, err := user.Current()
    if err != nil {
        log.Fatal( err )
    }
    goDiscCfgDir := usr.HomeDir + "/.config/godisc"
    os.Setenv("goDiscCfgDir", goDiscCfgDir + "/")
	cfgDirExists, err := exists(goDiscCfgDir)
	if err != nil {
		panic(err.Error())
	}
	if cfgDirExists == false {
		err := os.Mkdir(goDiscCfgDir,0770)
		if err != nil {
			panic(err.Error())
		}
	}
}

func main() {
	cdb := cache.New(5*time.Minute, 30*time.Second)
	cdb.Set("runIdle", 0, cache.DefaultExpiration)
	goDiscInit()
	rand.Seed( time.Now().UTC().UnixNano())
	conn, err := net.Dial("tcp", "discworld.starturtle.net:4242")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)
	msgchan := make(chan string)
	dir, found := cdb.Get("runIdle")
	if found && dir.(int) == 1 {
		ticker := time.NewTicker(time.Duration(returnRand()) * time.Second)
		quit := make(chan struct{})
		go func() {
		    for {
		       select {
		        case <- ticker.C:
		            fmt.Fprintf(conn, "hide\n")
		            fmt.Fprintf(conn, "unhide\n")
		            fmt.Fprintf(conn, "drop anaconda\n")
		            fmt.Fprintf(conn, "case anaconda\n")
		            fmt.Fprintf(conn, "peek anaconda\n")
		            fmt.Fprintf(conn, "rifle purse of anaconda\n")
		            fmt.Fprintf(conn, "palm dagger from component pouch\n")
		            fmt.Fprintf(conn, "slip dagger to component pouch\n")
		            fmt.Fprintf(conn, "get anaconda\n")
					fmt.Fprintf(conn, "look\n")
		        case <- quit:
		            ticker.Stop()
		            return
		        }
		    }
		 }()
	}

	go printMessages(msgchan, conn)
	go readKeyboardInput(conn, cdb)
	for {
		str, err := connbuf.ReadString('\n')
		if err != nil {
			break
		}
		str = triggerOn(str, conn)
		str = highLight(str)
		msgchan <- str
	}
}

func triggerOn(str string, conn net.Conn) string {
	triAct = make(map[string]string)

	triAct["enter your current character's name"] = "hate"
	triAct["Enter password:"] = "snuttefilten"
	triAct["are already playing"] = "y"

	for k, v := range triAct {
		if strings.Contains(str, k) == true {
			str = strings.Replace(str, k, ansi.Color(k, "red+b:white"), -1)
			wlog("Found", ansi.Color(k, "red+b:white"))
			fmt.Fprintf(conn, v+"\n")
		}
	}
	return str
}

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
		wlog("Could not find", os.Getenv("goDiscCfgDir") + "highlight.list")
	}
	return str
}

func delaySecond(n int) {
	time.Sleep(time.Duration(n) * time.Second)
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
		file, err := os.Open("alias.list")
		if err != nil {
		    log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			splitScan := strings.Split(scanner.Text(), "->")
		    if len(splitScan) > 1 {
		    	for _,v := range str {
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

func readKeyboardInput(c net.Conn, cdb *cache.Cache) {
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
			switch joinText {
			case "idleon":
				fmt.Println(ansi.Color("Activating idle", "red+b"))
				cdb.Set("runIdle", 1, cache.DefaultExpiration)
				linenoise.AddHistory("idleon")
				fmt.Fprintf(c, "look\n")
			case "idleoff":
				fmt.Println(ansi.Color("Deactivating idle", "red+b"))
				cdb.Set("runIdle", 0, cache.DefaultExpiration)
				linenoise.AddHistory("idleoff")
				fmt.Fprintf(c, "look\n")
			default:
				if strings.Contains(joinText, "|") {
					splitText := strings.Split(joinText, "|")
					for _,sv := range splitText {
						fmt.Fprintf(c, sv+"\n")
					}
					linenoise.AddHistory(joinText)
				} else {
					fmt.Fprintf(c, joinText+"\n")
					linenoise.AddHistory(joinText)
				}
			}
		} else {
			joinText := strings.Join(inputText, " ")
			wlog("[ CMD ]:", joinText)
			if strings.Contains(cmd, "|") {
				splitText := strings.Split(cmd, "|")
				for _,v := range splitText {
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

func quit() {
	os.Exit(0)
}

func regComp(str string, reg string) []string {
	r, _ := regexp.Compile(reg)
	res := r.FindStringSubmatch(str)
	return res
}

func tellSaver(str string) bool {
	res := regComp(str, "(You tell|You ask|You exclaim) (.+):(.+)")
	if len(res) > 1 {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] (%s) %s : %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "yellow+b"), ansi.Color(res[3], "green+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir") + "tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] (%s) %s : %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(strings.Replace(strings.TrimSpace(res2[1]), ">", "", -1), "blue+b"), ansi.Color(res2[3], "yellow+b"), ansi.Color(res2[4], "green+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir") + "tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

func chatSaver(str string) bool {
	res := regComp(str, "\\((\\D+)\\) (.+)wisps(.+)")
	if len(res) > 1 {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] (%s) %s : %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "yellow+b"), ansi.Color(res[3], "green+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir") + "talkerChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

func groupSaver(str string) bool {
	res3 := regComp(str, "\\[(.+)\\]\\s(.+) (.+)")
	if len(res3) > 2 {
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] [%s] %s %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res3[1], "blue+b"), ansi.Color(res3[2], "magenta+b"), ansi.Color(res3[3], "cyan+b"))
		f, err := os.OpenFile(os.Getenv("goDiscCfgDir") + "tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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


func printMessages(msgchan <-chan string, c net.Conn) {
	fmt.Printf("\n")
	for msg := range msgchan {
		ignoreChatPrint := chatSaver(msg)
		ignoreTellPrint := tellSaver(msg)
		ignoreGroupPrint := groupSaver(msg)
		if ignoreChatPrint == false && ignoreTellPrint == false && ignoreGroupPrint == false {
			fmt.Printf("%s", msg)
		}
	}
}

func isAcceptedExit(exit string) (accepted bool) {
	accepted = false
	switch exit {
	case "north":
		accepted = true
	case "northeast":
		accepted = true
	case "northwest":
		accepted = true
	case "south":
		accepted = true
	case "southeast":
		accepted = true
	case "southwest":
		accepted = true
	case "east":
		accepted = true
	case "west":
		accepted = true
	default:
		accepted = false
	}
	return
}

func wlog(s ...interface{}) {
	f, err := os.OpenFile(os.Getenv("goDiscCfgDir") + "godisc.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: %v", err.Error())
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(s)
}
