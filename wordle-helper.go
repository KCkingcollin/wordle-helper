package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type word struct{
	W 		string
	Rank 	uint16
}

var (
	screen 				tcell.Screen
	defaultStyle 		= tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	posIs, isIn, notIn 	string
	search 				[][26]map[uint16]uint8
	dict 				[]word
	mainTextBuf 		[][]rune
)

func input() {
	ev := screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC:
			drawText("\nCtrl+C pressed. Exiting.\n")
			time.Sleep(time.Millisecond*500)
			screen.Fini()
			os.Exit(0)
		case tcell.KeyBackspace:
			drawText(string(rune(0x08)))
			input()
		case tcell.KeyEnter:
			drawText("\n")
			return
		case tcell.KeyRune:
			drawText(string(ev.Rune()))
			input()
		}
	case *tcell.EventResize:
		_, screenSizeY := screen.Size()
		mainTextBuf = make([][]rune, screenSizeY)
		screen.Sync()
	}
}


// drawText function (modified to take screen explicitly)
func drawText(text string) {
	sizeX, sizeY := screen.Size()
	for _, r := range text {
		switch r {
		case '\n':
			slice := make([]rune, sizeX-len(mainTextBuf[sizeY-1]))
			mainTextBuf[sizeY-1] = append(mainTextBuf[sizeY-1], slice...)
			slice = make([]rune, 0)
			mainTextBuf = mainTextBuf[1:]
			mainTextBuf = append(mainTextBuf, slice)
			for x := range sizeX {
				screen.SetContent(x, sizeY-1, ' ', nil, defaultStyle)
			}
		case 0x08:
			mainTextBuf = mainTextBuf[:1]
		default:
			mainTextBuf[sizeY-1] = append(mainTextBuf[sizeY-1], r)
		}
	}
	for y := range mainTextBuf {
		for x, r := range mainTextBuf[y] {
			screen.SetContent(x, y, r, nil, defaultStyle)
		}
	}
	screen.Show()
}

func containsAny(s, targetBytes string) bool {
	if len(targetBytes) == 0 {return false}
	byteSet := make(map[rune]bool)
	for _, r := range targetBytes {
		if int(r) <= 123 && int(r) >= 97 {
			byteSet[r] = true
		}
	}
	for _, r := range s {
		if byteSet[r] {return true}
	}
	return false
}

func containsAll(s, requiredBytes string) bool {
	if len(requiredBytes) == 0 {return true}
	for i, r := range requiredBytes {
		if int(r) <= 123 && int(r) >= 97 {
			if s[i] == byte(r) {return false}
		}
	}
	foundBytes := make(map[rune]bool)
	for _, r := range requiredBytes {
		if int(r) <= 123 && int(r) >= 97 {
			foundBytes[r] = false
		}
	}

	for _, byteInS := range s {
		if _, ok := foundBytes[byteInS]; ok {
			foundBytes[byteInS] = true
		}
	}

	for _, found := range foundBytes {
		if !found {return false}
	}
	return true
}

func runSearch(notInRaw ...bool) {
		if len(posIs) <= 0 {posIs+="_"}
		size := len(search)
		if len(posIs) > size {posIs = posIs[:size]}
		if len(isIn) > size {isIn = isIn[:size]}
		list := make([]map[uint16]uint8, len(posIs))
		add2List := func(i int, j uint16) {
			ok := true
			if i > 0 {_, ok = list[i-1][j]}
			if ok && containsAll(dict[int(j)].W[:size], isIn) && !containsAny(dict[int(j)].W[:size], notIn) {
				list[i][j]++
			}
		}
		for i, r := range []rune(posIs) {
			list[i] = make(map[uint16]uint8)
			if int(r) <= 123 && int(r) >= 97 {
				for j := range search[i][int(r-97)] {
					add2List(i, j)
				}
			} else {
				for k := range 26 {
					if !containsAny(string(rune(k+97)), notIn) {
						for j := range search[i][k] {
							add2List(i, j)
						}
					}
				}
			}
		}
		output := make([]word, len(list[len(list)-1]))
		var i int
		for s := range list[len(list)-1] {
			output[i] = dict[s]
			i++
		}
		sort.Slice(output, func(i, j int) bool {
			return output[i].Rank > output[j].Rank
		})
		var textBuf string
		for _, r := range output {
			textBuf += string(r.W[:size])+"\n"
			// textBuf += string(r.W[:size])+" "
			// textBuf += fmt.Sprintln(r.Rank)
		}
		if len(notInRaw) > 0 && notInRaw[0] {
			fmt.Print(textBuf)
		} else {
			drawText(textBuf)
		}
}

