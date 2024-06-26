package parsing

import (
	"encoding/json"
	"fmt"
	"os"
)

type ParseConfigError struct {
	cause error
}

func (this *ParseConfigError) Error() string {
	return fmt.Sprintf("Error while parsing configuration file: %s", this.cause)
}

func NewParseConfigError(cause error) *ParseConfigError {
	return &ParseConfigError{
		cause,
	}
}

func ParseConfig(path string) (*Config, error) {
	config := Config{
		Tasks: []Task{},
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NewParseConfigError(err)
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, NewParseConfigError(err)
	}
	return &config, nil
}
