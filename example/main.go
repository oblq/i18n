package main

import (
	"net/http"

	"github.com/xenolf/lego/log"

	"github.com/oblq/i18n"
)

func main() {
	localizer := i18n.NewI18n(
		"",
		&i18n.Config{
			Locales:           []string{"en", "it"},
			LocalizationsPath: "./example/localizations",
		},
	)

	http.HandleFunc("/one", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(localizer.AutoT(r, false, "GEM", "Marco")))
	})

	http.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(localizer.AutoT(r, true, "GEM", "Marco")))
	})

	// http://localhost:8888/hardcoded?plural=true
	http.HandleFunc("/hardcoded", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		plural := r.FormValue("plural") != "true"
		w.Write([]byte(localizer.T("it", plural, "GEM", "Marco")))
	})

	log.Fatal(http.ListenAndServe(":8888", nil))
}
