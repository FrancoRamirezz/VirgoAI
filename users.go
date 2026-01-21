package models

import (
	"backend/utils"
	"errors"
	"regexp"
	"time"
)

// for users were here asking what does your users look like
type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Provider      string    `json:"provider"`
	ProviderID    string    `json:"providerId,omitempty"`
	Created       time.Time `json:"createdAt"` // when was user created
	EmailVerified bool      `json:"emailVerified"`
	Role          string    `json:"role"` // this will handle both teachers and students roles
	Upated        time.Time `json:"updated"`
}

// note we they reqister we want to
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

// login struct
// here were take the two options
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// here we would use the logic that takes two into the struct oldpassword and newpassword
type ChangePasswordRequest struct {
	Oldpassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}
type ForgotPasswordRequest struct {
	Email string `json:"email"`
	// we will add SendGrid to this
}
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}
type AuditLog struct {
	ID            int       `json:"id"`
	UserID        *int      `json:"userId,omitempty"` // Nullable
	Action        string    `json:"action"`
	IPAddress     string    `json:"ipAddress"`
	UserAgent     string    `json:"userAgent"`
	Success       bool      `json:"success"`
	FailureReason string    `json:"failureReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// this below would be from the stripe payment
type StripePayment struct {
	PaymentID      int       `json:"paymentid"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	PaymentCreated time.Time `json:"PaymentCreated"`
}

// sendgrid struct

// here we will make a registration function that will check a few items
// what if the email is empty, or password is empty or if the first name and lastname are empty
// Validate registration request
func (request *RegisterRequest) Validate() error {
	// Check required fields for the user
	if request.Email == "" || request.Password == "" || request.FirstName == "" || request.LastName == "" {
		return errors.New("all fields are required")
	}

	// here we validate the email
	if err := utils.ValidateEmail(request.Email); err != nil {
		return err
	}
	// now we check if the request is either teacher and student
	if request.Role != "student" && request.Role != "teacher" {
		return errors.New("invalid role: must be either student or teacher")
	}
	//
	if len(request.Password) < 8 {
		return errors.New("Password must be longer than 8")
	}
	// check for email validation
	emailCheck := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// check if the email is empty
	if !emailCheck.MatchString(request.Email) {
		return errors.New("email is not valid")
	}

	return nil

}

// here we point to the login struct that contains email and password
func (checkrequest *LoginRequest) Validate() error {
	if checkrequest.Email == "" || checkrequest.Password == "" {
		return errors.New("email and password are required")
	}
	// do a quick check of the email
	if err := utils.ValidateEmail(checkrequest.Email); err != nil {
		return err
	}
	return nil
}

// This happens after the user signs up
// now we check if they can change their password.
func (passwordrequest *ChangePasswordRequest) Validate() error {
	if passwordrequest.Oldpassword == "" || passwordrequest.NewPassword == "" {
		return errors.New("old password or new password are required")
	}
	// if the user makes a new password then we need the new users to stay consistent with the password measurement
	if len(passwordrequest.NewPassword) < 8 {
		return errors.New("new password must be atleast 8 characters")
	}
	// we also need to check if the new password matches the old password
	if passwordrequest.Oldpassword == passwordrequest.NewPassword {
		return errors.New("passwords must be different from each other")
	}
	return nil
}
