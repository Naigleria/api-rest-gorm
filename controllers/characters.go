package controllers

import (
	"fmt"
	"gorm/initializers"
	"gorm/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ChooseFirstCharacterOrBuyCharacter(c *gin.Context) {
	var body struct {
		UserID             uint `json:"user_id"`
		CharacterCatalogID uint `json:"character_catalog_id"`
		IsBuy              bool `json:"is_buy"`
	}

	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	character_name := ""

	if body.CharacterCatalogID == 1 {
		character_name = "Warrior"
	} else if body.CharacterCatalogID == 2 {
		character_name = "Paladin"
	} else if body.CharacterCatalogID == 3 {
		character_name = "Merchand"
	} else if body.CharacterCatalogID == 4 {
		character_name = "Guardian"
	} else if body.CharacterCatalogID == 5 {
		character_name = "Berserker"
	} else if body.CharacterCatalogID == 6 {
		character_name = "Thief"
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid character catalog ID",
		})
		return
	}

	is_active:=true

	if body.IsBuy{
		is_active=false
	}

	character := models.Character{
		IsActive: is_active,
		Level:    1,
		Attack:   10,
		Defense:  10,
		Name:     character_name,
	}

	fmt.Println(body.UserID)
	fmt.Println(body.CharacterCatalogID)

	if err := initializers.DB.Create(&character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save your choosed character 1",
		})
		return
	}

	usersCharacters := models.UsersCharacters{
		UserID:      body.UserID,
		CharacterID: character.ID,
	}
	fmt.Println(character.ID)

	if err := initializers.DB.Create(&usersCharacters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save your choosed character 2",
		})
		return
	}
	

	if body.IsBuy {

		type User struct {
			ID       uint `gorm:"column:id;primaryKey" json:"user_id"`
			Satoshis uint `json:"satoshis"`
		}

		var user User
		result := initializers.DB.Debug().Select("id, satoshis").Where("id = ?", body.UserID).First(&user)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unable to get user satoshis",
			})
			return
		}

		type CharactersCatalog struct {
			ID             uint   `gorm:"column:characters_catalog_id;primaryKey" json:"characters_catalog_id"`
			CostInSatoshis uint    `json:"cost_in_satoshis"`
		}

		

		var characterCatalog CharactersCatalog
		result = initializers.DB.Debug().Table("characters_catalog").Select("characters_catalog_id, cost_in_satoshis").Where("characters_catalog_id = ?", body.CharacterCatalogID).First(&characterCatalog)
		
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unable to get cost in satoshis of character in catalog",
			})
			return
		}


		if user.Satoshis >= characterCatalog.CostInSatoshis {
			fmt.Println("haciendo resta")
			fmt.Println(user.Satoshis)
			fmt.Println(characterCatalog.CostInSatoshis)
			user.Satoshis -= characterCatalog.CostInSatoshis
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficent satoshis to buy character",
			})
			return
		}
		result = initializers.DB.Debug().Model(&user).Updates(map[string]interface{}{"satoshis": user.Satoshis})

		//aqui lo que sigue es mandar el costo en satoshis a la wallet del admin

		//final response
		c.JSON(http.StatusCreated, gin.H{
			"message": "Character bought sucessfully",
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Character choosed sucessfully",
		})
	}
}
