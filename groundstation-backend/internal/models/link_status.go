package models

import (
	"time"

	"gorm.io/gorm"
)

type LinkType uint8

const (
	LinkTypeRadio LinkType = 1
	LinkTypeLTE   LinkType = 2
	LinkTypeDual  LinkType = 3
)

type LinkState uint8

const (
	LinkStateDisconnected LinkState = 0
	LinkStateConnecting   LinkState = 1
	LinkStateConnected    LinkState = 2
	LinkStateDegraded     LinkState = 3
)

type LinkQuality struct {
	RSSI       int8    `json:"rssi"`
	SNR        float64 `json:"snr"`
	PacketLoss float64 `json:"packet_loss"`
	LatencyMs  uint32  `json:"latency_ms"`
}

type LinkStatus struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64         `gorm:"index;not null" json:"uav_id"`
	ActiveLink      LinkType       `gorm:"not null;default:1" json:"active_link"`
	RadioRSSI       int8           `json:"radio_rssi"`
	RadioState      LinkState      `json:"radio_state"`
	RadioConnected  bool           `json:"radio_connected"`
	LteRSSI         int8           `json:"lte_rssi"`
	LteState        LinkState      `json:"lte_state"`
	LteConnected    bool           `json:"lte_connected"`
	LteNetworkType  string         `gorm:"type:varchar(20)" json:"lte_network_type"`
	PacketLoss      float64        `json:"packet_loss"`
	LatencyMs       uint32         `json:"latency_ms"`
	BytesSent       uint64         `json:"bytes_sent"`
	BytesReceived   uint64         `json:"bytes_received"`
	Timestamp       time.Time      `gorm:"index;not null" json:"timestamp"`
	AutoSwitchEnabled bool         `json:"auto_switch_enabled"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (LinkStatus) TableName() string {
	return "link_status"
}
