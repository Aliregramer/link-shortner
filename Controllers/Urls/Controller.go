package UrlController

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	// local
	"shorjiga/Database/Connection"
	h "shorjiga/Helper"
)

var db = Connection.Connection()
var Redis = Connection.RedisClient()

func Index(c *gin.Context) {
	url := Connection.Url{}
	var urls []Connection.Url

	db.Preload("States").Model(url).Scopes(Connection.Paginate(c.Request)).Find(&urls)

	// create response
	type Response struct {
		Id          uint      `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Short_url   string    `json:"short_url"`
		Full_url    string    `json:"full_url"`
		Clicked     int       `json:"clicked"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	var response []Response

	for _, url := range urls {
		response = append(response, Response{
			Id:          url.ID,
			Title:       url.Title,
			Description: url.Description,
			Short_url:   h.Getenv("BASE_SHORT_URL", "jajiga.com") + "/" + url.ShortUrl,
			Full_url:    h.Getenv("BASE_FULL_URL", "https://www.jajiga.com") + "/" + url.FullUrl,
			Clicked:     len(url.States),
			CreatedAt:   url.CreatedAt,
			UpdatedAt:   url.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}

func Store(c *gin.Context) {
	fullUrl := c.PostForm("fullUrl")

	var count int64
	db.Where("full_url = ?", fullUrl).Count(&count)
	if count != 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "link exists",
		})
	}

	if !strings.HasPrefix(fullUrl, os.Getenv("BASE_FULL_URL")) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "inserted url must start with " + os.Getenv("BASE_FULL_URL"),
		})
		return
	}

	// read len of short url if not set default 3
	var shortUrlLen int
	if c.PostForm("short_url_len") != "" {
		shortUrlLen, _ = strconv.Atoi(c.PostForm("short_url_len"))
		if shortUrlLen < 1 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "short url len must be greater than 0",
			})
			return
		}
	} else {
		shortUrlLen = 3
	}

	// json decode valid_char
	var valid_char []string
	json.Unmarshal([]byte(c.PostForm("valid_char")), &valid_char)

	shortUrl := createShortUrl(shortUrlLen, valid_char)

	url := Connection.Url{}
	url.Title = c.PostForm("title")
	url.Description = c.PostForm("description")
	url.ShortUrl = shortUrl
	url.FullUrl = fullUrl[len(os.Getenv("BASE_FULL_URL"))+1:]
	url.CreatedAt = time.Now()
	url.UpdatedAt = time.Now()

	result := db.Model(&url).Save(&url)
	if result.Error != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Error while saving",
		})
		return
	}

	go saveInRedis(url)

	c.JSON(http.StatusOK, gin.H{
		"short_url": os.Getenv("BASE_SHORT_URL") + "/" + shortUrl,
	})

}

