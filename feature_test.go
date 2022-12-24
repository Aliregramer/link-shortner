package main

import (
	"net/http"
	"net/http/httptest"
	UrlController "shorjiga/Controllers/Urls"
	middleware "shorjiga/Middleware"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	return router
}

// ------------------- Test Index -------------------
func TestIndexUnAuthentication(t *testing.T) {
	r := SetUpRouter()
	r.Use(middleware.Handle)
	r.GET("/l", UrlController.Index)
	req, _ := http.NewRequest("GET", "/l", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// func TestIndexValidation(t *testing.T) {
// 	r := SetUpRouter()
// 	// test without middleware
// 	r.GET("/l", UrlController.Index)

// 	// insert data
// 	db := Connection.Connection()
// 	url := Connection.Url{}
// 	url.Title = "test title"
// 	url.Description = "test description"
// 	url.ShortUrl = "tet"
// 	url.FullUrl = "test/testUrlForTest"
// 	url.CreatedAt = time.Now()
// 	url.UpdatedAt = time.Now()
// 	db.Model(&url).Save(&url)

// 	http.NewRequest("GET", "/l", nil)

// 	req, _ := http.NewRequest("GET", "/l", nil)

// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	type Response struct {
// 		Id          uint      `json:"id"`
// 		Title       string    `json:"title"`
// 		Description string    `json:"description"`
// 		Short_url   string    `json:"short_url"`
// 		Full_url    string    `json:"full_url"`
// 		Clicked     int       `json:"clicked"`
// 		CreatedAt   time.Time `json:"created_at"`
// 		UpdatedAt   time.Time `json:"updated_at"`
// 	}

// 	var response []Response

// 	response = append(response, Response{
// 		Id:          url.ID,
// 		Title:       url.Title,
// 		Description: url.Description,
// 		Short_url:   h.Getenv("BASE_SHORT_URL", "jajiga.com") + "/" + url.ShortUrl,
// 		Full_url:    h.Getenv("BASE_FULL_URL", "https://www.jajiga.com") + "/" + url.FullUrl,
// 		Clicked:     len(url.States),
// 		CreatedAt:   url.CreatedAt,
// 		UpdatedAt:   url.UpdatedAt,
// 	})
// 	panic(w.Body)
// 	assert.Equal(t, response, w.Body.String())

// }

// ------------------- Test Store -------------------

func TestStoreUnAuthentication(t *testing.T) {
	r := SetUpRouter()
	r.Use(middleware.Handle)
	r.POST("/l", UrlController.Store)
	req, _ := http.NewRequest("POST", "/l", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ------------------- Test Update -------------------
func TestUpdateUnAuthentication(t *testing.T) {
	r := SetUpRouter()
	r.Use(middleware.Handle)
	r.PUT("/l/1", UrlController.Store)
	req, _ := http.NewRequest("PUT", "/l/1", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ------------------- Test Show -------------------
func TestShowUnAuthentication(t *testing.T) {
	r := SetUpRouter()
	r.Use(middleware.Handle)
	r.GET("/l/1", UrlController.Store)
	req, _ := http.NewRequest("GET", "/l/1", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ------------------- Test Destroy -------------------
func TestDestroyUnAuthentication(t *testing.T) {
	r := SetUpRouter()
	r.Use(middleware.Handle)
	r.DELETE("/l/1", UrlController.Store)
	req, _ := http.NewRequest("DELETE", "/l/1", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
