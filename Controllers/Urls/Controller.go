package UrlController

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martinlindhe/base36"
	"gorm.io/gorm"

	// local
	"shorjiga/Database/Connection"
	h "shorjiga/Helper"
)

func Index(c *gin.Context) {
	url := Connection.Url{}
	var urls []Connection.Url

	db := Connection.Connection()
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
			Short_url:   h.Getenv("BASE_SHORT_URL", "j1g.com") + "/" + url.ShortUrl,
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
	type params struct {
		FullUrl     string `json:"full_url" form:"full_url" binding:"required"`
		Title       string `json:"title" form:"title" binding:"required|max=255"`
		ShortUrlLen int    `json:"short_url_len" form:"short_url_len" binding:"min=1|max=255"`
		Description string `json:"description" form:"description" binding:"max=255"`
	}
	// validate input
	var inputs params
	if err := c.ShouldBind(&inputs); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": err.Error(),
		})
		return
	}

	db := Connection.Connection()
	fullUrl, err := handleFullUrl(c, db, inputs.FullUrl)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": err.Error(),
		})
		return
	}

	// validate expire_at
	var expire_at time.Time
	if len(c.PostForm("expire_at")) != 0 {
		println(c.PostForm("expire_at"))
		expire_at, _ = time.Parse("2006-01-02 15:04", c.PostForm("expire_at"))
		if expire_at.Format("2006-01-02 15:04:00") == "0001-01-01 00:00:00" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Invalid expire_at format",
			})
			return
		}
	}

	// read len of short url if not set default 3
	if inputs.ShortUrlLen == 0 {
		inputs.ShortUrlLen = 3
	}

	// json decode valid_char
	var valid_char []string
	json.Unmarshal([]byte(c.PostForm("valid_char")), &valid_char)

	shortUrl := createShortUrl(inputs.ShortUrlLen, valid_char)
	url := Connection.Url{}
	url.Title = inputs.Title
	url.Description = inputs.Description
	url.ShortUrl = shortUrl
	url.FullUrl = fullUrl
	url.CreatedAt = time.Now()
	url.UpdatedAt = time.Now()
	url.ExpireAt = expire_at.Format("2006-01-02 15:04:00")

	result := db.Model(&url).Save(&url)
	if result.Error != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Error while saving",
		})
		return
	}

	saveInRedis(url)

	c.JSON(http.StatusOK, gin.H{
		"short_url": os.Getenv("BASE_SHORT_URL") + "/" + shortUrl,
	})

}

func Show(c *gin.Context) {
	id := c.Param("id")
	url := Connection.Url{}
	db := Connection.Connection()

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
		ExpireAt    string          `json:"expire_at"`
		Clicked     int             `json:"clicked"`
		State       []StateResponse `json:"state"`
	}
	response := Response{
		Id:          url.ID,
		Title:       url.Title,
		Description: url.Description,
		Short_url:   h.Getenv("BASE_SHORT_URL", "j1g.com") + "/" + url.ShortUrl,
		Full_url:    h.Getenv("BASE_FULL_URL", "https://www.jajiga.com") + "/" + url.FullUrl,
		Clicked:     len(statesResponse),
		State:       statesResponse,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
		ExpireAt:    url.ExpireAt,
	}

	saveInRedis(url)

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func Update(c *gin.Context) {
	type params struct {
		FullUrl     string `json:"full_url" form:"full_url"`
		Title       string `json:"title" form:"title" binding:"max=255"`
		ShortUrlLen int    `json:"short_url_len" form:"short_url_len" binding:"min=1|max=255"`
		Description string `json:"description" form:"description" binding:"max=255"`
	}
	// validate input
	var inputs params
	if err := c.ShouldBind(&inputs); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": err.Error(),
		})
		return
	}

	id := c.Param("id")
	url := Connection.Url{}
	db := Connection.Connection()

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
		} else if key == "description" {
			url.Description = value[0]
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
	saveInRedis(url)

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
		ShortUrl    string          `json:"short_url"`
		FullUrl     string          `json:"full_url"`
		CreatedAt   time.Time       `json:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at"`
		ExpireAt    string          `json:"expire_at"`
		Clicked     int             `json:"clicked"`
		State       []StateResponse `json:"state"`
	}
	response := Response{
		Id:          url.ID,
		Title:       url.Title,
		Description: url.Description,
		ShortUrl:    h.Getenv("BASE_SHORT_URL", "j1g.com") + "/" + url.ShortUrl,
		FullUrl:     h.Getenv("BASE_FULL_URL", "https://www.jajiga.com") + "/" + url.FullUrl,
		Clicked:     len(statesResponse),
		State:       statesResponse,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
		ExpireAt:    url.ExpireAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated",
		"data":    response,
	})
}

