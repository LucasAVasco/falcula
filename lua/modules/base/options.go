package base

import (
	"errors"

	"github.com/LucasAVasco/falcula/multiplexer"
)

// Options is the options that a module requires to work
type Options struct {
	Multiplexer *multiplexer.Multiplexer
	OnError     func(err error)
	OnDebug     func(msg string)
}

var ErrNoMultiplexer = errors.New("You must provide a multiplexer in the module options!")

func FillOptionsWithDefaults(opts *Options) (*Options, error) {
	if opts == nil {
		opts = &Options{}
	}

	if opts.Multiplexer == nil {
		return nil, ErrNoMultiplexer
	}

	if opts.OnError == nil {
		opts.OnError = func(err error) {}
	}

	if opts.OnDebug == nil {
		opts.OnDebug = func(msg string) {}
	}

	return opts, nil
}
