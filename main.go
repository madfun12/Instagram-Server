package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

type CodeRequest struct {
    Code string `json:"code"`
}

type Request struct{
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType string `json:"grant_type"`
	RedirectURI string `json:"redirect_uri"`
	Code string `json:"code"`
}

func getTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func postAccount(c *gin.Context) {
	var requestBody CodeRequest

	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Println(err)
		return
	}

	code := requestBody.Code

	fmt.Println("Auth Code:" + code)

	requestShortTokenFromInstagram(code)
}

func requestShortTokenFromInstagram(code string) error{
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return err
	}

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	grantType := "authorization_code"
	redirectURI := "https://instagram-test-omega.vercel.app/authorization-success"

	requestBody := Request{
		ClientId: clientId,
		ClientSecret: clientSecret,
		GrantType: grantType,
		RedirectURI: redirectURI,
		Code: code,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	url := "https://api.instagram.com/oauth/access_token"

	fmt.Println("Request URL:" + url)
	fmt.Println("Request Body:" + string(requestBodyBytes))
	return nil
}

func requestLongTermTokenFromInstagram(shortLivedToken string) error{
	return nil
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000"} // Add your frontend URL here
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    router.Use(cors.New(config))

	router.GET("/", getTest)
	router.POST("/account", postAccount)
	router.Run()
}