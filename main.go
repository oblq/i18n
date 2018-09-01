package i18n

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/oblq/sprbox"
	"golang.org/x/text/language"
)

// Config is the i18n config struct.
type Config struct {
	// Locales order is important, the first one is the default,
	// they must be ordered from the most preferred to te least one.
	// A localization file for any given locale must be provided.
	Locales []string `yaml:"locales"`
	// LocalizationsPath is the path of localization files.
	LocalizationsPath string `yaml:"localizations_path"`
}

// localization represent a key value with localized strings
// for both single and plural results
type localization struct {
	One, Other string
}

// I18n is the i18n instance.
type I18n struct {
	// Config struct
	Config *Config

	// GetLocaleOverride override the default method to get the user locale.
	// If nothing is returned the default method will be called anyway.
	GetLocaleOverride func(r *http.Request) string `json:"-"`

	// Tags is automatically generated using Config.Locales.
	// The first one is the default, they must be ordered from the most preferred to te least one.
	Tags []language.Tag

	// matcher is a language.Matcher configured for all supported languages.
	// Automatically generated using Config.Locales.
	matcher language.Matcher

	// language, key -> localization
	localizations map[string]map[string]localization

	// localizedHandlers is used by the FileServer
	localizedHandlers map[string]http.Handler
}

// NewI18n create a new instance of i18n.
// configFilePath take precedence over config.
func NewI18n(configFilePath string, config *Config) *I18n {
	if config == nil {
		config = &Config{}
	}

	i18n := &I18n{Config: config}

	if len(configFilePath) > 0 {
		if compsConfigFile, err := ioutil.ReadFile(configFilePath); err != nil {
			log.Fatalln("Wrong config path", err)
		} else if err = sprbox.Unmarshal(compsConfigFile, i18n.Config); err != nil {
			log.Fatalln("Can't unmarshal config file", err)
		}
	}

	if err := i18n.setup(); err != nil {
		fmt.Printf("[i18n] error: %v", err)
	}

	return i18n
}

// SpareConfig is the sprbox 'configurable' interface implementation.
func (i18n *I18n) SpareConfig(configFiles []string) error {
	if err := sprbox.LoadConfig(&i18n.Config, configFiles...); err != nil {
		return err
	}
	if err := i18n.setup(); err != nil {
		return err
	}

	return nil
}

func (i18n *I18n) setup() error {
	i18n.Tags = parseLocalesToTags(i18n.Config.Locales)
	if len(i18n.Tags) == 0 {
		i18n.Tags = []language.Tag{
			language.English,
		}
	}

	i18n.matcher = language.NewMatcher(i18n.Tags)
	i18n.localizations = make(map[string]map[string]localization)

	for _, lang := range i18n.Tags {
		var newLanguage map[string]localization

		//if locFile, err := ioutil.ReadFile(filepath.Join(i18n.Config.LocalizationsPath, lang.String()+".yml")); err != nil {
		//	return fmt.Errorf("[i18n] missing %s.* file in static folder: %v", lang.String(), err)
		//} else if err = sprbox.Unmarshal(locFile, &newLanguage); err != nil {
		//	return fmt.Errorf("[i18n] error: %v", err)
		//}

		if err := LoadLocalization(
			&newLanguage,
			filepath.Join(i18n.Config.LocalizationsPath, lang.String()),
		); err != nil {
			return fmt.Errorf("[i18n] error: %v", err)
		}
		i18n.localizations[lang.String()] = newLanguage
	}

	return nil
}
