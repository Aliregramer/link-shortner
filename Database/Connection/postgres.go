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
	UrlID     uint   `json:"url_id" gorm:"Index" gorm:"ForeignKey:urls_id"`
	Ip        string `json:"ip" gorm:"Index"`
	UserAgent string `json:"user_agent" gorm:"Index"`
}

func Migration() {
	appMode := h.Getenv("APP_MODE", "development")

	var db *gorm.DB
	if appMode == "test" {
		dbname := h.Getenv("DB_NAME", "shorjiga_test")
		createDatabase(dbname)
		db = testConnection()
	} else {
		dbname := h.Getenv("DB_NAME", "shorjiga")
		createDatabase(dbname)
		db = Connection()
	}

	// auto migrate all models
	db.AutoMigrate(&Url{})
	db.AutoMigrate(&State{})
}

func createDatabase(databaseName string) {
	host := h.Getenv("DB_HOST", "127.0.0.1")
	port := h.Getenv("DB_PORT", "5432")
	user := h.Getenv("DB_USER", "postgres")
	password := h.Getenv("DB_PASSWORD", "postgres")
	timezone := h.Getenv("DB_TIMEZONE", "Asia/Tehran")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable TimeZone=%s", host, port, user, password, timezone)
	DB, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// createDatabaseCommand := fmt.Sprintf("DROP DATABASE %s", databaseName)
	// DB.Exec(createDatabaseCommand)

	createDatabaseCommand := fmt.Sprintf("CREATE DATABASE %s", databaseName)
	DB.Exec(createDatabaseCommand)
}

func Connection() *gorm.DB {
	host := h.Getenv("DB_HOST", "127.0.0.1")
	port := h.Getenv("DB_PORT", "5432")
	dbname := h.Getenv("DB_NAME", "shorjiga")
	user := h.Getenv("DB_USER", "postgres")
	password := h.Getenv("DB_PASSWORD", "postgres")
	sslmode := h.Getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, dbname, user, password, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Silent),
	})

	CheckError(err)

	return db
}

func testConnection() *gorm.DB {
	host := h.Getenv("DB_HOST", "127.0.0.1")
	port := h.Getenv("DB_PORT", "5432")
	user := h.Getenv("DB_USER", "postgres")
	password := h.Getenv("DB_PASSWORD", "postgres")
	sslmode := h.Getenv("DB_SSLMODE", "disable")
	dbname := h.Getenv("DB_NAME", "shorjiga_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", host, port, user, dbname, password, sslmode)

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
