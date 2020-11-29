package gitlab

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  Username  `json:"username"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	AvatarURL string    `json:"avatar_url"`
	WebURL    string    `json:"web_url"`
}

var cache = map[Username]*User{}

func (gl *GitlabService) FindUser(username Username) *User {
	user, ok := cache[username]
	if ok {
		return user
	}

	body, err := gl.callMethod(fmt.Sprintf("users?username=%s", username))
	if err != nil {
		return nil
	}

	var userList []User

	err = json.Unmarshal(body, &userList)
	if err != nil {
		return nil
	}

	for userIndex := range userList {
		cache[user.Username] = &userList[userIndex]
	}

	return cache[username]
}
