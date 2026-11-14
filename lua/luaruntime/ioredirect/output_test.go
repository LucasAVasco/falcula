package ioredirect

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"

	lua "github.com/yuin/gopher-lua"
)

// fakeWriter records the messages sent to the debug logs. `os.execute` writes to it from the command goroutines
type fakeWriter struct {
	mutex  sync.Mutex
	output strings.Builder
}

func (w *fakeWriter) Append(messages ...any) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, m := range messages {
		if err, ok := m.(error); ok {
			fmt.Fprintf(&w.output, "ERROR(%v)", err)
			continue
		}
		fmt.Fprint(&w.output, m)
	}
}

func (w *fakeWriter) Write(b []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.output.Write(b)
}

func (w *fakeWriter) String() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.output.String()
}

func newRedirectedState(t *testing.T) (*lua.LState, *fakeWriter) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(L.Close)

	out := &fakeWriter{}
	Redirect(L, out)
	return L, out
}

func TestPrintAndIoWrite_AreSentToTheWriter(t *testing.T) {
	L, out := newRedirectedState(t)

	if err := L.DoString(`print("a", 1, true); print("single"); print(); io.write("x", 2)`); err != nil {
		t.Fatal(err)
	}

	if got, want := out.String(), "a\t1\ttrue\nsingle\nx2"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestOsExecute_ReturnsStatusAndCapturesOutput(t *testing.T) {
	testutil.RequireShell(t)
	L, out := newRedirectedState(t)

	err := L.DoString(`
		assert(os.execute() == true, "the shell must be available")

		local ok, kind, code = os.execute("echo out; echo err >&2")
		assert(ok == true and kind == "exit" and code == 0)

		ok, kind, code = os.execute("exit 4")
		assert(ok == false and kind == "exit" and code == 4, tostring(code))
	`)
	if err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"out\n", "err\n", "ERROR(exit status 4)"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}
