package adapter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

func newProcess(t *testing.T, shellCommand string) *process.Process {
	t.Helper()
	mux := multiplexer.New(func(_ *multiplexer.Client, b []byte) (int, error) { return len(b), nil })
	p, err := process.New(&process.Options{ManualStart: true, Shell: true, Multiplexer: mux, Name: "test"}, shellCommand)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// waitEnd waits for the step callback with a timeout guard
func waitEnd(t *testing.T, ended <-chan *iface.ExitInfo) *iface.ExitInfo {
	t.Helper()
	select {
	case info := <-ended:
		return info
	case <-time.After(10 * time.Second):
		t.Fatal("the step did not end")
		return nil
	}
}

func TestSerialProcesses_RunInOrderUntilTheEnd(t *testing.T) {
	testutil.RequireShell(t)
	log := filepath.Join(t.TempDir(), "log")

	ended := make(chan *iface.ExitInfo, 1)
	step := SerialProcessesToStep([]*process.Process{
		newProcess(t, "echo first >> "+log),
		newProcess(t, "echo second >> "+log),
	}, func(info *iface.ExitInfo, err error) {
		if err != nil {
			t.Errorf("unexpected start error: %v", err)
		}
		ended <- info
	})

	if info := waitEnd(t, ended); info.HasError() {
		t.Fatalf("unexpected error: %+v", info)
	}
	if info, err := step.Wait(); err != nil || info.HasError() {
		t.Fatalf("Wait = %+v, %v", info, err)
	}
	if got := testutil.ReadFile(t, log); got != "first\nsecond\n" {
		t.Fatalf("log = %q", got)
	}
}

func TestSerialProcesses_StopAtTheFirstFailure(t *testing.T) {
	testutil.RequireShell(t)
	marker := filepath.Join(t.TempDir(), "marker")

	step := SerialProcessesToStep([]*process.Process{
		newProcess(t, "exit 2"),
		newProcess(t, "touch "+marker),
	}, nil)

	info, err := step.Wait()
	if err != nil || info.Code != 2 {
		t.Fatalf("Wait = %+v, %v; want the exit code 2", info, err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("the process after the failure must not run")
	}
}

func TestSerialProcesses_StartError_IsReported(t *testing.T) {
	mux := multiplexer.New(func(_ *multiplexer.Client, b []byte) (int, error) { return len(b), nil })
	missing, err := process.New(&process.Options{ManualStart: true, Multiplexer: mux}, "/nonexistent/command")
	if err != nil {
		t.Fatal(err)
	}

	_, err = SerialProcessesToStep([]*process.Process{missing}, nil).Wait()
	if err == nil {
		t.Fatal("expected the start error")
	}
}

func TestSerialProcesses_Abort_StopsTheRunningAndTheNextProcesses(t *testing.T) {
	testutil.RequireShell(t)
	marker := filepath.Join(t.TempDir(), "marker")
	ready := filepath.Join(t.TempDir(), "ready")

	step := SerialProcessesToStep([]*process.Process{
		newProcess(t, "touch "+ready+"; sleep 30"),
		newProcess(t, "touch "+marker),
	}, nil)

	// Waits until the first process is running (polls the file it creates)
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the first process did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if _, err := step.Abort(true); err != nil {
		t.Fatal(err)
	}
	info, err := step.Wait()
	if err != nil || !info.Stopped {
		t.Fatalf("Wait = %+v, %v; want a stopped process", info, err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("the process after the abort must not run")
	}
}

func TestProcessToStep(t *testing.T) {
	testutil.RequireShell(t)

	info, err := ProcessToStep(newProcess(t, "exit 5")).Wait()
	if err != nil || info.Code != 5 {
		t.Fatalf("Wait = %+v, %v", info, err)
	}

	step := ProcessToStep(newProcess(t, "sleep 30"))
	if info, err := step.Abort(true); err != nil || !info.Stopped {
		t.Fatalf("Abort = %+v, %v", info, err)
	}
}
