package models

import (
	"fmt"
	"math"
	"math/rand"
	"starlabs/utils"
	"sync"
	"time"
)

type Satellite struct {
	ID          int     `json:"id"`
	Altitude    float64 `json:"altitude"`
	Theta       float64 `json:"theta"`
	Speed       float64 `json:"speed"`
	Inclination float64 `json:"inclination"`
	PlaneID     int     `json:"plane_id"`
}

type LogEntry struct {
	Timestamp string    `json:"timestamp"`
	Satellite Satellite `json:"satellite"`
}

type SatelliteManager struct {
	Satellites []Satellite
	Logs       []LogEntry
	Mutex      *sync.Mutex
	Broadcast  chan struct{}
}

func NewSatelliteManager() *SatelliteManager {
	return &SatelliteManager{
		Satellites: make([]Satellite, 50),
		Logs:       make([]LogEntry, 0),
		Mutex:      &sync.Mutex{},
		Broadcast:  make(chan struct{}),
	}
}

func (sm *SatelliteManager) InitializeSatellites() {
	rand.Seed(time.Now().UnixNano())
	numPlanes := 5
	planes := make([]struct {
		Inclination float64
		Altitude    float64
	}, numPlanes)
	for i := 0; i < numPlanes; i++ {
		planes[i] = struct {
			Inclination float64
			Altitude    float64
		}{
			Inclination: rand.Float64()*math.Pi/2 - math.Pi/4,
			Altitude:    550 + float64(rand.Intn(50)),
		}
	}

	for i := 0; i < 50; i++ {
		planeID := i % numPlanes
		sm.Satellites[i] = Satellite{
			ID:          i + 1,
			Altitude:    planes[planeID].Altitude,
			Theta:       rand.Float64() * 2 * math.Pi,
			Speed:       0.005 + rand.Float64()*0.01,
			Inclination: planes[planeID].Inclination,
			PlaneID:     planeID,
		}
	}
}

func (sm *SatelliteManager) StartSimulation() {
	logChan := make(chan LogEntry, 100)
	for i := 0; i < len(sm.Satellites); i++ {
		go sm.updateSatellitePosition(i, logChan)
	}

	go func() {
		for {
			select {
			case logEntry := <-logChan:
				sm.Mutex.Lock()
				if len(sm.Logs) > 20 {
					sm.Logs = sm.Logs[1:]
				}
				sm.Logs = append(sm.Logs, logEntry)

				// Запись лога в файл
				logText := logEntry.Timestamp +
					" - Sat " +
					string(logEntry.Satellite.ID) +
					": Theta=" +
					fmt.Sprintf("%.2f°", logEntry.Satellite.Theta*180/math.Pi)
				utils.WriteLog(logText)

				sm.Mutex.Unlock()
				sm.Broadcast <- struct{}{}
			}
		}
	}()
}

func (sm *SatelliteManager) updateSatellitePosition(index int, logChan chan LogEntry) {
	for {
		sat := &sm.Satellites[index]
		sat.Theta += sat.Speed
		if sat.Theta > 2*math.Pi {
			sat.Theta -= 2 * math.Pi
		}
		logChan <- LogEntry{
			Timestamp: time.Now().Format("15:04:05"),
			Satellite: *sat,
		}
		time.Sleep(500 * time.Millisecond)
	}
}
