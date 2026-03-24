package model

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID         int64  `json:"id"`
	Login      string `json:"login"`
	Password   string `json:"-"`
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	Patronymic string `json:"patronymic"`
	RoleID     int64  `json:"roleId"`
	RoleName   string `json:"role"`
}

// FullName returns "Фамилия И.О." format for display
func (u User) FullName() string {
	name := u.LastName
	if u.FirstName != "" {
		name += " " + string([]rune(u.FirstName)[:1]) + "."
	}
	if u.Patronymic != "" {
		name += string([]rune(u.Patronymic)[:1]) + "."
	}
	return name
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
