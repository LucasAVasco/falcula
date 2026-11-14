//go:build integration

package moddockercompose

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/modules/modmanager"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
)

const composeFile = `
services:
  web:
    image: example/web
    build: { context: . }
  db:
    image: postgres
`

// runComposeScript runs a Lua script next to a compose file, with a fake `docker` executable. Returns the logged docker calls and the
// prefix of the compose commands ("compose --project-directory <dir> -f <file>")
func runComposeScript(t *testing.T, script string) (calls []string, composePrefix string) {
	t.Helper()
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	dockerLog := testutil.FakeDocker(t)
	dir := t.TempDir()
	compose := testutil.WriteFile(t, filepath.Join(dir, "compose.yaml"), composeFile)
	scriptFile := testutil.WriteFile(t, filepath.Join(dir, "main.lua"), `
		local compose = require("falcula.compose")
		local manager = require("falcula.manager")
		local provider = compose.Provider:new("app", "compose.yaml")
		local function run(service)
			local man = manager.ServiceManager:new("m")
			man:add_service(service)
			local err = man:run()
			man:close()
			return err
		end
	`+script)

	r := testruntime.New(t, nil, nil)
	noop := &modmanager.Callbacks{
		OnNewManager:           func(*manager.Manager) {},
		OnDeleteManager:        func(*manager.Manager) {},
		OnAddService:           func(*manager.Manager, *enhanced.EnhancedService) {},
		OnServiceStatusChanged: func(*manager.Manager, *enhanced.EnhancedService) {},
	}
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{
		"falcula.compose": testruntime.Configure(r, New()),
		"falcula.manager": testruntime.Configure(r, modmanager.New(noop)),
	})

	if err := L.DoFile(scriptFile); err != nil {
		t.Fatal(err)
	}

	return testutil.ReadLines(t, dockerLog), "compose --project-directory " + dir + " -f " + compose
}

func assertCalls(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("docker calls:\n%q\nwant:\n%q", got, want)
	}
}

func TestIntegrationUpService_BuildsThenStartsForThePlatform(t *testing.T) {
	calls, prefix := runComposeScript(t, `assert(run(provider:new_up_service("linux/arm64")) == nil)`)

	assertCalls(t, calls, []string{
		"platform=linux/arm64 " + prefix + " up --build --no-start",
		"platform=linux/arm64 " + prefix + " up",
	})
}

func TestIntegrationDownService_RemovesVolumesWhenRequested(t *testing.T) {
	calls, prefix := runComposeScript(t, `
		assert(run(provider:new_down_service()) == nil)
		assert(run(provider:new_down_service({ volumes = true })) == nil)
	`)

	assertCalls(t, calls, []string{
		"platform= " + prefix + " down",
		"platform= " + prefix + " down --volumes",
	})
}

func TestIntegrationBuildService_PullsBuildsAndTagsEachPlatform(t *testing.T) {
	calls, prefix := runComposeScript(t, `
		assert(run(provider:new_build_service({
			builds = { { services = { "web" }, platforms = { "linux/amd64", "linux/arm64" } } },
		})) == nil)
	`)

	assertCalls(t, calls, []string{
		"platform= " + prefix + " pull --ignore-buildable",
		"platform=linux/amd64 " + prefix + " build web",
		"platform= tag example/web example/web:linux-amd64",
		"platform=linux/arm64 " + prefix + " build web",
		"platform= tag example/web example/web:linux-arm64",
	})
}

func TestIntegrationBuildService_DefaultsToAllBuildableServicesWithoutPull(t *testing.T) {
	calls, prefix := runComposeScript(t, `assert(run(provider:new_build_service({ no_pull = true })) == nil)`)

	assertCalls(t, calls, []string{"platform= " + prefix + " build web"})
}

func TestIntegrationPushService_SinglePlatformPushesTheTag(t *testing.T) {
	calls, _ := runComposeScript(t, `
		assert(run(provider:new_push_service({
			pushes = { { images = { "example/web" }, registries = { "reg.io" }, platforms = { "linux/amd64" }, tag = "1.0" } },
		})) == nil)
	`)

	assertCalls(t, calls, []string{
		"platform= tag example/web:linux-amd64 reg.io/example/web:1.0",
		"platform= push reg.io/example/web:1.0",
	})
}

func TestIntegrationPushService_SeveralPlatformsCreateAManifest(t *testing.T) {
	calls, _ := runComposeScript(t, `
		provider:add_push_images({ "example/web" })
		assert(run(provider:new_push_service({
			pushes = { { registries = { "reg.io" }, platforms = { "linux/amd64", "linux/arm64" }, tag = "1.0" } },
		})) == nil)
	`)

	assertCalls(t, calls, []string{
		"platform= tag example/web:linux-amd64 reg.io/example/web:linux-amd64",
		"platform= push reg.io/example/web:linux-amd64",
		"platform= tag example/web:linux-arm64 reg.io/example/web:linux-arm64",
		"platform= push reg.io/example/web:linux-arm64",
		"platform= manifest create reg.io/example/web:1.0 --amend reg.io/example/web:linux-amd64 --amend reg.io/example/web:linux-arm64",
	})
}

func TestIntegrationPushService_ImageWithTag_ReturnsError(t *testing.T) {
	calls, _ := runComposeScript(t, `
		local err = run(provider:new_push_service({ pushes = { { images = { "example/web:1.0" }, registries = { "reg.io" } } } }))
		assert(err ~= nil and err:find("cannot contain ':'"), tostring(err))
	`)

	assertCalls(t, calls, nil)
}

func TestIntegrationFailingDockerCommand_ReturnsError(t *testing.T) {
	t.Setenv("FAKE_DOCKER_EXIT", "1")

	calls, prefix := runComposeScript(t, `
		local err = run(provider:new_down_service())
		assert(err ~= nil and err:find("exit code: 1"), tostring(err))
	`)

	assertCalls(t, calls, []string{"platform= " + prefix + " down"})
}
