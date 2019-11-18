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
//
// The Locales order is important,
// they must be ordered from the most preferred to te least one,
// the first one is the default.
//
// Set Path OR Locs, Path takes precedence over Locs.
// Set Locs if you want to use hardcoded localizations,
// useful to embed i18n in other library packages.
// Otherwise set Path to load localization files.
type Config struct {
	// Locales order is important, the first one is the default,
	// they must be ordered from the most preferred to te least one.
	// A localization file for any given locale must be provided.
	Locales []string

	// Path is the path of localization files.
	// Files will be searched automatically based on Locales.
	Path string

	// Locs contains hardcoded localizations.
	// Use it if you want to use hardcoded localizations,
	// useful to embed i18n in other library packages.
	Locs map[string]map[string]*Localization
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
	// If an empty string is returned the default methods
	// will be used anyway (request's cookies and header).
	GetLocaleOverride func(r *http.Request) string `json:"-"`

	// Tags is automatically generated using Config.Locales.
	// The first one is the default, they must be
	// ordered from the most preferred to te least one.
	Tags []language.Tag

	// matcher is a language.Matcher configured for all supported languages.
	// Automatically generated using Config.Locales.
	matcher language.Matcher

	// localizations[<language>][<key>] -> Localization
	localizations map[string]map[string]*Localization

	// localizedHandlers is used by the FileServer
	localizedHandlers map[string]http.Handler
}

// New create a new instance of i18n.
func NewWithConfig(config *Config) *I18n {
	i18n := &I18n{Config: config}

	if err := i18n.setup(); err != nil {
		log.Fatal("[i18n] error:", err.Error())
	}

	return i18n
}

// New create a new instance of i18n.
// configFilePath is the path of the config file (like i18n.yaml).
func NewWithConfigFile(configFilePath string) *I18n {
	i18n := &I18n{Config: &Config{}}

	if len(configFilePath) == 0 {
		log.Fatal("[i18n] error: invalid config file path")
	}

	if compsConfigFile, err := ioutil.ReadFile(configFilePath); err != nil {
		log.Fatalln("Wrong config path", err)
	} else if err = sprbox.Unmarshal(compsConfigFile, i18n.Config); err != nil {
		log.Fatalln("Can't unmarshal config file", err)
	}

	if err := i18n.setup(); err != nil {
		log.Fatal("[i18n] error:", err.Error())
	}

	return i18n
}

// SpareConfig is the sprbox 'configurable' interface implementation.
func (i18n *I18n) SpareConfig(configFiles []string) (err error) {
	if err = sprbox.LoadConfig(&i18n.Config, configFiles...); err == nil {
		err = i18n.setup()
	}
	return
}

func (i18n *I18n) setup() error {
	i18n.Tags = parseLocalesToTags(i18n.Config.Locales)
	i18n.matcher = language.NewMatcher(i18n.Tags)

	if len(i18n.Config.Path) > 0 {
		return i18n.LoadLocalizationFiles(i18n.Config.Path)
	} else if i18n.Config.Locs != nil {
		i18n.localizations = i18n.Config.Locs
		return nil
	} else {
		return errors.New("configFilePath or config or locs: one of the two must be provided")
	}
}
