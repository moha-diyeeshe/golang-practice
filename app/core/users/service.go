package users

import (
	"database/sql"
	"time"
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"strings"
)


type UserService struct {

	DB *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{DB: db}
}

func (s *UserService) EnsureUserTable() error {

	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT UNIQUE,
		password TEXT NOT NULL,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	);
	`

	_, err := s.DB.Exec(query)
	return err
}

func (s *UserService) CreateUser(user *User) error {
	fmt.Println("🔍 Creating user:", user)
	rawPassword := strings.TrimSpace(user.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	query := `INSERT INTO users (username, email, phone, password, is_active, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = s.DB.QueryRow(query, user.Username, user.Email, user.Phone, user.Password, user.IsActive, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	return err
}

func (s *UserService) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, phone, is_active, created_at, updated_at FROM users WHERE id = $1`
	err := s.DB.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(user *User) error {
	query := `UPDATE users SET username = $1, email = $2, phone = $3, is_active = $4, updated_at = $5 WHERE id = $6`
	_, err := s.DB.Exec(query, user.Username, user.Email, user.Phone, user.IsActive, user.UpdatedAt, user.ID)
	return err
}

func (s *UserService) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.DB.Exec(query, id)
	return err
}

func (s *UserService) ListUsers() ([]*User, error) {
	query := `SELECT id, username, email, phone, is_active, created_at, updated_at FROM users`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *UserService) DeactivateUser(id int) error {
	query := `UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2`
	_, err := s.DB.Exec(query, time.Now(), id)
	return err
}

func (s *UserService) ActivateUser(id int) error {
	query := `UPDATE users SET is_active = true, updated_at = $1 WHERE id = $2`
	_, err := s.DB.Exec(query, time.Now(), id)
	return err
}



// func (s *UserService) Login(email, password string) (*User, error) {

// 	user := &User{}

// 	query := `SELECT id, username, email, password, is_active FROM users WHERE email=$1`

// 	err := s.DB.QueryRow(query, email).Scan(
// 		&user.ID,
// 		&user.Username,
// 		&user.Email,
// 		&user.Password,
// 		&user.IsActive,
// 	)

// 	if err != nil {
// 		return nil, fmt.Errorf("invalid credentials")
// 	}


	

// 	// check password
// 	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid credentials")
// 	}

// 	// optional: check active user
// 	if !user.IsActive {
// 		return nil, fmt.Errorf("user inactive")
// 	}

// 	return user, nil
// }

func (s *UserService) Login(email, password string) (*User, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	user := &User{}

	query := `
		SELECT id, username, email, password, is_active 
		FROM users 
		WHERE email=$1
	`

	fmt.Println("🔍 LOGIN ATTEMPT")
	fmt.Println("Email:", email)

	err := s.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsActive,
	)

	// ❌ user not found OR DB error
	if err != nil {
		fmt.Println("❌ DB ERROR:", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	fmt.Println("✅ USER FOUND")
	fmt.Printf("DB Email: [%s]\n", user.Email)
	fmt.Printf("DB Password Raw: [%s]\n", user.Password)

	// 🔥 CLEAN HASH (IMPORTANT FIX)
	hashedPassword := strings.TrimSpace(user.Password)

	fmt.Println("🔐 Comparing password...")

	// check password
	err = bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)

	if err != nil {
		fmt.Println("❌ PASSWORD MISMATCH:", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	fmt.Println("✅ PASSWORD MATCH SUCCESS")

	// check active user
	if !user.IsActive {
		fmt.Println("❌ USER IS INACTIVE")
		return nil, fmt.Errorf("user inactive")
	}

	fmt.Println("🎉 LOGIN SUCCESS")

	return user, nil
}