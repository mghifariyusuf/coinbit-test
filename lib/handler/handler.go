package handler

import (
	"net/http"

	"github.com/lovoo/goka"
	"github.com/lovoo/goka/storage"
)

type Handler struct {
	Emitter *goka.Emitter
	Storage storage.Storage
}

type HandlerFunc func(*Handler, http.ResponseWriter, *http.Request)

func HandlerFuncWrapper(h *Handler, f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(h, w, r)
	}
}
