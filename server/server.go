package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"starlabs/models"

	"github.com/gorilla/mux"
)

type Server struct {
	Manager *models.SatelliteManager
}

func NewServer(manager *models.SatelliteManager) *Server {
	return &Server{
		Manager: manager,
	}
}

func (s *Server) ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.ServeFile(w, r, "./templates/index.html")
}

func (s *Server) GetSatellites(w http.ResponseWriter, r *http.Request) {
	s.Manager.Mutex.Lock()
	defer s.Manager.Mutex.Unlock()

	// Преобразуем Satellites в SatelliteDTO
	satellitesDTO := make([]models.SatelliteDTO, len(s.Manager.Satellites))
	for i, sat := range s.Manager.Satellites {
		satellitesDTO[i] = models.SatelliteDTO{
			ID:          sat.ID,
			Altitude:    sat.Altitude,
			Theta:       sat.Theta,
			Speed:       sat.Speed,
			Inclination: sat.Inclination,
			PlaneID:     sat.PlaneID,
		}
	}

	data, err := json.Marshal(struct {
		Satellites []models.SatelliteDTO `json:"satellites"`
		Logs       []models.LogEntry     `json:"logs"`
	}{
		Satellites: satellitesDTO,
		Logs:       s.Manager.Logs,
	})
	if err != nil {
		log.Println("JSON marshal failed:", err)
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) UpdateSatellite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 || id > len(s.Manager.Satellites) {
		http.Error(w, "Invalid satellite ID", http.StatusBadRequest)
		return
	}

	var cmd models.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Отправляем команду в канал спутника
	s.Manager.Satellites[id-1].CommandChan <- cmd

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Command sent"})
}
