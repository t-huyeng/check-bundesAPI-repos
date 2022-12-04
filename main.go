package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pterm/pterm"
)

var pages = 1
var apiCount = 0
var apiCountMissingURL = 0

type Dictionary map[string]interface{}

var apiListMissingURL = []Dictionary{}
var apiCountMissingDescription = 0
var apiListMissingDescription = []Dictionary{}
var apiHTMLreport = []Dictionary{}

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
						name := dict["name"].(string)
						htmlUrl := dict["html_url"].(string)
						var infos = map[string]interface{}{"name": name, "html_url": htmlUrl}
						// if name ends with -api
						if name[len(name)-4:] == "-api" {
							pterm.Info.Println(name)
							// increase apiCount
							apiCount++
							// check if the repo has a url
							if dict["homepage"] == nil || dict["homepage"].(string) == "" {
								pterm.Warning.Println("- No URL")
								apiCountMissingURL++
								apiListMissingURL = append(apiListMissingURL, infos)
							}
							// check if the repo has a description
							if dict["description"] == nil || dict["description"].(string) == "" {
								pterm.Warning.Println("- No description")
								apiCountMissingDescription++
								apiListMissingDescription = append(apiListMissingDescription, infos)
							}
							// generate vaccum report of openapi
							// create raw openapi url
							url := "https://raw.githubusercontent.com/bundesAPI/" + name + "/main/openapi.yaml"
							GenerateHtml(url)
							apiHTMLreport = append(apiHTMLreport, infos)

						}
					}

				}

			}

		} else {
			break
		}
	}
	// generate json list
	jsonListMissingURL, _ := json.Marshal(apiListMissingURL)
	jsonListMissingDescription, _ := json.Marshal(apiListMissingDescription)
	// combine json lists to one json
	jsonList := []byte(`{"missingURL":` + string(jsonListMissingURL) + `,"missingDescription":` + string(jsonListMissingDescription) + `}`)

	// save json list to file
	ioutil.WriteFile("output.json", []byte(jsonList), 0644)

	var output = "# BundesAPI Repositories\n"

	output += ("### APIs found: " + fmt.Sprintln(apiCount))
	output += "\n"
	base_url := "https://t-huyeng.github.io/check-bundesAPI-repos/vacuum-reports/"

	// print html report links
	for _, item := range apiHTMLreport {
		output += fmt.Sprintln("- [" + item["name"].(string) + "](" + item["html_url"].(string) + ")")
		output += fmt.Sprintln(" : [OpenAPI Report (Vacuum)](" + base_url + item["name"].(string) + ".html)")
	}
	output += ("### APIs without URL: " + fmt.Sprintln(apiCountMissingURL) + "\n")
	// for each dict in apiListMissingURL add the name to the output
	for _, dict := range apiListMissingURL {
		output += ("- [" + dict["name"].(string) + "](" + dict["html_url"].(string) + ")\n")
	}

	output += fmt.Sprintln("")

	output += ("### APIs without description: " + fmt.Sprintln(apiCountMissingDescription) + "\n")

	// print list of APIs without description
	for _, dict := range apiListMissingDescription {
		output += ("- [" + dict["name"].(string) + "](" + dict["html_url"].(string) + ")\n")
	}

	fmt.Println(output)
	// add the print outs to the Readme.md file
	ioutil.WriteFile("README.md", []byte(output), 0644)

}
