package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ParseError struct {
	cause string
}

func (this ParseError) Error() string {
	return fmt.Sprintf("Error while parsing configuration file: %s", this.cause)
}

func NewParseError(cause string) ParseError {
	return ParseError{cause}
}

func Parse(path string) (*Config, error) {
	if !strings.HasSuffix(path, ".json") {
		return nil, NewParseError("Invalid config file format (expected a json file)")
	}

	config := Config{
		Tasks:  []Task{},
		LogDir: "/var/log/taskmaster",
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NewParseError(err.Error())
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, NewParseError(err.Error())
	}

	if len(config.Tasks) == 0 {
		return nil, NewParseError("No task to run")
	}

	if err := os.MkdirAll(config.LogDir, os.ModePerm); err != nil {
		return nil, NewParseError(fmt.Sprintf("Failed to open log directory (%s)", err))
	}

	return &config, nil
}
