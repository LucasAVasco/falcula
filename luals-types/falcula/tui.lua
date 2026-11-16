---@meta

---@class FalculaTui Text user interface module
local M = {}

---Show the TUI (creates it if needed). Does nothing in raw mode (`--raw`, CI environments or without a terminal).
function M.show() end

---Show the TUI on the arguments page: the user selects one of the `args` of the task (falcula.yaml) and the script runs again with them.
---Does nothing in raw mode. Only `lua_file` tasks can run again with other arguments.
---
---Example: let the user choose the arguments when the script runs without any.
---```lua
---if #require('falcula.cmd').args == 0 then
---  require('falcula.tui').show_args()
---  return
---end
---```
function M.show_args() end

---Hide the TUI.
function M.hide() end

return M
