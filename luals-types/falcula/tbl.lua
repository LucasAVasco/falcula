---@meta

---@class FalculaTbl Module that provides table related functions
local M = {}

---@alias FalculaTblBehavior
---| "keep"
---| "force"
---| "error"

---Extends a table with another tables.
---@param behavior FalculaTblBehavior The behavior to use when extending the table.
---@param dest_table table The destination table. Only this table is modified.
---@param ... table Extend the destination table with these tables
function M.extend(behavior, dest_table, ...) end

---Deeply extends (recursively) a table with another tables.
---@param behavior FalculaTblBehavior The behavior to use when extending the table.
---@param dest_table table The destination table. Only this table is modified.
---@param ... table Extend the destination table with these tables
function M.deep_extend(behavior, dest_table, ...) end

return M
