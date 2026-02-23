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
	tview.Styles.ContrastBackgroundColor = tcell.ColorGreen
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorGreen
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

	formValues := make(map[string]string)
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	var fields []tview.Primitive

	cmdLabel := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]" + cmd.Template + "[-]")
	cmdLabel.SetBackgroundColor(tcell.ColorDefault)
	container.AddItem(cmdLabel, 1, 0, false)
	container.AddItem(nil, 1, 0, false)

	for i, input := range cmd.Inputs {
		formValues[input.Key] = input.DefaultValue
		key := input.Key

		if i > 0 {
			container.AddItem(nil, 1, 0, false)
		}

		label := tview.NewTextView().
			SetDynamicColors(true).
			SetText("[gray]" + key + "[-]")
		label.SetBackgroundColor(tcell.ColorDefault)
		container.AddItem(label, 1, 0, false)

		if len(input.Choices) > 0 {
			dropdown := createDropdown(formValues, input)
			dropdown.SetFieldBackgroundColor(tcell.ColorDefault)
			dropdown.SetFocusFunc(func() {
				dropdown.SetFieldBackgroundColor(tcell.ColorGreen)
				dropdown.SetFieldTextColor(tcell.ColorWhite)
			})
			dropdown.SetBlurFunc(func() {
				dropdown.SetFieldBackgroundColor(tcell.ColorDefault)
				dropdown.SetFieldTextColor(tcell.ColorWhite)
			})
			container.AddItem(dropdown, 1, 0, true)
			fields = append(fields, dropdown)
		} else {
			field := tview.NewInputField().
				SetLabel("> ").
				SetText(input.DefaultValue).
				SetFieldWidth(30).
				SetFieldBackgroundColor(tcell.ColorDefault).
				SetFieldTextColor(tcell.ColorWhite)
			field.SetChangedFunc(func(text string) {
				formValues[key] = text
			})
			field.SetFocusFunc(func() {
				field.SetFieldBackgroundColor(tcell.ColorGreen)
			})
			field.SetBlurFunc(func() {
				field.SetFieldBackgroundColor(tcell.ColorDefault)
			})
			container.AddItem(field, 1, 0, true)
			fields = append(fields, field)
		}
	}

	container.AddItem(nil, 1, 0, false)

	submitBtn := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[ Submit ]")
	submitBtn.SetBackgroundColor(tcell.ColorDefault)
	submitBtnRow := tview.NewFlex().
		AddItem(submitBtn, 10, 0, true).
		AddItem(nil, 0, 1, false)
	container.AddItem(submitBtnRow, 1, 0, true)
	fields = append(fields, submitBtn)

	container.AddItem(nil, 0, 1, false)

	submit := func() {
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
	}

	focusField := func(index int) {
		if index >= 0 && index < len(fields) {
			state.app.SetFocus(fields[index])
		}
	}

	keybindBar := createKeybindBar("")

	submitBtn.SetFocusFunc(func() {
		submitBtn.SetBackgroundColor(tcell.ColorGreen)
		submitBtn.SetTextColor(tcell.ColorWhite)
		keybindBar.SetText(" [yellow]Tab/S-Tab[white] next/prev  [yellow]Enter[white] submit  [yellow]Esc[white] back")
	})
	submitBtn.SetBlurFunc(func() {
		submitBtn.SetBackgroundColor(tcell.ColorDefault)
		submitBtn.SetTextColor(tcell.ColorDefault)
		keybindBar.SetText(" [yellow]Tab/S-Tab[white] next/prev  [yellow]Esc[white] back")
	})
	submitBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			state.pages.SwitchToPage("Commands")
			return nil
		case tcell.KeyEnter:
			submit()
			return nil
		case tcell.KeyTab:
			focusField(0)
			return nil
		case tcell.KeyBacktab:
			focusField(len(fields) - 2)
			return nil
		}
		return event
	})

	for i, f := range fields {
		idx := i
		switch field := f.(type) {
		case *tview.InputField:
			field.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEsc:
					state.pages.SwitchToPage("Commands")
					return nil
				case tcell.KeyTab:
					focusField(idx + 1)
					return nil
				case tcell.KeyBacktab:
					focusField(idx - 1)
					return nil
				}
				return event
			})
		case *tview.DropDown:
			choices := cmd.Inputs[idx].Choices
			field.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEsc:
					state.pages.SwitchToPage("Commands")
					return nil
				case tcell.KeyTab:
					focusField(idx + 1)
					return nil
				case tcell.KeyBacktab:
					focusField(idx - 1)
					return nil
				}
				if event.Key() == tcell.KeyRune {
					switch event.Rune() {
					case 'j':
						cur, _ := field.GetCurrentOption()
						next := (cur + 1) % len(choices)
						field.SetCurrentOption(next)
						formValues[cmd.Inputs[idx].Key] = choices[next]
						return nil
					case 'k':
						cur, _ := field.GetCurrentOption()
						next := cur - 1
						if next < 0 {
							next = len(choices) - 1
						}
						field.SetCurrentOption(next)
						formValues[cmd.Inputs[idx].Key] = choices[next]
						return nil
					}
				}
				return event
			})
		}
	}

	wrapper := tview.NewFlex().
		AddItem(nil, 1, 0, false).
		AddItem(container, 0, 1, true).
		AddItem(nil, 1, 0, false)
	keybindBar.SetText(" [yellow]Tab/S-Tab[white] next/prev  [yellow]Esc[white] back")
	state.inputPage.AddItem(wrapper, 0, 1, true)
	state.inputPage.AddItem(keybindBar, 1, 0, false)
	state.pages.SwitchToPage("Input")
	if len(fields) > 0 {
		state.app.SetFocus(fields[0])
	}
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

		fmt.Print("\n\033[2m(Esc to quit, Enter to continue)\033[0m")
		key := waitForKey()
		fmt.Println()
		fmt.Println()
		if key == 27 {
			state.app.Stop()
		}
	})
}

func (state *State) initCommandPage() {
	state.commandsPage = tview.NewFlex().SetDirection(tview.FlexRow)

	commandsList := tview.NewList()
	commandsList.SetSelectedBackgroundColor(tcell.ColorGreen)
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
		fmt.Print("\n\033[2m(Esc to quit, Enter to continue)\033[0m")
		key := waitForKey()
		fmt.Println()
		fmt.Println()
		if key == 27 {
			state.app.Stop()
		}
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
	maxLen := len("[ Select ]")
	for _, c := range input.Choices {
		if w := len(c) + 4; w > maxLen {
			maxLen = w
		}
	}

	dropdown := tview.NewDropDown().
		SetFieldWidth(maxLen).
		SetFieldTextColor(tcell.ColorWhite).
		SetTextOptions("  ", "", "  ", "", "[ Select ]").
		SetListStyles(
			tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorWhite),
			tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite),
		)

	dropdown.SetCurrentOption(0)
	formValues[input.Key] = input.Choices[0]

	dropdown.SetOptions(input.Choices, func(option string, index int) {
		formValues[input.Key] = option
	})

	return dropdown
}

func waitForKey() byte {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Scanln()
		return 0
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	return buf[0]
}
