package process

import (
	"strings"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/service/iface"
)

func newProvider(prepare, main *Command) *Provider {
	mux := multiplexer.New(func(_ *multiplexer.Client, b []byte) (int, error) { return len(b), nil })
	return New(&ProviderConfig{Multiplexer: mux, Name: "prov"}, prepare, main)
}

func TestService_NilCommand_EndsImmediately(t *testing.T) {
	svc := newProvider(nil, nil).NewService(nil)

	called := false
	step, err := svc.Prepare(func(info *iface.ExitInfo, err error) {
		called = info != nil && !info.HasError() && err == nil
	})
	if err != nil || step != nil || !called {
		t.Fatalf("Prepare = %v, %v; callback called = %v", step, err, called)
	}
}

func TestService_RunsTheCommandAndReportsItsExit(t *testing.T) {
	testutil.RequireShell(t)
	svc := newProvider(nil, &Command{Shell: true, Command: []string{"exit 7"}}).NewService(nil)

	exited := make(chan *iface.ExitInfo, 1)
	step, err := svc.Start(func(info *iface.ExitInfo, err error) { exited <- info })
	if err != nil {
		t.Fatal(err)
	}

	select {
	case info := <-exited:
		if info.Code != 7 {
			t.Fatalf("callback exit information = %+v", info)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the callback was not called")
	}
	if info, _ := step.Wait(); info.Code != 7 {
		t.Fatalf("Wait = %+v", info)
	}
}

func TestService_NamesAndOptions(t *testing.T) {
	disabled := true
	p := newProvider(nil, nil)

	svc := p.NewService(&base.ServiceOpts{StartDisabled: &disabled})
	if svc.GetName() != "prov" || p.GetName() != "prov" || !svc.GetOpts().StartDisabled {
		t.Fatalf("name = %q, opts = %+v", svc.GetName(), svc.GetOpts())
	}
	if named := p.Provider.NewService("api", nil); named.GetName() != "prov.api" {
		t.Fatalf("named service = %q", named.GetName())
	}
}

func TestService_CommandThatCannotStart_ReportsTheErrorWhenWaiting(t *testing.T) {
	svc := newProvider(nil, &Command{Command: []string{"/nonexistent/command"}}).NewService(nil)

	exited := make(chan *iface.ExitInfo, 1)
	step, err := svc.Start(func(info *iface.ExitInfo, _ error) { exited <- info })
	if err != nil {
		t.Fatalf("the process starts inside the step, Start must not fail: %v", err)
	}

	if _, err := step.Wait(); err == nil || !strings.Contains(err.Error(), "error starting process") {
		t.Fatalf("Wait must return the start error, got %v", err)
	}
	select {
	case info := <-exited:
		if info.Error == nil {
			t.Fatal("the callback must receive the start error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the callback was not called")
	}
}
