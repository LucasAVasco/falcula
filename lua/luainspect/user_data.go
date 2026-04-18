package luainspect

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// inspectUserData inspects a user data and writes it to the writer. The first line is not indented. The caller must manually indent it if
// needed
func inspectUserData(writer *enhancedWriter, indentLevel uint, value *lua.LUserData) error {
	err := writer.WriteString("UserData: {\n")
	if err != nil {
		return fmt.Errorf("error writing first line: %w", err)
	}

	err = writer.WriteIndent(indentLevel + 1)
	if err != nil {
		return fmt.Errorf("error writing indent before value: %w", err)
	}

	err = writer.WriteString("Value: ")
	if err != nil {
		return fmt.Errorf("error writing value prefix: %w", err)
	}

	if luaValue, ok := value.Value.(lua.LValue); ok {
		err = inspect(writer, 0, indentLevel+1, luaValue)
		if err != nil {
			return fmt.Errorf("error inspecting value: %w", err)
		}
	} else {
		err = writer.WriteString(fmt.Sprintf("%#v\n", value.Value))
		if err != nil {
			return fmt.Errorf("error writing value: %w", err)
		}
	}

	err = writer.WriteString("\n")
	if err != nil {
		return fmt.Errorf("error writing new line after value: %w", err)
	}

	err = writer.WriteIndent(indentLevel)
	if err != nil {
		return fmt.Errorf("error writing indent before close brace: %w", err)
	}

	err = writer.WriteString("}")
	if err != nil {
		return fmt.Errorf("error writing last line of user data (close brace): %w", err)
	}

	return nil
}
