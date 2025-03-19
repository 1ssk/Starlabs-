package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	logFile  *os.File
	logMutex sync.Mutex
)

// Настройка логирования
func SetupLogging() {
	if err := os.MkdirAll("log", os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logFileName := fmt.Sprintf("log/server-%s.log", time.Now().Format("2006-01-02"))
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.Println("Logging initialized")

	// Запуск горутины, которая будет записывать лог каждые 30 секунд
	go logEvery30Seconds()
}

// Горутина для логирования каждые 30 секунд
func logEvery30Seconds() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			WriteLog("Logging every 30 seconds...")
		}
	}
}

// Безопасная запись в лог
func WriteLog(entry string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		_, err := logFile.WriteString(entry + "\n")
		if err != nil {
			log.Printf("Failed to write to log file: %v", err)
		}
	}
}

// Закрытие файла логов при завершении
func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}
