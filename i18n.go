package i18n

import (
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"

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
	// Could be easily generated using `language.English.String()`
	// from `golang.org/x/text/language` package.
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
func NewWithConfig(config *Config) (*I18n, error) {
	i18n := &I18n{Config: config}

	if err := i18n.setup(); err != nil {
		return nil, err
	}

	return i18n, nil
}

// New create a new instance of i18n.
// configFilePath is the path of the config file (like i18n.yaml).
func NewWithConfigFile(configFilePath string) (*I18n, error) {
	i18n := &I18n{Config: &Config{}}

	if len(configFilePath) == 0 {
		return nil, errors.New("invalid config file path")
	}

	if compsConfigFile, err := ioutil.ReadFile(configFilePath); err != nil {
		return nil, err
	} else if err = sprbox.ConfigParser.Unmarshal(compsConfigFile, i18n.Config); err != nil {
		return nil, err
	}

	if err := i18n.setup(); err != nil {
		return nil, err
	}

	return i18n, nil
}

// SpareConfig is the sprbox 'configurable' interface implementation.
func (i18n *I18n) SpareConfig(configFiles []string) (err error) {
	if err = sprbox.ConfigParser.Load(&i18n.Config, configFiles...); err == nil {
		err = i18n.setup()
	}
	return
}

func (i18n *I18n) setup() error {
	if len(i18n.Config.Locales) == 0 {
		return errors.New("i18n.Locales can't be left empty, at least one locale must be provided")
	}

	i18n.Tags = parseLocalesToTags(i18n.Config.Locales)
	i18n.matcher = language.NewMatcher(i18n.Tags)

	if len(i18n.Config.Path) > 0 {
		return i18n.LoadLocalizationFiles(i18n.Config.Path)
	} else if i18n.Config.Locs != nil {
		i18n.localizations = i18n.Config.Locs
		return nil
	} else {
		return errors.New("config.Path or config.Locs must be provided")
	}
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
		if err := sprbox.ConfigParser.Load(&langLocalizations, locFileName); err != nil {
			return err
		}

		i18n.localizations[lang.String()] = langLocalizations
	}

	return
}
