package main

import (
	"context"
	"fmt"
	rs "github.com/go-redis/redis"
	"log"
	"strings"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/route/user/sign"
	"url-shortener/internal/server"
	"url-shortener/internal/service/mail"
)

func main() {
	// TODO: connect to database at startup
	// TODO: make env variable configurable
	// TODO: do TDD
	env := config.ReadEnv()

	/**
	Database configuration
	*/
	dbConfig := database.Config{
		Username: env.DBUser,
		Password: env.DBPass,
		Host:     env.DBHost,
		Port:     env.DBPort,
		DBName:   env.DBName,
		DBParams: env.DBParams,
	}
	db, err := database.NewMySQLDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Unable to set up database | Reason: %v\n", err)
	}

	/**
	Caching configuration
	*/
	cache := cache.New(&rs.Options{
		Addr:         fmt.Sprintf("%v:%v", env.RedisHost, env.RedisPort),
		Password:     env.RedisPassword,
		DB:           0,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	})

	/**
	jwtKey configuration
	*/
	jwtKey := []byte(env.JwtKey)

	/**
	googleOauth configuration
	*/
	gConf := sign.GoogleOauthConfig{
		ClientId:     env.GoogleOauthClientId,
		ClientSecret: env.GoogleOauthClientSecret,
	}

	/**
	Email service
	*/
	emailRequestChannel := make(chan mail.SendEmailOptions, 20)
	var emailSetupOptions *mail.EmailServiceOptions
	if env.EmailServiceEnabled {
		emailSetupOptions = &mail.EmailServiceOptions{
			Email:    env.EmailUserName,
			Password: env.EmailUserPassword,
			Server:   env.EmailServerAddr,
		}
	}

	go mail.StartEmailService(context.Background(), emailSetupOptions, emailRequestChannel)

	serverOptions := server.ServerOptions{
		Database:                 db,
		Cache:                    cache,
		JwtKey:                   jwtKey,
		UseHttps:                 env.UseHttps,
		BaseUrl:                  env.BaseUrl.String(),
		Domain:                   strings.Split(env.BaseUrl.Host, ":")[0],
		HtmlTemplate:             "./internal/template",
		GoogleOauthConf:          gConf,
		EmailVerificationIgnored: !env.EmailServiceEnabled,
		EmailRequest:             emailRequestChannel,
	}
	r := server.SetupServer(serverOptions) // blocking if starting with success
	log.Printf("Server is listening...")
	port := fmt.Sprintf(":%v", env.Port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
