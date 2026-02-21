package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	configPath := defaultConfigPath()
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	state := initState(config)

	if err := state.app.SetRoot(state.pages, true).SetFocus(state.pages).Run(); err != nil {
		log.Fatal(err)
	}
}

func defaultConfigPath() string {
	// Prefer ./commander.yaml if it exists in the current directory
	if _, err := os.Stat("./commander.yaml"); err == nil {
		return "./commander.yaml"
	}

	// Fall back to XDG config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "./commander.yaml"
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "commander", "commander.yaml")
}
