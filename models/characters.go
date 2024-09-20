package models

import ()



type Character struct {
    ID uint   `gorm:"column:boardroom_id;primaryKey" json:"boardroom_id"`
    IsActive    bool   `json:"is_active"`
    Level       int    `json:"level"`
    Attack      int    `json:"attack"`
    Defense     int    `json:"defense"`
    Name        string `json:"name"`
    
}


type UsersCharacters struct {
	UserID      uint `gorm:"primaryKey"`
	CharacterID uint `gorm:"primaryKey"`
}

/*
type UserCharacter struct {
    UserID     uint `gorm:"column:user_id;primaryKey"`
    CharacterID uint `gorm:"column:character_id;primaryKey"`
}
*/