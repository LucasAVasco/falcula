package multiplexer

import (
	"sync"
	"testing"

	"github.com/fatih/color"
)

func TestNewClient_ConcurrentClients_HaveUniqueIds(t *testing.T) {
	m := New(func(*Client, []byte) (int, error) { return 0, nil })

	var mutex sync.Mutex
	ids := map[uint]bool{}
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			id := m.NewClient("name", "stdout", nil).GetId()
			mutex.Lock()
			ids[id] = true
			mutex.Unlock()
		})
	}
	wg.Wait()

	if len(ids) != 50 {
		t.Fatalf("expected 50 unique ids, got %d", len(ids))
	}
}

func TestClientWrite_CallsCallbackWithClientMetadata(t *testing.T) {
	type write struct {
		name, level, data string
		clr               *color.Color
	}
	var writes []write
	m := New(func(c *Client, b []byte) (int, error) {
		writes = append(writes, write{c.GetName(), c.GetLevel(), string(b), c.GetColor()})
		return len(b), nil
	})

	clr := color.New(color.FgRed)
	client := m.NewClient("svc", "stderr", clr)
	n, err := client.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write = %d, %v", n, err)
	}

	want := write{"svc", "stderr", "hello", clr}
	if len(writes) != 1 || writes[0] != want {
		t.Fatalf("callback received %+v, want %+v", writes, want)
	}
}
