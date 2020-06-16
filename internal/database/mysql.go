package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"time"
)

type MySQLService interface {
	CreateUser(user User) error
	CreateGoogleUser(user User, gUser GoogleUser) error
	GetUserWithEmail(email string) (*User, error)
	GetUserWithID(userId string) (*User, error)
	GetURLIfExistsWithUser(user User, oriURL string) (*URL, error)
	CreateURL(oriURL string, shortenURL string, user User) error
	GetURLWithShortenURL(shortenURL string) (*URL, error)
	GetURLsWithUser(user User, offset uint64, limit uint64) (uint64, []URL, error)
	DeleteURL(shortenURL string) error
	DeleteUser(user User) error
}

/**
 * Gorm* representing implementation of MySQLService.
 */
type gormUser struct {
	UserID    string `gorm:"primary_key"`
	Email     string
	Type      string
	Password  string
	UpdatedAt time.Time
}

type gormGoogleUser struct {
	UserID     string
	GoogleUUID string `gorm:"primary_key"`
	UpdatedAt  time.Time
}

type gormURL struct {
	OriginURL  string
	Owner      string
	ShortenURL string `gorm:"primary_key"`
	UpdatedAt  time.Time
}

type gormService struct {
	db *gorm.DB
}

func newGormSerice(user string, password string, host string, port string, dbName string, dbParams string) (*gormService, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@(%v:%v)/%v?%v", user, password, host, port, dbName, dbParams))
	if err != nil {
		log.Printf("Unable to init database connection %v", err)
		return nil, err
	}

	return &gormService{
		db: db,
	}, nil
}

func (g *gormService) Init() {
	if hasUserTable := g.db.HasTable(&gormUser{}); !hasUserTable {
		g.db.CreateTable(&gormUser{})
		g.db.Model(&gormUser{}).AddIndex("idx_user_id", "user_id")
		g.db.Model(&gormUser{}).AddIndex("idx_email", "email")

		g.db.CreateTable(&gormGoogleUser{})
		g.db.Model(&gormGoogleUser{}).AddIndex("idx_user_id", "user_id")
		g.db.Model(&gormGoogleUser{}).AddIndex("idx_google_uuid", "google_uuid")
	}

	if hasURLTable := g.db.HasTable(&gormURL{}); !hasURLTable {
		g.db.CreateTable(&gormURL{})
		g.db.Model(&gormURL{}).AddIndex("idx_shorten_url", "shorten_url")
	}
}

func (g *gormService) CreateUser(user User) error {
	u := gormUser{
		UserID:    user.UserID,
		Email:     user.Email,
		Type:      user.Type,
		Password:  user.Password,
		UpdatedAt: time.Now(),
	}
	if err := g.db.Create(&u).Error; err != nil {
		log.Printf("Unable to create user in table")
		return err
	}

	return nil
}

func (g *gormService) CreateGoogleUser(user User, gUser GoogleUser) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		g := gormGoogleUser{
			UserID:     gUser.UserID,
			GoogleUUID: gUser.GoogleUUID,
			UpdatedAt:  time.Now(),
		}
		if err := tx.Create(&g).Error; err != nil {
			log.Printf("Unable to create google user in table")
			return err
		}

		u := gormUser{
			UserID:   user.UserID,
			Email:    user.Email,
			Type:     user.Type,
			Password: user.Password,
		}
		if err := tx.Create(&u).Error; err != nil {
			log.Printf("Unable to create user in table")
			return err
		}

		return nil
	})
}

func (g *gormService) GetUserWithEmail(email string) (*User, error) {
	var gormUser gormUser
	execute := g.db.Where("email = ?", email).First(&gormUser)
	return g.queryUser(&gormUser, execute)
}

func (g *gormService) GetUserWithID(userId string) (*User, error) {
	var gormUser gormUser
	execute := g.db.Where("user_id = ?", userId).First(&gormUser)
	return g.queryUser(&gormUser, execute)
}

func (g *gormService) queryUser(userInfo *gormUser, execute *gorm.DB) (*User, error) {
	if execute.RecordNotFound() {
		return nil, NewRecordNotFoundError()
	}

	if err := execute.Error; err != nil {
		return nil, err
	}

	return &User{
		UserID:   userInfo.UserID,
		Email:    userInfo.Email,
		Type:     userInfo.Type,
		Password: userInfo.Password,
	}, nil
}