func isError(err error) {
	if err != nil {
		panic(err)
	}
}

func createDict(wordSize int) {
		dict = make([]word, 0)
		resp, err := http.Get("https://raw.githubusercontent.com/dwyl/english-words/refs/heads/master/words_alpha.txt")
		isError(err)
		defer resp.Body.Close()
		tempBuf, err := io.ReadAll(resp.Body)
		isError(err)
		buf := bytes.NewReader(tempBuf)
		reader := bufio.NewReader(buf)
		resp, err = http.Get("https://raw.githubusercontent.com/david47k/top-english-wordlists/refs/heads/master/top_english_words_mixed_100000.txt")
		isError(err)
		rankTempBuf, err := io.ReadAll(resp.Body)
		isError(err)
		rankBuf := bytes.NewReader(rankTempBuf)
		rankReader := bufio.NewReader(rankBuf)
		wordRank := make(map[string]uint16)
		var i uint16
		for {
			s, err := rankReader.ReadString('\n')
			if err != nil {
				break
			}
			s = strings.TrimSpace(s)
			if len(s) == wordSize {
				i++
				wordRank[s] = i
			}
		}
		for {
			b, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			b = strings.TrimSpace(b)
			if len(b) == wordSize {
				var word word
				word.W = b
				_, ok := wordRank[word.W]
				if !ok {
					word.Rank = 65535
				} else {
					word.Rank = wordRank[word.W]
				}
				dict = append(dict, word)
			}
		}
		file, err := os.Create("dict.gob")
		defer file.Close()
		isError(err)
		encoder := gob.NewEncoder(file)
		encoder.Encode(dict)
}

func createSearch(wordSize int) {
		search = make([][26]map[uint16]uint8, wordSize)
		for i := range wordSize {
			for charNum := range 26 {
				search[i][charNum] = make(map[uint16]uint8)
				for j, r := range dict {
					if strings.Contains(string(r.W[i]), string(rune(charNum+97))) {
						search[i][charNum][uint16(j)]++
					}
				}
			}
		}
		file, err := os.Create("search.gob")
		defer file.Close()
		isError(err)
		encoder := gob.NewEncoder(file)
		encoder.Encode(search)
}

func main() {
	file, err := os.Open("dict.gob")
	defer file.Close()
	if errors.Is(err, os.ErrNotExist) {
		createDict(5)
	} else {
		decode := gob.NewDecoder(file)
		err := decode.Decode(&dict)
		isError(err)
	}

	file, err = os.Open("search.gob")
	if errors.Is(err, os.ErrNotExist) {
		createSearch(5)
	} else {
		decode := gob.NewDecoder(file)
		err := decode.Decode(&search)
		isError(err)
	}

	help := func() {
		fmt.Print(
			"Argument 1: Known letter locations, use _ in the place of any unknown characters before the search term\n"+
			"Argument 2: Characters known to be in the word, but not the position, use _ to put them in the place that know they aren't in\n"+
			"Argument 3: Characters you know aren't in the word\n"+
			"Sending a number as an argument sets the size of the word\n", 
		)
	}

	if len(os.Args) > 1 {
		var i int
		for _, a := range os.Args {
			wordSize, err := strconv.Atoi(a)
			switch {
			case a == "-h" || a == "--help":
				help()
				return
			case err == nil:
				if len(search) != wordSize {
					createDict(wordSize)
					createSearch(wordSize)
				}
			default:
				i++
			}
			switch i-1 {
			case 1:
				posIs = strings.TrimSpace(a)
			case 2:
				isIn = strings.TrimSpace(a)
			case 3:
				notIn = strings.TrimSpace(a)
			case 4:
				runSearch(true)
				fmt.Println("Too many args, attempted to use the first 3")
				return
			}
		}
		runSearch(true)
		return
	}

	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create screen: %v\n", err)
		os.Exit(1)
	}

	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize screen: %v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	screen.SetStyle(defaultStyle)
	screen.Clear()

	_, screenSizeY := screen.Size()
	mainTextBuf = make([][]rune, screenSizeY)
	for {
		drawText("Type phrase, use _ in the place of any unknown characters before the search term\nMake sure the characters are all lowercase\n")
		drawText("is in pos > ")
		input()
		drawText("is in word > ")
		input()
		drawText("not in word > ")
		input()
		runSearch()
	}
}
