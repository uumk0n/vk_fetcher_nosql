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

	driver, err := neo4j.NewDriver(neo4jUri, neo4j.BasicAuth("neo4j", "YOUR_PASSWORD", ""))
	if err != nil {
		log.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	userId := "your_initial_user_id"

	vkFetcher := fetcher.NewVKFetcher(accessToken)

	data, err := vkFetcher.FetchData(userId, 2)
	if err != nil {
		log.Fatalf("Error fetching VK data: %v", err)
	}

	service := storage.NewNeo4jStorage(driver)

	service.SaveData(data)

	fmt.Println("Data saved to Neo4j")
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
