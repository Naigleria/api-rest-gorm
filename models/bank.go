package models

import (
	//"gorm.io/gorm"
)


type Bank struct {
    UserID         uint `gorm:"foreignKey:user_id"`
    CurrentBalance uint `gorm:"not null;default:0"`
    LimitBalance   uint `gorm:"not null;default:0"`
}

func (Bank) TableName() string {
	return "bank"
}