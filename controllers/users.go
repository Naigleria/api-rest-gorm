package controllers

import (
	
	"crypto/rand"
	"encoding/base64"
	"fmt"
	
	"os"
	"time"

	"gorm/initializers"
	"gorm/models"
	"net/http"
	"net/smtp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Signup(c *gin.Context) {
	//Get the email/pass off req body
	var body struct {
		Email             string
		Password          string
		Username          string
		VerificationToken string
		IsVerified        bool
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	//generate verification token
	verificationToken, err := generateVerificationToken()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to generate verification token",
		})
		fmt.Println(err)
	}

	//Create the user in db
	user := models.User{Email: body.Email, Password: string(hash), Username: body.Username, VerificationToken: verificationToken, IsVerified: false}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	//send verification email
	err = sendVerificationEmail(user.Email, verificationToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error when sending verification email",
		})
		fmt.Println(err)
	}
	//respond
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
	})
}

func generateVerificationToken() (string, error) {
	//define token length
	tokenLength := 32 //length in bytes

	//create a slice of bytes to store generated token
	tokenBytes := make([]byte, tokenLength)

	//generate random bytes for the token
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	//convert the random bytes in base64 format
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

func sendVerificationEmail(email, token string) error {
	// configure smtp authentication
	smtp_pass:= os.Getenv("SMTP_PASS")
	smtp_server:= os.Getenv("SMTP_SERVER")
	smtp_port:=os.Getenv("SMTP_PORT")
	smtp_email:= os.Getenv("SMTP_EMAIL")

	auth := smtp.PlainAuth("", smtp_email, smtp_pass, smtp_server)

	// build email message body
	message := fmt.Sprintf("Subject: Satoshi Fighter - Verificación de correo electrónico\n"+
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"+
		"<html><body>"+
		"<p>Estimado usuario,</p>"+
		"<p>Su cuenta ha sido creada exitosamente </p>"+
		"<p>Por favor, haz clic en el botón para verificar tu correo electrónico: </p>"+
		"<a href=\"%s/verify?token=%s\"><button style=\"background-color: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; border-radius: 5px;\">Reestablecer Contraseña</button></a>"+
		"<p>Gracias,<br>Satoshi Fighters</p>"+
		"<p>Este es un correo automático,<br>Por favor no responda a éste correo</p>"+
		"</body></html>", os.Getenv("APP_URL"), token)

	// Enviar el correo electrónico
	err := smtp.SendMail(smtp_server+":"+smtp_port, auth, smtp_email, []string{email}, []byte(message))
	if err != nil {
		return err
	}

	return nil
}

func VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Verification token not provided",
		})
		return
	}

	var user models.User
	err := initializers.DB.Where("verification_token = ?", token).First(&user).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token does not match wich any user",
		})
		return
	}
	user.IsVerified = true
	user.VerificationToken = ""

	if err = initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error trying to verify user with token"})
		return
	}
	redirectURL := os.Getenv("FRONTEND_URL")
	c.Redirect(http.StatusFound, redirectURL)
	
}

func ForgotPassword(c *gin.Context){
	type ForgotPassword struct {
		Email string `json:"email" binding:"required,email"`
	}

	var request ForgotPassword
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	//check if email is asociated to a user
	var user models.User
	result := initializers.DB.Where("email = ?", request.Email).First(&user)
	if result.Error != nil {
		if result.Error ==gorm.ErrRecordNotFound{
			c.JSON(http.StatusOK, gin.H{"message": "Password reset instructions sent"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
    }

	//generate token to restore user password
	claims:= jwt.MapClaims{
		"email": request.Email,
		"exp":   time.Now().Add(time.Minute * 10).Unix(),
	}

	//sign token with secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("RESET_PASSWORD_SECRET")
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	// configure smtp authentication
	smtp_pass:= os.Getenv("SMTP_PASS")
	smtp_server:= os.Getenv("SMTP_SERVER")
	smtp_port:=os.Getenv("SMTP_PORT")
	smtp_email:=os.Getenv("SMTP_EMAIL")
	auth := smtp.PlainAuth("", smtp_email, smtp_pass, smtp_server)

	//https://tu-dominio.com/verify-password?token=%s
	//"<p><a href=\"%s/api/user/verify-password?token=%s\"><button style=\"background-color: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; border-radius: 5px;\">Reestablecer Contraseña</button></a></p>"+

	message := fmt.Sprintf("Subject: Satoshi Fighter - Restore Password\n"+
    "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"+
    "<html><body>"+
    "<p>Estimado usuario,</p>"+
    "<p>Para reestablecer tu contraseña, haga click en el siguiente botón </p>"+
    "<a href=\"%sapi/user/verify-password?token=%s\"><button style=\"background-color: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; border-radius: 5px;\">Reestablecer Contraseña</button></a>"+
	"<p>Gracias,<br>Satoshi Fighters</p>"+
    "<p>Este es un correo automático,<br>Por favor no responda a éste correo</p>"+
    "</body></html>", os.Getenv("APP_URL"), signedToken)



	err = smtp.SendMail(smtp_server+":"+smtp_port, auth, smtp_email, []string{request.Email}, []byte(message))
	
	// response to client
    c.JSON(http.StatusOK, gin.H{"message": "Reset email sent successfully"})
}

func VerifyPasswordToken(c *gin.Context) {
	
	tokenString := c.Query("token")
	if tokenString == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
        return
    }

	// check token validity
	secret := os.Getenv("RESET_PASSWORD_SECRET")
    claims := jwt.MapClaims{}
    _, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })


	if err != nil {
	  c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
	  return
	}
  
	// extract user email and expiration info from  token
	email, exp := claims["email"].(string), claims["exp"].(float64)
	
	// verify if token is expired
	if time.Now().Unix() > int64(exp) {
	  c.JSON(http.StatusBadRequest, gin.H{"error": "Expired token"})
	  return
	}
  
	//search user asociated to the email from token
	var user models.User
	result := initializers.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
	  c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in searching user"})
	  return
	}
  
	// redirect user to retore password form in frontend
	// (assumming you have a path for this form for)
	c.Redirect(http.StatusFound, "/reset-password/" + strconv.Itoa(int(user.ID)))
}

