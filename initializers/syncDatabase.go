package initializers

import "gorm/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
	//DB.AutoMigrate(&models.UsersCharacters{})
	//DB.AutoMigrate(&models.Character{})
	//DB.AutoMigrate(&models.UserResponse{})
}
