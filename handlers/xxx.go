package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	// "go.mongodb.org/mongo-driver/mongo"
	// "strech-server/logger"
	"strech-server/models"
	// "strech-server/db"
)

// var collection *mongo.Collection = db.GetCollection(db.Client, "xxx")

type XxxHandler struct{}

var albums = []models.Xxx{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func (ah XxxHandler) GetAlbums(c *gin.Context) {
	c.JSON(http.StatusOK, albums)
}

func (ah XxxHandler) PostAlbums(c *gin.Context) {
	var newAlbum models.Xxx
	err := c.BindJSON(&newAlbum)
	cookie, err1 := c.Cookie("Cookie_1")
	if err != nil || err1 != nil {
		errorMessage := fmt.Sprint(err1)
		fmt.Println(errorMessage)
		return
	}

	fmt.Println(cookie)
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func (ah XxxHandler) GetAlbumByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
