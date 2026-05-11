package repository

import (
	"RTL-SDR/engine/internal/models"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.Session{}, &models.Detection{})
}

func (r *Repository) SaveSessionWithDetections(session models.Session, detections []models.Detection) error {
	start := time.Now()
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&session).Error; err != nil {
			return err
		}
		for _, det := range detections {
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&det).Error; err != nil {
				return err
			}
		}
		return nil
	})
	duration := time.Since(start)
	ids := make([]string, 0, len(detections))
	for i, det := range detections {
		if i >= 5 {
			break
		}
		ids = append(ids, det.ID)
	}
	idList := ""
	if len(ids) > 0 {
		idList = ids[0]
		for _, id := range ids[1:] {
			idList += ", " + id
		}
		if len(detections) > len(ids) {
			idList += ", ..."
		}
	}
	if err != nil {
		log.Printf("Repository: failed save session=%s detections=%d sample_detection_ids=[%s] duration=%s error=%v", session.ID, len(detections), idList, duration, err)
		return err
	}
	log.Printf("Repository: saved session=%s detections=%d sample_detection_ids=[%s] duration=%s", session.ID, len(detections), idList, duration)
	return nil
}

func (r *Repository) GetHistory(limit int) ([]models.Session, error) {
	start := time.Now()
	var sessions []models.Session
	if err := r.db.Preload("Detections").Order("session_start desc").Limit(limit).Find(&sessions).Error; err != nil {
		log.Printf("Repository: failed history query limit=%d duration=%s error=%v", limit, time.Since(start), err)
		return nil, err
	}
	log.Printf("Repository: history query returned=%d limit=%d duration=%s", len(sessions), limit, time.Since(start))
	return sessions, nil
}
