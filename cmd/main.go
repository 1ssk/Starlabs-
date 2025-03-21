package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"starlabs/models"
	"starlabs/server"
	"starlabs/utils"

	"github.com/gorilla/mux"
)

func main() {
	// Настройка логирования
	utils.SetupLogging()
	defer utils.CloseLogFile()

	// Инициализация спутников
	manager := models.NewSatelliteManager()
	manager.InitializeSatellites()
	manager.StartSimulation()

	// Инициализация сервера
	srv := server.NewServer(manager)

	// Настройка маршрутов с gorilla/mux
	router := mux.NewRouter()
	router.HandleFunc("/", srv.ServeHome).Methods("GET")
	router.HandleFunc("/api/satellites", srv.GetSatellites).Methods("GET") // Новый маршрут для данных
	router.HandleFunc("/api/satellite/{id}", srv.UpdateSatellite).Methods("POST")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Настройка HTTP-сервера
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Грейсфул-шатдаун
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("HTTP server running on port 8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	httpServer.Close()
}
