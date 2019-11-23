package sprbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type configParser struct {
	// struct field tag key
	sftKey string

	// set the default value
	sffDefault string

	// return error if missing value
	sffRequired string

	// sffEnv environment var value can be in json format, it will override also the default value
	sffEnv string

	// Files search ----------------------------------------------------------------------------------------------------

	// FileSearchCaseSensitive determine config files search mode, false by default.
	FileSearchCaseSensitive bool

	// files type regexp
	regexpValidExt *regexp.Regexp
	regexpYAML     *regexp.Regexp
	regexpTOML     *regexp.Regexp
	regexpJSON     *regexp.Regexp
}

func defaultConfigParser() configParser {
	return configParser{
		sftKey:      "sprc",
		sffDefault:  "default",
		sffRequired: "required",
		sffEnv:      "env",

		FileSearchCaseSensitive: false,
		regexpValidExt:          regexp.MustCompile(`(?i)(.y(|a)ml|.toml|.json)`), // `(?i)(\..{3,4})`
		regexpYAML:              regexp.MustCompile(`(?i)(.y(|a)ml)`),
		regexpTOML:              regexp.MustCompile(`(?i)(.toml)`),
		regexpJSON:              regexp.MustCompile(`(?i)(.json)`),
	}
}

// Unmarshal will unmarshal []byte to interface
// for yaml, toml and json data formats.
// Will also parse fmt template keys and struct flags.
func (c configParser) Unmarshal(data []byte, config interface{}) (err error) {
	switch {
	case c.unmarshalJSON(data, config) == nil:
		break
	case c.unmarshalYAML(data, config) == nil:
		break
	case c.unmarshalTOML(data, config) == nil:
		break
	default:
		return fmt.Errorf("the provided data is incompatible with an interface of type %T:\n%s",
			config, strings.TrimSuffix(string(data), "\n"))
	}

	if err = c.parseTemplateBytes(data, config); err != nil {
		return err
	}
	return c.parseConfigTags(config)
}

// LoadConfig will unmarshal all the matched
// config files to the config interface.
// The latest files will override the earliest.
// Will also parse fmt template keys and struct flags.
func (c configParser) Load(config interface{}, files ...string) (err error) {
	return c.LoadWithEnv(config, "", files...)
}

// LoadEnvConfig will unmarshal all the matched
// config files for the given environment-id to the config interface.
// environment specific files will override generic files.
// The latest files passed will override the former.
// Will also parse fmt template keys and struct flags.
func (c configParser) LoadWithEnv(config interface{}, envId string, files ...string) (err error) {
	foundFiles, err := c.configFilesByEnv(envId, files...)
	if err != nil {
		return fmt.Errorf("no config file found for '%s': %s", strings.Join(files, " | "), err.Error())
	}
	if len(foundFiles) == 0 {
		return fmt.Errorf("no config file found for '%s'", strings.Join(files, " | "))
	}

	for _, file := range foundFiles {
		var in []byte
		if in, err = ioutil.ReadFile(file); err != nil {
			return err
		}

		ext := filepath.Ext(file)

		switch {
		case c.regexpYAML.MatchString(ext):
			err = c.unmarshalYAML(in, config)
		case c.regexpTOML.MatchString(ext):
			err = c.unmarshalTOML(in, config)
		case c.regexpJSON.MatchString(ext):
			err = c.unmarshalJSON(in, config)
		default:
			err = fmt.Errorf("unknown data format, can't unmarshal file: '%s'", file)
		}

		if err != nil {
			return err
		}

		if err = c.parseTemplateFile(file, config); err != nil {
			return err
		}
	}

	//configB, err := yaml.Marshal(config)
	//if err != nil {
	//	return err
	//}
	//if err = parseTemplate(configB, config); err != nil {
	//	return err
	//}

	return c.parseConfigTags(config)
}

// File search ---------------------------------------------------------------------------------------------------------

// walkConfigPath look for a file matching the passed regex skipping sub-directories.
func (c configParser) walkConfigPath(configPath string, regex *regexp.Regexp) (matchedFile string, err error) {
	err = filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
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

	return
}

// configFilesByEnv will search for the given file names in the given path
// returning all the eligible files (eg.: <path>/config.yaml or <path>/config.<environment>.json)
//
// Files name can also be passed without file extension,
// configFilesByEnv is agnostic and will match any
// supported extension using the regex: `(?i)(.y(|a)ml|.toml|.json)`.
//
// The 'file' name will be searched as (in that order):
//  - '<path>/<file>(.* || <the_provided_extension>)'
//  - '<path>/<file>.<environment>(.* || <the_provided_extension>)'
//
// The latest found files will override previous.
func (c configParser) configFilesByEnv(envId string, files ...string) (foundFiles []string, err error) {
	for _, file := range files {
		configPath, fileName := filepath.Split(file)
		if len(configPath) == 0 {
			configPath = "./"
		}

		ext := filepath.Ext(fileName)
		extTrimmed := strings.TrimSuffix(fileName, ext)
		if len(ext) == 0 {
			ext = c.regexpValidExt.String() // search for any compatible file
		}

		format := "^%s%s$"
		if !c.FileSearchCaseSensitive {
			format = "(?i)(^%s)%s$"
		}
		// look for the config file in the config path (eg.: tool.yml)
		regex := regexp.MustCompile(fmt.Sprintf(format, extTrimmed, ext))
		var foundFile string
		foundFile, err = c.walkConfigPath(configPath, regex)
		if err != nil {
			break
		}
		if len(foundFile) > 0 {
			foundFiles = append(foundFiles, foundFile)
		}

		if len(envId) > 0 {
			// look for the env config file in the config path (eg.: tool.development.yml)
			//regexEnv := regexp.MustCompile(fmt.Sprintf(format, fmt.Sprintf("%s.%s", extTrimmed, Env().ID()), ext))
			regexEnv := regexp.MustCompile(fmt.Sprintf(format, fmt.Sprintf("%s.%s", extTrimmed, envId), ext))
			foundFile, err = c.walkConfigPath(configPath, regexEnv)
			if err != nil {
				break
			}
			if len(foundFile) > 0 {
				foundFiles = append(foundFiles, foundFile)
			}
		}
	}

	if err == nil && len(foundFiles) == 0 {
		err = fmt.Errorf("no config file found for '%s'", strings.Join(files, " | "))
	}
	return
}

