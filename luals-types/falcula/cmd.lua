---@meta

---@class FalculaCmd Module to access the command line arguments.
---@field args string[] The command line arguments provided by the user with `falcula run`, `falcula run-raw` or the TUI.
local M = {}

---Set the available command line arguments.
---This function informs the user about the command line arguments that can be user with `falcula run` and `falcula run-raw`. It also sets
---the commands that can the TUI can run.
---@param available_args string[][] The available command line arguments.
function M.set_available_args(available_args) end

---Get the available command line arguments (configured with `set_available_args()`).
---@return string[][] The available command line arguments.
function M._available_args() end

return M
