package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/twitchtv/twirp"
	"github.com/yiplee/airdrop/handler/api"
	"github.com/yiplee/airdrop/handler/pb"
	taskhandler "github.com/yiplee/airdrop/handler/task"
	"github.com/yiplee/airdrop/pkg/task"
)

type Config struct {
	TargetLimit int    `valid:"required"`
	BrokerID    string `valid:"uuid"`
	Debug       bool
}

func Route(tasks task.Store, cfg Config) http.Handler {
	if _, err := govalidator.ValidateStruct(cfg); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.AllowAll().Handler)
	r.Use(middleware.Logger)

	rpcHandler := r.Group(func(r chi.Router) {
		hook := &twirp.ServerHooks{}
		servers := []pb.TwirpServer{
			// task service
			pb.NewTaskServiceServer(
				taskhandler.New(tasks, cfg.TargetLimit, cfg.BrokerID),
				hook,
			),
		}

		for _, server := range servers {
			r.Mount(server.PathPrefix(), server)
		}
	})

	r.Mount("/twirp", rpcHandler)
	r.Mount("/api", api.Handle(rpcHandler))

	if cfg.Debug {
		r.Mount("/debug", middleware.Profiler())
		api.ResponseErrorMessageAsHint = true
	}

	return r
}
