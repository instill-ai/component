package asana

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router(middlewares ...func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	for _, m := range middlewares {
		r.Use(m)
	}

	r.Get("/goals/{goalGID}", getGoal)
	r.Put("/goals/{goalGID}", updateGoal)
	r.Delete("/goals/{goalGID}", deleteGoal)
	r.Post("/goals", createGoal)
	return r
}
