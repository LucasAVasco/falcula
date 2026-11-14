package manager

import (
	"sync"
	"testing"

	"github.com/LucasAVasco/falcula/service/iface"
)

type fakeService struct{ name string }

func (f *fakeService) GetName() string                                  { return f.name }
func (f *fakeService) GetOpts() *iface.Opts                             { return &iface.Opts{} }
func (f *fakeService) Prepare(iface.OnExitCallback) (iface.Step, error) { return nil, nil }
func (f *fakeService) Start(iface.OnExitCallback) (iface.Step, error)   { return nil, nil }

func TestRemoveServicesRemovesAllWhenPassedGetServices(t *testing.T) {
	m := New("test")
	for _, name := range []string{"a", "b", "c", "d"} {
		m.AddService(&fakeService{name: name}, nil)
	}

	if err := m.RemoveServices(m.GetServices(), false, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(m.GetServices()); got != 0 {
		t.Fatalf("expected no services left, got %d", got)
	}
}

func TestConcurrentServiceListAccess(t *testing.T) {
	m := New("test")

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			svc := m.AddService(&fakeService{name: string(rune('a' + i))}, nil)
			_ = m.Prepare(nil).Wait()
			_ = m.GetServices()
			_ = m.RemoveService(svc, true, nil)
		})
	}
	wg.Wait()

	if got := len(m.GetServices()); got != 0 {
		t.Fatalf("expected no services left, got %d", got)
	}
}
