package server

import (
	"encoding/json"
	"log"
	"net/http"

	"starlabs/models"

	"github.com/gorilla/websocket"
)

type Server struct {
	Manager *models.SatelliteManager
	Clients map[*websocket.Conn]bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewServer(manager *models.SatelliteManager) *Server {
	return &Server{
		Manager: manager,
		Clients: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	s.Manager.Mutex.Lock()
	s.Clients[conn] = true
	s.Manager.Mutex.Unlock()

	for range s.Manager.Broadcast {
		data, err := json.Marshal(struct {
			Satellites []models.Satellite `json:"satellites"`
			Logs       []models.LogEntry  `json:"logs"`
		}{
			Satellites: s.Manager.Satellites,
			Logs:       s.Manager.Logs,
		})
		if err != nil {
			log.Println("JSON marshal failed:", err)
			continue
		}
		s.Manager.Mutex.Lock()
		for client := range s.Clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				client.Close()
				delete(s.Clients, client)
			}
		}
		s.Manager.Mutex.Unlock()
	}
}

func (s *Server) ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.ServeFile(w, r, "index.html")
}
