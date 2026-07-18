---@meta

---@class FalculaProject Module to work with Falcula projects.
local M = {}

---@class FalculaProjectProject Definition of a project.
M.Project = {}

---@class FalculaProjectTask Definition of a task to run.
M.Task = {}

---@class FalculaProjectCommand
---Shell command executed by a task. May be defined by a string interpreted as a shell command, or by a list of arguments to execute.
M.Command = {}

---Get the current project.
---@return FalculaProjectProject
function M.get_current_project() end

---Get the absolute path to the file of the project.
---@return string
function M.Project:get_file() end

---Get the absolute path to the folder of the project.
---@return string
function M.Project:get_folder() end

---Get the minimum Falcula version required by the project.
---@return string
function M.Project:get_minimum_version() end

---Get the working directory for the project.
---@return string
function M.Project:get_cwd() end

---Get the list of projects that this project extends.
---@return string[]
function M.Project:get_extends() end

function M.Project:get_child_projects() end

---Get the list of sub-projects names of the project.
---@return string[]
function M.Project:get_child_projects_names() end

---Get the list of sub-projects names of the project and its sub-projects. Similar to the `get_child_projects_names` method, but recursive.
---@return string[]
function M.Project:get_all_child_projects_names() end

---Get a sub-project by its name.
---@param name string Full name of the project. One of the values returned by the `get_all_child_projects_names` method.
---@return FalculaProjectProject
---@see M.Project.get_all_child_projects_names
function M.Project:get_child_project_by_name(name) end

---Get the list of sub-projects of the project.
---@return table<string, FalculaProjectProject> map_of_name_to_project
function M.Project:get_child_projects() end

---Get the list of sub-projects of the project and its sub-projects. Similar to the `get_child_projects` method, but recursive.
---@return table<string, FalculaProjectProject> map_of_name_to_project
function M.Project:get_all_child_projects() end

---Check if the project is the root project.
---@return boolean
function M.Project:get_root() end

---Get the names of the tasks defined in the project.
---@return string[]
function M.Project:get_tasks_names() end

---Get the names of the tasks defined in the project and its sub-projects. Similar to the `get_tasks_names` method, but recursive.
---@return string[]
function M.Project:get_all_tasks_names() end

---Get a task by its name.
---@param name string Name of the task. One of the values returned by the `get_all_tasks_names` method.
---@return FalculaProjectTask
---@see M.Project.get_all_tasks_names
function M.Project:get_task_by_name(name) end

---Get the list of tasks defined in the project.
---@return table<string, FalculaProjectTask> map_of_name_to_task
function M.Project:get_tasks() end

---Get the list of all tasks defined in the project and its sub-projects.
---@return table<string, FalculaProjectTask> map_of_name_to_task
function M.Project:get_all_tasks() end

---Get the name of the fallback task defined in the project.
---@return string
function M.Project:get_fallback_task() end

---Get the project that this task belongs to.
---@return FalculaProjectProject
function M.Task:get_project() end

---Get the working directory for this task.
---@return string
function M.Task:get_cwd() end

---Get the shell command for this task. Return an empty command if no shell command is defined.
---@return FalculaProjectCommand
function M.Task:get_shell() end

---Get the Lua code for this task. Return an empty string if no Lua code is defined.
---@return string
function M.Task:get_lua() end

---Get the shell file for this task. Return an empty string if no shell file is defined.
---@return string
function M.Task:get_shell_file() end

---Get the Lua file for this task. Return an empty string if no Lua file is defined.
---@return string
function M.Task:get_lua_file() end

---Run the task.
---@param subTask? string Name of the sub-task to run. If not specified, the task is run.
---@param args? string[] Arguments to pass to the task.
function M.Task:run(subTask, args) end

---Get the string (shell command). If the command is not defined by a string, "" is returned.
---@return string
function M.Command:get_string() end

---Get the command and its arguments as a list. If the command is not defined by a list, {} is returned.
---@return string[]
function M.Command:get_list() end

return M
