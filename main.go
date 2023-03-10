package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// local
	"shorjiga/Controllers/Urls"
	db "shorjiga/Database/Connection"
	middleware "shorjiga/Middleware"
)


func main() {

	// loadDotEnv()

	db.Migration()

	r := gin.Default()
	r.GET("/l/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	if (gin.Mode() == gin.DebugMode) {
		r.Use(gin.Logger())
	} // goodbye copilot :(
	
	authorized := r.Group(os.Getenv("PREFIX_URL"))

	// per group middleware! in this case we use the custom created
	// AuthRequired() middleware just in the "authorized" group.
	authorized.Use(middleware.Handle)
	{
		authorized.GET("/", UrlController.Index)
		authorized.POST("/", UrlController.Store)
		authorized.PUT("/:id", UrlController.Update)
		authorized.GET("/show/:id", UrlController.Show)
		authorized.DELETE("/:id", UrlController.Destroy)
	}

	r.GET(os.Getenv("PREFIX_URL") + "/:url", UrlController.Redirect)
	r.GET(os.Getenv("PREFIX_URL") + "/r/:url", UrlController.RoomRedirect)
	r.GET(os.Getenv("PREFIX_URL") + "/r/:url/:utm", UrlController.RoomRedirect)

	r.Run("0.0.0.0:9090")
}

func loadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occurred. Err: %s", err)
	}
}
