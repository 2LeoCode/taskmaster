package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"taskmaster/utils"
)

type Config struct {
	Tasks  []Task
	LogDir string
}

type Task struct {
	Name               *string
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
	Environment        map[string]string
	WorkingDirectory   string
	Permissions        *uint
}

func (this *Config) String() string {
	return fmt.Sprintf(
		"{\n"+
			"  Tasks: %s\n"+
			"  LogDir: %s\n"+
			"}",
		"[\n    "+
			strings.Join(utils.Transform(this.Tasks, func(i int, task *Task) string {
				return fmt.Sprintf(
					"{\n"+
						"      Name: %s\n"+
						"      Command: %s\n"+
						"      Arguments: %s\n"+
						"      StartAtLaunch: %t\n"+
						"      Instances: %d\n"+
						"      Restart: %s\n"+
						"      RestartAttempts: %d\n"+
						"      ExpectedExitStatus: %d\n"+
						"      StartTime: %d\n"+
						"      StopTime: %d\n"+
						"      StopSignal: %s\n"+
						"      Stdout: %s\n"+
						"      Stderr: %s\n"+
						"      Environment: %+v\n"+
						"      WorkingDirectory: %s\n"+
						"      Permissions: %s\n"+
						"    }",
					utils.PointerFormat(task.Name),
					utils.PointerFormat(task.Command),
					task.Arguments,
					task.StartAtLaunch,
					task.Instances,
					task.Restart,
					task.RestartAttempts,
					task.ExpectedExitStatus,
					task.StartTime,
					task.StopTime,
					task.StopSignal,
					task.Stdout,
					task.Stderr,
					task.Environment,
					task.WorkingDirectory,
					utils.PointerFormat(task.Permissions),
				)
			}), "\n    ")+
			"\n  ]",
		this.LogDir,
	)
}

func (this *Task) String() string {
	return fmt.Sprintf(
		"{\n"+
			"  Name: %s\n"+
			"  Command: %s\n"+
			"  Arguments: %s\n"+
			"  StartAtLaunch: %t\n"+
			"  Instances: %d\n"+
			"  Restart: %s\n"+
			"  RestartAttempts: %d\n"+
			"  ExpectedExitStatus: %d\n"+
			"  StartTime: %d\n"+
			"  StopTime: %d\n"+
			"  StopSignal: %s\n"+
			"  Stdout: %s\n"+
			"  Stderr: %s\n"+
			"  Environment: %+v\n"+
			"  WorkingDirectory: %s\n"+
			"  Permissions: %s\n"+
			"}",
		utils.PointerFormat(this.Name),
		utils.PointerFormat(this.Command),
		this.Arguments,
		this.StartAtLaunch,
		this.Instances,
		this.Restart,
		this.RestartAttempts,
		this.ExpectedExitStatus,
		this.StartTime,
		this.StopTime,
		this.StopSignal,
		this.Stdout,
		this.Stderr,
		this.Environment,
		this.WorkingDirectory,
		utils.PointerFormat(this.Permissions),
	)
}

type TaskPropertyError struct {
	property string
}

type TaskMissingPropertyError struct {
	TaskPropertyError
}

type TaskInvalidPropertyError struct {
	TaskPropertyError
	value string
	info  string
}

type TaskEnumPropertyError struct {
	TaskInvalidPropertyError
}

func (this TaskMissingPropertyError) Error() string {
	return fmt.Sprintf("Missing required property for task: %s", this.property)
}

func (this TaskInvalidPropertyError) Error() string {
	return fmt.Sprintf(
		"Invalid value for property %s: %s (%s)",
		this.property, this.value, this.info,
	)
}

func newTaskMissingPropertyError(property string) TaskMissingPropertyError {
	return TaskMissingPropertyError{
		TaskPropertyError{property},
	}
}

func newTaskInvalidPropertyError(property, value, info string) TaskInvalidPropertyError {
	return TaskInvalidPropertyError{
		TaskPropertyError{property},
		value, info,
	}
}

func newTaskEnumPropertyError(property, value string, enum []string) TaskEnumPropertyError {
	return TaskEnumPropertyError{
		newTaskInvalidPropertyError(
			property,
			value,
			fmt.Sprintf("must be one of %s", "'"+strings.Join(enum, "', '")+"'"),
		),
	}
}

func (this *Task) UnmarshalJSON(data []byte) error {
	type LocalTask Task

	task := LocalTask{
		Name:               nil,
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
		Environment:        map[string]string{},
		WorkingDirectory:   ".",
		Permissions:        nil,
	}

	if err := json.Unmarshal(data, &task); err != nil {
		return err
	}

	restartValues := []string{"always", "never", "on-failure", "unless-stopped"}
	stopSignalValues := []string{"SIGINT", "SIGQUIT", "SIGTERM", "SIGUSR1", "SIGUSR2", "SIGSTOP", "SIGTSTP"}
	stdioValues := []string{"ignore", "inherit", "redirect"}

	switch {
	case task.Name == nil:
		return newTaskMissingPropertyError("name")

	case regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(*task.Name) == false:
		return newTaskInvalidPropertyError("name", *task.Name, "must match the pattern /^[a-zA-Z0-9_-]$/")

	case task.Command == nil:
		return newTaskMissingPropertyError("command")

	case task.Instances == 0:
		return newTaskInvalidPropertyError("instances", "0", "must be 1 or greater")

	case !slices.Contains(restartValues, task.Restart):
		return newTaskEnumPropertyError("restart", task.Restart, restartValues)

	case !slices.Contains(stopSignalValues, task.StopSignal):
		return newTaskEnumPropertyError("stopSignal", task.StopSignal, stopSignalValues)

	case !slices.Contains(stdioValues, task.Stdout):
		return newTaskEnumPropertyError("stdout", task.Stdout, stdioValues)

	case !slices.Contains(stdioValues, task.Stderr):
		return newTaskEnumPropertyError("stderr", task.Stderr, stdioValues)

	case task.Permissions != nil && *task.Permissions > 777:
		return newTaskInvalidPropertyError("permissions", strconv.Itoa(int(*task.Permissions)), "umask value cannot be greater than 777")
	}

	if fileInfo, err := os.Stat(task.WorkingDirectory); err != nil {
		return errors.New(fmt.Sprintf("Failed to get information on the current working directory (%s)", err))
	} else if !fileInfo.IsDir() {
		return newTaskInvalidPropertyError("workingDirectory", task.WorkingDirectory, "not a directory")
	}

	*this = Task(task)
	return nil
}
