package model

import "gorm.io/gorm"

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
	Find(id string) (SensorData, error)
	Save(data *SensorData) error
}

type SQLSensorDataRepo struct {
	DB *gorm.DB
}

func (r SQLSensorDataRepo) FindLast(t int) (SensorData, error) {
	var stored SensorData
	err := r.DB.Where("type = ?", t).Last(&stored).Error

	return stored, err
}

func (r SQLSensorDataRepo) Save(data *SensorData) error {
	return r.DB.Create(data).Error
}
