# i18n

[![GitHub tag](https://img.shields.io/github/tag/oblq/i18n.svg)](https://github.com/oblq/i18n)
[![Build Status](https://travis-ci.org/oblq/i18n.svg?branch=master)](https://travis-ci.org/oblq/i18n)
[![codecov](https://codecov.io/gh/oblq/i18n/branch/master/graph/badge.svg)](https://codecov.io/gh/oblq/i18n)
[![Go Report Card](https://goreportcard.com/badge/github.com/oblq/i18n)](https://goreportcard.com/report/github.com/oblq/i18n)
[![GoDoc](https://godoc.org/github.com/oblq/i18n?status.svg)](https://godoc.org/github.com/oblq/i18n)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)
[![sprbox](https://img.shields.io/badge/sprbox-ready-green.svg)](https://github.com/oblq/sprbox)  

**i18n** is a minimal, flexible and simple to use localizations package.  
It support yaml, json or toml localization files.

i18n is ready to be used in sprbox.

## Install

```sh
$ go get github.com/oblq/i18n
```

## Quickstart
Get a new instance:
```go
localizer := i18n.NewI18n(
    "", // optionally pass the config file path and nil config
    &i18n.Config{
        Locales: []string{
        	language.English.String(), // "en"
        	language.Italian.String(), // "it"
        },
        LocalizationsPath: "./example/localizations",
    },
)
```

Optionally override the GetLocale func to automatically determine the locale to be used from http requests (using i18n.AutoT func), if an empty string is returned the default method will be used anyway:
```go
localizer.GetLocaleOverride = func(r *http.Request) string {
    user := Auth.UserFromRequest(r)
    locale := ""
    if user != nil {
        Postgres.QueryRow("SELECT locale FROM user_data WHERE uid = $1", user.Uid).Scan(&locale)
    }
    return locale
}
```          

Automatically localize a key based on the http request, i18n will first look for the locale by the GetLocaleOverride func, then in cookies (`language` and/or `lang` keys), then in `Accept-Language` header:

```go
localizer.AutoT(r, false, "MY_KEY")
```

Localize a key:
```go
localizer.T("en", false, "MY_KEY")
```
         
Optionally pass parameters to be parsed with fmt package, given the `en.yaml` localization file: 
```yaml
"SAY_HELLO": 
    one: "Hello, %s!"
    other: "Hello, %s and %s!"
```

...localize a key using plural and multiple parameters:
```go
localizer.T("en", true, "SAY_HELLO", "Marco", "Valentina") // -> "Hello, Marco and Valentina!"
```
         
## Localized file server:
```go
landingHandler := localizer.FileServer(
    map[string]http.Handler{
        language.English.String(): http.FileServer(http.Dir("./landing_en")),
        language.Italian.String(): http.FileServer(http.Dir("./landing_ita")),
    })

mux.Handle("/", landingHandler)
```

## Author

- [Marco Muratori](mailto:marcomrtr@gmail.com) 

## License

i18n is available under the MIT license. See the [LICENSE](./LICENSE) file for more information.
