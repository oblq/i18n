package i18n

import (
	"net/http"
)

// SetFileServer set a different handler for any specific language.
// The default language will be used if no i18n.Tags tag is matched.
//
// EXAMPLE:
//	i18nInstance.SetFileServer(
//		map[string]http.Handler{
//			language.English.String(): http.FileServer(http.Dir("./landing_en")),
//			language.Italian.String(): http.FileServer(http.Dir("./landing_ita")),
//		})
//
//	mux.Handle("/", i18nInstance)
func (i18n *I18n) SetFileServer(handlers map[string]http.Handler) {
	i18n.localizedHandlers = handlers
}

func (i18n *I18n) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	locale := i18n.GetLocale(r)
	i18n.localizedHandlers[locale].ServeHTTP(w, r)
}
