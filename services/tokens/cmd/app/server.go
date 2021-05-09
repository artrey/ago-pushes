package app

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"tokens/cmd/app/dto"
	"tokens/pkg/tokens"
)

type Server struct {
	tokensSvc *tokens.Service
	mux       chi.Router
}

func NewServer(tokensSvc *tokens.Service, mux chi.Router) *Server {
	return &Server{tokensSvc: tokensSvc, mux: mux}
}

func (s *Server) Init() error {
	s.mux.Use(middleware.Logger)

	s.mux.Route("/api", func(r chi.Router) {
		r.Route("/tokens", func(tokensRouter chi.Router) {
			tokensRouter.Post("/register", s.register)
		})
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) register(writer http.ResponseWriter, request *http.Request) {
	var data dto.RegisterToken
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.tokensSvc.Register(request.Context(), data.UserID, data.PushToken)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
}
