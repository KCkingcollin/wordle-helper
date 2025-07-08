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
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type word struct{
	W 		[5]byte
	Rank 	uint16
}

var (
	screen 				tcell.Screen
	defaultStyle 		= tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	posIs, isIn, notIn 	string
	search 				[5][26]map[uint16]uint8
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
		case tcell.KeyUp:
			// goUpInLog()
			drawText("Up arrow pressed!\n")
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

func goUpInLog() {
	posIs = "test"
	isIn = "test"
	notIn = "test"
	drawText(fmt.Sprintln(posIs, isIn, notIn))
}

func containsAny(s, targetBytes []byte) bool {
	if len(targetBytes) == 0 {return false}
	byteSet := make(map[byte]bool)
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

func containsAll(s, requiredBytes []byte) bool {
	if len(requiredBytes) == 0 {return true}
	for i, r := range requiredBytes {
		if int(r) <= 123 && int(r) >= 97 {
			if s[i] == r {return false}
		}
	}
	foundBytes := make(map[byte]bool)
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
		size := len(posIs)
		list := make([]map[uint16]uint8, size)
		for i, r := range []rune(posIs) {
			list[i] = make(map[uint16]uint8)
			if int(r) <= 123 && int(r) >= 97 {
				for j, _ := range search[i][int(r-97)] {
					ok := true
					if i > 0 {_, ok = list[i-1][j]}
					if ok && containsAll(dict[int(j)].W[:5], []byte(isIn)) && !containsAny(dict[int(j)].W[:5], []byte(notIn)) {
						list[i][j]++
					}
				}
			} else {
				for k := range 26 {
					if !containsAny([]byte{byte(k+97)}, []byte(notIn)) {
						for j, _ := range search[i][k] {
							ok := true
							if i > 0 {_, ok = list[i-1][j]}
							if ok && containsAll(dict[int(j)].W[:5], []byte(isIn)) && !containsAny(dict[int(j)].W[:5], []byte(notIn)) {
								list[i][j]++
							}
						}
					}
				}
			}
		}
		output := make([]word, len(list[len(list)-1]))
		var i int
		for s, _ := range list[len(list)-1] {
			output[i] = dict[s]
			i++
		}
		sort.Slice(output, func(i, j int) bool {
			return output[i].Rank > output[j].Rank
		})
		var textBuf string
		for _, r := range output {
			textBuf += string(r.W[:5])+"\n"
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

func main() {
	file, err := os.Open("dict.gob")
	defer file.Close()
	if errors.Is(err,  os.ErrNotExist) {
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
		wordRank := make(map[[5]byte]uint16)
		var i uint16
		for {
			s, err := rankReader.ReadString('\n')
			if err != nil {
				break
			}
			if len([]rune(strings.TrimSpace(s))) == 5 {
				i++
				var word [5]byte
				for i := range word {
					word[i] = s[i]
				}
				wordRank[word] = i
			}
		}
		for {
			b, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if len([]rune(strings.TrimSpace(b))) == 5 {
				var word word
				for i := range word.W {
					word.W[i] = byte(b[i])
				}
				_, ok := wordRank[word.W]
				if !ok {
					word.Rank = 65535
				} else {
					word.Rank = wordRank[word.W]
				}
				dict = append(dict, word)
			}
		}
		file, err = os.Create("dict.gob")
		isError(err)
		encoder := gob.NewEncoder(file)
		encoder.Encode(dict)
	} else {
		decode := gob.NewDecoder(file)
		err := decode.Decode(&dict)
		isError(err)
	}

	file, err = os.Open("search.gob")
	if errors.Is(err, os.ErrNotExist) {
		for i := range 5 {
			for charNum := range 26 {
				search[i][charNum] = make(map[uint16]uint8)
				for j, r := range dict {
					if strings.Contains(string(r.W[i]), string(rune(charNum+97))) {
						search[i][charNum][uint16(j)]++
					}
				}
			}
		}
		file, err = os.Create("search.gob")
		isError(err)
		encoder := gob.NewEncoder(file)
		encoder.Encode(search)
	} else {
		decode := gob.NewDecoder(file)
		err := decode.Decode(&search)
		isError(err)
	}

	help := func() {
		fmt.Print(
			"Argument 1: Type phrase, use _ in the place of any unknown characters before the search term\n"+
			"Argument 2: In the second argument type any characters you know are in the word, but not the position\n"+
			"            Use _ to put the characters in the place you know they are not in\n"+
			"Argument 3: Type in any characters you know aren't in the word\n"+
			"Make sure the characters are all lowercase\n", 
		)
	}

	if len(os.Args) > 1 {
		for i, a := range os.Args {
			switch a {
			case "-h", "--help":
				help()
				return
			}
			switch i {
			case 1:
				posIs = strings.TrimSpace(a)
			case 2:
				isIn = strings.TrimSpace(a)
			case 3:
				notIn = strings.TrimSpace(a)
			case 4:
				runSearch(true)
				fmt.Println("Too many args, using the first 3")
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
