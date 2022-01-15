package i18n

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/oblq/swap"
	"golang.org/x/text/language"
)

type HTTPLocalePositionID string

const (
	HTTPLocalePositionIDHeader HTTPLocalePositionID = "header"
	HTTPLocalePositionIDCookie HTTPLocalePositionID = "cookie"
	HTTPLocalePositionIDQuery  HTTPLocalePositionID = "query"
)

type HTTPLocalePosition struct {
	ID  HTTPLocalePositionID
	Key string
}

var DefaultHTTPLookUpStrategy = []HTTPLocalePosition{
	{HTTPLocalePositionIDHeader, "Accept-Language"},
	{HTTPLocalePositionIDCookie, "lang"},
	{HTTPLocalePositionIDQuery, "lang"},
}

// Config is the i18n config struct.
//
// The Locales order is important,
// they must be ordered from the most preferred to te least one,
// the first one is the default.
//
// Set Path OR Locs, Path takes precedence over Locs.
// Set Locs if you want to use hardcoded localizations,
// useful to embed i18n in other library packages.
// Otherwise, set Path to load localization files.
type Config struct {
	// HTTPLookUpStrategy represent the strategy to extract the language from the request.
	// The order of element is important, the first one is the default.
	// Default is DefaultMiddlewareLookUpStrategy.
	HTTPLookUpStrategy []HTTPLocalePosition

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
	Locs map[string]map[string]string
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
	localizations map[string]map[string]string

	// localizedHandlers is used by the FileServer
	localizedHandlers map[string]http.Handler
}

// NewWithConfig create a new instance of i18n.
func NewWithConfig(config *Config) (*I18n, error) {
	i18n := &I18n{Config: config}

	if err := i18n.setup(); err != nil {
		return nil, err
	}

	return i18n, nil
}

// NewWithConfigFile create a new instance of i18n.
// configFilePath is the path of the config file (like i18n.yaml).
func NewWithConfigFile(configFilePath string) (*I18n, error) {
	i18n := &I18n{Config: &Config{}}

	if len(configFilePath) == 0 {
		return nil, errors.New("invalid config file path")
	}

	if err := swap.Parse(i18n.Config, configFilePath); err != nil {
		return nil, err
	}

	if err := i18n.setup(); err != nil {
		return nil, err
	}

	return i18n, nil
}

// New is the github.com/oblq/swap`Factory` interface.
// must be a singleton otherwise a lot of connection to the db will remain open
func (i18n *I18n) New(configFiles ...string) (instance interface{}, err error) {
	config := new(Config)
	if err = swap.Parse(config, configFiles...); err != nil {
		return nil, err
	}

	if instance, err = NewWithConfig(config); err != nil {
		return nil, err
	}

	return
}

// Configure is the github.com/oblq/swap`Configurable` interface implementation.
func (i18n *I18n) Configure(configFiles ...string) (err error) {
	if err = swap.Parse(i18n.Config, configFiles...); err == nil {
		err = i18n.setup()
	}
	return
}

func (i18n *I18n) setup() error {
	if i18n.Config.HTTPLookUpStrategy == nil {
		i18n.Config.HTTPLookUpStrategy = DefaultHTTPLookUpStrategy
	}

	if len(i18n.Config.Locales) == 0 {
		return errors.New("i18n.Locales can't be left empty, at least one locale must be provided")
	}

	i18n.Tags = parseLocalesToTags(i18n.Config.Locales)
	i18n.matcher = language.NewMatcher(i18n.Tags)

	if i18n.Config.Locs != nil {
		i18n.localizations = i18n.Config.Locs
		return nil
	} else if len(i18n.Config.Path) > 0 {
		return i18n.LoadLocalizationFiles(i18n.Config.Path)
	}

	return nil
}

// LoadLocalizationFiles will unmarshal all the matched
// localization files in i18n.Config.LocalizationsPath for the given i18n.Tags,
// localization files must be named as the <language.Tag>.String()
// (locale, e.g.: `en.yml` for `language.English`).
func (i18n *I18n) LoadLocalizationFiles(localizationsPath string) (err error) {
	i18n.localizations = make(map[string]map[string]string)

	for _, lang := range i18n.Tags {
		var langLocalizations map[string]string
		locFileName := filepath.Join(localizationsPath, lang.String())
		if err := swap.Parse(&langLocalizations, locFileName); err != nil {
			return err
		}

		i18n.localizations[lang.String()] = langLocalizations
	}

	return
}
