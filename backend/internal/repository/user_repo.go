package repository

import (
	"database/sql"

	"shoe-store/internal/model"
)

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

// FindByLogin returns the user with the given login, joining with roles to populate RoleName.
// Returns (nil, nil) if no user is found.
func (r *UserRepo) FindByLogin(login string) (*model.User, error) {
	query := `
		SELECT u.id, u.login, u.password, u.last_name, u.first_name, u.patronymic,
		       u.role_id, r.name
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.login = ?`

	var u model.User
	err := r.DB.QueryRow(query, login).Scan(
		&u.ID, &u.Login, &u.Password,
		&u.LastName, &u.FirstName, &u.Patronymic,
		&u.RoleID, &u.RoleName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID returns the user with the given ID, joining with roles to populate RoleName.
// Returns (nil, nil) if no user is found.
func (r *UserRepo) FindByID(id int64) (*model.User, error) {
	query := `
		SELECT u.id, u.login, u.password, u.last_name, u.first_name, u.patronymic,
		       u.role_id, r.name
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = ?`

	var u model.User
	err := r.DB.QueryRow(query, id).Scan(
		&u.ID, &u.Login, &u.Password,
		&u.LastName, &u.FirstName, &u.Patronymic,
		&u.RoleID, &u.RoleName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
