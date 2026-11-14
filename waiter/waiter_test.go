package waiter

import (
	"errors"
	"testing"
)

func TestWaiter_NoErrors_ReturnsNil(t *testing.T) {
	w := Waiter{}
	w.Go(func() {})
	if err := w.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaiter_ConcurrentErrors_AreAllReturned(t *testing.T) {
	errs := []error{errors.New("a"), errors.New("b"), errors.New("c")}

	w := Waiter{}
	for _, err := range errs {
		w.Go(func() { w.AddError(err) })
	}

	got := w.Wait()
	for _, err := range errs {
		if !errors.Is(got, err) {
			t.Errorf("the returned error %v does not wrap %v", got, err)
		}
	}
}
