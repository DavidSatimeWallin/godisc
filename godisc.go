package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mgutz/ansi"
	"github.com/pmylund/go-cache"
)

var (
	triAct     map[string]string
	hiLi     map[string]bool
	foundExits map[int]string
	killNpcs   []string
	unwantedNpcsInRoom []string
)

func tickr(conn net.Conn) {
    ticker := time.NewTicker(time.Minute * time.Duration(random(4,8)))
    go func() {
        for t := range ticker.C {
            wlog("Tick at", t)
            fmt.Fprintf(conn, "look\n")
        }
    }()
}

func fightingTimer(conn net.Conn, cdb *cache.Cache) {
	timer := time.NewTimer(time.Second * time.Duration(random(40, 70)))
    go func() {
    	fmt.Println(ansi.Color("Started fighting timer", "blue+b"))
        <- timer.C
        cdb.Set("fighting", 0, cache.NoExpiration)
        cdb.Set("haveAttacked", 0, cache.NoExpiration)
        fmt.Fprintf(conn, "ba\n")
        fmt.Fprintf(conn, "look\n")
    }()
}

func main() {
	rand.Seed( time.Now().UTC().UnixNano())
	cdb := cache.New(5*time.Minute, 30*time.Second)
	cdb.Set("amove", 0, cache.NoExpiration)
	cdb.Set("fighting", 0, cache.NoExpiration)
    cdb.Set("haveAttacked", 0, cache.NoExpiration)
	conn, err := net.Dial("tcp", "discworld.starturtle.net:4242")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)
	msgchan := make(chan string)
	go printMessages(msgchan, conn)
	go readKeyboardInput(conn, cdb)
	go tickr(conn)
	for {
		var attacked bool
		var wantedNpcsString string
		str, err := connbuf.ReadString('\n')
		if err != nil {
			break
		}
		str = triggerOn(str, conn)
		str = highLight(str)
		// if strings.Contains(str, "obvious exit") == true {
			wlog("Contains it", str)
			fStr, foundExits := handleStream(str, conn)
			msgchan <- fStr
			toWalk := doWalking(foundExits, cdb, conn)
			dir, _ := cdb.Get("amove")
			if toWalk != "none" && dir.(int) == 1{
				wlog("Searching", str)
				re3, _ := regexp.Compile("(.+) (is|are) standing here")
				result3 := re3.FindStringSubmatch(fStr)
				if len(result3) > 0 {
					attacked, wantedNpcsString = hunting(result3[1], conn, cdb)
				} else {
					wlog("Found nothing in", fStr)
				}
				wlog("Attacked", attacked)
				wlog("wantedNpcsString", wantedNpcsString)
				fdir, _ := cdb.Get("fighting")
				if attacked == true || fdir.(int) == 1 {
						cdb.Set("fighting", 1, cache.NoExpiration)
						adir, afound := cdb.Get("haveAttacked")
						if !afound || adir.(int) == 0 {
							cdb.Set("haveAttacked", 1, cache.NoExpiration)
							fmt.Fprintf(conn, "k " + wantedNpcsString + "\n")
							go fightingTimer(conn, cdb)
						}
					} else {
						walk(toWalk, conn, cdb, 0)
					}
			}
			wlog(toWalk)
		// } else {
		// 	msgchan <- str
		// }
	}
}

func cameFrom(walked string) (cameFrom string) {
	wlog(ansi.Color(walked, "green+b"))
	switch walked {
	case "north":
		cameFrom = "south"
	case "northeast":
		cameFrom = "southwest"
	case "northwest":
		cameFrom = "southeast"
	case "south":
		cameFrom = "north"
	case "southeast":
		cameFrom = "northwest"
	case "southwest":
		cameFrom = "northeast"
	case "east":
		cameFrom = "west"
	case "west":
		cameFrom = "east"
	default:
		cameFrom = "backwards"
	}
	return
}

