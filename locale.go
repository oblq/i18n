package i18n

import (
	"net/http"

	"log"

	"golang.org/x/text/language"
)

const (
	cookieKey1 = "language"
	cookieKey2 = "lang"

	headerKey = "Accept-Language"
)

// GetLocale return the request language Tag.
// tag.String() -> locale
// First look for user language by the GetLocaleOverride func,
// then in cookies ("language" and/or "lang" keys),
// then in 'Accept-Language' header.
func (i18n *I18n) GetLocale(r *http.Request) language.Tag {
	locale := ""

	if i18n.GetLocaleOverride != nil {
		locale = i18n.GetLocaleOverride(r)
	}

	if len(locale) == 0 {
		if cookieLang, err := r.Cookie(cookieKey1); err == nil {
			locale = cookieLang.Value
		} else if cookieLang, err := r.Cookie(cookieKey2); err == nil {
			locale = cookieLang.Value
		} else if acceptLang := r.Header.Get(headerKey); len(acceptLang) > 0 {
			locale = acceptLang
		}
	}

	t, _, _ := language.ParseAcceptLanguage(locale) // We ignore the error: the default language will be selected for t == nil.
	// we don't return tag anymore since it has some bugs, we can retrieve it from supported languages with index
	//tag, _, _ := matcher.Match(t...)
	_, i, _ := i18n.matcher.Match(t...)

	if len(i18n.Tags) > i {
		return i18n.Tags[i]
	}
	return i18n.Tags[0]
}

// parseLocalesToTags convert an array of locales to an array of language.Tag.
func parseLocalesToTags(locales []string) (tags []language.Tag) {
	for _, locale := range locales {
		newTag, err := language.Parse(locale)
		if err != nil {
			log.Fatal("can't parse one or more locale identifier:", err.Error())
		}

		tags = append(tags, newTag)
	}
	return
}
