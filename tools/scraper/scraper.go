// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/pelletier/go-toml/v2"
)

var (
	baseURL = "https://en.wikipedia.org/wiki/Wikipedia:Lists_of_common_misspellings/"
	page    = []string{
		"0%E2%80%939",
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	}
	correctionsFile = "./configs/correctionsv2.toml"
)

func main() {
	// Array containing all the known URLs in a sitemap
	mispellings := []string{}

	// Create a Collector specifically for Shopify
	c := colly.NewCollector(colly.AllowedDomains("en.wikipedia.org"))

	q, _ := queue.New(
		2, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
		r2, err := r.New("GET", r.URL.String(), nil)
		if err == nil {
			q.AddRequest(r2)
		}
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("/html/body/div[2]/div/div[3]/main/div[3]/div[3]/div[1]/ul[2]", func(e *colly.XMLElement) {
		scanner := bufio.NewScanner(strings.NewReader(e.Text))
		for scanner.Scan() {
			mispellings = append(mispellings, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	})

	for i := 0; i < len(page); i++ {
		// Add URLs to the queue
		q.AddURL(fmt.Sprintf("%s%s", baseURL, page[i]))
	}
	// Consume URLs
	q.Run(c)

	m := make(map[string]string)
	for _, url := range mispellings {
		pattern := regexp.MustCompile(`(?m)(?P<mispell>[^\n^\r]+)\s\((?P<correction>[^)^(^\n]+)\).*$`)
		matches := pattern.FindAllSubmatchIndex([]byte(url), -1)
		for _, loc := range matches {
			misspell := string([]byte(url)[loc[2]:loc[3]])
			corrections := strings.Split(string([]byte(url)[loc[4]:loc[5]]), ",")
			m[misspell] = corrections[0]
			// fmt.Printf("misspell %s, correction %v\n", misspell, corrections[0])
		}
	}

	n := cleanCorrections(m)

	b, err := toml.Marshal(&n)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(correctionsFile, b, 0644)
	if err != nil {
		panic(err)
	}
	// fmt.Println(string(b))
	// fmt.Println("Collected", len(mispellings), "URLs")
}

func cleanCorrections(m map[string]string) map[string]string {
	n := make(map[string]string)
	removeType := regexp.MustCompile(`\s\[[\w\s]+\]$`)
	// multipleWord := regexp.MustCompile(`\w+\s\w+`)

	for k, v := range m {
		// remove mispellings that are multiple words. autocorrector cannot handle
		// multi-word keys.
		if strings.ContainsRune(k, ' ') {
			continue
		}
		// remove corrections that are just indicating a variant
		if strings.Contains(v, "variant of") {
			continue
		}
		// remove alternatives
		if strings.Contains(v, "alternative spelling") {
			continue
		}
		// remove acceptables
		if strings.Contains(v, "acceptable spellings") {
			continue
		}
		// remove synonyms
		if strings.Contains(v, "acceptable synonym") {
			continue
		}
		// remove either/or spellings
		if strings.Contains(v, " or ") {
			continue
		}
		// remove conditionals
		if strings.Contains(v, "but not") {
			continue
		}
		// remove the type appended on the end of a correction (e.g.,
		// '[plural]').
		v = removeType.ReplaceAllString(v, "")
		// remove corrections with conditional explanations
		if strings.ContainsRune(v, ';') {
			continue
		}
		n[k] = v

		fmt.Printf("mispelling: %s correction: %s\n", k, v)
	}
	return n
}
