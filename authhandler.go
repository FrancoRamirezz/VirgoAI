// backend/handlers/auth_handlers.go
package handlers

import (
	"backend/config"
	"backend/database"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for auth operations
type AuthHandler struct {
	db *database.Postgres // Use your postgres instance
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *database.Postgres) *AuthHandler {
	return &AuthHandler{db: db}
}

// ============================================
// 1. TRADITIONAL EMAIL/PASSWORD AUTHENTICATION
// ============================================

// RegisterHandler creates a new user account with email/password
// POST /api/auth/register
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Registration request received")

	// Step 1: Parse incoming JSON
	// Step 2: make a var called req that calls the users struct
	// look at the struct in the users files
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Step 2: Validate all required fields and format. This now includes the
	if err := req.Validate(); err != nil {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Step 3: Check if email already exists using the new DB method
	_, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		// User exists
		utils.ErrorResponseJSON(w, http.StatusConflict, "Email already registered")
		return
	}

	// Step 4: Hash the password using the bcrypt
	// Note the we are also passing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), config.BcryptCost)
	if err != nil {
		log.Printf("Password hashing failed: %v", err)
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "Server error")
		return
	}

	// Step 5: Create user role. Note we want the user connection to have multiple connections
	user, err := h.db.CreateUser(
		r.Context(),
		req.Email,
		string(hashedPassword),
		req.FirstName,
		req.LastName,
		req.Role, // this role will be used for student or teacher, can add more such as payment users
		"local",  // Provider
		"",       // ProviderID (empty for local users)
	)
	log.Print("Registration was successful")

	if err != nil {
		log.Printf("Failed to create user: %v", err)
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	// Step 6: Log the registration in audit log
	h.db.CreateAuditLog(
		r.Context(),
		&user.ID,
		"register",
		utils.GetIPAddress(r),
		r.UserAgent(),
		true,
		"",
	)

	log.Printf("Congrats users, you registered successfully: %s", user.Email)

	// Step 7: Send success response
	utils.ResponseJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Registration successful! Please check your email to verify your account.",
		"user":    user,
	})
}

