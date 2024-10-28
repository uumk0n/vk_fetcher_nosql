package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var accessToken string
var driver neo4j.DriverWithContext // Изменено на DriverWithContext

func init() {
	godotenv.Load()
	accessToken = os.Getenv("ACCESS_TOKEN")

	// Подключение к Neo4j
	var err error
	driver, err = neo4j.NewDriverWithContext(
		os.Getenv("NEO4J_URI"),
		neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""),
	)
	if err != nil {
		log.Fatal("Ошибка при подключении к Neo4j:", err)
	}
}

// Функция для запроса к VK API
func vkRequest(method string, params map[string]string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.vk.com/method/%s?access_token=%s&v=5.131", method, accessToken)
	for k, v := range params {
		url += fmt.Sprintf("&%s=%s", k, v)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при запросе к VK API: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if _, ok := result["error"]; ok {
		return nil, fmt.Errorf("Ошибка VK API: %v", result["error"])
	}

	return result["response"].(map[string]interface{}), nil
}

// Сохранение пользователя в Neo4j
func saveUser(ctx context.Context, user map[string]interface{}) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, `
			MERGE (u:User {id: $id})
			SET u.name = $name, u.screen_name = $screen_name, u.sex = $sex, u.home_town = $home_town
		`, map[string]interface{}{
			"id":          user["id"],
			"name":        user["first_name"].(string) + " " + user["last_name"].(string),
			"screen_name": user["screen_name"],
			"sex":         user["sex"],
			"home_town":   user["home_town"],
		})
		return nil, err
	})
	return err
}

// Рекурсивная функция для получения данных
func fetchData(ctx context.Context, userID string, level, maxLevel int) {
	if level > maxLevel {
		return
	}

	log.Printf("Загрузка данных для пользователя %s на уровне %d\n", userID, level)
	userInfo, err := vkRequest("users.get", map[string]string{"user_ids": userID, "fields": "city,screen_name,home_town,sex"})
	if err != nil {
		log.Println("Ошибка при получении информации о пользователе:", err)
		return
	}

	// Извлечение и сохранение данных пользователя
	user := userInfo["items"].([]interface{})[0].(map[string]interface{})
	if err := saveUser(ctx, user); err != nil {
		log.Println("Ошибка при сохранении пользователя:", err)
	}

	// Получение подписчиков и подписок аналогично получению данных пользователя
}

func main() {
	ctx := context.Background()
	defer driver.Close(ctx) // Закрытие драйвера с использованием контекста

	userID := "12345" // Пример ID пользователя
	fetchData(ctx, userID, 1, 2)
}
