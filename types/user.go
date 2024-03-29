package types

type User struct {
	ID       int64  `json:"id,omitempty"`
	FullName string `json:"firstName,omitempty"`
	Password string `json:"-"`
	Email    string
}

type UserUpdateRequest struct {
	ID       int64
	FullName string
}

func CreateUser(id int64, fullName string, email string, password string) *User {
	return &User{
		ID:       id,
		FullName: fullName,
		Email:    email,
		Password: password,
	}
}

func CopyUser(user *User) *User {
	return CreateUser(user.ID, user.FullName, user.Email, user.Password)
}
