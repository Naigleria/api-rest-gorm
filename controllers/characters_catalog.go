package controllers

import (
	"gorm/initializers"
	"gorm/models"

	"github.com/gin-gonic/gin"
)

func GetCharactersCatalog(c *gin.Context) {
    
    var characters_catalog []models.CharactersCatalog
    if err := initializers.DB.Preload("Bonuses").Find(&characters_catalog).Error; err != nil {
        c.JSON(500, gin.H{"error": "Error fetching characters"})
        return
    }

    c.JSON(200, characters_catalog)
}

