package controllers

import (
	"fmt"
	"gorm/initializers"
	"gorm/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DepositInBank(c *gin.Context) {

	var body struct {
		UserID        uint `json:"user_id"`
		DepositAmount uint `json:"deposit_amount"`
	}

	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	bank := models.Bank{
		UserID:         body.UserID,
		CurrentBalance: 0,
		LimitBalance:   5000,
	}
	result := initializers.DB.Debug().FirstOrCreate(&bank, bank.UserID)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unable to create or retrieve bank data for user",
		})
		return
	}
	fmt.Println(bank.UserID)
	fmt.Println(bank.CurrentBalance)
	fmt.Println(bank.LimitBalance)

	type User struct {
		ID          uint   `gorm:"column:id;primaryKey" json:"user_id"`
		Satoshis    uint `json:"satoshis"`
	}

	var user User
	result = initializers.DB.Debug().Select("id, satoshis").Where("id = ?", body.UserID).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to get user satoshis",
		})
		return
	}
	fmt.Println(&user)
	if user.Satoshis >= body.DepositAmount {
		user.Satoshis -= body.DepositAmount
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Deposit amount cannot be higher than available satoshis",
		})
		return
	}

	result = initializers.DB.Debug().Model(&user).Updates(map[string]interface{}{"satoshis": user.Satoshis})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to update user satoshis",
		})
		return
	}

	bank.CurrentBalance += body.DepositAmount
	//add where clause if model doesnt have a primary key
	result = initializers.DB.Debug().Model(&bank).Where("user_id = ?", bank.UserID).Updates(map[string]interface{}{"current_balance": bank.CurrentBalance})
	if result.Error != nil {
		fmt.Println(result.Error.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unable to update for deposit amount",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deposited in bank sucessfully",
	})
}

func WithdrawFromBankToUser(c *gin.Context){
	var body struct {
		UserID        uint `json:"user_id"`
		WithdrawAmount uint `json:"withdraw_amount"`
	}

	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	
	bank := models.Bank{}

	result := initializers.DB.Where("user_id = ?", body.UserID).First(&bank)
	fmt.Println(bank)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to get balance from bank",
		})
		return
	}

	if body.WithdrawAmount > bank.CurrentBalance {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Insufficient satoshi amount to withdraw",
		})
		return
	}

	bank.CurrentBalance-= body.WithdrawAmount
	result = initializers.DB.Debug().Model(&bank).Where("user_id = ?", bank.UserID).Updates(map[string]interface{}{"current_balance": bank.CurrentBalance})
	
	if body.WithdrawAmount > bank.CurrentBalance {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to update bank current balance",
		})
		return
	}

	type User struct {
		ID          uint   `gorm:"column:id;primaryKey" json:"user_id"`
		Satoshis    uint `json:"satoshis"`
	}

	var user User
	result = initializers.DB.Debug().Select("id, satoshis").Where("id = ?", body.UserID).First(&user)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to get user satoshis",
		})
		return
	}

	user.Satoshis += body.WithdrawAmount

	result = initializers.DB.Debug().Model(&user).Updates(map[string]interface{}{"satoshis": user.Satoshis})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to update user satoshis",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Withdrawn from bank sucessfully",
	})
}