package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/lovelaze/nebula-sync/internal/sync"
)

const (
	port              = 8080
	readHeaderTimeout = 10 * time.Second
)

type Server struct {
	state  *sync.State
	router *chi.Mux
}

func NewServer(state *sync.State) *Server {
	router := chi.NewRouter()
	server := &Server{
		state:  state,
		router: router,
	}

	router.Get("/health", server.healthHandler)

	return server
}

func (s *Server) Start() {
	go func() {
		log.Debug().Msg("Starting http server")

		server := &http.Server{
			Handler:           s.router,
			Addr:              fmt.Sprintf(":%d", port),
			ReadHeaderTimeout: readHeaderTimeout,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start http server")
		}
	}()
}
