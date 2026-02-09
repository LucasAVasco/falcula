---@meta

---@class FalculaCompose Docker/Podman compose module.
local M = {}

---@class (exact) FalculaComposeOpts Docker/Podman compose service.
---@field push_images? string[] List of images that the 'Push' service will push.

---@class FalculaComposeProvider Service provider for Docker/Podman compose service.
M.Provider = {}

---Create a new Docker/Podman compose service provider.
---@param name string Name of the service.
---@param compose_file string Path to compose file.
---@param opts? FalculaComposeOpts Options.
---@return FalculaComposeProvider
function M.Provider:new(name, compose_file, opts) end

---Create a new list of Docker/Podman compose service providers.
---@param arg_list table[] List of arguments for each provider.
---@return FalculaComposeProvider[]
function M.Provider:new_list(arg_list) end

---Get the name of the Docker/Podman compose service provider.
---@return string
function M.Provider:get_name() end

---Add an image to push when using the 'Push' service.
---@param image string Image name.
function M.Provider:add_push_image(image) end

---Add images to push when using the 'Push' service.
---Equivalent to calling `add_push_image` multiple times.
---@param images string[] List of image names.
function M.Provider:add_push_images(images) end

---@class FalculaComposeProvider.BuildInfo Information to build a service.
---@field services? string[] List of services to build. Defaults to all buildable services.
---@field platforms? string[] List of platforms to build. In the format `os/arch`. Example: `linux/amd64`. Defaults to current platform.

---@class FalculaComposeProvider.BuildServiceOpts Options for the 'Build' service.
---@field no_pull? boolean Only build the images defined in the compose file (does not pull not required images).
---@field builds? FalculaComposeProvider.BuildInfo[] List of build information.

---Create a new 'Build' service.
---This service builds the images defined in the compose file.
---@param opts? FalculaComposeProvider.BuildServiceOpts Options for the service.
---@return FalculaServiceService
function M.Provider:new_build_service(opts) end

---Create a new 'Up' service.
---This service builds and starts the containers defined in the compose file.
---@param platform? string Platform to run the containers on.
---@return FalculaServiceService
function M.Provider:new_up_service(platform) end

---Create a new 'Down' service.
---This service stops and removes the containers defined in the compose file.
---@return FalculaServiceService
function M.Provider:new_down_service() end

---@class FalculaComposeProvider.PushInfo Information to push services images to a registry.
---@field services? string[] List of services to push. Defaults to all buildable services.
---@field images? string[] List of images to push. Defaults to all buildable images.
---@field platforms? string[] List of platforms to push. In the format `os/arch`. Example: `linux/amd64`. Defaults to current platform.
---@field registries string[] Registries to push the images to.
---@field tag? string Tag to append to the images in the registry. Example: if tag is 'v1.0.0', the image will be pushed as 'image:v1.0.0'.

---@class FalculaComposeProvider.PushServiceOpts Options for the 'Push' service.
---@field pushes FalculaComposeProvider.PushInfo[] List of push information.

---Create a new 'Push' service.
---This service pushes the images defined in the compose file to a registry.
---You must be logged in to the registry before using this service.
---@param opts FalculaComposeProvider.PushServiceOpts Options for the service.
---@return FalculaServiceService
function M.Provider:new_push_service(opts) end

return M
