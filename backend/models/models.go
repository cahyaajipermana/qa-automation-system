package models

import (
	"time"
)

// Base model with common fields
type Base struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Site represents a website to be tested
type Site struct {
	Base
	Name string `json:"name" gorm:"unique;not null"`
}

// Device represents a device for testing
type Device struct {
	Base
	Name string `json:"name" gorm:"unique;not null"`
}

// Feature represents a test feature
type Feature struct {
	Base
	Name string `json:"name" gorm:"unique;not null"`
} 