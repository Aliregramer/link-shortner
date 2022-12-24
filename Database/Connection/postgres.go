package Connection

import (
	"fmt"
	"net/http"

	// "os"
	"strconv"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	h "shorjiga/Helper"
)

type Url struct {
	gorm.Model
	ShortUrl    string `json:"short_url" gorm:"uniqueIndex"`
	FullUrl     string `json:"full_url" gorm:"Index"`
	Title       string `json:"title" gorm:"unique, default:null"`
	Description string `json:"description" gorm:"default:null"`
	States      []State
}

type State struct {
	gorm.Model
	UrlID     uint   `json:"url_id" gorm:"Index" gorm:"ForeignKey:users_id"`
	Ip        string `json:"ip" gorm:"Index"`
	UserAgent string `json:"user_agent" gorm:"Index"`
}

func Migration() {
	db := Connection()

	// auto migrate all models
	db.AutoMigrate(&Url{})
	db.AutoMigrate(&State{})
}

func Connection() *gorm.DB {
	host := h.Getenv("DB_HOST", "127.0.0.2")
	port := h.Getenv("DB_PORT", "5432")
	user := h.Getenv("DB_USER", "postgres")
	password := h.Getenv("DB_PASSWORD", "postgres")
	sslmode := h.Getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", host, port, user, password, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Silent),
	})

	CheckError(err)

	return db
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Paginate(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page == 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(q.Get("per_page"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
