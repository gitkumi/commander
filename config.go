package main

import (
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
	Title       string            `yaml:"title"`
	Template    string            `yaml:"command"`
	Description string            `yaml:"description"`
	Inputs      []CommandInput    `yaml:"inputs"`
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
