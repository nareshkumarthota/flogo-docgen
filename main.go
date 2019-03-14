package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
)

func main() {
	fmt.Println("Flogo doc generation from async api spec")

	templateFile := flag.String("template", "ws-template", "template file name")
	asyncFile := flag.String("input", "", "input file(async spec)")

	flag.Parse()

	fmt.Println("templateFile: ", *templateFile, "\nasyncFile: ", *asyncFile)

	// Read the template and have it in a string
	templateData, err := ioutil.ReadFile(*templateFile)
	if err != nil {
		log.Fatal("error occured in reading template: ", err)
	}

	// Read async spec
	asyncFileData, err := ioutil.ReadFile(*asyncFile)
	if err != nil {
		log.Fatal("error occured in reading async spec: ", err)
	}

	var asyncData map[string]interface{}
	if err := yaml.Unmarshal(asyncFileData, &asyncData); err != nil {
		log.Fatal("error occured in unmarshaling async spec: ", err)
	}

	// create template
	t := template.Must(template.New("top").Parse(string(templateData)))
	buf := &bytes.Buffer{}

	// data for template
	dataMap := make(map[string]interface{})

	// assigning data from async to template data
	if val, ok := asyncData["info"].(map[string]interface{})["title"]; ok {
		dataMap["title"] = val.(string)
	}

	if val, ok := asyncData["info"].(map[string]interface{})["version"]; ok {
		dataMap["version"] = val.(string)
	}

	if val, ok := asyncData["info"].(map[string]interface{})["description"]; ok {
		dataMap["description"] = val.(string)
	}

	if val, ok := asyncData["servers"].([]interface{}); ok {

		servrDtls := val[0].(map[string]interface{})

		// get port and path from url
		fields := strings.Split(servrDtls["url"].(string), ":")
		ParamFields := strings.Split(fields[1], "/")

		// derive values from fields with variables
		ParamFields = deriveValues(servrDtls["variables"].(map[string]interface{}), ParamFields)

		dataMap["url"] = "ws://" + fields[0] + ":" + ParamFields[0] + "/" + ParamFields[1]

	}

	// creating template content
	if err := t.Execute(buf, dataMap); err != nil {
		log.Printf("error while rendering Markdown file: %s", err.Error())
	}
	s := buf.String()

	outPutFilePath, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory: ", err)
	}
	outPutFilePath = filepath.Join(outPutFilePath, "flogo.json")

	// create output file
	createFileWithContent(outPutFilePath, s)

}

func createFileWithContent(filename, content string) error {

	// Create a file on disk
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("error while creating file: %s", err.Error())
		return fmt.Errorf("error while creating file: %s", err.Error())
	}
	defer file.Close()

	// Open the file to write
	file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("error while opening file: %s", err.Error())
		return fmt.Errorf("error while opening file: %s", err.Error())
	}

	// Write the Markdown doc to disk
	_, err = file.Write([]byte(content))
	if err != nil {
		log.Printf("error while writing Markdown to disk: %s", err.Error())
		return fmt.Errorf("error while writing Markdown to disk: %s", err.Error())
	}

	return nil
}

func deriveValues(variables map[string]interface{}, fields []string) []string {

	out := make([]string, len(fields))

	for index, elem := range fields {
		if strings.Contains(elem, "{") {
			elem = strings.Replace(elem, "{", "", -1)
			elem = strings.Replace(elem, "}", "", -1)
			elem = variables[elem].(map[string]interface{})["default"].(string)
		}
		out[index] = elem
	}

	return out
}
