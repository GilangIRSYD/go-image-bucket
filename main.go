package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "image/jpeg"

	"github.com/gin-gonic/gin"
)

var PATH string

func main() {

	//Init folder
	PATH = initFolder()
	fmt.Println("IMAGE ROOT PATH =>", PATH)

	// Initialize Router
	r := gin.Default()
	r.Use(CORSMiddleware())
	initRouter(r)
}

func initRouter(router *gin.Engine) {
	PORT := os.Getenv("PORT")
	router.GET("/health-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	router.POST("v1/image", uploadImage)
	router.GET("v1/image/:path", getImage)
	router.GET("bucket-image/halloguru/:file", getImage)

	router.Run(":" + PORT)
}

func initFolder() (path string) {
	path = filepath.Join(".", "images")
	err := os.MkdirAll(path, os.ModePerm) // If path is already a directory, MkdirAll does nothing

	if err != nil {
		log.Fatal(err)
	}

	return path
}

func uploadImage(ctx *gin.Context) {
	var (
		isError bool
		message string
		code    int
	)

	defer func() {
		if isError {
			ctx.JSON(code, gin.H{
				"status":  false,
				"code":    code,
				"message": message,
			})
		}
	}()

	body, err := ctx.GetRawData()
	if err != nil {
		isError = true
		message = err.Error()
		return
	}

	var base64Image map[string]string
	json.Unmarshal(body, &base64Image)

	// fmt.Println("baes64 =>>>", base64Image["image"])

	image, err := base64.StdEncoding.DecodeString(string(base64Image["image"]))
	if err != nil {
		isError = true
		message = err.Error()
		code = 500
		return
	}
	uniqueName := strconv.FormatInt(time.Now().UTC().UnixNano(), 10) + ".png"
	file, err := os.Create(PATH + "/" + uniqueName)
	defer file.Close()

	if _, err = file.Write(image); err != nil {
		isError = true
		message = err.Error()
		code = 500
		return
	}

	if err = file.Sync(); err != nil {
		isError = true
		message = err.Error()
		code = 500
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"code":    http.StatusOK,
		"message": "Success",
		"link":    "https://halloguru.herokuapp.com/bucket-image/halloguru/" + uniqueName,
	})
}

func getImage(c *gin.Context) {
	file := c.Param("file")
	fmt.Println("file ==>", file)
	c.File("./images/" + file)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
