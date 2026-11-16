---@meta

---@class FalculaCmd Module to access the command line arguments.
---@field args string[] The command line arguments provided by the user with `falcula run` or selected in the TUI.
local M = {}

---Get the available command line arguments: the `args` option of the task in the project configuration file (falcula.yaml).
---The TUI offers them to re-run the script with other arguments.
---@return string[][] The available command line arguments.
function M.get_available_args() end

return M
