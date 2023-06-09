package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"bufio"
	"strings"
	"io/ioutil"
)

func build(srcdir string) {
	
	fmt.Println("Compiling passages")
	content, err := compile(path.Join(srcdir, SRC_PASSAGES))
	if err != nil {
		stop(fmt.Sprint(err))
	}
	
	os.RemoveAll(path.Join(srcdir, GAME_DIST))
	err = os.Mkdir(path.Join(srcdir, GAME_DIST), 0755)
	if err != nil {
		stop(fmt.Sprint(err))
	}
	fmt.Printf("Creating %s/%s\n", GAME_DIST, GAME_HTML)
	file, err := os.Open(path.Join(srcdir, SRC_HTML))
	if err != nil {
		stop(fmt.Sprint(err))
	}
	defer file.Close()
	fileout, err := os.Create(path.Join(srcdir, GAME_DIST, GAME_HTML))
	if err != nil {
		stop(fmt.Sprint(err))
	}
	defer fileout.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		index := strings.Index(line, "</body>")
		if index >= 0 { 
			fmt.Fprintln(fileout, line[:index])
			fmt.Fprintln(fileout, "<script>")
			dumpCoreJS(fileout)
			fmt.Fprintln(fileout, "</script>")
			fmt.Fprintln(fileout, "<script>")
			fmt.Fprintln(fileout, content)
			fmt.Fprintln(fileout, "</script>")
			fmt.Fprintln(fileout, "<script>")
			dumpGameJS(fileout, path.Join(srcdir, SRC_JSON))
			fmt.Fprintln(fileout, "</script>")
			fmt.Fprintln(fileout, "<script> document.querySelector('head > title').innerText = game.title; engine.run(game, content); </script>")
			fmt.Fprintln(fileout, line[index:])
		} else { 
			fmt.Fprintln(fileout, line)
		}
	}
	if err := scanner.Err(); err != nil {
		stop(fmt.Sprint(err))
	}
	fileout.Sync()

	stat, err := os.Stat(path.Join(srcdir, SRC_ASSETS))
	if err == nil && stat.IsDir() {
		// we got an assets folder - does it contain stuff?
		files, err := ioutil.ReadDir(path.Join(srcdir, SRC_ASSETS))
		if err == nil && len(files) > 0 {
			// we got content! copy it
			fmt.Printf("Creating %s/%s\n", GAME_DIST, GAME_ASSETS)
			err = CopyDir(path.Join(srcdir, SRC_ASSETS), path.Join(srcdir, GAME_DIST, GAME_ASSETS))
			if err != nil {
				stop(fmt.Sprint(err))
			}
		}
	}
}

func dumpGameJS(fileout io.Writer, gameJson string) {
	file, err := os.Open(gameJson)
	if err != nil {
		stop(fmt.Sprint(err))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	fmt.Fprint(fileout, "const game = ")
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(fileout, line)
	}
}

// func dumpContentJS(fileout io.Writer, contentJS string) {
// 	file, err := os.Open(contentJS)
// 	if err != nil {
// 		stop(fmt.Sprint(err))
// 	}
// 	defer file.Close()
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		fmt.Fprintln(fileout, line)
// 	}
// }

func joinText(items []Text) string {
	texts := make([]string, len(items))
	for i, item := range(items) {
		switch (item.Kind) {
		case TEXT_WORD:
			texts[i] = item.Word
		case TEXT_QUOTE:
			texts[i] = "<q>" + joinText(item.Content) + "</q>"
		default:
			stop(fmt.Sprintf("Unrecognized Text kind %d", item.Kind))
		}
	}
	return strings.Join(texts, " ")
}

func compile(srcdir string) (string, error) {
	passages, err := getPassageNames(srcdir)
	if err != nil {
		return "", err
	}
	contentList := make([]string, 0)
	for _, passageName := range passages {
		fmt.Println(" Processing", passageName)
		passage, err := readPassage(srcdir, passageName)
		if err != nil {
			return "", err
		}
		r := strings.NewReader(passage)
		p := NewParser(r)
		psg, err := p.Parse()
		if err != nil {
			return "", err
		}
		body := "let c = io.choices(); "
		for _, b := range(psg.Blocks) {
			if b.Kind == TEXT {
				body += fmt.Sprintf("io.p(\"%s\"); ", joinText(b.Content))
			} else if b.Kind == IMAGE {
				body += fmt.Sprintf("io.img(\"%s\", \"%s\"); ", b.Image, b.Style)
			}
		}
		if len(psg.Options) > 0 {
			for _, option := range(psg.Options) {
				body += fmt.Sprintf("c = c.option(\"%s\", function() { engine.goPassage(state, content, \"%s\", true); }); ", joinText(option.Content), option.Target)
			}
		}
		body += "c.show();"
		contentList = append(contentList, fmt.Sprintf("content[\"%s\"] = (function(state) { %s });", passageName, body))
	}
	content := strings.Join(contentList, "\n")
	return fmt.Sprintf("const content = function(fn) { let content = {};\n%s;\nreturn content;}\n", content), nil
}

