package models

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             	int
	FirstName		string
	LastName		string
	Email          	string
	HashedPassword 	[]byte
	Created        	time.Time
}


type UserModel struct {
	DB *sql.DB
}



func generateResetPasswordCode () (string,error) {
	digits :="0123456789"
	passwordLength := 6
	
	var resetCode []byte

	for i:= 0;i<passwordLength;i++ {
		num,err := rand.Int(rand.Reader,big.NewInt(int64(len(digits))))
		if err != nil {
			return "",err
		}
		resetCode = append(resetCode, digits[num.Int64()])
	}

	return string(resetCode),nil
}

func sendEmailVerificationCode(recipientEmail, verificationCode string) error {
	// Set up authentication information.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	senderEmail := os.Getenv("MAIL_SENDER")
	senderPassword := os.Getenv("MAIL_PASSWORD")


	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)

	// Compose the message.
	subject := "Subject: Email Verification Code\n"
	body := fmt.Sprintf("From CODEVAULT\nYour verification code is: %s", verificationCode)
	message := []byte(subject + "\n" + body)

	// Send the email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		return err
	}

	fmt.Println("Verification code sent to:", recipientEmail)
	return nil
}


 func (m *UserModel) GetUser(id int) (*User, error) {
	stmt := `SELECT id, first_name, last_name, email FROM users WHERE id = $1`

	row := m.DB.QueryRow(stmt, id)
	u := &User{}

	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}


func (m *UserModel) Insert(firstName, lastName, email, password string) error {
	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	fullName := firstName + " " + lastName

	stmt := `INSERT INTO users (first_name, last_name, full_name, email, hashed_password, created) VALUES($1, $2, $3, $4, $5, NOW())`

	_, err = m.DB.Exec(stmt, firstName, lastName, fullName, email, hashedPassword)
	if err != nil {
		// Check if the error is a PostgreSQL-specific error
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			// Check for unique violation error
			if pgErr.Code == "23505" && strings.Contains(pgErr.Message, "users_email_key") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}


func (m *UserModel) UpdateUserProfile(userId int, firstName, lastName, email string) error {
	stmt := `UPDATE users SET first_name = $1, last_name = $2, email = $3 WHERE id = $4`

	// Execute the query
	_, err := m.DB.Exec(stmt, firstName, lastName, email, userId)
	if err != nil {
		// Check if the error is a PostgreSQL-specific error
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			// Check for unique violation error (duplicate email)
			if pgErr.Code == "23505" && strings.Contains(pgErr.Message, "users_email_key") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}


func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = $1"
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)

	if err != nil {
		// Handle no matching rows case
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Compare the hashed password with the plain-text password provided
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		// Handle password mismatch case
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}


func (m *UserModel) IsPasswordCorrect(id int ,password string) (int, error){
	var userId int
	var hashedPassword []byte

// RETRIEVE ID AND HASHED PASSWORD ASSOCIATED WITH GIVEN EMAIL IF NO MATCHING EMAIL RETURN ERRINVALIDCREDENTIAL ERROR

	stmt := "SELECT id, hashed_password FROM users WHERE id = ?"

	err := m.DB.QueryRow(stmt,id).Scan(&userId, &hashedPassword)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return 0, ErrInvalidCredentials
		}else{
			return 0,err
		}
	}

// CHECK WHETHER THE HASHED PASSWORD AND PLAIN TEXT PASSWORD PROVIDED MATCH IF THEY DONT RETURN ERRINVALIDCREDENTIAL ERROR
err = bcrypt.CompareHashAndPassword(hashedPassword,[]byte(password))
if err != nil {
	if errors.Is(err,bcrypt.ErrMismatchedHashAndPassword){
		return 0, ErrInvalidCredentials
	}else {
		return 0, err
	}
}
	return userId, nil
}


 func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"

	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}


func (m *UserModel) IsEmailExists(email string) (bool, int, error) {
	var exists bool
	var userID int

	// Use PostgreSQL's placeholder syntax
	stmt := "SELECT id FROM users WHERE email = $1"

	// Execute the query
	err := m.DB.QueryRow(stmt, email).Scan(&userID)

	if err == nil {
		// Email exists in the database
		exists = true
	} else if err == sql.ErrNoRows {
		// No matching user found
		exists = false
		userID = 0
		err = nil
	} else {
		// Unexpected error
		return false, 0, err
	}

	return exists, userID, nil
}


func (m *UserModel) ResetPassword(id int, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// Update the hashed password
	stmt := `UPDATE users SET hashed_password = $1 WHERE id = $2`
	_, err = m.DB.Exec(stmt, hashedPassword, id)
	if err != nil {
		return err
	}

	// Clear the reset code
	stmt = `UPDATE users SET reset_code = NULL WHERE id = $1`
	_, err = m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}


 func (m *UserModel) SendResetCode(userId int, email string) error {
	// Generate a reset password code
	resetCode, err := generateResetPasswordCode()
	if err != nil {
		return err
	}

	// Send the email with the reset code
	err = sendEmailVerificationCode(email, resetCode)
	if err != nil {
		return err
	}

	// Update the user record with the reset code
	stmt := `UPDATE users SET reset_code = $1 WHERE id = $2`
	_, err = m.DB.Exec(stmt, resetCode, userId)
	if err != nil {
		return err
	}

	return nil
}



func (m *UserModel) VerifyResetCode(id int, code string) error {
	var resetCode string

	// Query to retrieve the reset code for the specified user ID
	stmt := `SELECT reset_code FROM users WHERE id = $1`

	// Execute the query and scan the result into resetCode
	err := m.DB.QueryRow(stmt, id).Scan(&resetCode)
	if err != nil {
		return err
	}

	// Compare the provided code with the stored reset code
	if subtle.ConstantTimeCompare([]byte(resetCode), []byte(code)) != 1 {
		return errors.New("wrong code")
	}

	return nil
}





