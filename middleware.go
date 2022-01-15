package i18n

import (
	"context"
	"net/http"
)

const MiddlewareContextLocaleKey = "locale"

// Middleware looks for a language setting in the request
// and sets the request locale in context.
// It looks for the language using the `i18n.Config.HTTPLookUpStrategy`.
func (i18n *I18n) Middleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := i18n.GetLocale(r)
		updatedContext := context.WithValue(r.Context(), MiddlewareContextLocaleKey, locale)
		updatedRequest := r.WithContext(updatedContext)
		nextHandler.ServeHTTP(w, updatedRequest)
	})
}
