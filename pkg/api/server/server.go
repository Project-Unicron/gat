package server

import (
	"fmt"
	"net/http"
	"time"
)

// Config holds the configuration for the API server
type Config struct {
	Port      int
	Host      string
	ConfigDir string
}

// Server represents the GAT API server
type Server struct {
	config  Config
	mux     *http.ServeMux
	server  *http.Server
	running bool
}

// NewServer creates a new API server with the given configuration
func NewServer(config Config) *Server {
	// Set defaults if not provided
	if config.Port == 0 {
		config.Port = 9999
	}
	if config.Host == "" {
		config.Host = "localhost"
	}

	mux := http.NewServeMux()

	return &Server{
		config: config,
		mux:    mux,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		running: false,
	}
}

// GetServeMux returns the server's HTTP request multiplexer
func (s *Server) GetServeMux() *http.ServeMux {
	return s.mux
}

// RegisterHandler registers a handler for a specific path
func (s *Server) RegisterHandler(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
}

// RegisterHandlerFunc registers a handler function for a specific path
func (s *Server) RegisterHandlerFunc(path string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(path, handler)
}

// Start starts the API server
func (s *Server) Start() error {
	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Add health check endpoint
	s.RegisterHandlerFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	s.running = true
	fmt.Printf("GAT API server started on %s\n", s.server.Addr)
	return nil
}

// Stop stops the API server
func (s *Server) Stop() error {
	if !s.running {
		return fmt.Errorf("server is not running")
	}

	s.running = false
	return s.server.Close()
}
