package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	config "testjwt.com/internal/config"
)

func main() {
	opts := options.Client().ApplyURI(config.Mongo)
	_, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic("can't create db")
	}
	fmt.Println("Connected to db successfully")
	_ = gin.Default()

}
