---@meta

---@class FalculaService Service type definitions.
local M = {}

---@class FalculaServiceServiceOpts Service options.
---@field start_disabled? boolean If the service should not be automatically started. The user must enable the service manually.

---@class FalculaServiceService Generic service.

---@class FalculaServiceProviderOpts Provider options.
---@field service_opts? FalculaServiceServiceOpts Default options for the generated services.

---@class FalculaServiceProvider Generic provider.

return M
