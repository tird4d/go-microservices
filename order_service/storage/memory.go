package storage

type User struct {
	ID    string
	Email string
	Name  string
}

var users = make(map[string]User)

func SaveUser(u User) {
	users[u.ID] = u
}

func GetUser(id string) (User, bool) {
	u, ok := users[id]
	return u, ok
}
