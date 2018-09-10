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

var (
	// files type regexp
	regexpValidExt = regexp.MustCompile(`(?i)(.y(|a)ml|.toml|.json)`) // `(?i)(\..{3,4})`
	regexpYAML     = regexp.MustCompile(`(?i)(.y(|a)ml)`)
	regexpTOML     = regexp.MustCompile(`(?i)(.toml)`)
	regexpJSON     = regexp.MustCompile(`(?i)(.json)`)

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

// findLocalizationFiles will search for the given file in the given path
// returning all the eligible files (eg.: <path>/en.yaml or <path>/en.json)
//
// 'files' can also be passed without file extension,
// configFiles is agnostic and will match any
// supported extension in that case.
func findLocalizationFiles(files ...string) (foundFiles []string, err error) {
	for _, file := range files {
		configPath, fileName := filepath.Split(file)
		if len(configPath) == 0 {
			configPath = "./"
		}

		var regex *regexp.Regexp

		ext := filepath.Ext(fileName)
		extTrimmed := strings.TrimSuffix(fileName, ext)
		if len(ext) == 0 {
			ext = regexpValidExt.String()
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

	if len(foundFiles) == 0 {
		return foundFiles, fmt.Errorf("no localization file found for '%s'", strings.Join(files, " | "))
	}
	return
}

func unmarshalJSON(data []byte, loc interface{}) (err error) {
	return json.Unmarshal(data, loc)
}

func unmarshalTOML(data []byte, loc interface{}) (err error) {
	_, err = toml.Decode(string(data), loc)
	return err
}

func unmarshalYAML(data []byte, loc interface{}) (err error) {
	return yaml.Unmarshal(data, loc)
}

// LoadLocalizationFiles will unmarshal all the matched
// localization files in i18n.Config.LocalizationsPath for the given i18n.Tags,
// localization files must be named as the <language.Tag>.String()
// (locale, eg.: `en.yml` for `language.English`).
func (i18n *I18n) LoadLocalizationFiles(localizationsPath string) (err error) {
	i18n.localizations = make(map[string]map[string]*Localization)

	for _, lang := range i18n.Tags {

		var langLocalizations map[string]*Localization

		locFileName := filepath.Join(localizationsPath, lang.String())
		foundFiles, err := findLocalizationFiles(locFileName)
		if err != nil {
			return err
		}

		for _, file := range foundFiles {
			var data []byte
			if data, err = ioutil.ReadFile(file); err != nil {
				return err
			}

			ext := filepath.Ext(file)

			switch {
			case regexpYAML.MatchString(ext):
				err = unmarshalYAML(data, &langLocalizations)
			case regexpTOML.MatchString(ext):
				err = unmarshalTOML(data, &langLocalizations)
			case regexpJSON.MatchString(ext):
				err = unmarshalJSON(data, &langLocalizations)
			default:
				err = fmt.Errorf("unknown data format, can't unmarshal file: '%s'", file)
			}

			if err != nil {
				return err
			}
		}

		i18n.localizations[lang.String()] = langLocalizations
	}

	return
}
