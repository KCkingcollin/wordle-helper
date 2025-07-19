package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	// number characters in the lowercase alphabet
	alphaSize = int(26)
	// default word size
	defWsize = int(5)
)

var (
	// 2D searchMap data structure that has sorted words (represented as uint16 index) in the wordDict
	// into different maps each representing a set of words with a rune in the position
	//
	// the first dimension represents the position in the word.
	//
	// the second dimension represents the list of words with the alpha character (represented as a int from 0-25) in that x position
	searchMap 				[][alphaSize]map[uint16]struct{}
	wordDict 				[]word
	setEmpty 				= struct{}{}
	screen 					tcell.Screen
	defaultStyle 			= tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	posIs, isIn, notIn 		string
	mainTextBuf 			[][]rune
	ev 						tcell.Event
)

type word struct {
	W 		string
	Rank 	uint16
}

func isError(err error) {
	if err != nil {
		panic(err)
	}
}

// wip and not ready for use
func input() {
	ev = screen.PollEvent()
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

// wip and not ready for use
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

func containsAny(s, targetString string) bool {
	if len(targetString) == 0 {return false}
	runeSet := make(map[rune]bool)
	for _, r := range targetString {
		if r >= 'a' && r <= 'z' {
			runeSet[r] = true
		}
	}
	for _, r := range s {
		if runeSet[r] {return true}
	}
	return false
}

func containsAll(s, requiredRunes string) bool {
	// need to alter how this function works so that you can have 2 characters in the same position
	if len(requiredRunes) == 0 {return true}
	for i, r := range requiredRunes {
		if r >= 'a' && r <= 'z' {
			if s[i] == byte(r) {return false}
		}
	}
	foundRunes := make(map[rune]bool)
	for _, r := range requiredRunes {
		if r >= 'a' && r <= 'z' {
			foundRunes[r] = false
		}
	}

	for _, r := range s {
		if _, ok := foundRunes[r]; ok {
			foundRunes[r] = true
		}
	}

	for _, found := range foundRunes {
		if !found {return false}
	}
	return true
}

func runSearch(notInRaw ...bool) {
	if len(posIs) <= 0 {posIs+="_"}
	size := len(searchMap)
	if len(posIs) > size {posIs = posIs[:size]}
	if len(isIn) > size {isIn = isIn[:size]}
	list := make([]map[uint16]struct{}, len(posIs))
	add2List := func(i int, j uint16) {
		ok := true
		if i > 0 {_, ok = list[i-1][j]}
		if ok && containsAll(wordDict[int(j)].W[:size], isIn) && !containsAny(wordDict[int(j)].W[:size], notIn) {
			list[i][j] = setEmpty
		}
	}
	for i, r := range []rune(posIs) {
		list[i] = make(map[uint16]struct{})
		if r >= 'a' && r <= 'z' {
			for j := range searchMap[i][int(r-'a')] {
				add2List(i, j)
			}
		} else {
			for k := range alphaSize {
				if !containsAny(string(rune(k+'a')), notIn) {
					for j := range searchMap[i][k] {
						add2List(i, j)
					}
				}
			}
		}
	}
	output := make([]word, len(list[len(list)-1]))
	var i int
	for s := range list[len(list)-1] {
		output[i] = wordDict[s]
		i++
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].Rank > output[j].Rank
	})
	var textBuf string
	for _, r := range output {
		textBuf += string(r.W[:size])+"\n"
	}
	if len(notInRaw) > 0 && notInRaw[0] {
		fmt.Print(textBuf)
	} else {
		drawText(textBuf)
	}
}

func saveSearch(fileName string, search [][alphaSize]map[uint16]struct{}) {
	bufSizeGeuss := len(search[0][0])*alphaSize*len(search)*2+1
	buf := make([]byte, 0, bufSizeGeuss)
	buf = append(buf, byte(len(search)))
	for _, mapSlice := range search {
		for _, m := range mapSlice {
			for key := range m {
				buf = binary.BigEndian.AppendUint16(buf, key)
			}
			buf = append(buf, byte('A'))
		}
		buf = append(buf, byte('B'))
	}
	file, err := os.Create(fileName)
	defer file.Close()
	isError(err)
	io.Copy(file, bytes.NewReader(buf))
}

func loadSearch(fileName string) [][alphaSize]map[uint16]struct{} {
	file, err := os.Open(fileName)
	defer file.Close()
	isError(err)
	reader := bufio.NewReader(file)
	wordSize, _ := reader.ReadByte()
	search := make([][alphaSize]map[uint16]struct{}, wordSize)
	key := make([]byte, 2)
	for i := range wordSize {
		buf, err := reader.ReadBytes('B')
		if err != nil {break}
		buf = buf[:len(buf)-1]
		reader2 := bufio.NewReader(bytes.NewBuffer(buf))
		for j := range alphaSize {
			buf2, err := reader2.ReadBytes('A')
			if err != nil {break}
			buf2 = buf2[:len(buf2)-1]
			reader3 := bufio.NewReader(bytes.NewBuffer(buf2))
			for {
				_, err := reader3.Read(key)
				if err != nil {break}
				search[i][j] = make(map[uint16]struct{})
				search[i][j][binary.BigEndian.Uint16(key)] = setEmpty
			}
		}
	}
	return search
}

