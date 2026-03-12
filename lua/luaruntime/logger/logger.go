// Package logger implements a logger for Falcula that supports multiple log levels and saves the logs to a file
package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LucasAVasco/falcula/logfile"
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/service/enhanced"
)

// Logger is a logger that supports multiple log levels and saves the logs to a file
type Logger struct {
	mutex               sync.Mutex
	logFile             *os.File
	servicesMultiplexer *multiplexer.Multiplexer
	onServiceLog        func(b []byte) (int, error)
	onDebugLog          func(b []byte) (int, error)
	onErrorLog          func(error) (int, error)
}

func New() (*Logger, error) {
	l := &Logger{}

	// Log file
	logFile, err := logfile.New()
	if err != nil {
		return nil, fmt.Errorf("error creating log file: %w", err)
	}
	l.logFile = logFile

	// Service multiplexer
	l.ResetOnServiceLog()
	l.servicesMultiplexer = multiplexer.New(func(client *multiplexer.Client, b []byte) (int, error) {
		l.mutex.Lock()
		defer l.mutex.Unlock()

		level := client.GetLevel()
		name := client.GetName()
		color := client.GetColor()
		id := client.GetId()

		str := string(b)

		// Removes the trailing newline
		if str[len(str)-1] == '\n' {
			str = str[:len(str)-1]
		}
		lines := strings.SplitSeq(str, "\n")

		for line := range lines {
			// Uses the 'syslog_log' format recognized by 'Lnav'
			logFileLine := time.Now().Format(time.RFC3339) + " " + name + " " + level + "[" + strconv.Itoa(int(id)) + "]: " + line + "\n"

			_, err := l.logFile.Write([]byte(logFileLine))
			if err != nil {
				return 0, err
			}

			// Adds the log to the multiplexer
			uiLine := color.Sprint(name+" "+level+": ") + line + "\n"
			l.onServiceLog([]byte(uiLine))
		}

		return len(b), nil
	})

	// Debug log
	l.ResetOnDebugLog()

	// Error log
	l.ResetOnErrorLog()

	return l, nil
}

// Close closes the logger. Can be called multiple times
func (l *Logger) Close() error {
	// Multiplexer
	if l.servicesMultiplexer != nil {
		err := l.servicesMultiplexer.Close()
		if err != nil {
			return fmt.Errorf("error closing multiplexer: %w", err)
		}
		l.servicesMultiplexer = nil
	}

	// Log file
	if l.logFile != nil {
		logFileName := l.logFile.Name()

		// Closing
		err := l.logFile.Close()
		if err != nil {
			return fmt.Errorf("error closing log file: %w", err)
		}

		// Deletes the file
		err = os.Remove(logFileName)
		if err != nil {
			return fmt.Errorf("error removing log file: %w", err)
		}

		l.logFile = nil
	}

	return nil
}

// GetLogFilePath returns the path to the log file
func (l *Logger) GetLogFilePath() string {
	return l.logFile.Name()
}

func (l *Logger) GetServicesMultiplexer() *multiplexer.Multiplexer {
	return l.servicesMultiplexer
}

// SetOnServiceLog sets the handler for the service log
func (l *Logger) SetOnServiceLog(handler func(b []byte) (int, error)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onServiceLog = handler
}

// ResetOnServiceLog resets the handler for the service log to the default handler
func (l *Logger) ResetOnServiceLog() {
	l.SetOnServiceLog(func(b []byte) (int, error) {
		return fmt.Print(string(b))
	})
}

// LogService logs a message to the service log
func (l *Logger) LogService(svc enhanced.EnhancedService, message string) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.onServiceLog([]byte(message))
}

// SetOnDebugLog sets the handler for the debug log
func (l *Logger) SetOnDebugLog(handler func(b []byte) (int, error)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onDebugLog = handler
}

// ResetOnDebugLog resets the handler for the debug log to the default handler
func (l *Logger) ResetOnDebugLog() {
	l.SetOnDebugLog(func(b []byte) (int, error) {
		return fmt.Print(string(b))
	})
}

// LogDebug logs a message to the debug log
func (l *Logger) LogDebug(message string) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.onDebugLog([]byte(message))
}

// SetOnErrorLog sets the handler for the error log
func (l *Logger) SetOnErrorLog(handler func(error) (int, error)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onErrorLog = handler
}

// ResetOnErrorLog resets the handler for the error log to the default handler
func (l *Logger) ResetOnErrorLog() {
	l.SetOnErrorLog(func(err error) (int, error) {
		return fmt.Print(err, "\n")
	})
}

// LogError logs an error
func (l *Logger) LogError(err error) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.onErrorLog(err)
}

// GetOnErrorWithoutReturn returns a function that logs an error and discards the return value
func (l *Logger) GetOnErrorWithoutReturn() func(error) {
	return func(err error) {
		l.onErrorLog(err)
	}
}

// Append appends messages to the debug log if they are not errors. If they are errors, they are logged to the error log
func (l *Logger) Append(messages ...any) {
	for _, message := range messages {
		if _, ok := message.(error); ok {
			l.LogError(message.(error))
			continue
		} else {
			l.LogDebug(fmt.Sprintf("%+v", message))
		}
	}
}

// Write writes messages to the debug log
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.LogDebug(string(message))
}

// ResetLoggers resets all loggers
func (l *Logger) ResetLoggers() {
	l.ResetOnServiceLog()
	l.ResetOnDebugLog()
	l.ResetOnErrorLog()
}
