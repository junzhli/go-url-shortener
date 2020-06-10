package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	url2 "net/url"
	"os"
	"reflect"
)

type Env struct {
	DBUser  string
	DBPass  string
	DBHost  string
	DBPort  string
	JwtKey  string
	BaseUrl *url2.URL
	Port    string
}

func ReadEnv() Env {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Unable to read .env | Reason: %v ...skipped\n", err)
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Printf("DB_USER is empty. Default as \"root\"\n")
		dbUser = "root"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		log.Printf("DB_PASSWORD is empty\n")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Printf("DB_HOST is empty. Default as \"localhost\"\n")
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_Port")
	if dbPort == "" {
		log.Printf("DB_PORT is empty. Default as \"3306\"\n")
		dbPort = "3306"
	}

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Printf("JWT_KEY is empty. Default as \"testKey\"\n")
		jwtKey = "testKey"
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Printf("PORT is empty. Default as \"8080\"\n")
		port = "8080"
	}

	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		log.Printf("BASE_URL is empty. Default as \"http://url-shortener.com:%v\"\n", port)
		baseUrl = fmt.Sprintf("http://url-shortener.com:%v", port)
	}

	u, err := url2.ParseRequestURI(baseUrl)
	if err != nil {
		panic("Invalid baseUrl")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		panic("Invalid baseUrl")
	}

	env := Env{
		DBUser:  dbUser,
		DBPass:  dbPass,
		DBHost:  dbHost,
		DBPort:  dbPort,
		JwtKey:  jwtKey,
		BaseUrl: u,
		Port:    port,
	}

	fmt.Printf("===========================\n")
	v := reflect.ValueOf(env)
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%v: %v\n", v.Type().Field(i).Name, v.Field(i).Interface())
	}
	fmt.Printf("===========================\n")

	return env
}
