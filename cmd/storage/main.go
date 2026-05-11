package main

import (
	"RTL-SDR/engine/internal/models"
	"RTL-SDR/engine/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var repo *repository.Repository

func main() {
	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("Не удалось загрузить .env: %v", err)
	}

	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "root")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	database := getEnv("DB_NAME", "RTL_SDR_DB")
	storagePort := getEnv("STORAGE_PORT", "8081")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, database)
	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}

	repo = repository.New(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	http.HandleFunc("/api/detections", saveDetectionHandler)
	http.HandleFunc("/api/history", getHistoryHandler)

	log.Printf("Storage API запущен на :%s", storagePort)
	if err := http.ListenAndServe(":"+storagePort, nil); err != nil {
		log.Fatal(err)
	}
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if key == "" {
			continue
		}

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// saveDetectionHandler сохраняет сессию и детекции
func saveDetectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		SessionID       string `json:"session_id"`
		SessionStart    string `json:"session_start"`
		TotalDetections int    `json:"total_detections"`
		Summary         struct {
			AverageConfidence float64 `json:"average_confidence"`
			ReliabilityRate   float64 `json:"reliability_rate"`
			MostCommonObject  string  `json:"most_common_object"`
		} `json:"summary"`
		Detections []struct {
			DetectionID string `json:"detection_id"`
			Timestamp   string `json:"timestamp"`
			ObjectInfo  struct {
				Type       string  `json:"type"`
				Confidence float64 `json:"confidence"`
				IsReliable bool    `json:"is_reliable"`
			} `json:"object_info"`
			Location struct {
				Azimuth  float64 `json:"azimuth"`
				Range    float64 `json:"range"`
				Altitude float64 `json:"altitude"`
			} `json:"location"`
			SignalAnalysis struct {
				RMS               float64 `json:"rms"`
				CrestFactor       float64 `json:"crest_factor"`
				DominantFrequency float64 `json:"dominant_frequency"`
			} `json:"signal_analysis"`
			RiskAssessment string `json:"risk_assessment"`
			SignalQuality  string `json:"signal_quality"`
		} `json:"detections"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Storage: receive POST /api/detections session=%s detections=%d", input.SessionID, len(input.Detections))

	startTime, err := time.Parse(time.RFC3339, input.SessionStart)
	if err != nil {
		http.Error(w, "invalid session_start", http.StatusBadRequest)
		return
	}

	session := models.Session{
		ID:                input.SessionID,
		Start:             startTime,
		TotalDetections:   input.TotalDetections,
		AverageConfidence: input.Summary.AverageConfidence,
		ReliabilityRate:   input.Summary.ReliabilityRate,
		MostCommonObject:  input.Summary.MostCommonObject,
	}

	var detections []models.Detection
	for _, d := range input.Detections {
		ts, err := time.Parse(time.RFC3339, d.Timestamp)
		if err != nil {
			http.Error(w, "invalid detection timestamp", http.StatusBadRequest)
			return
		}
		detections = append(detections, models.Detection{
			ID:                d.DetectionID,
			SessionID:         input.SessionID,
			Timestamp:         ts,
			ObjectType:        d.ObjectInfo.Type,
			Confidence:        d.ObjectInfo.Confidence,
			IsReliable:        d.ObjectInfo.IsReliable,
			LocationAzimuth:   d.Location.Azimuth,
			LocationRange:     d.Location.Range,
			LocationAltitude:  d.Location.Altitude,
			SignalRMS:         d.SignalAnalysis.RMS,
			SignalCrestFactor: d.SignalAnalysis.CrestFactor,
			DominantFrequency: d.SignalAnalysis.DominantFrequency,
			RiskAssessment:    d.RiskAssessment,
			SignalQuality:     d.SignalQuality,
		})
	}

	err = repo.SaveSessionWithDetections(session, detections)
	if err != nil {
		log.Printf("Storage: failed to save session=%s error=%v", input.SessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// getHistoryHandler возвращает последние сессии (количество = limit)
func getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	log.Printf("Storage: receive GET /api/history limit=%d", limit)

	sessions, err := repo.GetHistory(limit)
	if err != nil {
		log.Printf("Storage: failed history query limit=%d error=%v", limit, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Storage: returned %d sessions", len(sessions))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}
