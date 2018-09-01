package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

const (
	// files type regexp
	regexExt = `(?i)(.y(|a)ml|.toml|.json)` // `(?i)(\..{3,4})` //

	// files type regexp
	regexYAML = `(?i)(.y(|a)ml)`
	regexTOML = `(?i)(.toml)`
	regexJSON = `(?i)(.json)`

	fileSearchCaseSensitive = false
)

// walkConfigPath look for a file matching the passed regex skipping sub-directories.
func walkConfigPath(configPath string, regex *regexp.Regexp) (matchedFile string) {
	err := filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		// nil if the path does not exist
		if info == nil {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() != filepath.Base(configPath) {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if regex.MatchString(info.Name()) {
			matchedFile = path
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	return
}

// getLocalizationFiles will search for the given file in the given path
// returning all the eligible files (eg.: <path>/en.yaml or <path>/en.json)
//
// 'files' can also be passed without file extension,
// configFiles is agnostic and will match any
// supported extension in that case.
func getLocalizationFiles(files ...string) (foundFiles []string) {
	for _, file := range files {
		configPath, fileName := filepath.Split(file)
		if len(configPath) == 0 {
			configPath = "./"
		}

		var regex *regexp.Regexp

		ext := filepath.Ext(fileName)
		extTrimmed := strings.TrimSuffix(fileName, ext)
		if len(ext) == 0 {
			ext = regexExt
		}

		format := "^%s%s$"
		if !fileSearchCaseSensitive {
			format = "(?i)(^%s)%s$"
		}
		regex = regexp.MustCompile(fmt.Sprintf(format, extTrimmed, ext))

		// look for the config file in the config path (eg.: tool.yml)
		if matchedFiles := walkConfigPath(configPath, regex); len(matchedFiles) > 0 {
			foundFiles = append(foundFiles, matchedFiles)
		}
	}

	return
}

// LoadLocalization will unmarshal all the matched
// localization files to the passed interface.
func LoadLocalization(config interface{}, files ...string) (err error) {
	foundFiles := getLocalizationFiles(files...)
	if len(foundFiles) == 0 {
		return fmt.Errorf("[i18n] no localization file found for '%s'", strings.Join(files, " | "))
	}

	for _, file := range foundFiles {
		var in []byte
		if in, err = ioutil.ReadFile(file); err != nil {
			return err
		}

		ext := filepath.Ext(file)

		switch {
		case regexp.MustCompile(regexYAML).MatchString(ext):
			err = yaml.Unmarshal(in, config)
		case regexp.MustCompile(regexTOML).MatchString(ext):
			_, err = toml.Decode(string(in), config)
		case regexp.MustCompile(regexJSON).MatchString(ext):
			err = json.Unmarshal(in, config)
		default:
			err = fmt.Errorf("[i18n] unknown data format, can't unmarshal file: '%s'", file)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
