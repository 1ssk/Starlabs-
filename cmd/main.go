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
)

func main() {
	// Настройка логирования
	utils.SetupLogging()
	defer utils.CloseLogFile() // Закроем файл логов при завершении

	// Инициализация спутников
	manager := models.NewSatelliteManager()
	manager.InitializeSatellites()
	manager.StartSimulation()

	// Инициализация сервера
	srv := server.NewServer(manager)

	// Маршруты
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", srv.ServeHome)
	http.HandleFunc("/ws", srv.HandleWebSocket)

	// Грейсфул-шатдаун
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("HTTP server running on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
}
