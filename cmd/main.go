package main

import (
	"fmt"
	"lab4/internal/fetcher"
	"lab4/internal/storage"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	neo4jUri      = "bolt://localhost:7687"
	vkAPIBaseURL  = "https://api.vk.com/method/"
	apiVersion    = "5.131"
	rateLimitWait = 350 * time.Millisecond
)

func main() {
	loadEnv()
	accessToken := os.Getenv("ACCESS_TOKEN")

	driver, err := neo4j.NewDriver(
		fmt.Sprintf("bolt://%s:%s", os.Getenv("NEO4J_HOST"), os.Getenv("NEO4J_BOLT_PORT")),
		neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""),
	)

	if err != nil {
		log.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	userId := "162400179"

	service := storage.NewNeo4jStorage(driver)
	err = service.CreateSchema()

	vkFetcher := fetcher.NewVKFetcher(accessToken)

	data, err := vkFetcher.FetchData(userId, 2)
	if err != nil {
		log.Fatalf("Error fetching VK data: %v", err)
	}

	service.SaveData(data)

	fmt.Println("Data saved to Neo4j")

	service.PrintSavedData()
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
