package modproject

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/project"

	lua "github.com/yuin/gopher-lua"
)

type taskRun struct {
	task    string
	subTask string
	args    []string
}

func newProjectState(t *testing.T, runs *[]taskRun) (*lua.LState, string) {
	t.Helper()
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), `
minimum_version: v0.0.1
extends: [base.yaml]
projects: { child: sub }
fallback_task: build
tasks:
  build: { shell: [make, all] }
  script: { lua_file: main.lua }
`)
	testutil.WriteFile(t, filepath.Join(root, "base.yaml"), "tasks: { lint: { shell: 'go vet' } }\n")
	testutil.WriteFile(t, filepath.Join(root, "sub", "falcula.yaml"), "projects: { grand: g }\ntasks: { hello: { lua: 'print(1)' } }\n")
	testutil.WriteFile(t, filepath.Join(root, "sub", "g", "falcula.yaml"), "tasks: { deep: { shell_file: deep.sh } }\n")

	proj, err := project.LoadProject(root)
	if err != nil {
		t.Fatal(err)
	}

	r := testruntime.New(t, proj, func(task *project.Task, subTask string, args []string) error {
		*runs = append(*runs, taskRun{task.Name, subTask, args})
		return nil
	})
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"project": testruntime.Configure(r, New())})
	L.SetGlobal("root", lua.LString(root))

	return L, root
}

func TestProjectGetters(t *testing.T) {
	L, _ := newProjectState(t, &[]taskRun{})

	err := L.DoString(`
		local function sorted(list) table.sort(list); return table.concat(list, ",") end

		local p = require("project").get_current_project()
		assert(p:get_file() == root .. "/falcula.yaml")
		assert(p:get_folder() == root and p:get_cwd() == root)
		assert(p:get_minimum_version() == "v0.0.1")
		assert(p:get_extends()[1] == "base.yaml")
		assert(p:get_root() == false)
		assert(p:get_fallback_task() == "build")

		assert(sorted(p:get_tasks_names()) == "build,lint,script", sorted(p:get_tasks_names()))
		assert(sorted(p:get_all_tasks_names()) == "build,child:grand:deep,child:hello,lint,script")
		assert(sorted(p:get_child_projects_names()) == "child")
		assert(sorted(p:get_all_child_projects_names()) == "child,child:grand")

		assert(#p:get_child_projects() == 1 and #p:get_all_child_projects() == 2)
		assert(p:get_child_project_by_name("child:grand"):get_folder() == root .. "/sub/g")
		assert(#p:get_tasks() == 3)
		assert(p:get_all_tasks()["child:grand:deep"] ~= nil)

		local missing, err = p:get_child_project_by_name("missing")
		assert(missing == nil and err ~= nil)
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskAndCommandGetters(t *testing.T) {
	L, _ := newProjectState(t, &[]taskRun{})

	err := L.DoString(`
		local p = require("project").get_current_project()

		local build = p:get_task_by_name("build")
		local list = build:get_shell():get_list()
		assert(list[1] == "make" and list[2] == "all" and build:get_shell():get_string() == "")
		assert(build:get_cwd() == root and build:get_lua() == "" and build:get_shell_file() == "" and build:get_lua_file() == "")
		assert(build:get_project():get_folder() == root)

		assert(p:get_task_by_name("script"):get_lua_file() == root .. "/main.lua")
		assert(p:get_task_by_name("lint"):get_shell():get_string() == "go vet")
		assert(p:get_task_by_name("child:hello"):get_lua() == "print(1)")
		assert(p:get_task_by_name("unknown"):get_shell():get_list()[1] == "make", "uses the fallback task")

		local task, err = p:get_task_by_name("child:unknown")
		assert(task == nil and err ~= nil)
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskRun_CallsTheRuntimeCallback(t *testing.T) {
	var runs []taskRun
	L, _ := newProjectState(t, &runs)

	err := L.DoString(`
		local p = require("project").get_current_project()
		p:get_task_by_name("build"):run()
		p:get_task_by_name("child:hello"):run("sub.task", { "a", 2 })
	`)
	if err != nil {
		t.Fatal(err)
	}

	want := []taskRun{{"build", "", []string{}}, {"hello", "sub.task", []string{"a", "2"}}}
	if !slices.EqualFunc(runs, want, func(a, b taskRun) bool {
		return a.task == b.task && a.subTask == b.subTask && slices.Equal(a.args, b.args)
	}) {
		t.Fatalf("runs = %+v, want %+v", runs, want)
	}
}
