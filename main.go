package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/akamensky/argparse"
	"github.com/tidwall/gjson"
)

func GetProgramPagePaths(sessionToken string, privatesOnly bool) []string {
	allProgramsCount := 0
	currentProgramIndex := 0

	listEndpointURL := "https://bugcrowd.com/programs.json?hidden[]=false&sort[]=invited-desc&sort[]=promoted-desc&offset[]="

	if privatesOnly {
		listEndpointURL = "https://bugcrowd.com/programs.json?accepted_invite[]=true&hidden[]=false&sort[]=invited-desc&sort[]=promoted-desc&offset[]="
	}

	paths := []string{}

	for {
		req, err := http.NewRequest("GET", listEndpointURL+strconv.Itoa(currentProgramIndex), nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Cookie", "_crowdcontrol_session="+sessionToken)
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:81.0) Gecko/20100101 Firefox/81.0")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if allProgramsCount == 0 {
			allProgramsCount = int(gjson.Get(string(body), "meta.totalHits").Int())
		}

		chunkData := gjson.Get(string(body), "programs.#.program_url")
		for i := 0; i < len(chunkData.Array()); i++ {
			paths = append(paths, chunkData.Array()[i].Str)
		}
		currentProgramIndex += 25

		if allProgramsCount <= currentProgramIndex {
			break
		}
	}

	return paths
}

func PrintProgramScope(url string, sessionToken string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Cookie", "_crowdcontrol_session="+sessionToken)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:81.0) Gecko/20100101 Firefox/81.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Yeah, HTML parsing is a pain @archwhite do something damn it :D
	// Or at least, don't break this tool aka don't change HTML stuff <3

	var scope []string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		fmt.Println("No url found")
		log.Fatal(err)
	}

	// Find each table
	doc.Find("#user-guides__bounty-brief__targets-table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("tbody td").Each(func(indexth int, tablecell *goquery.Selection) {
				if indexth == 0 {
					scope = append(scope, strings.TrimSpace(tablecell.Text()))
				}
			})
		})
	})

	for _, s := range scope {
		fmt.Println(s)
	}
}

func main() {
	parser := argparse.NewParser("bcscope", "Get the scope of your Bugcrowd programs")

	sessionToken := parser.String("t", "token", &argparse.Options{Required: true, Help: "Bugcrowd session token (_crowdcontrol_session)"})
	privateInvitesOnly := parser.Flag("p", "private", &argparse.Options{Required: false, Default: false, Help: "Only show private invites"})
	listOnly := parser.Flag("l", "list", &argparse.Options{Required: false, Default: false, Help: "List programs instead of grabbing their scope"})
	concurrency := parser.Int("c", "concurrency", &argparse.Options{Required: false, Default: 2, Help: "Set concurrency"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	programPaths := GetProgramPagePaths(*sessionToken, *privateInvitesOnly)

	if *listOnly {
		for _, path := range programPaths {
			fmt.Println("https://bugcrowd.com" + path)
		}
	} else {

		urls := make(chan string, *concurrency)
		processGroup := new(sync.WaitGroup)
		processGroup.Add(*concurrency)

		for i := 0; i < *concurrency; i++ {
			go func() {
				for {
					url := <-urls

					if url == "" {
						break
					}

					PrintProgramScope(url, *sessionToken)
				}
				processGroup.Done()
			}()
		}

		for _, path := range programPaths {
			urls <- "https://bugcrowd.com" + path
		}

		close(urls)
		processGroup.Wait()

	}
}
