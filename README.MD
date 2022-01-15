# i18n

[![GitHub tag](https://img.shields.io/github/tag/oblq/i18n.svg)](https://github.com/oblq/i18n)
[![Build Status](https://travis-ci.org/oblq/i18n.svg?branch=master)](https://travis-ci.org/oblq/i18n)
[![codecov](https://codecov.io/gh/oblq/i18n/branch/master/graph/badge.svg)](https://codecov.io/gh/oblq/i18n)
[![Go Report Card](https://goreportcard.com/badge/github.com/oblq/i18n)](https://goreportcard.com/report/github.com/oblq/i18n)
[![GoDoc](https://godoc.org/github.com/oblq/i18n?status.svg)](https://godoc.org/github.com/oblq/i18n)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)

**i18n** is a minimal, flexible, simple to use and simple to embed (in your own packages) localizations package.  
It accept yaml, json or toml for both config and localizations files.

## Installation

```sh
go get -u github.com/oblq/i18n/v2
```

## Quickstart

Create a new instance:
```go
package main

import (
	"github.com/oblq/i18n/v2"
	"golang.org/x/text/language"
)

func init() {
	// Optionally pass the config file path and nil config.
	// The config file can be in yaml, json or toml format.
	localizer, err := i18n.NewWithConfig(&i18n.Config{
            Locales: []string{
                language.English.String(), // "en"
                language.Italian.String(), // "it"
            },
            Path: "./example/localizations",
        })	
}
```


Localize a key:
```go
localizer.T("en", "MY_KEY")
```


Localize a key using plural and multiple parameters:  
```yaml
# ./example/localizations/en.yaml file

WHOAMI: "I'm %s!"
```

```go
localizer.TP("en", "WHOAMI", "i18n") // -> "I'm i18n!"
```


Automatically localize a key based on the http request, i18n will first look for the locale by the GetLocaleOverride func, then in cookies (`language` and/or `lang` keys), then in `Accept-Language` header:

```go
func(w http.ResponseWriter, r *http.Request) {
	response := localizer.AutoT(r, "MY_KEY")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
```


Optionally override the GetLocale func to automatically determine the locale to be used from http requests (using i18n.AutoT func), if an empty string is returned the default methods will be used anyway:
```go
localizer.GetLocaleOverride = func(r *http.Request) string {
    user := MyCustomAuthLib.UserFromRequest(r)
    return user.Locale
}
```          

## Localized file server:

```go
localizer.SetFileServer(
    map[string]http.Handler{
        language.English.String(): http.FileServer(http.Dir("./landing_en")),
        language.Italian.String(): http.FileServer(http.Dir("./landing_it")),
    }, 
)

mux.Handle("/", localizer)
```

## Embed i18n in your package:

Use hardcoded localizations, they can be a json, yaml or toml string:
```go
package main

import (
	"github.com/oblq/i18n/v2"
	"golang.org/x/text/language"
)

// keys
const (
	GEM = "GEM" // generic_error_message
)

var hardcodedLocs = map[string]map[string]string{
	language.English.String(): {
		GEM: "Something went wrong, please try again later %s",
	},
	language.Italian.String(): {
		GEM: "Qualcosa è andato storto, riprova più tardi %s",
	},
}

var localizer *i18n.I18n

func init() {
	localizer, err := i18n.NewWithConfig(&i18n.Config{
		Locales: []string{
			language.English.String(), // "en"
			language.Italian.String(), // "it"
		},
		Locs: hardcodedLocs,
	})
}
```

## Middleware

```go
package main

import (
	"net/http"
	
	"github.com/oblq/i18n/v2"
	"golang.org/x/text/language"
)

func init() {
	config := &i18n.Config{
		HTTPLookUpStrategy: []i18n.HTTPLocalePosition{
			{i18n.HTTPLocalePositionIDHeader, "Accept-Language"},
			{i18n.HTTPLocalePositionIDCookie, "locale"},
			{i18n.HTTPLocalePositionIDQuery, "lang"},
		},
		Locales: []string{
			language.Italian.String(), 
			language.English.String(),
		},
	}

	localizer, err := i18n.NewWithConfig(config)
	if err != nil {
		panic(err)
	}

	http.Handle("/this-is-my-locale", localizer.Middleware(http.HandlerFunc(getLocale)))
	http.ListenAndServe(":8080", nil)
}

func getLocale(w http.ResponseWriter, r *http.Request) {
	locale := r.Context().Value(i18n.MiddlewareContextLocaleKey)
	_, _ = w.Write([]byte(locale.(string)))
}
```

## Vendored packages

- [`golang.org/x/text/language`](golang.org/x/text/language)
- [`gopkg.in/yaml.v2`](https://github.com/go-yaml/yaml)  
- [`github.com/BurntSushi/toml`](https://github.com/BurntSushi/toml)

## Author

- [Marco Muratori](mailto:marcomrtr@gmail.com) 

## License

i18n is available under the MIT license. See the [LICENSE](./LICENSE) file for more information.
