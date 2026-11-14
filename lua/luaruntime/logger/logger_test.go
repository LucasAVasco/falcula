package logger

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func newLogger(t *testing.T) *Logger {
	t.Helper()
	l, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l
}

func TestServiceLogs_AreWrittenToTheLogFileAndTheHandler(t *testing.T) {
	l := newLogger(t)

	var uiLogs strings.Builder
	l.SetOnServiceLog(func(b []byte) (int, error) { return uiLogs.Write(b) })

	client := l.GetServicesMultiplexer().NewClient("web", "stdout", color.New(color.FgRed))
	if _, err := client.Write([]byte("line 1\nline 2\n")); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(l.GetLogFilePath())
	if err != nil {
		t.Fatal(err)
	}

	// The 'syslog' format recognized by Lnav: "<RFC3339 time> <name> <level>[<client id>]: <line>"
	pattern := regexp.MustCompile(`^\S+ web stdout\[\d+\]: line 1\n\S+ web stdout\[\d+\]: line 2\n$`)
	if !pattern.Match(content) {
		t.Errorf("log file content = %q", content)
	}

	if got := uiLogs.String(); !strings.Contains(got, "line 1\n") || !strings.Contains(got, "line 2\n") || !strings.Contains(got, "web stdout: ") {
		t.Errorf("UI logs = %q", got)
	}
}

func TestDebugAndErrorLogs_UseTheirHandlers(t *testing.T) {
	l := newLogger(t)

	var debug, errs []string
	l.SetOnDebugLog(func(b []byte) (int, error) { debug = append(debug, string(b)); return len(b), nil })
	l.SetOnErrorLog(func(err error) (int, error) { errs = append(errs, err.Error()); return 0, nil })

	l.Append("message", 2, errors.New("failure"))
	_, _ = l.Write([]byte("written"))
	_, _ = l.LogDebug("debug")
	_, _ = l.LogError(errors.New("direct"))

	if strings.Join(debug, "|") != "message|2|written|debug" {
		t.Errorf("debug logs = %q", debug)
	}
	if strings.Join(errs, "|") != "failure|direct" {
		t.Errorf("error logs = %q", errs)
	}
}

func TestClose_RemovesTheLogFile(t *testing.T) {
	l, err := New()
	if err != nil {
		t.Fatal(err)
	}
	path := l.GetLogFilePath()

	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("the log file must be removed, stat error: %v", err)
	}
	if l.GetLogFilePath() != "" {
		t.Error("the log file path must be empty after closing")
	}
	if err := l.Close(); err != nil {
		t.Fatalf("closing again must not fail: %v", err)
	}

	// Writing after closing must not crash
	client := l.GetServicesMultiplexer().NewClient("late", "stdout", color.New(color.FgRed))
	if _, err := client.Write([]byte("x")); err == nil {
		t.Error("writing to a closed logger must return an error")
	}
}
