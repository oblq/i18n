package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/oblq/i18n/v2"
	"golang.org/x/text/language"
)

func main() {
	localizer, _ := i18n.NewWithConfig(&i18n.Config{
		Locales: []string{"en", "it"},
		Path:    "./example/i18n",
	})

	http.HandleFunc("/one", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := []byte(localizer.AutoT(r, "GEM", "Marco"))
		_, _ = w.Write(response)
	})

	http.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := []byte(localizer.AutoTP(r, true, "GEM", "Marco"))
		_, _ = w.Write(response)
	})

	// http://localhost:8888/manual?plural=true
	http.HandleFunc("/manual", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		plural := r.FormValue("plural") == "true"
		response := []byte(localizer.TP("it", plural, "GEM", "Marco"))
		_, _ = w.Write(response)
	})

	// localized FileServer
	localizer.SetFileServer(
		map[string]http.Handler{
			language.English.String(): http.FileServer(http.Dir("./example/web/en")),
			language.Italian.String(): http.FileServer(http.Dir("./example/web/it")),
		},
	)
	http.Handle("/", localizer)

	fmt.Println("Try: http://localhost:8888/one")
	fmt.Println("Try: http://localhost:8888/other")
	fmt.Println("Try: http://localhost:8888/manual?plural=true")
	fmt.Println("Try: http://localhost:8888/")

	log.Fatal(http.ListenAndServe(":8888", nil))
}
