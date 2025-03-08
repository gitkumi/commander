package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

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

type State struct {
	config        CommanderFile
	app           *tview.Application
	pages         *tview.Pages
	outputPage    *tview.Flex
	commandsPage  *tview.Flex
	parsedCommand string
}

func (state *State) buildOutputContent(cmd Command) {
	state.outputPage.Clear()

	if len(cmd.Inputs) > 0 {
		state.buildOutputContentWithInputs(cmd)
		return
	}

	state.buildOutputContentWithoutInputs(cmd)
}

func (state *State) buildOutputContentWithInputs(cmd Command) {
	form := tview.NewForm()
	formValues := make(map[string]string)

	for _, input := range cmd.Inputs {
		formValues[input.Key] = input.DefaultValue
		form.AddInputField(input.Key, input.DefaultValue, 40, nil, func(text string) {
			formValues[input.Key] = text
		})
	}

	form.AddButton("Submit", func() {
		tpl, err := template.New("command").Parse(cmd.Template)
		if err != nil {
			state.displayText(fmt.Sprintf("Template parse error: %v", err))
			return
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, formValues); err != nil {
			state.displayText(fmt.Sprintf("Template execution error: %v", err))
			return
		}

		state.parsedCommand = buf.String()
		state.runCommand()
	})

	form.SetBorder(true).SetTitle("Input")
	state.outputPage.AddItem(form, 0, 1, true)
	state.pages.SwitchToPage("Output")
}

func (state *State) buildOutputContentWithoutInputs(cmd Command) {
	state.parsedCommand = cmd.Template
	state.runCommand()
}

func (state *State) runCommand() {
	state.outputPage.Clear()

	content := tview.NewTextView().
		SetWrap(true).
		SetText(state.parsedCommand)

	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			state.pages.SwitchToPage("Commands")
			return nil
		}
		return event
	})

	state.outputPage.AddItem(content, 0, 1, true)
	state.pages.SwitchToPage("Output")
	state.app.SetFocus(content)
}

func (state *State) displayText(text string) {
	state.outputPage.Clear()

	content := tview.NewTextView().
		SetWrap(true).
		SetText(text)

	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			state.pages.SwitchToPage("Commands")
			return nil
		}
		return event
	})

	state.outputPage.AddItem(content, 0, 1, true)
	state.pages.SwitchToPage("Output")
	state.app.SetFocus(content)
}

func (state *State) buildCommandPage() {
	state.commandsPage = tview.NewFlex().SetDirection(tview.FlexRow)

	commandsList := tview.NewList()
	for _, cmd := range state.config.Commands {
		commandsList.AddItem(cmd.Title, "", 0, func() {
			state.buildOutputContent(cmd)
		})
	}

	commandsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				state.app.Stop()
				return nil

			case 'j':
				curr := commandsList.GetCurrentItem()
				if curr < commandsList.GetItemCount()-1 {
					commandsList.SetCurrentItem(curr + 1)
				}
				return nil

			case 'k':
				curr := commandsList.GetCurrentItem()
				if curr > 0 {
					commandsList.SetCurrentItem(curr - 1)
				}
				return nil
			}
		}
		return event
	})

	state.commandsPage.AddItem(commandsList, 0, 1, true)
	state.commandsPage.SetTitle(" Commands ").SetBorder(true)

	state.pages.AddPage("Commands", state.commandsPage, true, true)
}

func (state *State) buildOutputPage() {
	state.outputPage = tview.NewFlex().SetDirection(tview.FlexRow)
	state.outputPage.SetTitle(" Output ").SetBorder(true)

	state.pages.AddPage("Output", state.outputPage, true, false)
}

func initState(config CommanderFile) *State {
	state := &State{
		config: config,
		app:    tview.NewApplication(),
		pages:  tview.NewPages(),
	}

	state.buildOutputPage()
	state.buildCommandPage()

	return state
}

func main() {
	config, err := loadConfig("./commander.yaml")
	if err != nil {
		log.Fatal(err)
	}

	state := initState(config)

	if err := state.app.SetRoot(state.pages, true).SetFocus(state.pages).Run(); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(filePath string) (CommanderFile, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return CommanderFile{}, err
	}

	var commanderFile CommanderFile
	err = yaml.Unmarshal(file, &commanderFile)
	if err != nil {
		return CommanderFile{}, err
	}

	return commanderFile, nil
}
