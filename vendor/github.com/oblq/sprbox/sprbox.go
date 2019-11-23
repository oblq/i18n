// Package sprbox is an agnostic config parser
// (supporting YAML, TOML, JSON and environment vars)
// and a toolbox factory with automatic configuration
// based on your build environment.
package sprbox

import (
	"fmt"
	"path"
	"runtime"
)

// small slant
const banner = `
                __          
  ___ ___  ____/ / ___ __ __
 (_-</ _ \/ __/ _ / _ \\ \ /
/___/ .__/_/ /_.__\___/_\_\  %s
env/_/aware toolbox factory

`

var (
	ColoredLogs = true
	HideBanner  = false

	BuildEnvironment    = defaultBuildEnvironment()
	ConfigParser        = defaultConfigParser()
	NestedConfigsParser = defaultNestedConfigsParser()
)

// Info return some useful info about the sprbox version, the environment and git.
func Info() string {
	env := BuildEnvironment.Current().Info()
	vcs := BuildEnvironment.Git.Info()

	if !HideBanner {
		version := ""
		if _, filename, _, ok := runtime.Caller(0); ok {
			sprboxRepo := newGitRepository(path.Dir(filename))
			if sprboxRepo.Error == nil {
				version = sprboxRepo.tag + "(" + sprboxRepo.build + ")"
			}
		}
		parsedBanner := fmt.Sprintf(darkGrey(banner), version)
		return fmt.Sprintf("\n%s%s\n%s\n", parsedBanner, env, vcs)
	} else {
		return fmt.Sprintf("\n%s\n%s\n", env, vcs)
	}
}

// Helpers -------------------------------------------------------------------------------------------------------------

//func dump(v interface{}) string {
//	// To marshal directly with yaml produce a panic with unexported fields
//
//	jd, err := json.Marshal(v)
//	if err != nil {
//		//fmt.Printf("dump err on %+v: %v\n", v, err)
//		return fmt.Sprintf("%+v", v)
//	}
//
//	// Convert the JSON to an object.
//	var jsonObj interface{}
//	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
//	// Go JSON library doesn't try to pick the right number type (int, float,
//	// etc.) when unmarshalling to interface{}, it just picks float64
//	// universally. go-yaml does go through the effort of picking the right
//	// number type, so we can preserve number type throughout this process.
//	err = yaml.Unmarshal(jd, &jsonObj)
//	if err != nil {
//		//fmt.Printf("dump err on %+v: %v\n", v, err)
//		return fmt.Sprintf("%+v", v)
//	}
//
//	// Marshal this object into YAML.
//	yd, err := yaml.Marshal(jsonObj)
//	if err != nil {
//		//fmt.Printf("dump err on %+v: %v\n", v, err)
//		return fmt.Sprintf("%+v", v)
//	}
//	return string(yd)
//	//b, _ := json.MarshalIndent(v, "", "  ")
//	//return string(b)+"\n"
//}
