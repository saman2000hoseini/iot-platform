package model

import (
	"context"
	"github.com/opentracing/opentracing-go"
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
	Find(t int, ctx context.Context) (SensorThreshold, error)
	Save(threshold *SensorThreshold, ctx context.Context) error
}

type SQLSensorThresholdRepo struct {
	DB *gorm.DB
}

func (r SQLSensorThresholdRepo) Find(t int, ctx context.Context) (SensorThreshold, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "find-threshold")
	defer span.Finish()

	var stored SensorThreshold
	err := r.DB.Where(&SensorThreshold{Type: t}).First(&stored).Error

	return stored, err
}

func (r SQLSensorThresholdRepo) Save(threshold *SensorThreshold, ctx context.Context) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "save-threshold")
	defer span.Finish()

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
