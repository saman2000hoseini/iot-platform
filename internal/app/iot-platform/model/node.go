package model

import (
	"github.com/saman2000hoseini/iot-platform/internal/pkg/nodestate"
	"gorm.io/gorm"
)

type Node struct {
	ID        string `gorm:"primaryKey" json:"id"`
	IP        string `gorm:"not null" json:"ip"`
	EntryCode string `gorm:"not null" json:"entrycode"`
	Type      int    `gorm:"not null" json:"type"`
	State     int    `gorm:"default:0" json:"state"`
}

func NewNode(id, ip, code string, t int) *Node {
	return &Node{
		ID:        id,
		IP:        ip,
		EntryCode: code,
		Type:      t,
		State:     nodestate.ON,
	}
}

type NodeRepo interface {
	FindByType(t int) ([]Node, error)
	IsValid(id, ip, entrycode string) bool
	Save(node *Node) error
	Update(node Node) error
}

type SQLNodeRepo struct {
	DB *gorm.DB
}

func (r SQLNodeRepo) FindByType(t int) ([]Node, error) {
	var stored []Node
	err := r.DB.Where(&Node{Type: t}).Find(&stored).Error

	return stored, err
}

func (r SQLNodeRepo) IsValid(id, ip, entrycode string) bool {
	var stored Node
	err := r.DB.Where(&Node{ID: id, IP: ip, EntryCode: entrycode}).First(&stored).Error
	if err == nil {
		return true
	}
	return false
}

func (r SQLNodeRepo) Save(node *Node) error {
	return r.DB.Create(node).Error
}

func (r SQLNodeRepo) Update(node Node) error {
	var stored Node
	err := r.DB.Where(&Node{ID: node.ID}).First(&stored).Error
	if err != nil {
		return err
	}

	stored.State = node.State

	return r.DB.Save(stored).Error
}
