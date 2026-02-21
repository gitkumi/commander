package main

import (
	"encoding/json"
	"os"

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

	var commands []Command
	for name, script := range pkg.Scripts {
		commands = append(commands, Command{
			Title:    name,
			Template: script,
		})
	}

	return CommanderFile{Commands: commands}, nil
}
