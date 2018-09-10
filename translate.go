package i18n

import (
	"fmt"
	"net/http"
)

func (i18n *I18n) translate(locale string, plural bool, key string, params ...interface{}) string {
	var language map[string]*Localization
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

// EXPORTED ------------------------------------------------------------------------------------------------------------

// T translate the key based on the passed locale.
func (i18n *I18n) T(locale string, key string, params ...interface{}) string {
	return i18n.translate(locale, false, key, params...)
}

// AutoT automatically translate the key based on the http request:
// it will first look for user language by the GetLocaleOverride func,
// then in cookies ("language" and/or "lang" keys),
// then in 'Accept-Language' header.
func (i18n *I18n) AutoT(r *http.Request, key string, params ...interface{}) string {
	if r == nil {
		fmt.Println("[i18n] http request nil, key:", key)
		return key
	}
	locale := i18n.GetLocale(r)
	return i18n.translate(locale, false, key, params...)
}

// TP translate the key based on the passed locale
// and for possibly plural values.
func (i18n *I18n) TP(locale string, plural bool, key string, params ...interface{}) string {
	return i18n.translate(locale, plural, key, params...)
}

// AutoTP automatically translate the key based on the
// http request and for possibly plural values:
// it will first look for user language by the GetLocaleOverride func,
// then in cookies ("language" and/or "lang" keys),
// then in 'Accept-Language' header.
func (i18n *I18n) AutoTP(r *http.Request, plural bool, key string, params ...interface{}) string {
	if r == nil {
		fmt.Println("[i18n] http request nil, key:", key)
		return key
	}
	locale := i18n.GetLocale(r)
	return i18n.translate(locale, plural, key, params...)
}
