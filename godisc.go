package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/mgutz/ansi"
	cache "github.com/patrickmn/go-cache"
	"github.com/stesla/gotelnet"
	linenoise "pkg.re/essentialkaos/go-linenoise.v3"
)

type (
	XPObj struct {
		StartTS          string
		StartXP          int
		LastTS           string
		LastXP           int
		AverageXP        int
		TotalXP          int
		HighestAverageXP int
	}
	Connection struct {
		Host string
		Port int
	}
)

var (
	connections         []Connection
	tellSaverMaxLength  int = 35
	groupSaverMaxLength int = 35
	CLOGFile            *os.File
	WLOGFile            *os.File
	TellChatFile        *os.File
	AliasFile           *os.File
	HighlightFile       *os.File
	C                   *cache.Cache
	Clubs               *cache.Cache
	ChatItems           string
	conn                gotelnet.Conn
	err                 error
)

const (
	AliasDynamicPlaceholder = "##"
)

func init() {

	connections = []Connection{
		Connection{
			Host: "discworld.starturtle.net",
			Port: 4242,
		},
		Connection{
			Host: "disctemp.starturtle.net",
			Port: 23,
		},
	}

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
	excludeFileExists, _ := exists(os.Getenv("goDiscCfgDir") + "exclude.list")
	if excludeFileExists == false {
		excludeFile, err := os.Create(os.Getenv("goDiscCfgDir") + "exclude.list")
		if err != nil {
			wlog("could not create exclude.list", err.Error())
		} else {
			wlog("created ", excludeFile)
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

func main() {
	XP := XPObj{
		StartTS:          "",
		StartXP:          0,
		LastTS:           "",
		LastXP:           0,
		AverageXP:        0,
		TotalXP:          0,
		HighestAverageXP: 0,
	}

	C = cache.New(5*time.Minute, 10*time.Minute)
	cacheHighlights()

	Clubs = cache.New(5*time.Minute, 10*time.Minute)
	cacheClubNames()

	addTalkersToClubCache()

	var tmpItems []string
	for k, _ := range Clubs.Items() {
		tmpItems = append(tmpItems, k)
	}
	ChatItems = strings.Join(tmpItems, "|")
	fmt.Println("Now have", ChatItems)

	var err error
	CLOGFile, err = os.OpenFile(os.Getenv("goDiscCfgDir")+"clog", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		wlog(err.Error)
	}
	WLOGFile, err = os.OpenFile(os.Getenv("goDiscCfgDir")+"log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		wlog(err.Error)
	}
	TellChatFile, err = os.OpenFile(os.Getenv("goDiscCfgDir")+"tellChat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		wlog(err.Error)
	}

	defer CLOGFile.Close()
	defer WLOGFile.Close()
	defer TellChatFile.Close()

	msgchan := make(chan string)

	for _, connection := range connections {
		hostname := fmt.Sprintf("%s:%d", connection.Host, connection.Port)
		log.Println("connecting to", hostname)
		conn, err = gotelnet.Dial(hostname)
		if err != nil {
			log.Println("could not connect to", hostname, err)
			continue
		}
		break
	}
	if err != nil {
		fmt.Println("We ran out of hosts. Here's the last error we encountered.", err)
		os.Exit(1)
	}

	connbuf := bufio.NewReader(conn)
	go printMessages(msgchan, conn, &XP)
	go readKeyboardInput(conn)
	for {
		str, _ := connbuf.ReadString('\n')
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

func addTalkersToClubCache() {
	talkers := []string{
		"one",
		"two",
		"dead",
		"Wizards",
		"playtesters",
		"Adventurers",
		"playerkillers",
		"Catfish",
		"Warriors",
		"A'Tuin",
		"Apex",
		"Priests",
		"Thieves",
		"Witches",
		"Pishe",
		"Ankh-MorporkCouncil",
		"Ankh-MorporkCouncilMagistrate",
		"HublandishBarbarians",
		"igame",
		"inews",
		"Apex",
		"Fish",
		"Hat",
		"Sek",
		"Gufnork",
		"Gapp",
		"Sandelfon",
		"ConlegiumSicariorum",
		"Assassins",
		"theAgateanEmpireCouncil",
		"DjelibeybiCouncil",
		"Hashishim",
		"DjelibeybiCouncilMagistrate",
		"Quiz",
		"LancreHighlandRegiment",
		"Ninjas",
		"Witches",
	}
	for _, talker := range talkers {
		log.Println("Saving", talker)
		Clubs.Set(talker, nil, cache.NoExpiration)
	}
}

func cacheClubNames() {
	resp, err := soup.Get("http://discworld.starturtle.net/lpc/playing/clubs.c?type=club")
	if err != nil {
		os.Exit(1)
	}
	doc := soup.HTMLParse(resp)
	list := doc.Find("div", "id", "content").Find("ul").FindAll("li")
	for _, item := range list {
		link := item.Find("a")
		club := strings.Replace(link.Text(), " ", "_", -1)
		club = strings.ToLower(club)
		log.Println("Saving", link.Text())
		log.Println("Saving", club)
		Clubs.Set(link.Text(), nil, cache.NoExpiration)
		Clubs.Set(club, nil, cache.NoExpiration)
	}
}

func cacheHighlights() {
	var (
		err   error
		color string
		row   string
	)
	HighlightFile, err = os.Open(os.Getenv("goDiscCfgDir") + "highlight.list")
	if err != nil {
		wlog("could not open highlight.list -file", err.Error())
	}
	scanner := bufio.NewScanner(HighlightFile)
	for scanner.Scan() {
		color = "red"
		row = scanner.Text()
		if strings.Contains(row, ";;") || len(row) < 1 {
			continue
		}
		if strings.Contains(row, "#") {
			exp := strings.Split(row, "#")
			if len(exp) > 1 {
				row = exp[0]
				color = exp[1]
			}
		}
		log.Println("Saving", row, "with color", color)
		C.Set(row, color, cache.NoExpiration)
	}
}

func highLight(str string) string {
	cachedItems := C.Items()

	for k, v := range cachedItems {
		str = strings.Replace(str, k, ansi.Color(k, fmt.Sprintf("%s+b", v.Object)), -1)
	}
	return str

}

func findAlias(str []string) string {
	if len(str) < 1 {
		return "none"
	}

	var err error
	var b []byte
	var multipleInputVars bool

	if len(str) > 1 {
		multipleInputVars = true
	}

	b, err = ioutil.ReadFile(os.Getenv("goDiscCfgDir") + "alias.list")

	if err != nil {
		wlog(err.Error)
	}

	var fileContent string = string(b)

	var lines []string = strings.Split(fileContent, "\n")

	if len(lines) > 0 {
		for _, v := range lines {
			ex := strings.Split(v, "->")
			if len(ex) > 1 {
				if strings.Contains(ex[1], AliasDynamicPlaceholder) && multipleInputVars {
					var repStr string
					for i := 1; i < len(str); i++ {
						repStr = fmt.Sprintf("%s %s", repStr, str[i])
					}
					ex[1] = strings.Replace(ex[1], AliasDynamicPlaceholder, repStr, -1)
				}
			}
			if multipleInputVars {
				for _, v := range str {
					v = strings.TrimSpace(v)
					if len(v) >= len(ex[0]) {
						if v[0:len(ex[0])] == ex[0] && len(ex) > 1 {
							return ex[1]
						}
					}
				}
			} else {
				s := strings.TrimSpace(str[0])
				if s[0:len(s)] == ex[0] {
					return ex[1]
				}
			}
		}
	}
	return "none"
}

func getCurrentTime() string {
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
		if XP.AverageXP > XP.HighestAverageXP {
			XP.HighestAverageXP = XP.AverageXP
		}
		var hAvS string
		switch {
		case XP.HighestAverageXP > 9999:
			hAvS = fmt.Sprintf("%dK", (XP.HighestAverageXP / 1000))
		case XP.HighestAverageXP > 99999:
			hAvS = fmt.Sprintf("%dM", (XP.HighestAverageXP / 1000000))
		default:
			hAvS = fmt.Sprintf("%d", XP.HighestAverageXP)
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
		stringToWrite := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n%s\t%s\n%s\t%s\n%s\t%s", ansi.Color("Average XP / h", "blue+b"), ansi.Color(avS, "yellow+b"), ansi.Color("Highest Average XP / h", "blue+b"), ansi.Color(hAvS, "yellow+b"), ansi.Color("Total XP", "blue+b"), ansi.Color(totS, "yellow+b"))

		if err = ioutil.WriteFile(os.Getenv("goDiscCfgDir")+"xp.log", []byte(stringToWrite), 0664); err != nil {
			wlog(err.Error)
		}
	}
	return XP
}

func readKeyboardInput(c net.Conn) {
	for {
		str, err := linenoise.Line("")
		wlog(str)
		if err != nil {
			if err != linenoise.ErrKillSignal {
				fmt.Printf("Unexpected error: %s\n", err)
			}
			quit()
		}
		inputText := strings.Fields(str)
		joinText := strings.Join(inputText, " ")
		cmd := findAlias(inputText)
		if cmd == "none" {
			if strings.Contains(joinText, "|") {
				splitText := strings.Split(joinText, "|")
				for _, sv := range splitText {
					fmt.Fprintf(c, sv+"\n")
				}
				linenoise.AddHistory(joinText)
				clog(joinText)
			} else {
				fmt.Fprintf(c, joinText+"\n")
				linenoise.AddHistory(joinText)
				clog(joinText)
			}
		} else {
			if strings.Contains(cmd, "|") {
				splitText := strings.Split(cmd, "|")
				for _, v := range splitText {
					fmt.Fprintf(c, v+"\n")
					linenoise.AddHistory(joinText)
					clog(joinText)
				}
			} else {
				//fmt.Fprintf(c, cmd+"\n")
				_, err := c.Write([]byte(cmd + "\n"))
				wlog(err)
				linenoise.AddHistory(joinText)
				clog(joinText)
			}
		}
	}
}
func quit() {
	os.Exit(0)
}
func regComp(str string, reg string) []string {
	r, err := regexp.Compile(reg)
	if err != nil {
		fmt.Println("Could not compile regex", err)
	}
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
	ignoreNpcs := []string{"sailor", "Renee Palm", "The midnight hag", "deckhand", "Mr Werks", "Sister Rhelin", "Slim Stevie", "rogue", "nameless man", "obnoxious beggar", "silversmith", "urchin", "heavy", "Jones", "seagull", "barman", "samurai", "tramp", "Mihk-gran-bohp", "engineer", "warrior", "pickpocket", "Khepresh", "smuggler", "citadel", "guard", "hopelite", "hoplite", "poet", "lady", "giant", "schoolboy", "farmer", "soldier", "ceremonial", "Kang Wu", "rickshaw driver", "Imperial guard", "Ryattenoki", "Kyakenko", "actor", "youth"}
	for _, v := range ignoreNpcs {
		if strings.Contains(str, v) {
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
		if _, err := TellChatFile.WriteString(stringToWrite + "\n"); err != nil {
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
			if _, err := TellChatFile.WriteString(stringToWrite + "\n"); err != nil {
				wlog(err.Error)
			}
			return true
		}
	}
	return false
}

// chatSaver handles which strings to write to the talker history log.
func chatSaver(str string) bool {
	for cl, _ := range Clubs.Items() {
		if strings.Contains(str, fmt.Sprintf("(%s)", cl)) {
			str = strings.Replace(str, ": ", " ", -1)
			pattern := fmt.Sprintf(`(\(%s\)) ([a-zA-Z]+) (.+)`, cl)
			res := regComp(str, pattern)
			if len(res) > 2 {
				var stringToWrite string
				t := time.Now()
				stringToWrite = fmt.Sprintf("[ %d:%d:%d ] %s %s %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "yellow+b"), ansi.Color(res[3], "green+b"))
				if _, err := TellChatFile.WriteString(stringToWrite + "\n"); err != nil {
					wlog(err.Error)
				}
				return true
			}
		}
	}
	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// groupSaver handles which strings to write to the tell history log.
func groupSaver(str string) bool {
	res := regComp(str, "\\[(.+)\\](.+) (.+)")
	if len(res) > 2 && len(res) < groupSaverMaxLength {
		if len(res[1]) < 3 || strings.Contains(res[0], "Ench 5") || strings.Contains(res[2], "no destination") || strings.Contains(res[2], "is mounted on") {
			return false
		}
		if strings.Contains(res[1], "Discworld") || strings.Contains(res[0], "----------------------") || strings.Contains(res[2], "----------------------") {
			return false
		}
		exclude := []string{"job", "from", "for", "here", "main", "quests", "all", "details", "none"}
		if contains(exclude, res[1]) || contains(exclude, res[0]) {
			return false
		}
		var stringToWrite string
		t := time.Now()
		stringToWrite = fmt.Sprintf("[ %d:%d:%d ] [%s] %s %s", t.Hour(), t.Minute(), t.Second(), ansi.Color(res[1], "blue+b"), ansi.Color(res[2], "magenta+b"), ansi.Color(res[3], "cyan+b"))
		if _, err := TellChatFile.WriteString(stringToWrite + "\n"); err != nil {
			wlog(err.Error)
		}
		return true
	}
	return false
}

// printMessages listens on the msgchan and then filters the text. Everything not written to history files should be written to stdout.
func printMessages(msgchan <-chan string, c net.Conn, XP *XPObj) {
	fmt.Printf("\n")
	for msg := range msgchan {
		if len(msg) < 1 {
			fmt.Printf("%s", msg)
		} else {
			clog(msg)
			// Parse msg to see if it should be written to a file instead of being printed.
			ignoreChatPrint := chatSaver(msg)
			ignoreTellPrint := tellSaver(msg)
			ignoreGroupPrint := groupSaver(msg)
			if strings.Contains(msg, "letMeResetCounter") || XP.AverageXP < 0 || XP.TotalXP < 0 {
				newXP := XPObj{
					StartTS:          "",
					StartXP:          0,
					LastTS:           "",
					LastXP:           0,
					AverageXP:        0,
					TotalXP:          0,
					HighestAverageXP: 0,
				}
				XP = &newXP
			}
			XP = saveXp(msg, XP)
			// If the three Print filters above are all false print msg to screen.
			if !ignoreChatPrint && !ignoreTellPrint && !ignoreGroupPrint {
				fmt.Printf("%s", msg)
			}
		}
	}
}

// clog is the function that logs EVERYTHING!!
func clog(s ...interface{}) {
	t := time.Now()
	stringToWrite := fmt.Sprintf("[ %d-%d-%d %d:%d:%d ]", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	if _, err := CLOGFile.WriteString(stringToWrite + fmt.Sprintf(" %s", s...)); err != nil {
		wlog(err.Error)
	}
}

// wlog is the generic log function which writes and given arguments to a lot file.
func wlog(s ...interface{}) {
	WLOGFile.WriteString(fmt.Sprintf(" %s", s...))
}
