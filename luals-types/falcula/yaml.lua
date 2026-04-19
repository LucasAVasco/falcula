---@meta

---@class FalculaYaml Module to work with YAML
local M = {}

---Encode a Lua value to a YAML string
---@param value any The value to encode
---@return string
M.encode = function(value) end

---Decode a YAML string to a Lua value
---@param str string The YAML string to decode
---@return any
M.decode = function(str) end

return M
