package terminal

import (
	"os"
	"testing"
)

func newEnvironment(vars map[string]string, stdinIsTerminal, stdoutIsTerminal bool) environment {
	stdin, stdout := os.NewFile(0, "stdin"), os.NewFile(1, "stdout")
	return environment{
		getenv: func(key string) string { return vars[key] },
		isTerminal: func(file *os.File) bool {
			if file == stdin {
				return stdinIsTerminal
			}
			return stdoutIsTerminal
		},
		stdin:  stdin,
		stdout: stdout,
	}
}

func TestNonInteractiveReason(t *testing.T) {
	tests := []struct {
		name   string
		vars   map[string]string
		stdin  bool
		stdout bool
		want   Reason
	}{
		{"interactive terminal", nil, true, true, ReasonNone},
		{"CI=true", map[string]string{"CI": "true"}, true, true, ReasonCI},
		{"CI=1", map[string]string{"CI": "1"}, true, true, ReasonCI},
		{"CI=false is not CI", map[string]string{"CI": "false"}, true, true, ReasonNone},
		{"CI=0 is not CI", map[string]string{"CI": "0"}, true, true, ReasonNone},
		{"provider variable", map[string]string{"GITHUB_ACTIONS": "true"}, true, true, ReasonCI},
		{"provider variable with CI=false", map[string]string{"CI": "false", "JENKINS_URL": "http://ci"}, true, true, ReasonCI},
		{"output redirected", nil, true, false, ReasonOutputRedirected},
		{"input not a terminal", nil, false, true, ReasonInputNotTerminal},
		{"CI takes precedence", map[string]string{"CI": "true"}, false, false, ReasonCI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newEnvironment(tt.vars, tt.stdin, tt.stdout).nonInteractiveReason()
			if got != tt.want {
				t.Errorf("nonInteractiveReason() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNonInteractiveReason_Pipe(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reader.Close(); writer.Close() })

	env := environment{getenv: func(string) string { return "" }, isTerminal: isTerminal, stdin: reader, stdout: writer}
	if got := env.nonInteractiveReason(); got != ReasonOutputRedirected {
		t.Errorf("nonInteractiveReason() with pipes = %q, want %q", got, ReasonOutputRedirected)
	}
}
