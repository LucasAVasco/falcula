---@meta

---@class FalculaJson Module to work with JSON
local M = {}

---Encode a Lua value to a JSON string
---@param value any The value to encode
---@return string
M.encode = function(value) end

---Decode a JSON string to a Lua value
---@param str string The JSON string to decode
---@return any
M.decode = function(str) end

return M
