package config

import (
	"encoding/json"
	"fmt"
	"os"
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
	config := Config{
		Tasks: []Task{},
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

	return &config, nil
}
