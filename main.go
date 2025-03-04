package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

type CommanderFile struct {
	Commands    []Command         `yaml:"commands"`
	Environment map[string]string `yaml:"environment"`
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

	commandsPage := tview.NewFlex().
		SetDirection(tview.FlexRow)
	pages.AddPage("Commands", commandsPage, true, true)

	outputPage := tview.NewFlex().
		SetDirection(tview.FlexRow)
	pages.AddPage("Output", outputPage, true, false)

	commandsList := tview.NewList()

	for _, cmd := range commanderFile.Commands {
		commandsList.AddItem(cmd.Title, "", 0, func() {
			outputPage.Clear()

			outputPageContent := tview.NewTextView().
				SetWrap(true).
				SetText(cmd.Template)

			outputPageContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
					pages.SwitchToPage("Commands")
					return nil
				}
				return event
			})

			outputPage.AddItem(outputPageContent, 0, 1, true)

			pages.SwitchToPage("Output")
		})
	}

	commandsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		return event
	})

	commandsPage.AddItem(commandsList, 0, 1, true)

	commandsPage.
		SetTitle(" Commands ").
		SetBorder(true)

	outputPage.
		SetTitle(" Output ").
		SetBorder(true)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
