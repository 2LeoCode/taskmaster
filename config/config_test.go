package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValidFull(t *testing.T) {
	config, err := Parse("testdata/valid_full.json")

	assert.Nil(t, err)
	assert.Equal(
		t,
		"{\n"+
			"  Tasks: [\n"+
			"    {\n"+
			"      Name: valid-full\n"+
			"      Command: bash\n"+
			"      Arguments: [-c echo $WELCOME]\n"+
			"      StartAtLaunch: false\n"+
			"      Instances: 42\n"+
			"      Restart: on-failure\n"+
			"      RestartAttempts: 42\n"+
			"      ExpectedExitStatus: 42\n"+
			"      StartTime: 42\n"+
			"      StopTime: 42\n"+
			"      StopSignal: SIGTERM\n"+
			"      Stdout: inherit\n"+
			"      Stderr: inherit\n"+
			"      Environment: map[WELCOME:Hello world!]\n"+
			"      WorkingDirectory: /tmp\n"+
			"      Permissions: 777\n"+
			"    }\n"+
			"  ]\n"+
			"  LogDir: /tmp/taskmaster-logs\n"+
			"}",
		config.String(),
	)
}

func TestParseValidPartial(t *testing.T) {
	config, err := Parse("testdata/valid_partial.json")

	assert.Nil(t, err)
	assert.Equal(
		t,
		"{\n"+
			"  Tasks: [\n"+
			"    {\n"+
			"      Name: valid-partial\n"+
			"      Command: ls\n"+
			"      Arguments: []\n"+
			"      StartAtLaunch: true\n"+
			"      Instances: 1\n"+
			"      Restart: on-failure\n"+
			"      RestartAttempts: 5\n"+
			"      ExpectedExitStatus: 0\n"+
			"      StartTime: 0\n"+
			"      StopTime: 5000\n"+
			"      StopSignal: SIGQUIT\n"+
			"      Stdout: redirect\n"+
			"      Stderr: redirect\n"+
			"      Environment: map[]\n"+
			"      WorkingDirectory: .\n"+
			"      Permissions: (nil)\n"+
			"    }\n"+
			"  ]\n"+
			"  LogDir: /var/log/taskmaster\n"+
			"}",
		config.String(),
	)
}

func TestParseInvalidNoName(t *testing.T) {
	_, err := Parse("testdata/invalid_no_name.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: Missing required property for task: name", err.Error())
}

func TestParseInvalidName(t *testing.T) {
	_, err := Parse("testdata/invalid_name.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: Invalid value for property name: hello### (must match the pattern /^[a-zA-Z0-9_-]$/)", err.Error())
}

func TestParseInvalidNoCommand(t *testing.T) {
	_, err := Parse("testdata/invalid_no_command.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: Missing required property for task: command", err.Error())
}

func TestParseInvalidNoTask(t *testing.T) {
	_, err := Parse("testdata/invalid_no_task.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: No task to run", err.Error())
}

func TestParseInvalidEmptyTasks(t *testing.T) {
	_, err := Parse("testdata/invalid_empty_tasks.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: No task to run", err.Error())
}

func TestParseInvalidNoJson(t *testing.T) {
	_, err := Parse("testdata/invalid_no_json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: Invalid config file format (expected a json file)", err.Error())
}

func TestParseInvalidType1(t *testing.T) {
	_, err := Parse("testdata/invalid_type1.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: json: cannot unmarshal number into Go struct field Config.Tasks of type string", err.Error())
}

func TestParseInvalidType2(t *testing.T) {
	_, err := Parse("testdata/invalid_type2.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: json: cannot unmarshal number into Go struct field Config.Tasks of type bool", err.Error())
}

func TestParseInvalidType3(t *testing.T) {
	_, err := Parse("testdata/invalid_type3.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: Invalid value for property restart: sometimes (must be one of 'always', 'never', 'on-failure', 'unless-stopped')", err.Error())
}

func TestParseInvalidType4(t *testing.T) {
	_, err := Parse("testdata/invalid_type4.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: json: cannot unmarshal number -10 into Go struct field Config.Tasks of type uint", err.Error())
}

func TestParseInvalidType5(t *testing.T) {
	_, err := Parse("testdata/invalid_type5.json")

	assert.NotNil(t, err)
	assert.Equal(t, "Error while parsing configuration file: json: cannot unmarshal object into Go struct field Config.Tasks of type string", err.Error())
}
