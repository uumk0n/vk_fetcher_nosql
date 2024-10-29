package storage

import (
	"fmt"
	"lab4/internal/fetcher"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jStorage struct {
	driver neo4j.Driver
}

func NewNeo4jStorage(driver neo4j.Driver) *Neo4jStorage {
	return &Neo4jStorage{driver: driver}
}

// SaveData saves the VkData to Neo4j
func (s *Neo4jStorage) SaveData(data fetcher.VkData) {
	session := s.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	if err := s.saveUserNode(session, data.User); err != nil {
		log.Printf("Error saving user node: %v", err)
		return
	}

	if err := s.saveUserRelations(session, data.User, data.Followers, "FOLLOWS"); err != nil {
		log.Printf("Error saving user relations: %v", err)
	}

	if err := s.saveUserRelations(session, data.User, data.Subscriptions, "SUBSCRIBES_TO"); err != nil {
		log.Printf("Error saving subscription relations: %v", err)
	}

	if err := s.saveGroupRelations(session, data.User, data.Groups); err != nil {
		log.Printf("Error saving group relations: %v", err)
	}
}

func (s *Neo4jStorage) saveUserNode(session neo4j.Session, user fetcher.VkUser) error {
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MERGE (u:User {id: $id})
			ON CREATE SET u.screen_name = $screenName, u.name = $name, u.sex = $sex, u.city = $city
			ON MATCH SET u.screen_name = $screenName, u.name = $name, u.sex = $sex, u.city = $city`
		_, err := tx.Run(query, map[string]interface{}{
			"id":         user.ID,
			"screenName": user.ScreenName,
			"name":       user.FirstName,
			"sex":        user.Sex,
			"city":       user.City.Title,
		})
		return nil, err
	})
	return err
}

func (s *Neo4jStorage) saveUserRelations(session neo4j.Session, user fetcher.VkUser, connections []fetcher.VkUser, relation string) error {
	for _, connection := range connections {
		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			query := fmt.Sprintf(`
				MERGE (c:User {id: $connectionId})
				ON CREATE SET c.screen_name = $screenName, c.name = $name, c.sex = $sex, c.city = $city
				MERGE (u:User {id: $userId})
				MERGE (c)-[:%s]->(u)`, relation)
			_, err := tx.Run(query, map[string]interface{}{
				"connectionId": connection.ID,
				"screenName":   connection.ScreenName,
				"name":         connection.FirstName,
				"sex":          connection.Sex,
				"city":         connection.City.Title,
				"userId":       user.ID,
			})
			return nil, err
		})
		if err != nil {
			log.Printf("Error saving %s relation to Neo4j: %v", relation, err)
			return err
		}
	}
	return nil
}

func (s *Neo4jStorage) saveGroupRelations(session neo4j.Session, user fetcher.VkUser, groups []fetcher.VkGroup) error {
	for _, group := range groups {
		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			query := `
				MERGE (g:Group {id: $groupId})
				ON CREATE SET g.name = $name, g.screen_name = $screenName
				MERGE (u:User {id: $userId})
				MERGE (u)-[:SUBSCRIBES_TO]->(g)`
			_, err := tx.Run(query, map[string]interface{}{
				"groupId":    group.ID,
				"name":       group.Name,
				"screenName": group.ScreenName,
				"userId":     user.ID,
			})
			return nil, err
		})
		if err != nil {
			log.Printf("Error saving group %d to Neo4j: %v", group.ID, err)
			return err
		}
	}
	return nil
}
