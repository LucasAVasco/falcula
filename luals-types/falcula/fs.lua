---@meta

---@class FalculaFs Module to work with the file system.
local M = {}

---@enum FalculaFsOperation File system operations.
M.OPERATIONS = {
	CREATE = 1,
	WRITE = 2,
	REMOVE = 4,
	RENAME = 8,
	CHMOD = 16,
}
M.OP = M.OPERATIONS

---@enum FalculaFsOperationCodes File system operations codes.
M.OPERATIONS_CODES = {
	[1] = "CREATE",
	[2] = "WRITE",
	[4] = "REMOVE",
	[8] = "RENAME",
	[16] = "CHMOD",
}
M.OP_CODES = M.OPERATIONS_CODES

---@class FalculaFsWatchOpts Watch options.
---@field wildcard? boolean Expand wildcard patterns (use Golang `filepath.Glob` function).
---@field recursive? boolean Watch directories recursively Instead of watching only the first level.
---@field vcs? boolean Ignore files ignored by version control systems.

---Watch files or directories for changes.
---@param paths string[] Paths to watch (files or directories).
---@param callback fun(path: string, operation: FalculaFsOperation)
---@param opts? FalculaFsWatchOpts Options.
function M.watch(paths, callback, opts) end

---@class FalculaFsWatchParallelEntry Watch options.
---@field files string[] Files to watch.
---@field callback fun(path: string, operation: FalculaFsOperation)
---@field opts? FalculaFsWatchOpts Options.

---Watch files or directories for changes in parallel.
---@param config FalculaFsWatchParallelEntry[]
function M.watch_parallel(config) end

return M
