// Package i18n provides support for looking up messages
// according to a set of locale preferences.
//
// Localization files can be yaml, json or toml.
//
// Get a new instance:
// 	localizer := i18n.NewI18n(
//		"",
//		&i18n.Config{
//			Locales: []string{"en", "it"},
//			LocalizationsPath: "./localizations",
//		},
//	)
//
// Optionally override the GetLocale func,
// if an empty string is returned the default method will be used anyway:
//  localizer.GetLocaleOverride = func(r *http.Request) string {
// 		user := Auth.UserFromRequest(r)
// 		locale := ""
// 		if user != nil {
//			Postgres.QueryRow("SELECT locale FROM user_data WHERE uid = $1", user.Uid).Scan(&locale)
// 		}
// 		return locale
//  }
//
// Localize a key based on the http request,
// i18n will first look for user language by the GetLocaleOverride func,
// then in cookies ("language" and/or "lang" keys),
// then in 'Accept-Language' header:
// 	localizer.AutoT(r, "MY_KEY")
//
// Localize a key based on the given locale:
//  localizer.T("en", "MY_KEY")
//
// Optionally pass parameters to be parsed with fmt package:
//  // en.yaml -> "SAY_HELLO": "Hello, %s!"
//  localizer.T("en", "SAY_HELLO", "Marco")
//
// Optionally use a localized file server:
//	landingHandler := localizer.FileServer(
//		map[string]http.Handler{
//			language.English.String(): http.FileServer(http.Dir("./landing_en")),
//			language.Italian.String(): http.FileServer(http.Dir("./landing_ita")),
//		})
//
//  mux.Handle("/", landingHandler)
package i18n
