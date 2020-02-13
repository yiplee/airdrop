package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/yiplee/airdrop/handler/pb"
)

func Handle(twirpHandler http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Route("/tasks", func(r chi.Router) {
		r.HandleFunc("/", Twirp(twirpHandler, TwirpOpts{
			PathPrefix: pb.TaskServicePathPrefix,
			Method:     "Create",
		}))

		r.HandleFunc("/{trace_id}", Twirp(twirpHandler, TwirpOpts{
			PathPrefix: pb.TaskServicePathPrefix,
			Method:     "Find",
		}))
	})

	return r
}
