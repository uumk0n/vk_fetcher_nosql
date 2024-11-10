package fetcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	vkAPIBaseURL  = "https://api.vk.com/method/"
	apiVersion    = "5.131"
	rateLimitWait = 350 * time.Millisecond
)

type VkUser struct {
	ID         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ScreenName string `json:"screen_name"`
	Sex        int    `json:"sex"`
	City       struct {
		Title string `json:"title"`
	} `json:"city"`
}

type VkGroup struct {
	ID         int
	Name       string
	ScreenName string
}

type VkData struct {
	User          VkUser
	Followers     []VkUser
	Subscriptions []VkUser
	Groups        []VkGroup
}

type VKFetcher struct {
	AccessToken string
}

func NewVKFetcher(accessToken string) *VKFetcher {
	return &VKFetcher{AccessToken: accessToken}
}

// FetchData retrieves VK user data with specified depth
func (f *VKFetcher) FetchData(userID string, depth int) (VkData, error) {
	usersData := make(map[int]VkUser)
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return VkData{}, err
	}
	err = f.fetchRecursive(userIDInt, depth, usersData)
	if err != nil {
		return VkData{}, err
	}
	return f.mapToVkData(usersData), nil
}

func (f *VKFetcher) fetchRecursive(userID int, depth int, usersData map[int]VkUser) error {
	if depth <= 0 {
		return nil
	}

	user, err := f.fetchUserData(userID)
	if err != nil {
		return fmt.Errorf("error fetching user data for user %d: %w", userID, err)
	}
	usersData[user.ID] = user

	followers, err := f.fetchUsers("users.getFollowers", user.ID)
	if err != nil {
		return err
	}

	subscriptions, err := f.fetchUsers("users.getSubscriptions", user.ID)
	if err != nil {
		return err
	}

	for _, followerID := range followers {
		if _, exists := usersData[followerID]; !exists {
			follower, err := f.fetchUserData(followerID)
			fmt.Println("follower", follower)
			if err != nil {
				return err
			}
			usersData[followerID] = follower
			if err := f.fetchRecursive(followerID, depth-1, usersData); err != nil {
				return err
			}
		}
	}
	for _, sId := range subscriptions {
		if _, exists := usersData[sId]; !exists {
			s, err := f.fetchUserData(sId)
			fmt.Println("s", s)
			if err != nil {
				return err
			}
			usersData[sId] = s
			if err := f.fetchRecursive(sId, depth-1, usersData); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *VKFetcher) fetchUserData(userID int) (VkUser, error) {
	// Modify this to return VkUser based on the userID
	params := fmt.Sprintf("user_ids=%d&fields=screen_name,sex,city&access_token=%s&v=%s", userID, f.AccessToken, apiVersion)
	url := fmt.Sprintf("%susers.get?%s", vkAPIBaseURL, params)
	var response struct {
		Response []VkUser `json:"response"`
	}

	if err := f.makeVKRequest(url, &response); err != nil || len(response.Response) == 0 {
		return VkUser{}, fmt.Errorf("error fetching user data: %w", err)
	}
	return response.Response[0], nil
}

func (f *VKFetcher) fetchUsers(method string, userID int) ([]int, error) {
	params := fmt.Sprintf("user_id=%d&access_token=%s&v=%s", userID, f.AccessToken, apiVersion)
	url := fmt.Sprintf("%s%s?%s", vkAPIBaseURL, method, params)
	var response struct {
		Response struct {
			Items []int `json:"items"` // Change to []int to match the response structure
		} `json:"response"`
	}

	if err := f.makeVKRequest(url, &response); err != nil {
		log.Printf("Error fetching users for method %s: %v", method, err)
		return nil, err
	}
	return response.Response.Items, nil
}

func (f *VKFetcher) makeVKRequest(url string, result interface{}) error {
	time.Sleep(rateLimitWait)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (f *VKFetcher) mapToVkData(usersData map[int]VkUser) VkData {
	var vkData VkData
	for _, user := range usersData {
		if vkData.User.ID == 0 {
			vkData.User = user // Assign the first user as the main user
		}
		vkData.Followers = append(vkData.Followers, user)
	}
	return vkData
}