// LoginHandler authenticates user with email/password and creates session
// POST /api/auth/login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(" Login request received")

	// Step 1: Parse login credentials
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Step 2: Validate required fields
	if req.Email == "" || req.Password == "" {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Step 3: Find user in database using new DB method
	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Log failed login attempt
		h.db.CreateAuditLog(
			r.Context(),
			nil,
			"login_failed",
			utils.GetIPAddress(r),
			r.UserAgent(),
			false,
			"user not found",
		)
		utils.ErrorResponseJSON(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Step 4: Check if user is local provider
	if user.Provider != "local" {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Please use OAuth login for this account")
		return
	}

	// Step 5: Verify password matches stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Log failed login attempt
		h.db.CreateAuditLog(
			r.Context(),
			&user.ID,
			"login_failed",
			utils.GetIPAddress(r),
			r.UserAgent(),
			false,
			"invalid password",
		)
		utils.ErrorResponseJSON(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Step 6: Optional - Check if email is verified
	// if !user.EmailVerified {
	//     utils.RespondWithError(w, http.StatusForbidden, "Please verify your email before logging in")
	//     return
	// }

	// Step 7: Create session
	session, _ := config.GetSessionStore().Get(r, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["email"] = user.Email
	session.Values["provider"] = user.Provider

	if err := session.Save(r, w); err != nil {
		log.Printf("❌ Failed to save session: %v", err)
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Step 8: Log successful login
	h.db.CreateAuditLog(
		r.Context(),
		&user.ID,
		"login",
		utils.GetIPAddress(r),
		r.UserAgent(),
		true,
		"",
	)

	log.Printf(" User logged in successfully: %s", user.Email)

	// Don't send password hash to client
	user.PasswordHash = ""

	utils.ResponseJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})
}

// ============================================
// 2. OAUTH AUTHENTICATION
// ============================================

// BeginAuthHandler initiates OAuth flow
// GET /auth/{provider}
func (h *AuthHandler) BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	log.Printf("🔑 Starting OAuth flow with provider: %s", provider)

	q := r.URL.Query()
	q.Add("provider", provider)
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
}

// CallbackHandler processes OAuth callback
// GET /auth/{provider}/callback
func (h *AuthHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	log.Printf("🔄 OAuth callback received from: %s", provider)

	q := r.URL.Query()
	q.Add("provider", provider)
	r.URL.RawQuery = q.Encode()

	// Step 1: Complete OAuth flow
	// this comes from the OAUth config flow
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Printf("❌ OAuth authentication failed: %v", err)
		http.Redirect(w, r, config.GetFrontendURL()+"/Login?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	log.Printf("✅ OAuth user data received: %s (%s)", gothUser.Email, gothUser.Name)

	// Step 2: Check if user exists using new DB method
	user, err := h.db.GetUserByProviderID(r.Context(), provider, gothUser.UserID)

	// Step 3: Create new user if they don't exist
	if err != nil {
		log.Printf("👤 Creating new user from %s OAuth: %s", provider, gothUser.Email)

		firstName := gothUser.FirstName
		lastName := gothUser.LastName
		if firstName == "" {
			firstName = gothUser.Name
		}

		// Create user using new DB method
		user, err = h.db.CreateUser(
			r.Context(),
			gothUser.Email,
			"",
			firstName,
			lastName,
			"student",       //  Default role for OAuth users
			provider,        //  "google", "github", etc.
			gothUser.UserID, //  OAuth provider's user ID
		)
		if err != nil {
			log.Printf("❌ Failed to create OAuth user: %v", err)
			http.Redirect(w, r, config.GetFrontendURL()+"/Login?error=create_failed", http.StatusTemporaryRedirect)
			return
		}

		// Verify email automatically for OAuth users
		h.db.VerifyEmail(r.Context(), user.ID)

		// Log OAuth registration
		h.db.CreateAuditLog(
			r.Context(),
			&user.ID,
			"oauth_register",
			utils.GetIPAddress(r),
			r.UserAgent(),
			true,
			"",
		)
	}

	// Step 4: Create session
	session, _ := config.GetSessionStore().Get(r, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["email"] = user.Email
	session.Values["provider"] = provider

	if err := session.Save(r, w); err != nil {
		log.Printf("❌ Failed to save OAuth session: %v", err)
		http.Redirect(w, r, config.GetFrontendURL()+"/Login?error=session_failed", http.StatusTemporaryRedirect)
		return
	}

	// Log OAuth login
	h.db.CreateAuditLog(
		r.Context(),
		&user.ID,
		"oauth_login",
		utils.GetIPAddress(r),
		r.UserAgent(),
		true,
		"",
	)

	log.Printf("✅ OAuth login successful: %s via %s", user.Email, provider)

	http.Redirect(w, r, config.GetFrontendURL()+"/auth/callback?Login=success", http.StatusTemporaryRedirect)
}

// ============================================
// 3. SESSION MANAGEMENT
// ============================================

// GetCurrentUserHandler returns logged-in user's information
// GET /api/auth/me
func (h *AuthHandler) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "auth-session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		utils.ErrorResponseJSON(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Fetch user using new DB method
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("❌ Failed to fetch user: %v", err)
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "User not found")
		return
	}

	// Don't send password hash
	user.PasswordHash = ""

	utils.ResponseJSON(w, http.StatusOK, user)
}

