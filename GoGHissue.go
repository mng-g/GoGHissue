// Simple CLI tool to manage issues on GitHub

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Response struct {
	Number uint   `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

var TOKEN string = "YOUR_TOKEN"

const MAX_LABEL int = 3

var user = flag.String("u", "user", "username")
var repo = flag.String("r", "repo", "repository")

func main() {

	flag.Parse()
	var URL string = "https://api.github.com/repos/" + *user + "/" + *repo + "/issues"
	fmt.Println("URL: ", URL)

	if *user == "user" && *repo == "repo" {
		fmt.Fprintf(os.Stderr, "\nYou haven't set username and repository!\n")
		os.Exit(1)
	}

	fmt.Print("What do you want to do? [N]ew, [R]ead, [U]pdate, [C]lose: ")
	var action string
	fmt.Scanln(&action)
	method := ""
	var body []byte = nil
	var n uint = 1
	switch action {
	case "N":
		method, body = Create()

	case "R":
		method, URL = Read(URL)

	case "U":
		//	var choice bool
		fmt.Print("What is the number of the issue that you want to update? ")
		fmt.Scanln(&n)
		method = "PATCH"
		number := fmt.Sprint("/", n)
		URL = fmt.Sprint(URL, number)

		// GET the original values
		bodyBytes := Request(method, URL, body)
		var responseObject Response
		json.Unmarshal(bodyBytes, &responseObject)

		_title := UpdateValue("Title", &responseObject)
		body = []byte(fmt.Sprintf(`{
			"title": "%v",
		`, _title))

		_body := UpdateValue("Body", &responseObject)
		body2 := []byte(fmt.Sprintf(`
			"body": "%v",
		`, _body))
		body = append(body, body2...)

	case "C":
		method, URL, body = Close(URL)

	default:
		fmt.Fprintf(os.Stderr, "error: action %v not valid\n", action)
		os.Exit(1)
	}

	// Debug
	//fmt.Printf("\nThe request is:\nMethod: %s\nURL: %s\nBody: %s\n", method, URL, string(body))
	// Make Request
	bodyBytes := Request(method, URL, body)

	// Get response
	var responseObject Response
	json.Unmarshal(bodyBytes, &responseObject)
	if responseObject.Number != 0 {
		fmt.Printf("API Response as struct %+v\n", responseObject)
	} else {
		fmt.Fprintf(os.Stderr, "\nOps! Something went wrong. Empty Reply\n")
		os.Exit(1)
	}
}

func Create() (string, []byte) {
	fmt.Println("Creating new issue...")
	method := "POST"
	// GET user input
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Title: ")
	scanner.Scan()
	_title := scanner.Text()
	fmt.Print("Body: ")
	scanner.Scan()
	_body := scanner.Text()
	fmt.Println("You can add up to", MAX_LABEL, "labels.")
	_labels := make([]string, MAX_LABEL)
	for i := 0; i < MAX_LABEL; i++ {
		fmt.Print("Label: ")
		scanner.Scan()
		_labels[i] = scanner.Text()
		if i < MAX_LABEL-1 {
			fmt.Printf("Do you want add another label? ")
			isConfirmed := Ask4confirm()
			if !isConfirmed {
				break
			}
		}
	}
	_labels = deleteEmpty(_labels)
	_labelsStr := `["` + strings.Join(_labels, `", "`) + `"]`

	// JSON body
	body := []byte(fmt.Sprintf(`{
		"title": "%v",
		"body": "%v",
		"labels": %v
	}`, _title, _body, _labelsStr))
	return method, body
}

func Read(URL string) (string, string) {
	var n uint
	fmt.Print("What is the number of the issue that you want to read? ")
	fmt.Scanln(&n)
	method := "GET"
	number := fmt.Sprint("/", n)
	URL = fmt.Sprint(URL, number)
	return method, URL
}

func UpdateValue(key string, resp *Response) string {
	var value string
	fmt.Printf("Do you want to update the %v? ", key)
	isConfirmed := Ask4confirm()
	if isConfirmed {
		fmt.Printf("%v: ", key)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		value = scanner.Text()
	} else {
		switch strings.ToLower(key) {
		case "title":
			value = resp.Title
		case "body":
			value = resp.Body
		default:
			{
				fmt.Fprintf(os.Stderr, "error: %v not exists\n", key)
				os.Exit(1)
			}
		}
	}
	return value
}

func Close(URL string) (string, string, []byte) {
	var n uint
	fmt.Print("What is the number of the issue that you want to close? ")
	fmt.Scanln(&n)
	method := "PATCH"
	number := fmt.Sprint("/", n)
	URL = fmt.Sprint(URL, number)
	body := []byte(`{
		"state": "close"
	}`)
	return method, URL, body
}

func Request(m string, u string, b []byte) []byte {
	client := &http.Client{}
	req, err := http.NewRequest(m, u, bytes.NewBuffer(b))
	if err != nil {
		fmt.Print(err.Error())
	}
	// Add headers
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", TOKEN)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	return bodyBytes
}

func Ask4confirm() bool {
	var s string

	fmt.Printf("(y/N): ")
	_, err := fmt.Scan(&s)
	if err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
