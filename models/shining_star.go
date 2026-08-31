package models

import (
	"time"
)

type AchievementType string

const (
	TypeTaskMaster       AchievementType = "Task Master"
	TypeConsistencyStar  AchievementType = "Consistency Star"
	TypeSpeedPerformer   AchievementType = "Speed Performer"
	TypeDeadlineChampion AchievementType = "Deadline Champion"
	TypeProductivityPro  AchievementType = "Productivity Pro"
	Type1TripAround      AchievementType = "1 trip around"
)

type ShiningStar struct {
	ID             uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint            `gorm:"not null" json:"user_id"`
	AdminID        uint            `json:"admin_id"`
	Type           AchievementType `gorm:"type:varchar(50);not null" json:"type"`
	Description    string          `gorm:"type:text" json:"description"`
	CertificateURL string          `gorm:"type:varchar(255)" json:"certificate_url"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
