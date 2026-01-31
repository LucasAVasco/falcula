---@meta

---@class FalculaProcess Operational system processes module.
local M = {}

---@class FalculaProcessProvider Service provider for operational system processes.
M.Provider = {}

---Create a new process service provider.
---@param name string Name of the service.
---@param prepare_cmd? string|string[] Command to run before the main command. Ignored if `nil`.
---@param main_cmd? string|string[] Command to run. Ignored if `nil`.
---@return FalculaProcessProvider
function M.Provider:new(name, prepare_cmd, main_cmd) end

---Create a new list of process service providers.
---@param arg_list table[] List of arguments for each provider.
---@return FalculaProcessProvider[]
function M.Provider:new_list(arg_list) end

---Get the name of the service provider.
---@return string
function M.Provider:get_name() end

---Create a new service.
---This service runs
---@return FalculaServiceService
function M.Provider:new_service() end

return M