// LogoutHandler terminates user session
// POST /api/auth/logout
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("👋 Logout request received")

	session, _ := config.GetSessionStore().Get(r, "auth-session")

	// Log logout before clearing session
	if userID, ok := session.Values["user_id"].(int); ok {
		h.db.CreateAuditLog(
			r.Context(),
			&userID,
			"logout",
			utils.GetIPAddress(r),
			r.UserAgent(),
			true,
			"",
		)
	}

	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		log.Printf("⚠️  Failed to clear session: %v", err)
	}

	gothic.Logout(w, r)

	log.Println("✅ User logged out successfully")

	utils.ResponseJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// ============================================
// 4. PASSWORD MANAGEMENT
// ============================================

// ChangePasswordHandler updates user's password
// POST /api/auth/change-password
func (h *AuthHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔑 Password change request received")

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Oldpassword == "" || req.NewPassword == "" {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Both old and new passwords are required")
		return
	}

	if len(req.NewPassword) < 8 {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "New password must be at least 8 characters")
		return
	}

	session, _ := config.GetSessionStore().Get(r, "auth-session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		utils.ErrorResponseJSON(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user using new DB method
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "Server error")
		return
	}

	if user.Provider != "local" {
		utils.ErrorResponseJSON(w, http.StatusBadRequest, "Cannot change password for OAuth accounts")
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Oldpassword)); err != nil {
		utils.ErrorResponseJSON(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), config.BcryptCost)
	if err != nil {
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "Server error")
		return
	}

	// Update password using new DB method
	if err := h.db.UpdatePassword(r.Context(), user.Email, string(newHash)); err != nil {
		log.Printf("❌ Failed to update password: %v", err)
		utils.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// Log password change
	h.db.CreateAuditLog(
		r.Context(),
		&userID,
		"password_changed",
		utils.GetIPAddress(r),
		r.UserAgent(),
		true,
		"",
	)

	log.Printf("✅ Password changed for user ID: %d", userID)

	utils.ResponseJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// ForgotPasswordHandler initiates password reset
// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📧 Forgot password request received")

	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Email == "" {
		utils.ResponseJSON(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Check if user exists
	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err == nil && user.Provider == "local" {
		// Generate reset token
		token := utils.GenerateSecureToken(32)
		expiresAt := time.Now().Add(15 * time.Minute)

		// Save token using new DB method
		if err := h.db.CreatePasswordResetToken(r.Context(), user.ID, token, expiresAt); err != nil {
			log.Printf("❌ Failed to create reset token: %v", err)
		} else {
			// TODO: Send email with reset link
			// utils.SendPasswordResetEmail(user.Email, token)
			log.Printf("📧 Password reset token created for: %s", req.Email)
		}
	}

	// Always return success (don't reveal if email exists)
	utils.ResponseJSON(w, http.StatusOK, map[string]string{
		"message": "If an account exists with this email, you will receive password reset instructions.",
	})
}

// ResetPasswordHandler resets password using token
// POST /api/auth/reset-password
func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔓 Password reset request received")

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		utils.ResponseJSON(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	if len(req.NewPassword) < 8 {
		utils.ResponseJSON(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Validate token using new DB method
	userID, expiresAt, used, err := h.db.GetPasswordResetToken(r.Context(), req.Token)
	if err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	if used {
		utils.ResponseJSON(w, http.StatusBadRequest, "Token has already been used")
		return
	}

	if time.Now().After(expiresAt) {
		utils.ResponseJSON(w, http.StatusBadRequest, "Token has expired")
		return
	}

	// Get user
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, "User not found")
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), config.BcryptCost)
	if err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, "Server error")
		return
	}

	// Update password
	if err := h.db.UpdatePassword(r.Context(), user.Email, string(newHash)); err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	// Mark token as used
	h.db.MarkPasswordResetTokenAsUsed(r.Context(), req.Token)

	// Log password reset
	h.db.CreateAuditLog(
		r.Context(),
		&userID,
		"password_reset",
		utils.GetIPAddress(r),
		r.UserAgent(),
		true,
		"",
	)

	log.Println("✅ Password reset successfully")

	utils.ResponseJSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully. You can now login with your new password.",
	})
}
