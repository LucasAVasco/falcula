// Package tui implements the text user interface
package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LucasAVasco/falcula/logfile"
	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/tui/argsview"
	"github.com/LucasAVasco/falcula/tui/help"
	"github.com/LucasAVasco/falcula/tui/mainpage"

	"github.com/rivo/tview"
	lua "github.com/yuin/gopher-lua"
)

// Tui is the text user interface widget. If raw stdout mode is enabled, the user interface is not created and the logs are printed to the
// standard output
type Tui struct {
	logFile         *os.File
	rawMode         bool
	multi           *multiplexer.Multiplexer
	lastCommandArgs []string
	luaState        *lua.LState
	moduleLoader    *modules.Loader
	luaFileMutex    sync.Mutex

	// User interface

	app      *tview.Application
	pages    *tview.Pages
	mainPage *mainpage.MainPage
	argsView *argsview.ArgsView
	help     *help.HelpWidget
}

// New creates a new text user interface. If rawMode is true, the user interface is not created and the logs are printed to the standard
// output
func New(rawMode bool) (*Tui, error) {
	t := Tui{
		rawMode: rawMode,
	}

	// Log file
	logFile, err := logfile.New()
	if err != nil {
		return nil, fmt.Errorf("error creating log file: %w", err)
	}
	t.logFile = logFile

	// Multiplexer
	t.multi = multiplexer.New(func(client *multiplexer.Client, b []byte) (int, error) {
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

			_, err := t.logFile.Write([]byte(logFileLine))
			if err != nil {
				return 0, err
			}

			// Adds the log to the circular buffers
			uiLine := color.Sprint(name+" "+level+": ") + line + "\n"
			t.mainPage.ServiceLogs.Append(uiLine)
		}

		return len(b), nil
	})

	return &t, nil
}

// Close the text user interface. Can be called multiple times.
func (t *Tui) Close() error {
	// Multiplexer
	if t.multi != nil {
		err := t.multi.Close()
		if err != nil {
			return fmt.Errorf("error closing multiplexer: %w", err)
		}
		t.multi = nil
	}

	// Log file
	if t.logFile != nil {
		logFileName := t.logFile.Name()

		// Closing
		err := t.logFile.Close()
		if err != nil {
			return fmt.Errorf("error closing log file: %w", err)
		}

		// Deletes the file
		err = os.Remove(logFileName)
		if err != nil {
			return fmt.Errorf("error removing log file: %w", err)
		}

		t.logFile = nil
	}

	return nil
}

// Open the text user interface
func (t *Tui) Open() error {
	// Creates a new application
	err := t.newApp()
	if err != nil {
		return fmt.Errorf("error creating 'tview' application: %w", err)
	}

	// Runs the falcula file
	if os.Args[1] == "run" || os.Args[1] == "run-raw" {
		args := os.Args[2:]
		t.lastCommandArgs = args

		if t.rawMode {
			err := t.executeLuaFile(args)
			if err != nil {
				return fmt.Errorf("error executing Lua file: %w", err)
			}
		} else {
			t.createLuaRoutine(args)

			// Runs the application
			err = t.app.Run()
			if err != nil {
				return fmt.Errorf("error running 'tview' application: %w", err)
			}
		}
	}

	return nil
}
