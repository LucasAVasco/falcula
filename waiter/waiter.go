// Package waiter implements a WaitGroup with multiple errors handling
package waiter

import (
	"errors"
	"fmt"
	"sync"
)

// Waiter is a WaitGroup with multiple errors handling
type Waiter struct {
	sync.WaitGroup
	errs []error
}

func NewWaiter() *Waiter {
	return &Waiter{}
}

// AddError adds an error to the Waiter errors list
func (w *Waiter) AddError(err error) {
	w.errs = append(w.errs, err)
}

// Wait waits for the Waiter to end and returns all errors in a single error interface
func (w *Waiter) Wait() error {
	w.WaitGroup.Wait()

	if len(w.errs) > 0 {
		return fmt.Errorf("error waiting for services: [\n%w\n]", errors.Join(w.errs...))
	}

	return nil
}
