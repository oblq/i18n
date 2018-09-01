package i18n

import (
	"net/http"
)

// FileServer provides a different handler for any specific language.
// Must set an handler for any supported languages
// otherwise the primary language will be used.
//
// EXAMPLE:
//	landingHandler := I18n.FileServer(
//		map[string]http.Handler{
//			language.English.String(): http.FileServer(http.Dir("./landing_en")),
//			language.Italian.String(): http.FileServer(http.Dir("./landing_ita")),
//		})
//
//	mux.Handle("/", landingHandler)
func (i18n *I18n) FileServer(handlers map[string]http.Handler) *I18n {
	i18n.localizedHandlers = handlers
	return i18n
}

func (i18n *I18n) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	locale := i18n.GetLocale(r).String()
	if handler, ok := i18n.localizedHandlers[locale]; ok {
		handler.ServeHTTP(w, r)
	} else {
		i18n.localizedHandlers[i18n.Tags[0].String()].ServeHTTP(w, r)
	}
}
