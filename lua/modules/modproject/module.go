// Package modproject is a module that provides functions and classes for working with the current project.
package modproject

import (
	"fmt"
	"maps"
	"slices"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/project"

	lua "github.com/yuin/gopher-lua"
)

type Loader struct {
	base.BaseModule
	project *project.Project
}

func New() *Loader {
	return &Loader{}
}

func (l *Loader) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	l.project = l.Config.Runtime.GetProject()

	projectClass, err := luaclass.Register(L, mod, &luaclass.Info{Name: "Project"})
	if err != nil {
		return fmt.Errorf("error registering class of '%s' module: %w", name, err)
	}

	taskClass, err := luaclass.Register(L, mod, &luaclass.Info{Name: "Task"})
	if err != nil {
		return fmt.Errorf("error registering class of '%s' module: %w", name, err)
	}

	commandClass, err := luaclass.Register(L, mod, &luaclass.Info{Name: "Command"})
	if err != nil {
		return fmt.Errorf("error registering class of '%s' module: %w", name, err)
	}

	// Functions to wrap Go values in instances of the classes
	newProject := func(L *lua.LState, proj *project.Project) *lua.LTable {
		return newInstance(L, projectClass, proj)
	}

	newProjectList := func(L *lua.LState, projects map[string]*project.Project) *lua.LTable {
		list := L.NewTable()
		for _, proj := range projects {
			list.Append(newProject(L, proj))
		}
		return list
	}

	newTask := func(L *lua.LState, task *project.Task) *lua.LTable {
		return newInstance(L, taskClass, task)
	}

	newCommand := func(L *lua.LState, command *project.Command) *lua.LTable {
		return newInstance(L, commandClass, command)
	}

	// Global functions
	L.SetField(mod, "get_current_project", L.NewFunction(func(L *lua.LState) int {
		L.Push(newProject(L, l.project))
		return 1
	}))

	// Project methods
	methods := map[string]lua.LGFunction{
		"get_file": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(lua.LString(proj.File))
			return 1
		},

		"get_folder": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(lua.LString(proj.Folder))
			return 1
		},

		"get_minimum_version": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(lua.LString(proj.MinimumVersion))
			return 1
		},

		"get_cwd": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(lua.LString(proj.Cwd))
			return 1
		},

		"get_extends": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)
			extends := luatable.GetLuaTableFromStrings(L, proj.Extends)
			L.Push(extends)
			return 1
		},

		"get_child_projects_names": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)
			names := proj.GetChildProjectsNames()
			L.Push(luatable.GetLuaTableFromStrings(L, names))
			return 1
		},

		"get_all_child_projects_names": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)
			names, err := proj.GetAllChildProjectsNames()
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting all child projects names: %w", err))
			}
			L.Push(luatable.GetLuaTableFromStrings(L, names))
			return 1
		},

		"get_child_project_by_name": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)
			child, err := proj.GetChildProjectByName(L.CheckString(2))
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting child project by name: %w", err))
			}
			L.Push(newProject(L, child))
			return 1
		},

		"get_child_projects": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			childProjects, err := proj.GetChildProjects()
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting child projects: %w", err))
			}

			L.Push(newProjectList(L, childProjects))
			return 1
		},

		"get_all_child_projects": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			childProjects, err := proj.GetAllChildProjects()
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting all child projects: %w", err))
			}

			L.Push(newProjectList(L, childProjects))
			return 1
		},

		"get_root": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(lua.LBool(proj.Root))
			return 1
		},

		"get_tasks_names": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(luatable.GetLuaTableFromStrings(L, slices.Collect(maps.Keys(proj.Tasks))))
			return 1
		},

		"get_all_tasks_names": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			tasks, err := proj.GetAllTasksNames()
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting all tasks names: %w", err))
			}

			L.Push(luatable.GetLuaTableFromStrings(L, tasks))
			return 1
		},

		"get_task_by_name": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)
			task, err := proj.GetTaskByName(L.CheckString(2))
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting task by name: %w", err))
			}
			L.Push(newTask(L, task))
			return 1
		},

		"get_tasks": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			list := L.NewTable()
			for _, task := range proj.Tasks {
				list.Append(newTask(L, task))
			}

			L.Push(list)
			return 1
		},

		"get_all_tasks": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			tasks, err := proj.GetAllTasks()
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error getting all tasks: %w", err))
			}

			tasksMap := L.NewTable()
			for name, task := range tasks {
				tasksMap.RawSetString(name, newTask(L, task))
			}

			L.Push(tasksMap)
			return 1
		},

		"get_fallback_task": func(L *lua.LState) int {
			proj := luaclass.GetInstance[*project.Project](L)

			L.Push(lua.LString(proj.FallbackTask))
			return 1
		},
	}

	luaclass.SetMethods(L, projectClass, methods)

	// Task methods
	methods = map[string]lua.LGFunction{
		"get_project": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			L.Push(newProject(L, task.Project))
			return 1
		},

		"get_cwd": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			L.Push(lua.LString(task.Cwd))
			return 1
		},

		"get_shell": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			L.Push(newCommand(L, &task.Shell))
			return 1
		},

		"get_lua": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			L.Push(lua.LString(task.Lua))
			return 1
		},

		"get_shell_file": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			L.Push(lua.LString(task.ShellFile))
			return 1
		},

		"get_lua_file": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			L.Push(lua.LString(task.LuaFile))
			return 1
		},

		"run": func(L *lua.LState) int {
			task := luaclass.GetInstance[*project.Task](L)
			subTask := L.OptString(2, "")
			argsTable := L.OptTable(3, L.NewTable())
			args := luatable.GetStringsFromLuaTable(argsTable)

			err := l.Config.Runtime.GetOnRunTask()(task, subTask, args)
			if err != nil {
				return luaerror.Push(L, 0, fmt.Errorf("error running task: %w", err))
			}
			return 0
		},
	}

	luaclass.SetMethods(L, taskClass, methods)

	// Command methods
	methods = map[string]lua.LGFunction{
		"get_string": func(L *lua.LState) int {
			command := luaclass.GetInstance[*project.Command](L)
			L.Push(lua.LString(command.String))
			return 1
		},

		"get_list": func(L *lua.LState) int {
			command := luaclass.GetInstance[*project.Command](L)
			list := luatable.GetLuaTableFromStrings(L, command.List)
			L.Push(list)
			return 1
		},
	}

	luaclass.SetMethods(L, commandClass, methods)

	return nil
}

// newInstance creates an instance of a class that wraps a Go value
func newInstance(L *lua.LState, class *lua.LTable, goInstance any) *lua.LTable {
	instance := luaclass.CreateInstance(L, class)
	luaclass.SetInstance(L, instance, goInstance)
	return instance
}
