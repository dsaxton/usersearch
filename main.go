package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type InputData struct {
	License    []string `json:"license"`
	Authors    []string `json:"authors"`
	Categories []string `json:"categories"`
	Sites      []Site   `json:"sites"`
}

type Site struct {
	Name          string   `json:"name"`
	URL           string   `json:"uri_check"`
	ExistsCode    int      `json:"e_code"`
	ExistsString  string   `json:"e_string"`
	MissingCode   int      `json:"m_code"`
	MissingString string   `json:"m_string"`
	Known         []string `json:"known"`
	Category      string   `json:"cat"`
	Valid         bool     `json:"valid"`
}

const inputData = "https://raw.githubusercontent.com/WebBreacher/WhatsMyName/main/wmn-data.json"
const usage = "Usage: usersearch -u=<username> -o=<outfile>"

func main() {
	username := flag.String("u", "", "Username")
	outfile := flag.String("o", "", "Output file name")
	flag.Parse()

	if *username == "" || *outfile == "" {
		fmt.Println(usage)
		os.Exit(1)
	}

	file, err := os.Create(*outfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	response, err := http.Get(inputData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var data InputData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < len(data.Sites); i++ {
		wg.Add(1)
		go func(site Site) {
			defer wg.Done()

			// These sites are prone to false positives
			if site.Name == "aaha_chat" || site.Name == "ru_123rf" || site.Name == "Salon24" || site.Name == "olx" {
				return
			}

			replacedURL := strings.Replace(site.URL, "{account}", *username, 1)
			client := &http.Client{Timeout: 10 * time.Second}
			response, err := client.Get(replacedURL)
			if err != nil {
				return
			}
			defer response.Body.Close()

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return
			}

			if response.StatusCode == site.ExistsCode && strings.Contains(string(body), site.ExistsString) {
				fmt.Println(replacedURL)
				mu.Lock()
				file.WriteString(replacedURL)
				file.WriteString("\n")
				mu.Unlock()
			}
		}(data.Sites[i])
	}

	wg.Wait()
}
