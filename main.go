package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	cb "github.com/atotto/clipboard"
	"github.com/soloviev1d/ggrep/color"
)

type query struct {
	searchFor   string
	searchIn    string
	nestedDirs  []string
	firstSearch bool
	matchedStr  string
}

func newQuery(args []string) *query {
	q := query{
		searchFor:   args[0],
		searchIn:    args[1],
		firstSearch: true,
	}
	return &q
}

func (q *query) search() string {
	var (
		files []fs.FileInfo
		err   error
		wd, _ = os.Getwd()
	)

	if q.searchIn[0] == '/' || q.searchIn == "." {
		if q.searchIn == "." {
			q.searchIn = ""
		}
		files, err = ioutil.ReadDir(wd + q.searchIn)
		if err != nil {
			log.Fatalf("%sno such directory: %s%s\n", color.Red, wd+q.searchIn+"/", color.Reset)
		}
		for _, file := range files {
			if !file.IsDir() && file.Mode()&0111 != 0111 {
				f, err := os.Open(wd + q.searchIn + "/" + file.Name())
				if err != nil {
					log.Fatalf("failed to open file: %v\n", err)
				}
				defer f.Close()
				var (
					s  = bufio.NewScanner(f)
					ln = 1
				)
				for s.Scan() {
					matched, _ := regexp.MatchString(q.searchFor, s.Text())
					if matched {
						fmt.Printf("found in %s%s:%d%s: %s\n", color.Blue, f.Name(), ln, color.Reset, strings.TrimSpace(s.Text()))
						q.matchedStr += strings.TrimSpace(s.Text()) + "\n"
					}
					ln++
				}
			}
		}
	} else {
		f, err := os.Open(q.searchIn)
		if err != nil {
			log.Fatalf("%sno such file: %s%s\n", color.Red, q.searchIn, color.Reset)
		}
		defer f.Close()
		var (
			s  = bufio.NewScanner(f)
			ln = 1
		)
		for s.Scan() {
			matched, _ := regexp.MatchString(q.searchFor, s.Text())
			if matched {
				fmt.Printf("found in %s%s:%d%s: %s\n", color.Blue, f.Name(), ln, color.Reset, strings.TrimSpace(s.Text()))
				q.matchedStr += strings.TrimSpace(s.Text()) + "\n"
			}
			ln++
		}
	}

	return q.matchedStr
}

func (q *query) mustSearchRecursively(dir string) string {
	if dir == "." {
		dir = ""
	}
	// fmt.Println(dir, " ", q.nestedDirs)
	if len(q.nestedDirs) > 0 {
		q.nestedDirs = q.nestedDirs[1:]
	}
	wd, _ := os.Getwd()
	// fmt.Println(dir, " ", q.nestedDirs)
	files, err := ioutil.ReadDir(wd + dir + "/")
	if err != nil {
		log.Fatalf("%sno such directory: %s%s\n", color.Red, wd+dir+"/", color.Reset)
	}
	for _, file := range files {

		if file.IsDir() {
			q.nestedDirs = append(q.nestedDirs, dir+"/"+file.Name())
		} else if file.Mode()&0111 != 0111 {
			f, err := os.Open(wd + dir + "/" + file.Name())
			if err != nil {
				log.Fatalf("failed to open file: %v", err)
			}
			defer f.Close()
			var (
				s  = bufio.NewScanner(f)
				ln = 1
			)
			for s.Scan() {
				matched, _ := regexp.MatchString(q.searchFor, s.Text())
				if matched {
					fmt.Printf("found in %s%s:%d%s: %s\n", color.Blue, f.Name(), ln, color.Reset, strings.TrimSpace(s.Text()))
					q.matchedStr += strings.TrimSpace(s.Text()) + "\n"
				}
				ln++
			}
		}
	}
	if len(q.nestedDirs) == 0 {
		return q.matchedStr
	}
	return q.mustSearchRecursively(q.nestedDirs[0])
}

func main() {
	var (
		copyOutput = flag.Bool("co", false, "Copy outupt to clipboard")
		recursive  = flag.Bool("r", false, "search for expression recursively from given directory")
		matched    string
	)

	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Printf("Invalid arguments\n")
		os.Exit(0)
	}
	q := newQuery(args)
	if *recursive {
		matched = q.mustSearchRecursively(q.searchIn)
	} else {
		matched = q.search()
	}

	if *copyOutput {
		copyOutputToClipboard(matched)
	}

}

func copyOutputToClipboard(output string) {
	err := cb.WriteAll(output)
	if err != nil {
		log.Fatalf("failed to copy to clipboard!")
	}
}
