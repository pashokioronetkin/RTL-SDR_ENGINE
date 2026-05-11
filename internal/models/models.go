package models

import "time"

type Session struct {
	ID                string      `json:"session_id" gorm:"primaryKey;column:session_id"`
	Start             time.Time   `json:"session_start" gorm:"column:session_start"`
	TotalDetections   int         `json:"total_detections"`
	AverageConfidence float64     `json:"average_confidence"`
	ReliabilityRate   float64     `json:"reliability_rate"`
	MostCommonObject  string      `json:"most_common_object"`
	CreatedAt         time.Time   `json:"created_at"`
	Detections        []Detection `json:"detections,omitempty" gorm:"foreignKey:SessionID"`
}

type Detection struct {
	ID                string    `json:"detection_id" gorm:"primaryKey;column:detection_id"`
	SessionID         string    `json:"session_id" gorm:"index"`
	Timestamp         time.Time `json:"timestamp"`
	ObjectType        string    `json:"object_type" gorm:"column:object_type"`
	Confidence        float64   `json:"confidence"`
	IsReliable        bool      `json:"is_reliable" gorm:"column:is_reliable"`
	LocationAzimuth   float64   `json:"location_azimuth" gorm:"column:location_azimuth"`
	LocationRange     float64   `json:"location_range" gorm:"column:location_range"`
	LocationAltitude  float64   `json:"location_altitude" gorm:"column:location_altitude"`
	SignalRMS         float64   `json:"signal_rms" gorm:"column:signal_rms"`
	SignalCrestFactor float64   `json:"signal_crest_factor" gorm:"column:signal_crest_factor"`
	DominantFrequency float64   `json:"dominant_frequency" gorm:"column:dominant_frequency"`
	RiskAssessment    string    `json:"risk_assessment" gorm:"column:risk_assessment"`
	SignalQuality     string    `json:"signal_quality" gorm:"column:signal_quality"`
	CreatedAt         time.Time `json:"created_at"`
}

func (Session) TableName() string   { return "sessions" }
func (Detection) TableName() string { return "detections" }
