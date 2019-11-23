package sprbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// environment struct.
type environment struct {
	id         string
	Regexp     *regexp.Regexp
	inferredBy string

	// RunCompiled true means that the program run from
	// a precompiled binary for that environment.
	// CompiledPath() returns the path base if RunCompiled == true
	// so that static files can stay side by side with the executable
	// while it is possible to have a different location when the
	// program is launched with `go run`.
	//
	// By default only Production and Staging environments have RunCompiled = true.
	RunCompiled bool
}

// ID returns the environment id,
// which are also a valid tag for the current environment.
func (e *environment) ID() string {
	return e.id
}

// Info return some environment info.
func (e *environment) Info() string {
	return fmt.Sprintf("%s - tag: %s\n", green(strings.ToUpper(e.ID())), e.inferredBy)
}

//---------------------------------------------------------------------------------------------------------

type buildEnvironment struct {
	// HardcodedTag can be used to directly define the current environment.
	// Its value can interpolated with -ldflags at build/run time using the `InterpolableTag` value:
	// 	go build -ldflags "-X github.com/oblq/sprbox/env.InterpolableTag=develop" -v -o ./api_bin ./api
	//
	// If HardcodedTag is empty then the environment variable BuildEnvironment.EnvironmentKeyForTag
	// (which is 'BUILD_ENV' by default) will be checked.
	//
	// If also the environment variable is empty the Git.BranchName will be checked.
	//
	//
	// The default (customizable) regular expressions to match the current environment are:
	//  - Production: 	(production)|(master)
	//	- Staging:	    (staging)|(release/*)|(hotfix/*)
	//	- Testing: 	    (testing)|(test)
	//	- Development: 	(development)|(develop)|(dev)|(feature/*)
	//	- Local: 	    local
	HardcodedTag string

	// EnvironmentKeyForTag is the environment variable key
	// for the build environment tag.
	EnvironmentKeyForTag string

	// Git is the project version control system.
	// By default it uses the working directory.
	Git *gitRepository

	// Default environment's configuration
	Production  *environment
	Staging     *environment
	Testing     *environment
	Development *environment
	Local       *environment

	// currentTAG is the tag from which buildEnvironment
	// has determined the current environment
	currentTAG string

	mutex sync.Mutex
}

// InterpolableTag define the current environment.
// Can be defined by code or, since it is an exported string,
// can be interpolated with -ldflags at build/run time:
// 	go build -ldflags "-X github.com/oblq/sprbox/env.BuildEnvTag=develop" -v -o ./api_bin ./api
var InterpolableTag string

func defaultBuildEnvironment() *buildEnvironment {
	return &buildEnvironment{
		HardcodedTag:         InterpolableTag,
		EnvironmentKeyForTag: "BUILD_ENV",
		Git:                  newGitRepository("./"),

		Production: &environment{
			id:          "production",
			Regexp:      regexp.MustCompile("(production)|(master)"),
			RunCompiled: true,
		},
		Staging: &environment{
			id:          "staging",
			Regexp:      regexp.MustCompile("(staging)|(release/*)|(hotfix/*)"),
			RunCompiled: true,
		},
		Testing: &environment{
			id:          "testing",
			Regexp:      regexp.MustCompile("(testing)|(test)"),
			RunCompiled: true,
		},
		Development: &environment{
			id:          "development",
			Regexp:      regexp.MustCompile("(development)|(develop)|(dev)|(feature/*)"),
			RunCompiled: true,
		},
		Local: &environment{
			id:          "local",
			Regexp:      regexp.MustCompile("local"),
			RunCompiled: false,
		},
	}
}

func (e *buildEnvironment) SetGitPath(path string) {
	e.Git = newGitRepository(path)
}

var testingRegexp = regexp.MustCompile(`_test|(\.test$)|_Test`)

// Env returns the current selected environment by
// matching the privateTAG variable against the environments RegEx.
func (e *buildEnvironment) Current() *environment {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	inferredBy := ""

	// loading tag
	if len(e.HardcodedTag) > 0 {
		e.currentTAG = e.HardcodedTag
		inferredBy = fmt.Sprintf("'%s', inferred from 'HardcodedTag' var, set manually.", e.currentTAG)
		//return
	} else if e.currentTAG = os.Getenv(e.EnvironmentKeyForTag); len(e.currentTAG) > 0 {
		inferredBy = fmt.Sprintf("'%s', inferred from '%s' environment variable.", e.currentTAG, e.EnvironmentKeyForTag)
		//return
	} else if e.Git != nil {
		if e.Git.Error == nil {
			e.currentTAG = e.Git.branchName
			inferredBy = fmt.Sprintf("<empty>, inferred from git.BranchName (%s).", e.Git.branchName)
			//return
		}
	} else if testingRegexp.MatchString(os.Args[0]) {
		e.currentTAG = e.Testing.ID()
		inferredBy = fmt.Sprintf("'%s', inferred from the running file name (%s).", e.currentTAG, os.Args[0])
		//return
	} else {
		inferredBy = "<empty>, default environment is 'local'."
	}

	switch {
	case e.Production.Regexp.MatchString(e.currentTAG):
		e.Production.inferredBy = inferredBy
		return e.Production
	case e.Staging.Regexp.MatchString(e.currentTAG):
		e.Staging.inferredBy = inferredBy
		return e.Staging
	case e.Testing.Regexp.MatchString(e.currentTAG):
		e.Testing.inferredBy = inferredBy
		return e.Testing
	case e.Development.Regexp.MatchString(e.currentTAG):
		e.Development.inferredBy = inferredBy
		return e.Development
	case e.Local.Regexp.MatchString(e.currentTAG):
		e.Local.inferredBy = inferredBy
		return e.Local
	default:
		e.Local.inferredBy = inferredBy
		return e.Local
	}
}

// EnvSubDir returns <path>/<environment>
func (e *buildEnvironment) EnvSubDir(path string) string {
	return filepath.Join(path, e.Current().ID())
}

// CompiledPath returns the path base if RunCompiled == true
// for the environment in use so that static files can
// stay side by side with the executable
// while it is possible to have a different location when the
// program is launched with `go run`.
// This allow to manage multiple packages in one project during development,
// for instance using a config path in the parent dir, side by side with
// the packages, while having the same config folder side by side with
// the executable where needed.
//
// Can be used in:
//  sprbox.LoadToolBox(&myToolBox, sprbox.CompiledPath("../config"))
//
// Example:
//  sprbox.Development.RunCompiled = false
//  sprbox.Tag = sprbox.Development.ID()
//  sprbox.CompiledPath("../static_files/config") // -> "../static_files/config"
//
//  sprbox.Development.RunCompiled = true
//  sprbox.Tag = sprbox.Development.ID()
//  sprbox.CompiledPath("../static_files/config") // -> "config"
//
// By default only Production and Staging environments have RunCompiled = true.
func (e *buildEnvironment) CompiledPath(path string) string {
	if e.Current().RunCompiled {
		return filepath.Base(path)
	}
	return path
}
