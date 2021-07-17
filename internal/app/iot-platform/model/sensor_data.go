package model

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type SensorData struct {
	ID       int    `gorm:"primaryKey"`
	SensorID string `gorm:"not null"`
	Value    int    `gorm:"not null"`
	Type     int    `gorm:"not null"`
}

func NewSensorData(sid string, value, t int) *SensorData {
	return &SensorData{
		SensorID: sid,
		Value:    value,
		Type:     t,
	}
}

type SensorDataRepo interface {
	FindLast(t int, ctx context.Context) (SensorData, error)
	Save(data *SensorData, ctx context.Context) error
}

type SQLSensorDataRepo struct {
	DB *gorm.DB
}

func (r SQLSensorDataRepo) FindLast(t int, ctx context.Context) (SensorData, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "find-last-data")
	defer span.Finish()

	var stored SensorData
	err := r.DB.Where("type = ?", t).Last(&stored).Error

	return stored, err
}

func (r SQLSensorDataRepo) Save(data *SensorData, ctx context.Context) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "save-data")
	defer span.Finish()

	return r.DB.Create(data).Error
}
