package controllers

import (
	"net/http"
	"ren0503/talkive/src/models"
	"ren0503/talkive/src/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateSession(ctx *gin.Context) {
	db := ctx.MustGet("db").(*mongo.Client)
	collection := db.Database("meet").Collection("sessions")

	var session models.Session
	if err := ctx.ShouldBindJSON(&session); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session.Password = utils.HashPassword(session.Password)

	result, _ := collection.InsertOne(ctx, session)
	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	url := CreateSocket(session, ctx, insertedID)
	ctx.JSON(http.StatusOK, gin.H{"socket": url})
}
