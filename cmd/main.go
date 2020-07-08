package main

import (
	"context"
	"fmt"
	rs "github.com/go-redis/redis"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	ch "url-shortener/internal/cache"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/route/user/sign"
	"url-shortener/internal/server"
	"url-shortener/internal/service/mail"
)

func periodicallyCheckRedis(r ch.Redis, err chan error) {
	for {
		if e := r.Ping(); e != nil {
			err <- e
		}
		time.Sleep(5 * time.Second)
	}
}

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
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Warning: unable to close mysql connection properly | Reason: %v\n", err)
		}
	}()

	/**
	Caching configuration
	*/
	cache := ch.New(&rs.Options{
		Addr:         fmt.Sprintf("%v:%v", env.RedisHost, env.RedisPort),
		Password:     env.RedisPassword,
		DB:           0,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  250 * time.Second,
		MinIdleConns: 1,
	})
	redisErr := make(chan error) // return true indicates something is wrong
	go periodicallyCheckRedis(cache, redisErr)
	defer func() {
		if err := cache.Close(); err != nil {
			log.Printf("Warning: unable to close redis connection properly | Reason: %v\n", err)
		}
	}()

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

	serverErr := make(chan error) // return true indicates something is wrong
	go func(serverErr chan<- error) {
		r := server.SetupServer(serverOptions)
		log.Printf("Server is listening...")
		port := fmt.Sprintf(":%v", env.Port)
		if err := r.Run(port); err != nil {
			serverErr <- err
		}
	}(serverErr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	terminated := false
	for !terminated {
		select {
		case <-done:
			log.Printf("Gracefully shutting down...\n")
			terminated = true
		case err := <-serverErr:
			log.Fatalf("Unable to start server: %v\n", err)
		case err := <-redisErr:
			log.Fatalf("Connection check with redis failed: %v\n", err)
		default:
		}
	}
}