func removeCameFrom(came string, foundExits map[int]string) (map[int]string) {
	counter := 0
	newFoundExits := make(map[int]string)
	for _,v := range foundExits {
		if v != came {
			newFoundExits[counter] = v
			counter++
		}
	}
	return newFoundExits
}

func walk(toWalk string, conn net.Conn, cdb *cache.Cache, killWindow int) {
	go func() {
		fdir, _ := cdb.Get("fighting")
		hadir, _ := cdb.Get("haveAttacked")
			if fdir.(int) == 0 && hadir.(int) == 0 {
			if killWindow == 0 {
					delaySecond(random(2,6))
				} else {
					delaySecond(random(60,90))
				}
				cdb.Set("lastWalked", toWalk, cache.NoExpiration)
				fmt.Fprintf(conn, toWalk+"\n")
		}
	}()
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
	hiLi = make(map[string]bool)

	hiLi["bodyguards"] = true
	hiLi["bodyguard"] = true
	hiLi["merchants"] = true
	hiLi["merchant"] = true
	hiLi["dealers"] = true
	hiLi["dealer"] = true
	hiLi["traders"] = true
	hiLi["trader"] = true
	hiLi["noblemen"] = true
	hiLi["nobleman"] = true
	hiLi["samurais"] = true
	hiLi["samurai"] = true
	hiLi["warriors"] = true
	hiLi["warrior"] = true
	hiLi["Imperial Guards"] = true
	hiLi["Imperial Guard"] = true
	hiLi["judges"] = true
	hiLi["judge"] = true

	for k, v := range hiLi {
		if strings.Contains(str, k) == true && v == true {
			str = strings.Replace(str, k, ansi.Color(k, "red+b"), -1)
			wlog("Found", ansi.Color(k, "red+b"))
		}
	}
	return str
}

func delaySecond(n int) {
	time.Sleep(time.Duration(n) * time.Second)
}

func random(min, max int) int {
	return min + rand.Intn(max-min)
}

func calcNumExits(r string) int {
	switch r {
	case "one":
		return 1
	case "two":
		return 2
	case "three":
		return 3
	case "four":
		return 4
	case "five":
		return 5
	case "six":
		return 6
	case "seven":
		return 7
	case "eight":
		return 8
	case "nine":
		return 9
	case "ten":
		return 10
	default:
		return 999
	}
}

func checkResetWalk() bool {
	var chance map[int]bool
	chance = make(map[int]bool)
	chance[0] = false
	chance[1] = false
	chance[2] = false
	chance[3] = false
	chance[4] = false
	chance[5] = false
	chance[6] = false
	chance[7] = false
	chance[8] = false
	chance[9] = false
	if len(chance) < 1 {
		return false
	}
	return chance[rand.Intn(len(chance))]
}

