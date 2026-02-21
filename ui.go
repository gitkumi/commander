package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"golang.org/x/term"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type State struct {
	config       CommanderFile
	app          *tview.Application
	pages        *tview.Pages
	commandsPage *tview.Flex
	inputPage    *tview.Flex
}

func initState(config CommanderFile) *State {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDarkGray
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDarkGray
	tview.Styles.BorderColor = tcell.ColorDefault
	tview.Styles.TitleColor = tcell.ColorDefault
	tview.Styles.GraphicsColor = tcell.ColorDefault
	tview.Styles.PrimaryTextColor = tcell.ColorDefault
	tview.Styles.SecondaryTextColor = tcell.ColorDefault
	tview.Styles.TertiaryTextColor = tcell.ColorDefault
	tview.Styles.InverseTextColor = tcell.ColorDefault
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorDefault

	state := &State{
		config: config,
		app:    tview.NewApplication(),
		pages:  tview.NewPages(),
	}

	state.initCommandPage()
	state.initInputPage()

	return state
}

func (state *State) handleCommandSelection(cmd Command) {
	if len(cmd.Inputs) > 0 {
		state.buildInputPageContent(cmd)
		return
	}

	state.runCommand(cmd.Template, cmd.Environment)
}

func (state *State) buildInputPageContent(cmd Command) {
	state.inputPage.Clear()

	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)
	form.SetFieldTextColor(tcell.ColorBlack)
	form.SetButtonTextColor(tcell.ColorBlack)
	formValues := make(map[string]string)

	for _, input := range cmd.Inputs {
		formValues[input.Key] = input.DefaultValue

		if len(input.Choices) > 0 {
			dropdown := createDropdown(formValues, input)
			form.AddFormItem(dropdown)
			continue
		}

		form.AddInputField(input.Key, input.DefaultValue, 30, nil, func(text string) {
			formValues[input.Key] = text
		})
	}

	form.AddButton("Submit", func() {
		tpl, err := template.New("command").Parse(cmd.Template)
		if err != nil {
			state.suspendWithMessage(fmt.Sprintf("Template parse error: %v", err))
			return
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, formValues); err != nil {
			state.suspendWithMessage(fmt.Sprintf("Template execution error: %v", err))
			return
		}

		state.runCommand(buf.String(), cmd.Environment)
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			state.pages.SwitchToPage("Commands")
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			index, _ := form.GetFocusedItemIndex()
			if index > 0 {
				form.SetFocus(index - 1)
			}
			return nil
		}
		return event
	})

	state.inputPage.AddItem(form, 0, 1, true)
	state.inputPage.AddItem(createKeybindBar(" [yellow]Tab/S-Tab[white] next/prev  [yellow]Enter[white] submit  [yellow]Esc[white] back"), 1, 0, false)
	state.pages.SwitchToPage("Input")
	state.app.SetFocus(form)
}

func (state *State) runCommand(parsedCommand string, commandEnv map[string]string) {
	state.app.Suspend(func() {
		cmd := exec.Command("sh", "-c", parsedCommand)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		for key, value := range state.config.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
		for key, value := range commandEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}

		fmt.Printf("\033[1m$ %s\033[0m\n\n", parsedCommand)

		cmd.Run()

		fmt.Print("\n\033[2m(press any key)\033[0m")
		waitForKey()
		fmt.Println()
		fmt.Println()
	})
}

func (state *State) initCommandPage() {
	state.commandsPage = tview.NewFlex().SetDirection(tview.FlexRow)

	commandsList := tview.NewList()
	commandsList.SetSelectedBackgroundColor(tcell.ColorDarkGray)
	commandsList.SetSelectedTextColor(tcell.ColorWhite)
	commandsList.SetSecondaryTextColor(tcell.ColorGray)
	for _, cmd := range state.config.Commands {
		commandsList.AddItem(cmd.Title, cmd.Template, 0, func() {
			state.handleCommandSelection(cmd)
		})
	}

	commandsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			state.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
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

	listWrapper := tview.NewFlex().
		AddItem(nil, 1, 0, false).
		AddItem(commandsList, 0, 1, true).
		AddItem(nil, 1, 0, false)
	state.commandsPage.AddItem(listWrapper, 0, 1, true)
	state.commandsPage.AddItem(createKeybindBar(" [yellow]j/k[white] navigate  [yellow]Enter[white] select  [yellow]Esc[white] quit"), 1, 0, false)
	state.commandsPage.SetTitle(" Commands ").SetBorder(true)

	state.pages.AddPage("Commands", state.commandsPage, true, true)
}

func (state *State) suspendWithMessage(msg string) {
	state.app.Suspend(func() {
		fmt.Printf("\033[31m%s\033[0m\n", msg)
		fmt.Print("\n\033[2m(press any key)\033[0m")
		waitForKey()
		fmt.Println()
		fmt.Println()
	})
}

func (state *State) initInputPage() {
	state.inputPage = tview.NewFlex().SetDirection(tview.FlexRow)
	state.inputPage.SetTitle(" Input ").SetBorder(true)

	state.pages.AddPage("Input", state.inputPage, true, false)
}

func createKeybindBar(bindings string) *tview.TextView {
	bar := tview.NewTextView().
		SetDynamicColors(true).
		SetText(bindings)
	bar.SetBackgroundColor(tcell.ColorDefault)
	return bar
}

func createDropdown(formValues map[string]string, input CommandInput) *tview.DropDown {
	dropdown := tview.NewDropDown().
		SetLabel(input.Key + ": ").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDarkGray).
		SetFieldTextColor(tcell.ColorBlack)

	dropdown.SetCurrentOption(0)
	formValues[input.Key] = input.Choices[0]

	dropdown.SetOptions(input.Choices, func(option string, index int) {
		formValues[input.Key] = option
	})

	dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'j':
				currentIndex, _ := dropdown.GetCurrentOption()
				newIndex := currentIndex + 1
				if newIndex >= len(input.Choices) {
					newIndex = 0
				}
				dropdown.SetCurrentOption(newIndex)
				formValues[input.Key] = input.Choices[newIndex]
				return nil
			case 'k':
				currentIndex, _ := dropdown.GetCurrentOption()
				newIndex := currentIndex - 1
				if newIndex < 0 {
					newIndex = len(input.Choices) - 1
				}
				dropdown.SetCurrentOption(newIndex)
				formValues[input.Key] = input.Choices[newIndex]
				return nil
			}
		}
		return event
	})

	return dropdown
}

func waitForKey() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Scanln()
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 1)
	os.Stdin.Read(buf)
}