// File parse ----------------------------------------------------------------------------------------------------------

func (c configParser) unmarshalJSON(data []byte, config interface{}) (err error) {
	return json.Unmarshal(data, config)
}

func (c configParser) unmarshalTOML(data []byte, config interface{}) (err error) {
	_, err = toml.Decode(string(data), config)
	return err
}

func (c configParser) unmarshalYAML(data []byte, config interface{}) (err error) {
	return yaml.Unmarshal(data, config)
}

// parseTemplateBytes parse all text/template placeholders
// (eg.: {{.Key}}) in config files.
func (c configParser) parseTemplateBytes(file []byte, config interface{}) error {
	tpl, err := template.New("tpl").Parse(string(file))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, config); err != nil {
		return err
	}

	switch {
	case c.unmarshalJSON(buf.Bytes(), config) == nil:
		return nil
	case c.unmarshalYAML(buf.Bytes(), config) == nil:
		return nil
	case c.unmarshalTOML(buf.Bytes(), config) == nil:
		return nil
	default:
		return fmt.Errorf("the provided data is incompatible with an interface of type %T:\n%s",
			config, strings.TrimSuffix(string(file), "\n"))
	}
}

// parseTemplateFile parse all text/template placeholders
// (eg.: {{.Key}}) in config files.
func (c configParser) parseTemplateFile(file string, config interface{}) error {
	tpl, err := template.ParseFiles(file)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, config); err != nil {
		return err
	}

	ext := filepath.Ext(file)

	switch {
	case c.regexpYAML.MatchString(ext):
		return c.unmarshalYAML(buf.Bytes(), config)
	case c.regexpTOML.MatchString(ext):
		return c.unmarshalTOML(buf.Bytes(), config)
	case c.regexpJSON.MatchString(ext):
		return c.unmarshalJSON(buf.Bytes(), config)
	default:
		return fmt.Errorf("unknown data format, can't unmarshal file: '%s'", file)
	}
}

// Flags parse ---------------------------------------------------------------------------------------------------------

// parseConfigTags will process the struct field tags.
func (c configParser) parseConfigTags(elem interface{}) error {
	elemValue := reflect.Indirect(reflect.ValueOf(elem))

	switch elemValue.Kind() {

	case reflect.Struct:
		elemType := elemValue.Type()
		//fmt.Printf("%sProcessing STRUCT: %s = %+v\n", indent, elemType.Name(), elem)

		for i := 0; i < elemType.NumField(); i++ {

			ft := elemType.Field(i)
			fv := elemValue.Field(i)

			if !fv.CanAddr() || !fv.CanInterface() {
				//fmt.Printf("%sCan't addr or interface FIELD: CanAddr: %v, CanInterface: %v. -> %s = '%+v'\n", indent, fv.CanAddr(), fv.CanInterface(), ft.Name, fv.Interface())
				continue
			}

			tag := ft.Tag.Get(c.sftKey)
			tagFields := strings.Split(tag, ",")
			//fmt.Printf("\n%sProcessing FIELD: %s %s = %+v, tags: %s\n", indent, ft.Name, ft.Type.String(), fv.Interface(), tag)
			for _, flag := range tagFields {

				kv := strings.Split(flag, "=")

				if kv[0] == c.sffEnv {
					if len(kv) == 2 {
						if value := os.Getenv(kv[1]); len(value) > 0 {
							//debugPrintf("Loading configuration for struct `%v`'s field `%v` from env %v...\n", elemType.Name(), ft.Name, kv[1])
							if err := yaml.Unmarshal([]byte(value), fv.Addr().Interface()); err != nil {
								return err
							}
						}
					}
				}

				if empty := reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()); empty {
					if kv[0] == c.sffDefault {
						if len(kv) == 2 {
							if err := yaml.Unmarshal([]byte(kv[1]), fv.Addr().Interface()); err != nil {
								return err
							}
						}
					} else if kv[0] == c.sffRequired {
						return errors.New(ft.Name + " is required")
					}
				}
			}

			switch fv.Kind() {
			case reflect.Ptr, reflect.Struct, reflect.Slice, reflect.Map:
				if err := c.parseConfigTags(fv.Addr().Interface()); err != nil {
					return err
				}
			}

			//fmt.Printf("%sProcessed  FIELD: %s %s = %+v\n", indent, ft.Name, ft.Type.String(), fv.Interface())
		}

	case reflect.Slice:
		for i := 0; i < elemValue.Len(); i++ {
			if err := c.parseConfigTags(elemValue.Index(i).Addr().Interface()); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range elemValue.MapKeys() {
			if err := c.parseConfigTags(elemValue.MapIndex(key).Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}
