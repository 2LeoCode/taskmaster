package config

import (
	"fmt"
	"strconv"
	"taskmaster/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseValidFull(t *testing.T) {
	config, err := Parse("testdata/valid_full.json")

	require.Nil(t, err)
	require.Equal(
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
			"      Stdout: redirect\n"+
			"      Stderr: ignore\n"+
			"      Environment: map[WELCOME:Hello world!]\n"+
			"      WorkingDirectory: /tmp\n"+
			"      Permissions: "+fmt.Sprint(utils.Must(strconv.ParseUint("777", 8, 0)))+"\n"+
			"    }\n"+
			"  ]\n"+
			"  LogDir: /tmp/taskmaster-logs\n"+
			"}",
		config.String(),
	)
}

func TestParseValidPartial(t *testing.T) {
	config, err := Parse("testdata/valid_partial.json")

	require.Nil(t, err)
	require.Equal(
		t,
		"{\n"+
			"  Tasks: [\n"+
			"    {\n"+
			"      Name: valid-partial\n"+
			"      Command: ls\n"+
			"      Arguments: []\n"+
			"      StartAtLaunch: true\n"+
			"      Instances: 1\n"+
			"      Restart: unless-stopped\n"+
			"      RestartAttempts: 5\n"+
			"      ExpectedExitStatus: 0\n"+
			"      StartTime: 0\n"+
			"      StopTime: 5000\n"+
			"      StopSignal: SIGTERM\n"+
			"      Stdout: redirect\n"+
			"      Stderr: redirect\n"+
			"      Environment: map[]\n"+
			"      WorkingDirectory: .\n"+
			"      Permissions: "+fmt.Sprint(utils.GetUmask())+"\n"+
			"    }\n"+
			"  ]\n"+
			"  LogDir: /tmp/taskmaster-logs\n"+
			"}",
		config.String(),
	)
}

func TestParseInvalidNoName(t *testing.T) {
	_, err := Parse("testdata/invalid_no_name.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: Missing required property for task: name", err.Error())
}

func TestParseInvalidName(t *testing.T) {
	_, err := Parse("testdata/invalid_name.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: Invalid value for property name: hello### (must match the pattern ^[a-zA-Z0-9_-]+$)", err.Error())
}

func TestParseInvalidNoCommand(t *testing.T) {
	_, err := Parse("testdata/invalid_no_command.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: Missing required property for task: command", err.Error())
}

func TestParseInvalidNoTask(t *testing.T) {
	_, err := Parse("testdata/invalid_no_task.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: No task to run", err.Error())
}

func TestParseInvalidEmptyTasks(t *testing.T) {
	_, err := Parse("testdata/invalid_empty_tasks.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: No task to run", err.Error())
}

func TestParseInvalidNoJson(t *testing.T) {
	_, err := Parse("testdata/invalid_no_json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: Invalid config file format (expected a json file)", err.Error())
}

func TestParseInvalidType1(t *testing.T) {
	_, err := Parse("testdata/invalid_type1.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: json: cannot unmarshal number into Go struct field Config.Tasks of type string", err.Error())
}

func TestParseInvalidType2(t *testing.T) {
	_, err := Parse("testdata/invalid_type2.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: json: cannot unmarshal number into Go struct field Config.Tasks of type bool", err.Error())
}

func TestParseInvalidType3(t *testing.T) {
	_, err := Parse("testdata/invalid_type3.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: Invalid value for property restart: sometimes (must be one of 'always', 'never', 'on-failure', 'unless-stopped')", err.Error())
}

func TestParseInvalidType4(t *testing.T) {
	_, err := Parse("testdata/invalid_type4.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: json: cannot unmarshal number -10 into Go struct field Config.Tasks of type uint", err.Error())
}

func TestParseInvalidType5(t *testing.T) {
	_, err := Parse("testdata/invalid_type5.json")

	require.NotNil(t, err)
	require.Equal(t, "Error while parsing configuration file: json: cannot unmarshal object into Go struct field Config.Tasks of type string", err.Error())
}
