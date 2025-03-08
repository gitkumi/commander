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

type UI struct {
	app          *tview.Application
	config       CommanderFile
	pages        *tview.Pages
	outputPage   *tview.Flex
	commandsPage *tview.Flex
}

func (ui *UI) buildOutputContent(cmd Command) {
	ui.outputPage.Clear()

	if len(cmd.Inputs) > 0 {
		ui.buildOutputContentWithInputs(cmd)
		return
	}

	ui.buildOutputContentWithoutInputs(cmd)
}

func (ui *UI) buildOutputContentWithInputs(cmd Command) {
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
			ui.displayContent(fmt.Sprintf("Template parse error: %v", err))
			return
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, formValues); err != nil {
			ui.displayContent(fmt.Sprintf("Template execution error: %v", err))
			return
		}

		ui.displayContent(buf.String())
	})

	form.SetBorder(true).SetTitle("Input")
	ui.outputPage.AddItem(form, 0, 1, true)
	ui.pages.SwitchToPage("Output")
}

func (ui *UI) buildOutputContentWithoutInputs(cmd Command) {
	ui.displayContent(cmd.Template)
}

func (ui *UI) displayContent(text string) {
	ui.outputPage.Clear()

	content := tview.NewTextView().
		SetWrap(true).
		SetText(text)

	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			ui.pages.SwitchToPage("Commands")
			return nil
		}
		return event
	})

	ui.outputPage.AddItem(content, 0, 1, true)
	ui.pages.SwitchToPage("Output")
	ui.app.SetFocus(content)
}

func (ui *UI) buildCommandPage() {
	ui.commandsPage = tview.NewFlex().SetDirection(tview.FlexRow)

	commandsList := tview.NewList()
	for _, cmd := range ui.config.Commands {
		commandsList.AddItem(cmd.Title, "", 0, func() {
			ui.buildOutputContent(cmd)
		})
	}

	commandsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				ui.app.Stop()
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

	ui.commandsPage.AddItem(commandsList, 0, 1, true)
	ui.commandsPage.SetTitle(" Commands ").SetBorder(true)

	ui.pages.AddPage("Commands", ui.commandsPage, true, true)
}

func (ui *UI) buildOutputPage() {
	ui.outputPage = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.outputPage.SetTitle(" Output ").SetBorder(true)

	ui.pages.AddPage("Output", ui.outputPage, true, false)
}

func buildUI(app *tview.Application, config CommanderFile) *UI {
	ui := &UI{
		app:    app,
		config: config,
		pages:  tview.NewPages(),
	}

	ui.buildOutputPage()
	ui.buildCommandPage()

	return ui
}

func main() {
	config, err := loadConfig("./commander.yaml")
	if err != nil {
		log.Fatal(err)
	}

	app := tview.NewApplication()
	ui := buildUI(app, config)

	if err := app.SetRoot(ui.pages, true).SetFocus(ui.pages).Run(); err != nil {
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
