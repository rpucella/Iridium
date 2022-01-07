package main

import (
	"fmt"
	"path"
	"strconv"
	"strings"
)

func textContent(item Text) string {
	switch item.Kind {
	case TEXT_WORD:
		return item.Word

	default:
		stop(fmt.Sprintf("Unknown Text kind %d", item.Kind))
		return ""
	}
}

const (
	maxWidth = 78
)

type buffer struct {
	line string
	last string
	indent int
	firstLine bool
}

var buff = buffer{"", "", 0, true}

func emitReset(indent int) {
	buff.line = ""
	buff.last = ""
	buff.indent = indent
	buff.firstLine = true
}

func emitString(s string) {
	buff.last += s
}

func emitSpace() {
	indent := 0
	if buff.firstLine {
		indent = buff.indent
	}
	if indent + len(buff.line) + len(buff.last) + 1 > maxWidth {
		fmt.Println(buff.line)
		buff.line = strings.Repeat(" ", buff.indent) + buff.last + " "
		buff.last = ""
		buff.firstLine = false
	} else {
		buff.line += buff.last + " "
		buff.last = ""
	}
}

func emitDone() {
	indent := 0
	if buff.firstLine {
		indent = buff.indent
	}
	if indent + len(buff.line) + len(buff.last) + 1 > maxWidth {
		fmt.Println(buff.line)
		fmt.Println(buff.last)
	} else {
		fmt.Println(buff.line + buff.last)
	}
}

func printTexts(content []Text) {
	for i, t := range(content) {
		switch t.Kind {
		case TEXT_WORD:
			emitString(t.Word)
			if i < len(content) - 1 {
				emitSpace()
			}

		case TEXT_QUOTE:
			emitString("\"")
			printTexts(t.Content)
			emitString("\"")
			if i < len(content) - 1 {
				emitSpace()
			}
			
		default:
			stop(fmt.Sprintf("Unknown Text kind %d", t.Kind))
		}
	}
}

func run(srcdir string) {
	passagesDir := path.Join(srcdir, SRC_PASSAGES)
	config, err := readConfig(srcdir)
	if err != nil {
		stop(fmt.Sprint(err))
	}
	clear()
	fmt.Println(config.Title)
	fmt.Println(config.Subtitle)
	fmt.Println("By", config.Author)
	fmt.Println()

	currentPassage := config.InitialPassage
	for true {
		passage, err := readPassage(passagesDir, currentPassage)
		if err != nil {
			stop(fmt.Sprint(err))
		}
		//fmt.Println(passage)
		r := strings.NewReader(passage)
		p := NewParser(r)
		psg, err := p.Parse()
		if err != nil {
			stop(fmt.Sprint(err))
		}
		for _, x := range(psg.Blocks) {
			if x.Kind == TEXT {
				emitReset(0)
				printTexts(x.Content)
				emitDone()
				fmt.Println()
			}
		}
		if len(psg.Options) > 0 {
			for i, option := range(psg.Options) {
				fmt.Printf(" % 2d. ", i + 1)
				emitReset(5)
				printTexts(option.Content)
				emitDone()
			}
			fmt.Println()
			var input string
			for { 
				fmt.Print("? ")
				fmt.Scanln(&input)
				if input == "q" {
					fmt.Println("Bailing")
					return
				}
				if input == "" && len(psg.Options) == 1 {
					// one choice, so take it
					currentPassage = psg.Options[0].Target
					break
				}
				choice, err := strconv.Atoi(input)
				if err == nil && choice > 0 && choice <= len(psg.Options) {
					currentPassage = psg.Options[choice - 1].Target
					break
				}
			}
			clear()
			continue
		}
		break
	}
}


func clear() {
	fmt.Print("\033[H\033[2J\n")
}