func saveDict(fileName string, dict []word) {
	buf := make([]byte, 0, len(dict)*4)
	for _, elm := range dict {
		buf = append(buf, byte(len(elm.W)))
		buf = append(buf, []byte(elm.W)...)
		buf = binary.BigEndian.AppendUint16(buf, elm.Rank)
	}
	file, err := os.Create(fileName)
	defer file.Close()
	isError(err)
	io.Copy(file, bytes.NewReader(buf))
}

func loadDict(fileName string) []word {
	dict := make([]word, 0, 15000)
	file, err := os.Open(fileName)
	defer file.Close()
	isError(err)
	buf := bufio.NewReader(file)
	rank := make([]byte, 2)
	for i := 0; true; i++ {
		var elm word
		wordSize, err := buf.ReadByte()
		if err != nil {break}
		w := make([]byte, wordSize)
		buf.Read(w)
		elm.W = string(w)
		buf.Read(rank)
		elm.Rank = binary.BigEndian.Uint16(rank)
		dict = append(dict, elm)
	}
	return dict
}

func createDict() []word {
	var dict []word
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
		i++
		wordRank[s] = i
	}
	for {
		b, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		b = strings.TrimSpace(b)
		var word word
		word.W = b
		_, ok := wordRank[word.W]
		if !ok {
			word.Rank = math.MaxUint16
		} else {
			word.Rank = wordRank[word.W]
		}
		dict = append(dict, word)
	}
	saveDict("dict.bin", dict)
	return dict
}

func createWordDict(wordSize int) {
	var dict []word
	_, err := os.Stat("dict.bin")
	if errors.Is(err, os.ErrNotExist) {
		dict = createDict()
	} else {
		dict = loadDict("dict.bin")
	}
	wordDict = make([]word, 0)
	for _, word := range dict {
		if len(word.W) == wordSize {
			wordDict = append(wordDict, word)
		}
	}
	saveDict("word-dict.bin", wordDict)
}

func createSearch(wordSize int) {
	searchMap = make([][alphaSize]map[uint16]struct{}, wordSize)
	for i := range wordSize {
		for charNum := range alphaSize {
			searchMap[i][charNum] = make(map[uint16]struct{})
			for j, r := range wordDict {
				if strings.Contains(string(r.W[i]), string(rune(charNum+'a'))) {
					searchMap[i][charNum][uint16(j)] = setEmpty
				}
			}
		}
	}
	saveSearch("search-map.bin", searchMap)
}

func main() {
	start := time.Now()
	_, err := os.Stat("word-dict.bin")
	if errors.Is(err, os.ErrNotExist) {
		createWordDict(defWsize)
	} else {
		wordDict = loadDict("word-dict.bin")
	}

	_, err = os.Stat("search-map.bin")
	if errors.Is(err, os.ErrNotExist) {
		createSearch(defWsize)
	} else {
		loadSearch("search-map.bin")
	}
	fmt.Println(time.Since(start))

	help := func() {
		fmt.Print(
			"Argument 1: Known character locations, use _ in the place of any unknown characters\n"+
			"Argument 2: Characters known to be in the word, use _ in the place of possible positions\n"+
			"Argument 3: Characters you know aren't in the word\n"+
			"Sending a number as an argument sets the size of the word, otherwise it will default to 5\n"+
			"_ is only required in the positions before the known characters\n", 
		)
	}

	if len(os.Args) > 1 {
		var wordSize, i int
		outer:
		for _, a := range os.Args {
			size, err := strconv.Atoi(a)
			switch {
			case a == "-h" || a == "--help":
				help()
				return
			case err == nil:
				wordSize = size
				if len(searchMap) != wordSize {
					createWordDict(wordSize)
					createSearch(wordSize)
				}
			default:
				i++
			}
			switch i-1 {
			case 1:
				posIs = strings.TrimSpace(a)
				posIs = strings.ToLower(posIs)
			case 2:
				// need to alter how this argument works so that you can have 2 characters in the same position
				// most likely with a 2d slice of runes so that each position can have more than 1 rune associated to it
				isIn = strings.TrimSpace(a)
				isIn = strings.ToLower(isIn)
			case 3:
				notIn = strings.TrimSpace(a)
				notIn = strings.ToLower(notIn)
			case 4:
				fmt.Println("Too many args, attempted to use the first 3")
				break outer
			}
		}
		if len(searchMap) != defWsize && wordSize == 0 {
			createWordDict(defWsize)
			createSearch(defWsize)
		}
		runSearch(true)
		return
	}

	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create screen: %v\n", err)
		return
	}

	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize screen: %v\n", err)
		return
	}
	defer screen.Fini()

	screen.SetStyle(defaultStyle)
	screen.Clear()

	_, screenSizeY := screen.Size()
	mainTextBuf = make([][]rune, screenSizeY)
	ev = screen.PollEvent()
	for {
		drawText("Type known character locations, use _ in the place of any unknown characters\n")
		drawText("_ is only required in the positions before the known characters\n")
		drawText(" > ")
		input()
		drawText("Type characters known to be in the word, use _ in the place of possible positions\n")
		drawText(" > ")
		input()
		drawText("Type characters you know aren't in the word\n")
		drawText(" > ")
		input()
		runSearch()
	}
}
