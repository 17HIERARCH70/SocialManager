package services

import (
	"context"
	"github.com/17HIERARCH70/SocialManager/internal/domain/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceMethods interface {
	AuthenticateUser(email, password string) (*models.User, error)
	CreateUser(user *models.User) (int, error)
	GetAllUsers() ([]models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
}

type userService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) UserServiceMethods {
	return &userService{db: db}
}

func (s *userService) AuthenticateUser(email, password string) (*models.User, error) {
	user := &models.User{}
	// Query user by email
	err := s.db.QueryRow(context.Background(), "SELECT id, email, password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		return nil, err
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
	if err != nil {
		return nil, err // Password does not match
	}

	return user, nil
}

func (s *userService) CreateUser(user *models.User) (int, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PassHash), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Insert user into database
	var userId int
	err = s.db.QueryRow(context.Background(), "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id", user.Email, string(hashedPassword)).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil

}

func (s *userService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	rows, err := s.db.Query(context.Background(), "SELECT id, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(context.Background(), "SELECT id, email FROM users WHERE id = $1", id).Scan(&user.ID, &user.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUser(user *models.User) error {
	_, err := s.db.Exec(context.Background(), "UPDATE users SET email = $1, password = $2 WHERE id = $3", user.Email, user.PassHash, user.ID)
	return err
}

func (s *userService) DeleteUser(id string) error {
	_, err := s.db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
	return err
}
