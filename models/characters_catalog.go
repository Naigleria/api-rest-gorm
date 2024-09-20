package models

import ()

type CharactersCatalog struct {
    ID             uint   `gorm:"column:characters_catalog_id;primaryKey" json:"characters_catalog_id"`
    Name           string `json:"name"`
    CostInSatoshis uint    `json:"cost_in_satoshis"` 
	Bonuses        []Bonus `gorm:"many2many:characters_catalog_bonuses;" json:"bonuses"` 
	
}

func (CharactersCatalog) TableName() string {
    return "characters_catalog"
}

type Bonus struct {
    ID    uint `gorm:"column:bonus_id;primaryKey" json:"bonus_id"`
    Bonus string    `json:"bonus"`
}

func (Bonus) TableName() string {
    return "bonuses"
}

type CharactersCatalogBonuses struct {
    BonusID     uint `gorm:"primaryKey"`
    CharacterCatalogID uint `gorm:"primaryKey"`
}