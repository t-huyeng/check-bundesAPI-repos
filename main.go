package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var pages = 1
var apiCount = 0
var apiCountMissingURL = 0
var apiListMissingURL []string
var apiCountMissingDescription = 0
var apiListMissingDescription []string

func main() {

	for i := 1; i <= pages; i++ {
		// call https://api.github.com/users/BundesAPI/repos?page=+pages
		resp, err := http.Get("https://api.github.com/users/BundesAPI/repos?page=" + fmt.Sprint(pages))
		if err != nil {
			log.Fatalln(err)
		}
		// check if response is 200
		if resp.StatusCode == 200 {
			// format the response to json
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err.Error())
			}
			var data interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				panic(err.Error())
			}
			// check if data.length is 0
			if len(data.([]interface{})) == 0 {
				fmt.Println("No more pages")
				fmt.Println("---------------------")
				break
			}
			pages++
			//if data is a list
			if list, ok := data.([]interface{}); ok {
				// loop through the list
				for _, item := range list {
					// if item is a dict
					if dict, ok := item.(map[string]interface{}); ok {
						// get the name of the repo
						// if name ends with -api
						name := dict["name"].(string)

						if name[len(name)-4:] == "-api" {
							fmt.Println(name)
							// increase apiCount
							apiCount++
							// check if the repo has a url
							if dict["homepage"] == nil || dict["homepage"].(string) == "" {
								fmt.Println("- No URL")
								apiCountMissingURL++
								apiListMissingURL = append(apiListMissingURL, name)
							}
							// check if the repo has a description
							if dict["description"] == nil || dict["description"].(string) == "" {
								fmt.Println("- No description")
								apiCountMissingDescription++
								apiListMissingDescription = append(apiListMissingDescription, name)
							}
							fmt.Println("---------------------")
						}

					}
				}

			}

		} else {
			break
		}
	}
	var output = "# BundesAPI Repositories \n"

	output += ("### APIs found: " + fmt.Sprintln(apiCount))
	output += ("### APIs without URL: " + fmt.Sprintln(apiCountMissingURL) + "\n")
	// print list of APIs without URL
	output += fmt.Sprintln(strings.Join(apiListMissingURL, ", "))
	output += fmt.Sprintln("")

	output += ("### APIs without description: " + fmt.Sprintln(apiCountMissingDescription) + "\n")
	// print list of APIs without description
	output += fmt.Sprintln(strings.Join(apiListMissingDescription, ", "))

	fmt.Println(output)
	// add the print outs to the Readme.md file
	ioutil.WriteFile("README.md", []byte(output), 0644)

}