func (g *gormService) GetURLIfExistsWithUser(user User, oriURL string) (*URL, error) {
	var gormURL gormURL
	execute := g.db.Where("origin_url = ? AND owner = ?", oriURL, user.UserID).First(&gormURL)

	if execute.RecordNotFound() {
		return nil, NewRecordNotFoundError()
	}

	if err := execute.Error; err != nil {
		return nil, err
	}

	return &URL{
		OriginURL:  gormURL.OriginURL,
		Owner:      gormURL.Owner,
		ShortenURL: gormURL.ShortenURL,
	}, nil
}

func (g *gormService) CreateURL(oriURL string, shortenURL string, user User) error {
	u := gormURL{
		OriginURL:  oriURL,
		Owner:      user.UserID,
		ShortenURL: shortenURL,
		UpdatedAt:  time.Now(),
	}
	if err := g.db.Create(&u).Error; err != nil {
		log.Printf("Unable to create url in table")
		return err
	}

	return nil
}

func (g *gormService) GetURLWithShortenURL(shortenURL string) (*URL, error) {
	var gormURL gormURL
	execute := g.db.Where("shorten_url = ?", shortenURL).First(&gormURL)

	if execute.RecordNotFound() {
		return nil, NewRecordNotFoundError()
	}

	if err := execute.Error; err != nil {
		return nil, err
	}

	return &URL{
		OriginURL:  gormURL.OriginURL,
		Owner:      gormURL.Owner,
		ShortenURL: gormURL.ShortenURL,
	}, nil
}

func (g *gormService) GetURLsWithUser(user User, offset uint64, limit uint64) (uint64, []URL, error) {
	var gormUrl2 []gormURL
	var count int
	countExecute := g.db.Where("owner = ?", user.UserID).Find(&gormUrl2).Count(&count)
	if err := countExecute.Error; err != nil {
		return 0, nil, err
	}

	var gormUrls []gormURL
	queryExecute := g.db.Order("updated_at desc").Offset(offset).Where("owner = ?", user.UserID).Limit(limit).Find(&gormUrls)
	if err := queryExecute.Error; err != nil {
		return 0, nil, err
	}

	urls := make([]URL, len(gormUrls))
	for i, url := range gormUrls {
		urls[i] = URL{
			OriginURL:  url.OriginURL,
			Owner:      url.Owner,
			ShortenURL: url.ShortenURL,
		}
	}

	return uint64(count), urls, nil
}

func (g *gormService) DeleteURL(shortenURL string) error {
	var gormURL gormURL
	execute := g.db.Unscoped().Where("shorten_url = ?", shortenURL).Delete(&gormURL)
	if err := execute.Error; err != nil {
		return err
	}

	return nil
}

func (g *gormService) DeleteUser(user User) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		var gormUser gormUser
		execute := tx.Unscoped().Where("user_id = ?", user.UserID).Delete(&gormUser)
		if err := execute.Error; err != nil {
			return err
		}

		if user.Type == UserTypeGoogle {
			var gormGoogleUser gormGoogleUser
			execute := tx.Unscoped().Where("user_id = ?", user.UserID).Delete(&gormGoogleUser)
			if err := execute.Error; err != nil {
				return err
			}
		}

		return nil
	})
}

type Config struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
	DBParams string
}

// NewMySQLDatabase returns MySQLService.
// Error returns if occurred.
func NewMySQLDatabase(c Config) (MySQLService, error) {
	var user string
	var pass string
	var host string
	var port string
	var dbName string
	var dbParams string

	if c.Username == "" {
		user = "root"
	} else {
		user = c.Username
	}

	pass = c.Password

	if c.Host == "" {
		host = "localhost"
	} else {
		host = c.Host
	}

	if c.Port == "" {
		port = "3306"
	} else {
		port = c.Port
	}

	if c.DBName == "" {
		dbName = "url_shortener"
	} else {
		dbName = c.DBName
	}

	if c.DBParams == "" {
		dbParams = "charset=utf8&parseTime=True&loc=Local"
	} else {
		dbParams = c.DBParams
	}

	g, err := newGormSerice(user, pass, host, port, dbName, dbParams)
	if err != nil {
		log.Printf("Unable to create an instance of Gorm")
		return nil, err
	}

	g.Init()

	return g, nil
}
