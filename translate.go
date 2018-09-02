package i18n

import (
	"fmt"
	"log"
	"net/http"
)

// T translate the key based on the passed locale.
func (i18n *I18n) T(locale string, plural bool, key string, params ...interface{}) string {
	var language map[string]Localization
	var ok bool
	if language, ok = i18n.localizations[locale]; !ok {
		language, _ = i18n.localizations[i18n.Tags[0].String()]
	}

	if localization, ok := language[key]; ok {
		if plural {
			return fmt.Sprintf(localization.Other, params...)
		}
		return fmt.Sprintf(localization.One, params...)
	}
	return key
}

// AutoT automatically translate the key based on the http request:
// it will first look for user language by the GetLocaleOverride func,
// then in cookies ("language" and/or "lang" keys),
// then in 'Accept-Language' header.
func (i18n *I18n) AutoT(r *http.Request, plural bool, key string, params ...interface{}) string {
	if r == nil {
		log.Println("Loc: http request was nil for the key:", key)
		return key
	}
	tag := i18n.GetLocale(r)
	return i18n.T(tag.String(), plural, key, params...)
}
