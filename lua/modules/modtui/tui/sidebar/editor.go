package sidebar

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
)

// defaultEditor is the editor used when neither `$VISUAL` nor `$EDITOR` is set
const defaultEditor = "vi"

// editorCommand returns the command that opens the files in the user editor (`$VISUAL`, `$EDITOR` or `vi`). The editor runs through the
// shell, so it can have arguments (e.g. "code --wait"). The files are shell arguments: they are never parsed by the shell
func editorCommand(getenv func(string) string, files []string) (name string, args []string) {
	editor := getenv("VISUAL")
	if editor == "" {
		editor = getenv("EDITOR")
	}
	if editor == "" {
		editor = defaultEditor
	}

	return "sh", append([]string{"-c", editor + ` "$@"`, "sh"}, files...)
}

// appendFile appends the file to the list if it is not empty and not in the list yet
func appendFile(files []string, file string) []string {
	if file == "" || slices.Contains(files, file) {
		return files
	}
	return append(files, file)
}

// filesOfCurrentNode returns the files associated with the current node: the file of a service, or the files of the services of a manager
// (all managers if the current node is the root). The list has no duplicates
func (s *Sidebar) filesOfCurrentNode() ([]string, error) {
	var files []string
	err := s.executeFunctionOnCurrentNode(&nodeHandlers{
		OnManager: func(man *manager.Manager) error {
			for _, svc := range man.GetServices() {
				files = appendFile(files, svc.GetFile())
			}
			return nil
		},
		OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
			if svc.GetFile() == "" {
				return fmt.Errorf("service '%s' has no associated file", svc.GetName())
			}
			files = appendFile(files, svc.GetFile())
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("no service has an associated file")
	}

	return files, nil
}

// openCurrentNodeFiles opens the files associated with the current node in the user editor, suspending the text user interface while it
// runs
func (s *Sidebar) openCurrentNodeFiles() {
	files, err := s.filesOfCurrentNode()
	if err != nil {
		s.OnError(fmt.Errorf("error opening the associated file: %w", err))
		return
	}

	name, args := editorCommand(os.Getenv, files)
	s.app.Suspend(func() {
		err = runInteractive(name, args...)
	})

	if err != nil {
		s.OnError(fmt.Errorf("error running the editor: %w", err))
	}
}
