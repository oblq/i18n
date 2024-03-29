package i18n

import (
	"log"
	"net/http"

	"golang.org/x/text/language"
)

// getLocaleUnsafe return the request locale.
// It first looks for the locale by the
// GetLocaleOverride func (if defined),
// then if locale is empty it will look in:
// - cookies ("language" and/or "lang" keys)
// - 'Accept-Language' header
func (i18n *I18n) getLocaleUnsafe(r *http.Request) (locale string) {
	if r == nil {
		return
	}

	if i18n.GetLocaleOverride != nil {
		locale = i18n.GetLocaleOverride(r)
	}

	if len(locale) == 0 {
		for _, lookUpStrategy := range i18n.Config.HTTPLookUpStrategy {
			switch lookUpStrategy.ID {
			case HTTPLocalePositionIDHeader:
				locale = r.Header.Get(lookUpStrategy.Key)
			case HTTPLocalePositionIDCookie:
				if cookieLang, err := r.Cookie(lookUpStrategy.Key); err == nil {
					locale = cookieLang.Value
				}
			case HTTPLocalePositionIDQuery:
				locale = r.URL.Query().Get(lookUpStrategy.Key)
			}

			if len(locale) > 0 {
				break
			}
		}
	}
	return
}

// GetLanguageTag return the request language.Tag.
// A recognized tag is always returned.
//  <language.Tag>.String() // -> locale
// It first looks for the request locale (GetLocale func),
// then it will look for the corresponding language.Tag
// in the i18n predefined Tags, if no tag is matched
// the first one will be returned.
func (i18n *I18n) GetLanguageTag(r *http.Request) language.Tag {
	locale := i18n.getLocaleUnsafe(r)
	return i18n.MatchAvailableLanguageTag(locale)
}

// GetLocale return GetLanguageTag(r).String()
// It always returns a valid result.
func (i18n *I18n) GetLocale(r *http.Request) (locale string) {
	return i18n.GetLanguageTag(r).String()
}

// MatchAvailableLanguageTag return one of the available locales
// corresponding language.Tag.
//  <language.Tag>.String() // -> locale
// A recognized language is always returned.
// If no locale is matched the first one from the supported list
// will be returned.
func (i18n *I18n) MatchAvailableLanguageTag(locale string) language.Tag {
	if len(locale) > 0 {
		// We ignore the error: the default language will be selected for t == nil.
		t, _, _ := language.ParseAcceptLanguage(locale)
		// we don't return tag anymore since it has some bugs, we can retrieve it from supported languages with index
		//tag, _, _ := matcher.Match(t...)
		_, i, _ := i18n.matcher.Match(t...)

		if len(i18n.Tags) > i {
			return i18n.Tags[i]
		}
	}
	return i18n.Tags[0]
}

// parseLocalesToTags convert an array of locales to an array of language.Tag.
// If no language.Tag can be parsed for the provided locales
// then language.English will be returned by default in tags array.
func parseLocalesToTags(locales []string) (tags []language.Tag) {
	for _, locale := range locales {
		newTag, err := language.Parse(locale)
		if err != nil {
			log.Fatal("can't parse one or more locale identifier:", err.Error())
		}

		tags = append(tags, newTag)
	}

	if len(tags) == 0 {
		tags = []language.Tag{
			language.English,
		}
	}

	return
}
