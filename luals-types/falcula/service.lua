---@meta

---@class FalculaService Service type definitions.
local M = {}

---@class FalculaServiceServiceOpts Service options.
---@field start_disabled? boolean If the service should not be automatically started. The user must enable the service manually.
---@field file? string File associated with the service, opened with the `o` key in the TUI. Relative to the current Lua file. Docker compose services default to their compose file. Not supported in the provider `service_opts`.

---@class FalculaServiceService Generic service.

---@class FalculaServiceProviderOpts Provider options.
---@field service_opts? FalculaServiceServiceOpts Default options for the generated services.

---@class FalculaServiceProvider Generic provider.

return M
