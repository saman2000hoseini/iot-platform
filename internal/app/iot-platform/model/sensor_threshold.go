package model

import (
	"gorm.io/gorm"
)

type SensorThreshold struct {
	ID        int `gorm:"primaryKey"`
	Threshold int `gorm:"not null"`
	Type      int `gorm:"not null;unique"`
}

func NewSensorThreshold(threshold, t int) *SensorThreshold {
	return &SensorThreshold{
		Threshold: threshold,
		Type:      t,
	}
}

type SensorThresholdRepo interface {
	Find(t int) (SensorThreshold, error)
	Save(threshold *SensorThreshold) error
}

type SQLSensorThresholdRepo struct {
	DB *gorm.DB
}

func (r SQLSensorThresholdRepo) Find(t int) (SensorThreshold, error) {
	var stored SensorThreshold
	err := r.DB.Where(&SensorThreshold{Type: t}).First(&stored).Error

	return stored, err
}

func (r SQLSensorThresholdRepo) Save(threshold *SensorThreshold) error {
	if err := r.DB.Create(threshold).Error; err != nil {
		var stored SensorThreshold
		err := r.DB.Where(&SensorThreshold{Type: threshold.Type}).First(&stored).Error
		if err != nil {
			return err
		}

		threshold.ID = stored.ID

		return r.DB.Save(threshold).Error
	}

	return nil
}
