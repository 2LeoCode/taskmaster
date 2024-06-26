package parsing

import "encoding/json"

type Config struct {
	Tasks []Task
}

type Task struct {
	Command            string
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
	Permissions        int
}

func (this *Task) UnmarshalJSON(data []byte) error {
	type LocalTask Task

	task := LocalTask{
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
		Permissions:        -1,
	}

	if err := json.Unmarshal(data, &task); err != nil {
		return NewParseConfigError(err)
	}

	*this = Task(task)
	return nil
}
