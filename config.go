package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type CommanderFile struct {
	Commands    []Command         `yaml:"commands"`
	Environment map[string]string `yaml:"environment"`
}

type CommandInput struct {
	Key          string   `yaml:"key"`
	DefaultValue string   `yaml:"defaultValue"`
	Choices      []string `yaml:"choices"`
}

type Command struct {
	Title    string            `yaml:"title"`
	Template string            `yaml:"command"`
	Inputs   []CommandInput    `yaml:"inputs"`
	Environment map[string]string `yaml:"environment"`
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

func detectPackageManager() string {
	lockFiles := map[string]string{
		"pnpm-lock.yaml": "pnpm",
		"yarn.lock":      "yarn",
		"bun.lockb":      "bun",
		"bun.lock":       "bun",
	}
	for file, manager := range lockFiles {
		if _, err := os.Stat(file); err == nil {
			return manager
		}
	}
	return "npm"
}

func loadPackageJSON(filePath string) (CommanderFile, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return CommanderFile{}, err
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(file, &pkg); err != nil {
		return CommanderFile{}, err
	}

	pm := detectPackageManager()

	var commands []Command
	for name, script := range pkg.Scripts {
		commands = append(commands, Command{
			Title:    name,
			Template: pm + " run " + name,
		})
		_ = script
	}

	return CommanderFile{Commands: commands}, nil
}

func loadMakefile(filePath string) (CommanderFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return CommanderFile{}, err
	}
	defer file.Close()

	var commands []Command
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Match target lines like "build:" or "test: deps"
		// Skip lines starting with tab/space (recipes) and variables
		if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, ".") || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, ':'); idx > 0 {
			// Skip variable assignments (`:=`, `::=`)
			if idx+1 < len(line) && line[idx+1] == '=' {
				continue
			}
			target := strings.TrimSpace(line[:idx])
			if target == "" || strings.ContainsAny(target, " $%") {
				continue
			}
			commands = append(commands, Command{
				Title:    target,
				Template: "make " + target,
			})
		}
	}

	return CommanderFile{Commands: commands}, scanner.Err()
}
