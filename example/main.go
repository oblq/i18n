package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/oblq/i18n"
)

func main() {
	localizer := i18n.New(
		"",
		&i18n.Config{
			Locales: []string{"en", "it"},
			Path:    "./example/i18n",
		},
		nil,
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

	// http://localhost:8888/manual?plural=true
	http.HandleFunc("/manual", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		plural := r.FormValue("plural") == "true"
		response := []byte(localizer.TP("it", plural, "GEM", "Marco"))
		w.Write(response)
	})

	fmt.Println("Try: http://localhost:8888/one")
	fmt.Println("Try: http://localhost:8888/other")
	fmt.Println("Try: http://localhost:8888/manual?plural=true")

	log.Fatal(http.ListenAndServe(":8888", nil))
}
