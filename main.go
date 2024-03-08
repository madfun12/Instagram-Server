package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
)

type CodeRequest struct {
    Code string `json:"code"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID int64 `json:"user_id"`
}

type LongTermAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID int64 `json:"user_id"`
	ExpirationDate int64 `json:"expires_in"`
}

type Account struct {
	ID int
	UserID int64
	UserName string
	AccessToken string
	ExpirationSeconds int
}

type InstagramMedia struct {
    ID         string `json:"id"`
    MediaURL   string `json:"media_url"`
    Permalink  string `json:"permalink"`
    Caption    string `json:"caption"`
    Timestamp  string `json:"timestamp"`
    MediaType  string `json:"media_type"`
    ThumbnailURL string `json:"thumbnail_url"`
}

type InstagramAccount struct {
    UserID         string          `json:"user_id"`
    Media          []InstagramMedia `json:"media"`
}

func getAccount(c *gin.Context) {
    // Connect to PostgreSQL database
    connStr := "postgres://postgres:admin@localhost:5432/instagram?sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Println("GET /account - Error connecting to database: ", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }
    defer db.Close()

    // Extract userID from route parameter
    userID := c.Param("user_id")

    // Convert userID to int64
    userIDInt, err := strconv.ParseInt(userID, 10, 64)
    if err != nil {
        // Handle invalid userID
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Retrieve account from database
    var account Account
    err = db.QueryRow("SELECT id, user_id, access_token, expiration_seconds FROM account WHERE user_id=$1", userIDInt).
        Scan(&account.ID, &account.UserID, &account.AccessToken, &account.ExpirationSeconds)
    if err != nil {
        fmt.Println("Error querying account table: ", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account from database"})
        return
    }

	// Get username
	requestUrl := "https://graph.instagram.com/v11.0/me/media?fields=id,media_url,permalink,caption,timestamp,media_type,thumbnail_url&limit=10&access_token=" + account.AccessToken
    resp, err := http.Get(requestUrl)
	if err != nil {
		fmt.Println("Error getting account information:", err)
		return
	}
	defer resp.Body.Close()

    // Check response status code
    if resp.StatusCode != http.StatusOK {
         fmt.Println("Instagram API returned non-OK status code: ", resp.Status)
		 return
    }

    // Retrieve user's media posts from Instagram API
    media, err := getRecentPosts(account.AccessToken)
    if err != nil {
        fmt.Println("Error retrieving user's media posts: ", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user's media posts"})
        return
    }

    // Construct account information JSON object
    accountInfo := InstagramAccount{
        UserID:         fmt.Sprint(account.UserID),
        Media:          media,
    }

    // Return account information with media to client
    c.JSON(http.StatusOK, accountInfo)
}

func getRecentPosts(accessToken string) ([]InstagramMedia, error) {
    // Make request to Instagram API to fetch user's recent media posts
    requestUrl := "https://graph.instagram.com/v11.0/me/media?fields=id,media_url,permalink,caption,timestamp,media_type,thumbnail_url&limit=8&access_token=" + accessToken
    resp, err := http.Get(requestUrl)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Check response status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Instagram API returned non-OK status code: %s", resp.Status)
    }

    // Decode Instagram API response
    var mediaResponse struct {
        Data []InstagramMedia `json:"data"`
    }
    err = json.NewDecoder(resp.Body).Decode(&mediaResponse)
    if err != nil {
        return nil, err
    }

    return mediaResponse.Data, nil
}

func postAccount(c *gin.Context) {
	connStr := "postgres://postgres:admin@localhost:5432/instagram?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil{
		fmt.Println("Error connecting to database: ", err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS account (
		id SERIAL PRIMARY KEY,
		user_id BIGINT,
		access_token TEXT,
		expiration_seconds INT
	)`)
	if err != nil {
		fmt.Println("Error creating account table: ", err)
		return
	}

	// Create an empty CodRequest variable
	var requestBody CodeRequest

	// Take the request and bind the contents to requestBody
	if err := c.BindJSON(&requestBody); err != nil {
		fmt.Println(err)
		return
	}

	// Get authorization code and then retrieve short lived token
	code := requestBody.Code
	shortLivedToken, userID, err := requestShortTokenFromInstagram(code)
	if err != nil{
		fmt.Println(err)
		return
	}

	// Use short lived token to retrieve long lived token
	longLivedToken, expiresIn, err := requestLongTermTokenFromInstagram(shortLivedToken)
	if err != nil {
		fmt.Println("Error getting long lived token: ", err)
		return
	}

	// Insert data into the account table
	_, err = db.Exec(`INSERT INTO account (user_id, access_token, expiration_seconds) VALUES ($1, $2, $3)`, userID, longLivedToken, expiresIn)

	if err != nil {
		fmt.Println("Error inserting data into account table: ", err)
		return
	}

	fmt.Println("Data inserted into account table successfully")

	responseData := struct {
		UserID string `json:"userID"`
	}{
		UserID: fmt.Sprint(userID),
	}

	c.JSON(http.StatusOK, responseData)
}

func requestShortTokenFromInstagram(code string) (string, int64, error){
	// Load env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return "", 0, err
	}

	// Get client id & client secret from .env
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURI := "https://instagram-test-omega.vercel.app/authorization-success"

	// Create the POST form data
    formData := url.Values{}
    formData.Set("client_id", clientId)
    formData.Set("client_secret", clientSecret)
    formData.Set("redirect_uri", redirectURI)
    formData.Set("code", code)
    formData.Set("grant_type", "authorization_code")

    // Make the POST request
    resp, err := http.PostForm("https://api.instagram.com/oauth/access_token", formData)
    if err != nil {
        fmt.Println("Error making POST request:", err)
        return "", 0, err
    }

	responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", 0, err
    }

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := errors.New("non-OK response from Instagram")
		fmt.Println("Error", err)
		return "", 0, err
	}

	var accessTokenReponse AccessTokenResponse
	// Reset the response body reader to its initial state
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	// Parse the JSON response body
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accessTokenReponse)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return "", 0, err
	}

	return accessTokenReponse.AccessToken, accessTokenReponse.UserID, nil
}

func requestLongTermTokenFromInstagram(shortLivedToken string) (string, int64, error){
	clientSecret := os.Getenv("CLIENT_SECRET")

	url := "https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=" + clientSecret + "&access_token=" + shortLivedToken
	resp, err := http.Get(url)
	if err != nil{
		return "", 0, err
	}

	responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", 0,err
    }

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := errors.New("non-OK response from Instagram")
		fmt.Println("Error", err)
		return "", 0,err
	}

	var accessTokenReponse LongTermAccessTokenResponse
	// Reset the response body reader to its initial state
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	// Parse the JSON response body
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accessTokenReponse)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return "", 0, err
	}

	return accessTokenReponse.AccessToken, accessTokenReponse.ExpirationDate, nil
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000"} // Add your frontend URL here
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    router.Use(cors.New(config))

	router.GET("/account/:user_id", getAccount)
	router.POST("/account", postAccount)
	router.Run()
}