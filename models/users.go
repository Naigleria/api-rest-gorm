package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username          string `gorm:"unique"`
	Email             string `gorm:"unique"`
	Password          string
	VerificationToken string
	IsVerified        bool
	BattleStats       string
	SafetyTime        string
	Satoshis          uint
	ActiveCharacter   []Character `gorm:"many2many:users_characters;"`
}

type Users []User


//API REPONSES, DONT MIGRATE TO DB
type UserResponse struct {
	ID          uint   `gorm:"column:id;primaryKey" json:"user_id"`
	Username    string `json:"username"`
	BattleStats string `json:"battle_stats"`
	SafetyTime  string `json:"safety_time"`
	Satoshis    uint   `json:"satoshis"`
	ActiveCharacter []Character `gorm:"many2many:users_characters;foreignKey:id;joinForeignKey:user_id;References:id;joinReferences:character_id" json:"active_characters"`
	
}

func (UserResponse) TableName() string {
	return "users"
}