func Destroy(c *gin.Context) {
	id := c.Param("id")
	url := Connection.Url{}
	db := Connection.Connection()

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

	var Redis = Connection.RedisClient()
	// delete from redis
	Redis.Del("urls:" + url.ShortUrl)

	c.JSON(204, gin.H{
		"message": "Item deleted",
	})
}

func Redirect(c *gin.Context) {
	shortUrl := c.Param("url")

	shortUrl = strings.ReplaceAll(shortUrl, "/", "")

	db := Connection.Connection()

	var Redis = Connection.RedisClient()
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
		go saveInRedis(url)
	} else {
		full_url = result.Val()
	}

	go createState(full_url, c)

	c.Redirect(302, os.Getenv("BASE_FULL_URL")+"/"+full_url) // todo throw error if not found
}

func RoomRedirect(c *gin.Context) {
	roomIdBase36 := c.Param("url")
	utm := c.Param("utm")

	roomId := strconv.FormatUint(base36.Decode(roomIdBase36)+3139000, 10)

	var utmParam string
	if len(utm) != 0 {
		utmParam = readUtmPart(utm)
	}

	go createStateForRoom(roomId, utmParam, c)

	c.Redirect(302, os.Getenv("BASE_FULL_URL")+"/room/"+roomId+utmParam)
}

func readUtmPart(utm string) string {
	result := "?"
	switch utm[0] {
	case 'd':
		result += "utm_source=direct"
	case 'e':
		result += "utm_source=telegram"
	case 'f':
		result += "utm_source=facebook"
	case 's':
		result += "utm_source=sms"
	case 't':
		result += "utm_source=twitter"
	case 'w':
		result += "utm_source=whatsapp"
	}

	if len(utm) > 1 {
		result += "&"
		
		switch utm[1] {
		case 'r':
			result += "utm_medium=room"
		case 's':
			result += "utm_medium=search"
		case 'w':
			result += "utm_medium=wishes"
		}
	}

	if result == "?" {
		result = ""
	}

	return result
}

func handleFullUrl(c *gin.Context, db *gorm.DB, full_url string) (string, error) {
	// check if full url is valid
	if !strings.HasPrefix(full_url, os.Getenv("BASE_FULL_URL")) {

		return "", errors.New("inserted url must start with " + os.Getenv("BASE_FULL_URL"))
	}
	fullUrl := strings.TrimPrefix(full_url, os.Getenv("BASE_FULL_URL")) // remove base url from full url

	var count int64
	db.Where("full_url = ?", fullUrl).Count(&count)
	if count != 0 {
		return "", errors.New("url already exists")
	}

	return fullUrl, nil
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
	db := Connection.Connection()

	db.Table("urls").Where("short_url = ?", short_url).Count(&count)
	if count > 0 {
		return true
	}

	return false
}

func saveInRedis(url Connection.Url) {
	var Redis = Connection.RedisClient()
	Redis.Set("urls:"+url.ShortUrl, url.FullUrl, 0)
}

func createState(full_url string, c *gin.Context) {
	url := Connection.Url{}
	state := Connection.State{}
	db := Connection.Connection()

	db.Where("full_url = ?", full_url).First(&url)

	state.UrlID = &url.ID
	state.Ip = c.ClientIP()
	state.UserAgent = c.Request.UserAgent()
	state.CreatedAt = time.Now()
	state.UpdatedAt = time.Now()

	db.Model(&state).Save(&state)
}

func createStateForRoom(roomId string, utm string, c *gin.Context) {
	state := Connection.State{}
	db := Connection.Connection()

	if len(utm) != 0 {
		utm = strings.ReplaceAll(utm, "?", "")
		utmSlice:= strings.Split(utm, "&")

		for _,param := range utmSlice {
			if (strings.Contains(param, "utm_source")) {
				state.UtmSource = strings.ReplaceAll(param, "utm_source=", "")
			} else if  (strings.Contains(param, "utm_medium")) {
				state.UtmMedium = strings.ReplaceAll(param, "utm_medium=", "")
			}
		}
		
	}

	state.RoomId = roomId
	state.Ip = c.ClientIP()
	state.UrlID = nil
	state.UserAgent = c.Request.UserAgent()
	state.CreatedAt = time.Now()
	state.UpdatedAt = time.Now()

	db.Model(&state).Save(&state)
}
