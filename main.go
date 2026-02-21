package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	config, err := resolveConfig()
	if err != nil {
		log.Fatal(err)
	}

	state := initState(config)

	if err := state.app.SetRoot(state.pages, true).SetFocus(state.pages).Run(); err != nil {
		log.Fatal(err)
	}
}

func resolveConfig() (CommanderFile, error) {
	if len(os.Args) > 1 {
		return loadConfig(os.Args[1])
	}

	if _, err := os.Stat("./commander.yaml"); err == nil {
		return loadConfig("./commander.yaml")
	}

	if _, err := os.Stat("./package.json"); err == nil {
		return loadPackageJSON("./package.json")
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return CommanderFile{}, fmt.Errorf("no config found")
		}
		configDir = filepath.Join(home, ".config")
	}
	return loadConfig(filepath.Join(configDir, "commander", "commander.yaml"))
}