func RestoreUserPassword(c *gin.Context){
	//extract user id from url
	userID:=9

	type ChangePassword struct {
		Password string `json:"password"` 
		RepeatPassword string `json:"repeat_password"`
	}

	var body ChangePassword
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
	//validate if passwords are equal
	if body.Password!=body.RepeatPassword{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords are diferent"})
        return
	}
	//Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	type User struct {
		ID       uint   `gorm:"column:id;primaryKey" json:"user_id"`
		Email    string `json:"email"`
	}

	var user User
	result := initializers.DB.Debug().Select("id, email").Where("id = ?", userID).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unable to get user email",
		})
		return
	}

	//result = initializers.DB.Debug().Model(&user).Where("id = ?", userID).Updates(map[string]interface{}{"password": string(hash)})
	result = initializers.DB.Debug().Model(&user).Updates(map[string]interface{}{"password": string(hash)})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to restore user password",
		})
		return
	}

	// configure smtp authentication
	smtp_pass:= os.Getenv("SMTP_PASS")
	smtp_server:= os.Getenv("SMTP_SERVER")
	smtp_port:=os.Getenv("SMTP_PORT")
	smtp_email:=os.Getenv("SMTP_EMAIL")
	auth := smtp.PlainAuth("", smtp_email, smtp_pass, smtp_server)

	message := fmt.Sprintf("Subject: Satoshi Fighter - Password Restored Successfully\n"+
    "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"+
    "<html><body>"+
    "<p>Estimado usuario,</p>"+
    "<p>Su contraseña ha sido cambiada exitosamente</p>"+
    "<p>Si usted no ha realizado esta acción por favor contacte con soporte inmediatamente</p>"+
	"<p>Gracias,<br>Satoshi Fighters</p>"+
    "<p>Este es un correo automático,<br>Por favor no responda a éste correo</p>"+
    "</body></html>")


	fmt.Println(user.Email)
	err = smtp.SendMail(smtp_server+":"+smtp_port, auth, smtp_email, []string{user.Email}, []byte(message))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to send restore email",
		})
		return
	}

	// response to client
    c.JSON(http.StatusOK, gin.H{"message": "Restore email sent successfully"})
}


  

func Login(c *gin.Context) {
	//Get the email and pass off req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	//Look up requested user
	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)
	fmt.Println(body.Email + " : " + body.Password)
	fmt.Println(user)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password 1",
		})
		return
	}

	//Compare sent in  pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password 2",
		})
	}

	//Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"subject":    user.ID,
		"expiration": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	//Sign and get the complete encoded	token as a string using the secret
	tokenString, error := token.SignedString([]byte(os.Getenv("SECRET")))

	if error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}
	//Send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
}
func Validate(c *gin.Context) {

	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func Logout(c *gin.Context) {
	// Clear the JWT cookie
	c.SetCookie("Authorization", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func GetUsers(c *gin.Context) {

	users := models.Users{}
	initializers.DB.Find(&users)
	c.JSON(http.StatusOK, users)

}

func GetUser(c *gin.Context) {
	fmt.Println("In GetUser")
	user, err := getUserById(c)

	if err != nil {
		c.JSON(http.StatusNotFound, "")
	} else {
		c.JSON(http.StatusOK, user)
	}

}

func getUserById(c *gin.Context) (models.User, error) {

	userId, err := strconv.Atoi(c.Param("ID"))

	if err != nil {
		fmt.Println("ID must be a number ")
		return models.User{}, err
	}

	var user models.User
	err = initializers.DB.First(&user, userId).Error
	if err != nil {
		fmt.Println("Error trying retrieve user")
		return user, err
	}

	return user, nil
}

func 	UpdateUser(c *gin.Context) {

	user_ant, err := getUserById(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	//Decode request body in a new user
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		fmt.Println("Error binding JSON:", err.Error())
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid JSON"})
		return
	}

	// Update ID of the user
	user.ID = uint(user_ant.ID)

	// save in db updated user
	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error trying to save user"})
		return
	}

	// Response with updated user
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {

	user, err := getUserById(c)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else {
		initializers.DB.Delete(&user)
		c.JSON(http.StatusOK, user)
	}

}

func CharacterView(c *gin.Context) {
	fmt.Println("In characterWiew")
	userId, err := strconv.Atoi(c.Param("ID"))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID must be a number"})
		return
	}

	var user models.UserResponse
	result := initializers.DB.Debug().Preload("ActiveCharacter", "is_active = ?", true).Select("id, username, battle_stats, safety_time, satoshis").First(&user, userId)

	if result.Error != nil{
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}
	
	c.JSON(http.StatusOK, user)
}

//nextval('characters_character_id_seq'::regclass)
