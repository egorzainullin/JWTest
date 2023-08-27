package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testjwt.com/internal/config"
	"testjwt.com/internal/encryptor"
	"testjwt.com/internal/token"
	u "testjwt.com/internal/user"
)

const severUrl = "localhost:8080"

func main() {
	router := gin.Default()
	router.GET("/gen/:id", genTokensAndSave)
	router.GET("refresh", refreshTokenAndSave)
	router.Run(severUrl)
}

func getStorage() *mongo.Client {
	opts := options.Client().ApplyURI(config.Mongo)
	storage, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic("can't create db")
	}
	fmt.Println("Connected to db successfully")
	return storage
}

func genTokensAndSave(c *gin.Context) {
	userId := c.Param("id")
	storage := getStorage()
	collection := storage.Database("db").Collection("users")
	filter := bson.D{{Key: "id", Value: userId}}
	var result u.UserElement
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			tokens, error := token.GenTokens()
			encrypted, encryptionErr := encryptTokens(tokens)
			if encryptionErr != nil {
				c.IndentedJSON(http.StatusBadRequest, "can not encrypt tokens")
				return
			}
			user := u.UserElement{Encrypted: encrypted, Id: userId}
			_, err := collection.InsertOne(context.TODO(), user)
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, "can not add user with that id")
				return
			}
			if error != nil {
				c.IndentedJSON(http.StatusBadRequest, "can not gen tokens")
				return
			}
			c.IndentedJSON(http.StatusOK, tokens)
			return
		}
	}
	c.IndentedJSON(http.StatusBadRequest, "user with id exists")
}

func combineTokens(tokens token.AccessAndRefreshTokens) string {
	accessToken := tokens.AccessToken
	lastSix := accessToken[len(accessToken)-6:]
	stringFromTwoTokens := tokens.RefreshToken + lastSix
	return stringFromTwoTokens
}

func encryptTokens(tokens token.AccessAndRefreshTokens) (encrypted string, err error) {
	stringFromTwoTokens := combineTokens(tokens)
	encrypted, err = encryptor.GetHash(stringFromTwoTokens)
	fmt.Println(err)
	return encrypted, err
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	Id           string `json:"id"`
}

func refreshTokenAndSave(c *gin.Context) {
	var request RefreshRequest
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "incorrect json")
		return
	}
	storage := getStorage()
	collection := storage.Database("db").Collection("users")
	filter := bson.D{{Key: "id", Value: request.Id}}
	var userElement u.UserElement
	err := collection.FindOne(context.TODO(), filter).Decode(&userElement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusBadRequest, "no tokens with this user id")
			return
		}
		c.IndentedJSON(http.StatusBadRequest, "can not find element")
		return
	}
	encrypted := userElement.Encrypted
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "encryption failure")
		return
	}
	if !encryptor.CheckHash(combineTokens(token.AccessAndRefreshTokens{AccessToken: request.AccessToken, RefreshToken: request.RefreshToken}), encrypted) {
		c.IndentedJSON(http.StatusBadRequest, "tokens do not match")
		return
	}
	newTokens, err := token.GenTokens()
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "can not gen tokens")
		return
	}
	newEncrypted, encErr := encryptTokens(newTokens)
	if encErr != nil {
		c.IndentedJSON(http.StatusBadRequest, "encryption error")
		return
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "encrypted", Value: newEncrypted}}}}
	_, updateErr := collection.UpdateOne(context.TODO(), filter, update)
	if updateErr != nil {
		c.IndentedJSON(http.StatusBadRequest, "update failed")
		return
	}
	c.IndentedJSON(http.StatusOK, newTokens)
}
