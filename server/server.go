package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/repo-scm/proxy/config"
	"github.com/repo-scm/proxy/monitor"
)

//go:embed templates/index.html
var templateFS embed.FS

type Server struct {
	config  *config.Config
	monitor *monitor.Monitor
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config:  cfg,
		monitor: monitor.NewMonitor(cfg),
	}
}

func (s *Server) Handler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/ui", s.handleUI).Methods("GET")
	r.HandleFunc("/ui/", s.handleUI).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/status", s.handleAPIStatus).Methods("GET")
	api.HandleFunc("/sites", s.handleAPISites).Methods("GET")
	api.HandleFunc("/sites/{site}/health", s.handleAPISiteHealth).Methods("GET")
	api.HandleFunc("/sites/{site}/queues", s.handleAPISiteQueues).Methods("GET")
	api.HandleFunc("/sites/{site}/connections", s.handleAPISiteConnections).Methods("GET")

	return r
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) handleAPISites(w http.ResponseWriter, r *http.Request) {
	sites := s.monitor.GetAllSitesStatus()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sites)
}

func (s *Server) handleAPISiteHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteName := vars["site"]

	health := s.monitor.GetSiteHealth(siteName)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (s *Server) handleAPISiteQueues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteName := vars["site"]

	queues := s.monitor.GetSiteQueues(siteName)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(queues)
}

func (s *Server) handleAPISiteConnections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteName := vars["site"]

	connections := s.monitor.GetSiteConnections(siteName)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(connections)
}
