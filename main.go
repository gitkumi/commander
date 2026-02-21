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

	home, _ := os.UserHomeDir()

	dir, err := os.Getwd()
	if err != nil {
		return CommanderFile{}, err
	}

	for {
		type candidate struct {
			file string
			load func(string) (CommanderFile, error)
		}
		candidates := []candidate{
			{"commander.yaml", loadConfig},
			{"package.json", loadPackageJSON},
			{"Makefile", loadMakefile},
		}

		for _, c := range candidates {
			path := filepath.Join(dir, c.file)
			if _, err := os.Stat(path); err == nil {
				return c.load(path)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir || dir == home {
			break
		}
		dir = parent
	}

	return CommanderFile{}, fmt.Errorf("no config found (searched up to %s)", home)
}
