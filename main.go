package main

import (
	"log"
	"os"

	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

type CommanderFile struct {
	Commands    []Command
	Environment map[string]string
}

type CommandInput struct {
	Key          string `yaml:"key"`
	DefaultValue string `yaml:"defaultValue"`
}

type Command struct {
	Title       string         `yaml:"title"`
	Template    string         `yaml:"command"`
	Description string         `yaml:"description"`
	Inputs      []CommandInput `yaml:"inputs"`
}

func main() {
	file, err := os.ReadFile("./commander.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var commanderFile CommanderFile
	err = yaml.Unmarshal(file, &commanderFile)
	if err != nil {
		log.Fatal(err)
	}

	app := tview.NewApplication()
	pages := tview.NewPages()

	list := tview.NewBox().
		SetBorder(true).
		SetTitle(" Commander ")

	listWrapper := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true)

	pages.AddPage("List", listWrapper, true, true)

	output := tview.NewBox().
		SetBorder(true).
		SetTitle(" Output ")

	outputWrapper := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(output, 0, 1, true)

	pages.AddPage("Output", outputWrapper, true, true)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
