---@meta

---@class FalculaManager
local M = {}

---@class FalculaManagerManager Manager for services.
M.ServiceManager = {}

---Create a new service manager.
---@param name string The name of the manager.
---@return FalculaManagerManager
function M.ServiceManager:new(name) end

---Create a new list of service managers.
---@param arg_list table[] List of arguments for each manager.
---@return FalculaManagerManager[]
function M.ServiceManager:new_list(arg_list) end

---Add a service to the manager.
---@param service FalculaServiceService The service to add to the manager.
function M.ServiceManager:add_service(service, callbacks) end

---Add multiple services to the manager
---Equivalent to calling `add_service` multiple times.
---@param services FalculaServiceService[] The services to add to the manager.
function M.ServiceManager:add_services(services) end

---Start the prepare phase of the services in the manager.
function M.ServiceManager:start_prepare() end

---Wait for the prepare phase of the services in the manager to end.
function M.ServiceManager:wait_prepare() end

---Run the prepare phase of the services in the manager.
---Equivalent to calling `start_prepare` and `wait_prepare`.
function M.ServiceManager:prepare() end

---Abort the prepare phase of the services in the manager.
function M.ServiceManager:abort_prepare() end

---Start the services in the manager.
---Before starting the services, this function will prepare the services.
function M.ServiceManager:start() end

---Wait for the services in the manager to end.
function M.ServiceManager:wait() end

---Run the services in the manager
---Equivalent to calling `start` and `wait`.
function M.ServiceManager:run() end

---Run the services in the manager serially (one after the other ends).
function M.ServiceManager:run_serial() end

---Stop the services in the manager.
function M.ServiceManager:stop() end

---Close the manager. You can not use the manager anymore after this function is called.
function M.ServiceManager:close() end

return M
