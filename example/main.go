package main

import (
	"log"
	"net/http"

	"github.com/oblq/i18n"
)

func main() {
	localizer := i18n.New(
		"",
		&i18n.Config{
			Locales:           []string{"en", "it"},
			LocalizationsPath: "./example/i18n",
		},
	)

	http.HandleFunc("/one", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := []byte(localizer.AutoT(r, "GEM", "Marco"))
		w.Write(response)
	})

	http.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := []byte(localizer.AutoTP(r, true, "GEM", "Marco"))
		w.Write(response)
	})

	// http://localhost:8888/hardcoded?plural=true
	http.HandleFunc("/hardcoded", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		plural := r.FormValue("plural") == "true"
		response := []byte(localizer.TP("it", plural, "GEM", "Marco"))
		w.Write(response)
	})

	log.Fatal(http.ListenAndServe(":8888", nil))
}
