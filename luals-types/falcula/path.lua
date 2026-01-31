---@meta

---@class FalculaPaths Utilities functions to manage paths.
local M = {}

---Get the current file absolute path.
---@return string
function M.get_current_file() end

---Get the current directory absolute path.
---@return string
function M.get_current_dir() end

---Returns the absolute path of the given path.
---@param path string The path to get the absolute path of. Relative paths are relative to the current Lua script.
---@return string
function M.abs(path) end

---Returns the absolute paths of the given paths.
---@param path string[] The paths to get the absolute paths of. Relative paths are relative to the current Lua script.
---@return string[]
function M.abs_list(path) end

---Returns the relative path of the given path from the given base path.
---@param path string The path to get the relative path of. May be relative to the current Lua script.
---@param base? string The base path to get the relative path from. May be relative to the current Lua script.
---@return string
function M.rel(path, base) end

---Returns the relative paths of the given paths from the given base path.
---@param path string[] The paths to get the relative paths of. May be relative to the current Lua script.
---@param base? string The base path to get the relative paths from. May be relative to the current Lua script.
---@return string[]
function M.rel_list(path, base) end

return M
