package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	Tasks []Task
}

type Task struct {
	Command            *string
	Arguments          []string
	StartAtLaunch      bool
	Instances          uint
	Restart            string
	RestartAttempts    uint
	ExpectedExitStatus int
	StartTime          uint
	StopTime           uint
	StopSignal         string
	Stdout             string
	Stderr             string
	LogDir             string
	Environment        map[string]string
	WorkingDirectory   string
	Permissions        *int
}

type TaskParseError struct {
	property string
	value    string
	info     string
}

func (this TaskParseError) Error() string {
	return fmt.Sprintf(
		"Invalid value for property %s: %s (%s)",
		this.property, this.value, this.info,
	)
}

func NewTaskParseError(property, value, info string) TaskParseError {
	return TaskParseError{property, value, info}
}

type TaskEnumParseError struct {
	TaskParseError
}

func NewTaskEnumParseError(property, value string, enum []string) TaskEnumParseError {
	return TaskEnumParseError{
		NewTaskParseError(
			property,
			value,
			fmt.Sprintf("must be one of %s", "'"+strings.Join(enum, "', '")+"'"),
		),
	}
}

func (this *Task) UnmarshalJSON(data []byte) error {
	type LocalTask Task

	task := LocalTask{
		Command:            nil,
		Arguments:          []string{},
		StartAtLaunch:      true,
		Instances:          1,
		Restart:            "on-failure",
		RestartAttempts:    5,
		ExpectedExitStatus: 0,
		StartTime:          0,
		StopTime:           5000,
		StopSignal:         "SIGQUIT",
		Stdout:             "redirect",
		Stderr:             "redirect",
		LogDir:             "/var/log/taskmaster",
		Environment:        map[string]string{},
		WorkingDirectory:   ".",
		Permissions:        nil,
	}

	if err := json.Unmarshal(data, &task); err != nil {
		return NewParseError(err.Error())
	}

	restartValues := []string{"always", "never", "on-failure", "unless-stopped"}
	stopSignalValues := []string{"SIGINT", "SIGQUIT", "SIGTERM", "SIGUSR1", "SIGUSR2", "SIGSTOP", "SIGTSTP"}
	stdioValues := []string{"ignore", "inherit", "redirect"}

	if task.Command == nil {
		return errors.New("You need to provide a command to run for each task")
	}

	if task.Instances == 0 {
		return NewTaskParseError("instances", "0", "must be 1 or greater")
	}

	if !slices.Contains(restartValues, task.Restart) {
		return NewTaskEnumParseError("restart", task.Restart, restartValues)
	}

	if !slices.Contains(stopSignalValues, task.StopSignal) {
		return NewTaskEnumParseError("stopSignal", task.StopSignal, stopSignalValues)
	}

	if !slices.Contains(stdioValues, task.Stdout) {
		return NewTaskEnumParseError("stdout", task.Stdout, stdioValues)
	}

	if !slices.Contains(stdioValues, task.Stderr) {
		return NewTaskEnumParseError("stderr", task.Stderr, stdioValues)
	}

	if fileInfo, err := os.Stat(task.WorkingDirectory); err != nil {
		return errors.New(fmt.Sprintf("Failed to get information on the current working directory (%s)", err))
	} else if !fileInfo.IsDir() {
		return NewTaskParseError("workingDirectory", task.WorkingDirectory, "not a directory")
	}

	if err := os.MkdirAll(task.LogDir, os.ModePerm); err != nil {
		return errors.New(fmt.Sprintf("Failed to open log directory (%s)", err))
	}

	if task.Permissions != nil && *task.Permissions > 777 {
		return NewTaskParseError("permissions", strconv.Itoa(*task.Permissions), "umask value cannot be greater than 777")
	}

	*this = Task(task)
	return nil
}
