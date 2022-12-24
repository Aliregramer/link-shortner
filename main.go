package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// local
	"shorjiga/Controllers/Urls"
	db "shorjiga/Database/Connection"
	middleware "shorjiga/Middleware"
)


func main() {

	loadDotEnv()

	db.Migration()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	authorized := r.Group("/")
	// per group middleware! in this case we use the custom created
	// AuthRequired() middleware just in the "authorized" group.
	authorized.Use(middleware.Handle)
	{
		authorized.GET("/l/", UrlController.Index)
		authorized.POST("/l/", UrlController.Store)
		authorized.PUT("/l/:id", UrlController.Update)
		// authorized.GET("/:id", UrlController.Show) // TODO: validate id type
		authorized.DELETE("/l/:id", UrlController.Destroy)
	}

	r.GET("/:url", UrlController.Redirect)

	r.Run("0.0.0.0:9090")
}

func loadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occurred. Err: %s", err)
	}
}
