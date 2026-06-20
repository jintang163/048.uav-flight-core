package models

import (
	"time"

	"gorm.io/gorm"
)

type UnlockStatus string

const (
	UnlockStatusPending  UnlockStatus = "pending"
	UnlockStatusApproved UnlockStatus = "approved"
	UnlockStatusRejected UnlockStatus = "rejected"
	UnlockStatusExpired  UnlockStatus = "expired"
	UnlockStatusCancelled UnlockStatus = "cancelled"
)

type TemporaryUnlocking struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID         string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID        uint64         `gorm:"index;not null" json:"uav_id"`
	GeofenceID   uint64         `gorm:"index" json:"geofence_id"`
	ApplicantID  uint64         `gorm:"index;not null" json:"applicant_id"`
	ApproverID   uint64         `gorm:"index" json:"approver_id"`
	Title        string         `gorm:"type:varchar(200);not null" json:"title"`
	Reason       string         `gorm:"type:text;not null" json:"reason"`
	Status       UnlockStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Category     GeofenceCategory `gorm:"type:varchar(30);default:'custom'" json:"category"`
	UnlockType   string         `gorm:"type:varchar(20);default:'temporary'" json:"unlock_type"`
	StartTime    *time.Time     `json:"start_time"`
	EndTime      *time.Time     `json:"end_time"`
	MaxAltitude  float64        `gorm:"type:decimal(8,2)" json:"max_altitude"`
	MaxDistance  float64        `gorm:"type:decimal(10,2)" json:"max_distance"`
	CenterLat    float64        `gorm:"type:decimal(10,7)" json:"center_lat"`
	CenterLng    float64        `gorm:"type:decimal(10,7)" json:"center_lng"`
	Radius       float64        `gorm:"type:decimal(10,2)" json:"radius"`
	ApprovalRemark string       `gorm:"type:text" json:"approval_remark"`
	ApprovedAt   *time.Time     `json:"approved_at"`
	CancelledAt  *time.Time     `json:"cancelled_at"`
	MissionID    uint64         `gorm:"index" json:"mission_id"`
	ContactName  string         `gorm:"type:varchar(50)" json:"contact_name"`
	ContactPhone string         `gorm:"type:varchar(20)" json:"contact_phone"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	UAV        *UAV             `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Geofence   *Geofence        `gorm:"foreignKey:GeofenceID" json:"geofence,omitempty"`
	Applicant  *User            `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	Approver   *User            `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}

func (TemporaryUnlocking) TableName() string {
	return "temporary_unlockings"
}
