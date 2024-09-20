package main

import (
	"gorm/controllers"
	"gorm/initializers"
	"gorm/middleware"

	//"gorm/models"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {

	//endpoint

	/*
		router.HandleFunc("/api/user/", controllers.GetUsers).Methods("GET")
		router.HandleFunc("/api/user/{id: [0-9]+}", controllers.GetUser).Methods("GET")
		router.HandleFunc("/api/user/signup", controllers.Signup).Methods("POST")
		router.HandleFunc("/api/user/{id: [0-9]+}", controllers.UpdateUser).Methods("PUT")
		router.HandleFunc("/api/user/{id: [0-9]+}", controllers.DeleteUser).Methods("DELETE")
	*/

	r := gin.Default()

	r.POST("/api/user/signup", controllers.Signup)
	r.GET("/verify", controllers.VerifyEmail)
	r.POST("/api/user/login", controllers.Login)
	r.POST("/api/user/forgot-password", controllers.ForgotPassword)
	r.GET("/api/user/verify-password", controllers.VerifyPasswordToken)
	r.POST("/api/user/restore-password", controllers.RestoreUserPassword)

	//requireAuth middleware
	r.GET("/api/user/validate", middleware.RequireAuth, controllers.Validate)
	r.POST("/api/user/logout", middleware.RequireAuth, controllers.Logout)    //ok
	r.GET("/api/user/", middleware.RequireAuth, controllers.GetUsers)         //ok

	//r.GET("/api/user/:ID", middleware.RequireAuth, controllers.GetUser)       //ok
	r.PUT("/api/user/:ID", middleware.RequireAuth, controllers.UpdateUser)    //ok a medias (faltan muchas restricciones)
	r.DELETE("/api/user/:ID", middleware.RequireAuth, controllers.DeleteUser) //ok
	r.GET("/api/user/:ID", middleware.RequireAuth, controllers.CharacterView)

	//characters_catalog
	r.GET("/api/user/characters_catalog", middleware.RequireAuth, controllers.GetCharactersCatalog)
	r.POST("/api/user/choose_first_character", middleware.RequireAuth, controllers.ChooseFirstCharacterOrBuyCharacter)
	
	//bank
	r.POST("/api/user/deposit_bank",middleware.RequireAuth, controllers.DepositInBank)
	r.POST("/api/user/withdraw_bank", middleware.RequireAuth, controllers.WithdrawFromBankToUser)
	
	r.Run()
}
