package i18n

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/oblq/sprbox"
	"golang.org/x/text/language"
)

// Config is the i18n config struct.
type Config struct {
	// Locales order is important, the first one is the default,
	// they must be ordered from the most preferred to te least one.
	// A localization file for any given locale must be provided.
	Locales []string

	// Path is the path of localization files.
	Path string
}

// Localization represent a key value with localized strings
// for both single and plural results
type Localization struct {
	One, Other string
}

// I18n is the i18n instance.
type I18n struct {
	// Config struct
	Config *Config

	// GetLocaleOverride override the default method
	// to get the http request locale.
	// If nothing is returned the default method will be
	// called anyway (request's cookies and header).
	GetLocaleOverride func(r *http.Request) string `json:"-"`

	// Tags is automatically generated using Config.Locales.
	// The first one is the default, they must be
	// ordered from the most preferred to te least one.
	Tags []language.Tag

	// matcher is a language.Matcher configured for all supported languages.
	// Automatically generated using Config.Locales.
	matcher language.Matcher

	// localizations[<language>][<key>] -> Localization
	localizations map[string]map[string]Localization

	// localizedHandlers is used by the FileServer
	localizedHandlers map[string]http.Handler
}

// New create a new instance of i18n.
// 1. configFilePath is the path of the config file (like i18n.yaml in the root).
// 2. config is a config struct provided at run-time.
// 3. locs contains hardcoded localizations.
//
// Use 1 or 2 if you have localizations files in a predefined
// path, files will be searched by locales automatically.
//
// Use 3 if you want to use hardcoded localizations,
// useful to embed i18n in other library packages.
//
// `configFilePath` take precedence over `config` which takes precedence over `locs`.
//
// The order of Config.Locales or locs[<locale>] is important,
// they must be ordered from the most preferred to te least one,
// the first one is the default.
func New(configFilePath string, config *Config, locs map[string]map[string]Localization) *I18n {
	i18n := &I18n{Config: config}

	if len(configFilePath) > 0 {
		if i18n.Config == nil {
			i18n.Config = &Config{}
		}
		if compsConfigFile, err := ioutil.ReadFile(configFilePath); err != nil {
			log.Fatalln("Wrong config path", err)
		} else if err = sprbox.Unmarshal(compsConfigFile, i18n.Config); err != nil {
			log.Fatalln("Can't unmarshal config file", err)
		}
	}

	if err := i18n.setup(locs); err != nil {
		log.Fatal("[i18n] error:", err.Error())
	}

	return i18n
}

// SpareConfig is the sprbox 'configurable' interface implementation.
func (i18n *I18n) SpareConfig(configFiles []string) (err error) {
	if err = sprbox.LoadConfig(&i18n.Config, configFiles...); err == nil {
		err = i18n.setup(nil)
	}
	return
}

func (i18n *I18n) setup(locs map[string]map[string]Localization) error {
	if i18n.Config != nil {
		i18n.Tags = parseLocalesToTags(i18n.Config.Locales)
		i18n.matcher = language.NewMatcher(i18n.Tags)
		return i18n.LoadLocalizationFiles(i18n.Config.Path)
	} else if locs != nil {
		locales := make([]string, 0)
		for k := range locs {
			locales = append(locales, k)
		}
		i18n.Tags = parseLocalesToTags(locales)
		i18n.matcher = language.NewMatcher(i18n.Tags)
		i18n.localizations = locs
		return nil
	} else {
		return errors.New("configFilePath or config or locs: one of the three must be provided")
	}
}
