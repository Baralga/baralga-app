package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

func main() {
	data, err := ioutil.ReadFile("app.tpl.yaml")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	envVars := make(map[string]string)
	for _, element := range os.Environ() {
		variable := strings.Split(element, "=")
		envVars[variable[0]] = variable[1]
	}

	ut, err := template.New("users").Parse(string(data))
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile("app.yaml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = ut.Execute(file, envVars)

	if err != nil {
		panic(err)
	}
}
