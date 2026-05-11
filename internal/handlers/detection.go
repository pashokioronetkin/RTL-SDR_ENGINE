package handlers

import (
	"RTL-SDR/engine/internal/storage"
	"RTL-SDR/engine/internal/ws/protocol"
	"context"
	"encoding/json"
	"log"
	"time"
)

type DetectionData struct {
	SessionID       string         `json:"session_id"`
	SessionStart    string         `json:"session_start"`
	TotalDetections int            `json:"total_detections"`
	Detections      []RawDetection `json:"detections"`
	Summary         Summary        `json:"summary"`
}

type RawDetection struct {
	DetectionID    string         `json:"detection_id"`
	Timestamp      string         `json:"timestamp"`
	ObjectInfo     ObjectInfo     `json:"object_info"`
	Location       Location       `json:"location"`
	SignalAnalysis SignalAnalysis `json:"signal_analysis"`
	RiskAssessment string         `json:"risk_assessment"`
	SignalQuality  string         `json:"signal_quality"`
}

type ObjectInfo struct {
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	IsReliable bool    `json:"is_reliable"`
}

type Location struct {
	Azimuth  float64 `json:"azimuth"`
	Range    float64 `json:"range"`
	Altitude float64 `json:"altitude"`
}

type SignalAnalysis struct {
	RMS               float64 `json:"rms"`
	CrestFactor       float64 `json:"crest_factor"`
	DominantFrequency float64 `json:"dominant_frequency"`
}

type Summary struct {
	ObjectDistribution map[string]int `json:"object_distribution"`
	AverageConfidence  float64        `json:"average_confidence"`
	ReliableDetections int            `json:"reliable_detections"`
	ReliabilityRate    float64        `json:"reliability_rate"`
	MostCommonObject   string         `json:"most_common_object"`
}

type Broadcaster interface {
	Broadcast([]byte)
}

func HandleDetection(storageClient *storage.Client, broadcaster Broadcaster) protocol.HandlerFunc {
	return func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
		var data DetectionData
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}

		start := time.Now()
		log.Printf("Gateway: сохраняем session=%s detections=%d в storage", data.SessionID, len(data.Detections))
		inputMap := map[string]interface{}{
			"session_id":       data.SessionID,
			"session_start":    data.SessionStart,
			"total_detections": data.TotalDetections,
			"detections":       data.Detections,
			"summary":          data.Summary,
		}
		if err := storageClient.SaveDetectionData(inputMap); err != nil {
			log.Printf("Gateway: storage error session=%s duration=%s error=%v", data.SessionID, time.Since(start), err)
			return nil, err
		}
		log.Printf("Gateway: storage confirmed session=%s duration=%s", data.SessionID, time.Since(start))

		// Рассылка всем клиентам gateway
		if broadcaster != nil {
			broadcastMsg := map[string]interface{}{
				"type":    "new_detection_data",
				"payload": data,
			}
			if b, err := json.Marshal(broadcastMsg); err == nil {
				broadcaster.Broadcast(b)
			}
		}

		return map[string]string{"status": "ok", "session_id": data.SessionID}, nil
	}
}
