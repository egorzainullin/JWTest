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
			encrypted, encryptionErr := ecryptTokens(tokens)
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

func ecryptTokens(tokens token.AccessAndRefreshTokens) (encrypted string, err error) {
	accessToken := tokens.AccessToken
	lastSix := accessToken[len(accessToken)-6:]
	stringFromTwoTokens := tokens.RefreshToken + lastSix
	encrypted, err = encryptor.GetHash(stringFromTwoTokens)
	return encrypted, err
}
