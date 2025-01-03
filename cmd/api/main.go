package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/syumai/workers"
)

func main() {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
	workers.Serve(r)
}
