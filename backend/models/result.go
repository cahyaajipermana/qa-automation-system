package models

import (
	"time"
)

// Result represents a test result
type Result struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SiteID    uint      `json:"site_id" gorm:"not null"`
	DeviceID  uint      `json:"device_id" gorm:"not null"`
	FeatureID uint      `json:"feature_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"type:enum('processing','passed','failed','warning');not null"`
	Browser   string    `json:"browser" gorm:"type:varchar(255);null"`
	Location  string    `json:"location" gorm:"type:varchar(255);null"`
	Screenshot string    `json:"screenshot" gorm:"type:varchar(255);null"`
	ErrorLog  string    `json:"error_log" gorm:"type:varchar(255);null"`
	Duration  float64   `json:"duration" gorm:"type:float;null"`
	VideoPath string    `json:"video_path" gorm:"type:varchar(255);null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Site      Site      `json:"site" gorm:"foreignKey:SiteID"`
	Device    Device    `json:"device" gorm:"foreignKey:DeviceID"`
	Feature   Feature   `json:"feature" gorm:"foreignKey:FeatureID"`
	Details   []ResultDetail `json:"details" gorm:"foreignKey:ResultID"`
}

// ResultDetail represents detailed information about a test result
type ResultDetail struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ResultID    uint      `json:"result_id" gorm:"not null"`
	Screenshot  string    `json:"screenshot" gorm:"type:varchar(255);null"`
	Description string    `json:"description" gorm:"type:text;null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Result      Result    `json:"result" gorm:"foreignKey:ResultID"`
} 