func Show(c *gin.Context) {
	id := c.Param("id")
	url := Connection.Url{}
	result := db.Where("id = ?", id).First(&url)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Url not found",
		})
		return
	}

	type StateResponse struct {
		Ip        string `json:"ip"`
		UserAgent string `json:"user_agent"`
		CreatedAt string `json:"created_at"`
	}

	states := []Connection.State{}
	statesResponse := []StateResponse{}
	db.Model(states).Where("url_id = ?", id).Order("created_at desc").Limit(100).Find(&statesResponse)

	// create response
	type Response struct {
		Id          uint            `json:"id"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Short_url   string          `json:"short_url"`
		Full_url    string          `json:"full_url"`
		CreatedAt   time.Time       `json:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at"`
		Clicked     int             `json:"clicked"`
		State       []StateResponse `json:"state"`
	}
	response := Response{
		Id:          url.ID,
		Title:       url.Title,
		Description: url.Description,
		Short_url:   h.Getenv("BASE_SHORT_URL", "jajiga.com") + "/" + url.ShortUrl,
		Full_url:    h.Getenv("BASE_FULL_URL", "https://www.jajiga.com") + "/" + url.FullUrl,
		Clicked:     len(statesResponse),
		State:       statesResponse,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}

	//TODO: add finded url in redis

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func Update(c *gin.Context) {
	id := c.Param("id")
	url := Connection.Url{}
	result := db.Where("id = ?", id).First(&url)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Url not found",
		})
		return
	}

	// loop on present parm in request
	for key, value := range c.Request.PostForm {
		if key == "title" {
			url.Title = value[0]
		} else if key == "full_url" {
			url.FullUrl = value[0]
		} else if key == "short_url" {
			if checkShortUrlExists(value[0]) {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"message": "Short url already exists",
				})
				return
			}
			url.ShortUrl = value[0]
		}
	}

	result = db.Model(&url).Save(&url)

	if result.Error != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Error while saving",
		})
		return
	}

	// update in redis
	go saveInRedis(url)

	// create response
	type Response struct {
		Id          uint      `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Short_url   string    `json:"short_url"`
		Full_url    string    `json:"full_url"`
		Clicked     int       `json:"clicked"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	var response []Response
	response = append(response, Response{
		Id:          url.ID,
		Title:       url.Title,
		Description: url.Description,
		Short_url:   h.Getenv("BASE_SHORT_URL", "jajiga.com") + "/" + url.ShortUrl,
		Full_url:    h.Getenv("BASE_FULL_URL", "https://www.jajiga.com") + "/" + url.FullUrl,
		Clicked:     len(url.States),
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated",
		"data":    response,
	})
}

func Destroy(c *gin.Context) {
	id := c.Param("id")
	url := Connection.Url{}
	result := db.Where("id = ?", id).Delete(&url)
	if result.Error != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Error while saving",
		})
		return
	} else if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Url not found",
		})
		return
	}

	// delete from redis
	Redis.Del("urls:" + url.ShortUrl)

	c.JSON(204, gin.H{
		"message": "Item deleted",
	})
}

func Redirect(c *gin.Context) {
	shortUrl := c.Param("url")
	shortUrl = strings.ReplaceAll(shortUrl, "/", "")

	result := Redis.Get("urls:" + shortUrl)

	var full_url string
	if result.Val() == "" {
		url := Connection.Url{}
		result := db.Where("short_url = ?", shortUrl).First(&url)

		if result.Error != nil && result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Url not found",
			})
			return
		}
		full_url = url.FullUrl
	} else {
		full_url = result.Val()
	}

	go createState(full_url, c)

	c.Redirect(302, os.Getenv("BASE_FULL_URL")+"/"+full_url) // todo throw error if not found
}

func createShortUrl(length int, valid_char []string) string {

	var validItem string
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFJHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"

	for _, item := range valid_char {
		if item == "lowercase" {
			validItem += lowercase
		} else if item == "uppercase" {
			validItem += uppercase
		} else if item == "numbers" {
			validItem += numbers
		}
	}

	if validItem == "" {
		validItem = lowercase + uppercase + numbers
	}

	stringLen := len(validItem)

	var shortUrl string
	for i := 0; i < length; i++ {
		randomCharIndex := rand.Intn(stringLen)
		shortUrl += string(validItem[randomCharIndex])
	}

	if checkShortUrlExists(shortUrl) {
		return createShortUrl(length, valid_char)
	}

	return shortUrl
}

func checkShortUrlExists(short_url string) bool {
	var count int64
	db.Table("urls").Where("short_url = ?", short_url).Count(&count)
	if count > 0 {
		return true
	}

	return false
}

func saveInRedis(url Connection.Url) {
	Redis.Set("urls:"+url.ShortUrl, url.FullUrl, 0)
}

func createState(full_url string, c *gin.Context) {
	url := Connection.Url{}
	state := Connection.State{}

	db.Where("full_url = ?", full_url).First(&url)

	state.UrlID = url.ID
	state.Ip = c.ClientIP()
	state.UserAgent = c.Request.UserAgent()
	state.CreatedAt = time.Now()
	state.UpdatedAt = time.Now()

	db.Model(&state).Save(&state)
}
