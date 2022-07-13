package web

import (
	"coinbit-test/core/service"
	"coinbit-test/lib/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func API(h *handler.Handler, router chi.Router) {
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("pong"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	router.Route("/api", func(r chi.Router) {
		r.Post("/deposit", handler.HandlerFuncWrapper(h, service.Deposit))
		r.Get("/details", handler.HandlerFuncWrapper(h, service.Details))
	})
}