func doWalking(foundExits map[int]string, cdb *cache.Cache, c net.Conn) (string) {
	wlog(len(foundExits))
	var lastWalked string
	var toWalk string
	dir, found := cdb.Get("lastWalked")
	if found {
		lastWalked = dir.(string)
	} else {
		lastWalked = "none"
	}
	wlog(lastWalked)
	came := cameFrom(lastWalked)
	foundExits = removeCameFrom(came, foundExits)
	wlog("Removed", came)
	wlog("Found:", foundExits)

	lWCheck := mapSearch(lastWalked, foundExits)
	if lWCheck == true {
		toWalk = lastWalked
	} else {
		if len(foundExits) > 0 {
			toWalk = foundExits[rand.Intn(len(foundExits))]
		} else {
			wlog("Could not find any exits!", foundExits)
		}
	}
	return toWalk
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

func hunting(resultString string, c net.Conn, cdb *cache.Cache) (bool, string) {
	attacked := false
	killNpcs := []string{}
	killNpcs = append(killNpcs, "child")
	killNpcs = append(killNpcs, "merchant")
	killNpcs = append(killNpcs, "trader")
	killNpcs = append(killNpcs, "farmer")
	// killNpcs = append(killNpcs, "woman")
	// killNpcs = append(killNpcs, "judge")
	// killNpcs = append(killNpcs, "dressmaker")
	// killNpcs = append(killNpcs, "rich woman")
	// killNpcs = append(killNpcs, "warrior")

	unwantedNpcsInRoom := []string{}
	unwantedNpcsInRoom = append(unwantedNpcsInRoom, "Royal Guard")
	unwantedNpcsInRoom = append(unwantedNpcsInRoom, "brawler")
	unwantedNpcsInRoom = append(unwantedNpcsInRoom, "brawler")
	foundUnwantedNpcInRoom := false
	var wantedNpcsString string

	wantedNpcs := []string{}

	foundNpcs := cleanNpcString(resultString)
	// foundNpcsArray := strings.Split(foundNpcs, ", ")

	for _, un := range unwantedNpcsInRoom {
		// wlog("Looking for unwanted npcs in room")
		if strings.Contains(foundNpcs, un) == true {
			foundUnwantedNpcInRoom = true
		}
	}

	for _, sv := range killNpcs {
		wlog("Looking for", sv, "in", foundNpcs)
		// wlog(strings.Contains(foundNpcs, sv))
		if strings.Contains(foundNpcs, sv) == true {
			wantedNpcs = append(wantedNpcs, sv)
		}
	}
	// wlog("I found", foundNpcsArray)
	// wlog("I wanna kill", wantedNpcs)
	// wlog("Did i find unwanted npcs in room?", foundUnwantedNpcInRoom)
	if len(wantedNpcs) > 0 && foundUnwantedNpcInRoom == false {
		wantedNpcsString = strings.Join(wantedNpcs, "&")
		attacked = true
	}
	return attacked, wantedNpcsString
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

func handleStream(str string, c net.Conn) (fStr string, foundExits map[int]string) {

	var numExits int
	foundExits = make(map[int]string)

	// Find the string representation of how many exits are in the room
	re1, _ := regexp.Compile("There (is|are) [a-z]+ obvious (exit|exits): (.+).")
	result := re1.FindStringSubmatch(str)
	if len(result) > 0 {
		numExits = calcNumExits(result[2])
		wlog("Number of exits found:", numExits)
		wlog(result[3])
		myStringArray := strings.Split(strings.Replace(strings.Replace(strings.Replace(result[3], ".", "", -1), " and ", ", ", -1), " ", "", -1), ",")
		mCount := 0
		for _, v := range myStringArray {
			subre, _ := regexp.Compile("(^\\w+)")
			vRes := subre.FindStringSubmatch(v)
			rV := vRes[1]
			b := isAcceptedExit(rV)
			if b == true {
				foundExits[mCount] = rV
				mCount++
			} else {
				wlog("This is not acceptable:", rV)
			}
		}
		wlog("Acceptable exits found:", foundExits)
	}
	fStr = str
	return
}

func readKeyboardInput(c net.Conn, cdb *cache.Cache) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inputText := scanner.Text()
		switch inputText {
		case "amove":
			cdb.Set("amove", 1, cache.NoExpiration)
			fmt.Println(ansi.Color("Activated move", "red+b"))
		case "demove":
			cdb.Set("amove", 0, cache.NoExpiration)
			fmt.Println(ansi.Color("Deactivated move", "red+b"))
		default:
			if strings.Contains(inputText, "|") {
				splitText := strings.Split(inputText, "|")
				for _,v := range splitText {
					fmt.Fprintf(c, v+"\n")
				}
			} else {
				fmt.Fprintf(c, scanner.Text()+"\n")
			}
		}
	}
}

func printMessages(msgchan <-chan string, c net.Conn) {
	for msg := range msgchan {
		fmt.Printf("%s", msg)
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
	f, err := os.OpenFile("godisc.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: %v", err.Error())
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(s)
}
