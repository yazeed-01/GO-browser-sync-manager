package models

import (
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	Name     string
	URL      string
	Category string
}

type Email struct {
	gorm.Model
	Email string
}

type Extension struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `json:"name"`
}
