package database

//Go: a statically typed, compiled language developed by Google.
//PostgreSQL: a powerful, open-source relational database management system.
//pgx: the PostgreSQL Driver for Go, which provides a way to interact with PostgreSQL databases from Go code.
//CRUD: Create, Read, Update, Delete, which are the basic operations that can be performed on a database.

/// plz check for the db go
import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// this makes it global var connection pool
// connection pool = multiple reusable database connection better than one
// here we will make a struct to handle mutiple concurrent calls to the db
// safer for thread usage
// for go make sure the name of the struct is upper case because then they can be exported
type Postgres struct {
	db *pgxpool.Pool
}

// Uses sync.Once to guarantee the connection pool is initialized exactly once, even with concurrent access
var (
	pgInstance *Postgres
	pgoneConn  sync.Once
)

// intialize the db
func Newinit(ctx context.Context, connString string) (*Postgres, error) {
	pgoneConn.Do(func() {
		// In addition, a config struct can be created by ParseConfig and modified before establishing the connection with ConnectConfig to configure settings such as tracing that cannot be configured with a connection string.
		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			log.Fatal("Unable to parse  connection string%w", err)
			return
		}
		// here we set how many connections are allowed and for how long
		// we only get 4 max connections, default timeouts, default settings
		// No customization
		config.MaxConns = 25
		config.MinConns = 5
		config.MaxConnLifetime = time.Hour         // Connections live max 1 hour
		config.MaxConnIdleTime = 30 * time.Minute  // Idle connections closed after 30min
		config.HealthCheckPeriod = 1 * time.Minute // Idle connection

		// here we will create the connection pool, use new method to ensure one connections is being proccessed
		//The *pgx.Conn returned by pgx.Connect() represents a single connection and is not concurrency safe.
		// This is entirely appropriate for a simple command line example such as above. However, for many uses, such as a web application server, concurrency is required.
		// To use a connection pool replace the import github.com/jackc/pgx/v5 with github.com/jackc/pgx/v5/pgxpool and connect with pgxpool.New() instead of pgx.Connect().
		///
		db, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			log.Fatal("unable to create connection pool:%w", err)
			return
		}
		// test the connection using the ping method on the database
		//this is found from the article: this forces GO to make a conection
		if err := db.Ping(ctx); err != nil {
			log.Fatal("unable to ping%w", err)
			return
		}

		// 	create a new instance of posgres struct and get its memory address
		pgInstance = &Postgres{db: db}
		log.Print("Database connected")

	})
	return pgInstance, nil
}

// Ping checks if database is reachable
// the db ping forces us to connect to our database
func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

// this is the same as from the documentation closing the databse file
// the database defer db.Close() is from the documentation
func (pg *Postgres) Close() {
	pg.db.Close()
}

// here we will implement the schema for the project
func (pg *Postgres) CreateTables(ctx context.Context) error {
	log.Println("📋 Creating database tables...")

	// Users table
	// note this comes from users models and mirror them
	// users:= Create Table users( ID, email, passwordhash, Note: this means that the parameters must be in the same order)
	//
	usersTable := `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'student',  -- ✅ ADD THIS
    provider VARCHAR(50) DEFAULT 'local',
    provider_id VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_provider_user UNIQUE (provider, provider_id)
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role); 
`
	// check the edge case such that the user is empty
	if _, err := pg.db.Exec(ctx, usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	log.Println(" Users table ready")

	// Password reset tokens table
	resetTokensTable := `
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(255) NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_reset_tokens_token ON password_reset_tokens(token);
	CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON password_reset_tokens(user_id);
	CREATE INDEX IF NOT EXISTS idx_reset_tokens_expires ON password_reset_tokens(expires_at);
	`

	if _, err := pg.db.Exec(ctx, resetTokensTable); err != nil {
		return fmt.Errorf("failed to create password_reset_tokens table: %w", err)
	}
	log.Println(" Password reset tokens table ready")

	// Sessions table
	sessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(128) PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		data TEXT,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
	`

	if _, err := pg.db.Exec(ctx, sessionsTable); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}
	log.Println(" Sessions table ready")

	// Audit log table
	auditLogTable := `
	CREATE TABLE IF NOT EXISTS audit_log (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
		action VARCHAR(50) NOT NULL,
		ip_address VARCHAR(45),
		user_agent TEXT,
		success BOOLEAN,
		failure_reason VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);
	CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
	`

	if _, err := pg.db.Exec(ctx, auditLogTable); err != nil {
		return fmt.Errorf("failed to create audit_log table: %w", err)
	}
	log.Println(" Audit log table ready")

	log.Println("All tables created successfully")
	return nil
}

// here we will create the user. Note: the user will have the attributed from the methods
func (pg *Postgres) CreateUser(ctx context.Context, email, passwordHash, firstName, lastName, role, provider, providerID string) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, role, provider, provider_id, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, email, first_name, last_name, role, provider, provider_id, email_verified, created_at, updated_at
	`

	var user User
	err := pg.db.QueryRow(ctx, query, email, passwordHash, firstName, lastName, role, provider, providerID, false).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Provider,
		&user.ProviderID,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("user with email %s already exists", email)
		}
		return nil, fmt.Errorf("unable to create user: %w", err)
	}

	log.Printf("✅ Created user: %s (ID: %d, Role: %s)", user.Email, user.ID, user.Role)
	return &user, nil
}

// here i will refer to the users file for models
// file has the models
func (pg *Postgres) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, provider, provider_id, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := pg.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Provider,
		&user.ProviderID,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("unable to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (pg *Postgres) GetUserByID(ctx context.Context, userID int) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, provider, provider_id, email_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := pg.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Provider,
		&user.ProviderID,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("unable to get user: %w", err)
	}

	return &user, nil
}

// GetUserByProviderID retrieves a user by OAuth provider and provider ID
func (pg *Postgres) GetUserByProviderID(ctx context.Context, provider, providerID string) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, provider, provider_id, email_verified, created_at, updated_at
		FROM users
		WHERE provider = $1 AND provider_id = $2
	`

	var user User
	err := pg.db.QueryRow(ctx, query, provider, providerID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Provider,
		&user.ProviderID,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("unable to get user: %w", err)
	}

	return &user, nil
}

// UpdatePassword updates a user's password
func (pg *Postgres) UpdatePassword(ctx context.Context, email, newPasswordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE email = $2
	`

	result, err := pg.db.Exec(ctx, query, newPasswordHash, email)
	if err != nil {
		return fmt.Errorf("unable to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	log.Printf(" Password updated for user: %s", email)
	return nil
}

// UpdateUser updates user information
func (pg *Postgres) UpdateUser(ctx context.Context, userID int, firstName, lastName string) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := pg.db.Exec(ctx, query, firstName, lastName, userID)
	if err != nil {
		return fmt.Errorf("unable to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	log.Printf("✅ Updated user ID: %d", userID)
	return nil
}

// VerifyEmail marks a user's email as verified
func (pg *Postgres) VerifyEmail(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET email_verified = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := pg.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("unable to verify email: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	log.Printf(" Email verified for user ID: %d", userID)
	return nil
}

// DeleteUser deletes a user by ID
func (pg *Postgres) DeleteUser(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := pg.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	log.Printf(" Deleted user ID: %d", userID)
	return nil
}

// ListUsers retrieves all users with pagination
func (pg *Postgres) ListUsers(ctx context.Context, limit, offset int) ([]User, error) {
	query := `
		SELECT id, email, first_name, last_name, provider, provider_id, email_verified, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := pg.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("unable to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Provider,
			&user.ProviderID,
			&user.EmailVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// CountUsers returns the total number of users
func (pg *Postgres) CountUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := pg.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to count users: %w", err)
	}

	return count, nil
}

// BulkInsertUsers inserts multiple users using batch operations
// Note: this is  for 100s-1000s o
func (pg *Postgres) BulkInsertUsers(ctx context.Context, users []User) error {
	query := `INSERT INTO users (email, first_name, last_name, provider, provider_id) VALUES ($1, $2, $3, $4, $5)`
	batch := &pgx.Batch{} // this is insertion part of

	for _, user := range users {
		batch.Queue(query, user.Email, user.FirstName, user.LastName, user.Provider, user.ProviderID)
	}

	results := pg.db.SendBatch(ctx, batch)
	defer results.Close()

	for _, user := range users {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
				log.Printf("⚠️  User %s already exists, skipping", user.Email)
				continue
			}
			return fmt.Errorf("unable to insert user %s: %w", user.Email, err)
		}
	}

	return results.Close()
}

// CopyInsertUsers inserts multiple users using COPY protocol
// Fastest method for bulk inserts (10,000+ rows)
// Note: COPY doesn't handle constraint violations gracefully
func (pg *Postgres) CopyInsertUsers(ctx context.Context, users []User) error {
	entries := [][]any{}
	columns := []string{"email", "first_name", "last_name", "provider", "provider_id"}
	tableName := "users"

	for _, user := range users {
		entries = append(entries, []any{user.Email, user.FirstName, user.LastName, user.Provider, user.ProviderID})
	}

	rowsAffected, err := pg.db.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(entries),
	)
	if err != nil {
		return fmt.Errorf("error copying into %s table: %w", tableName, err)
	}

	log.Printf(" Inserted %d users using COPY", rowsAffected)
	return nil
}

// ============================================
// PASSWORD RESET TOKEN OPERATIONS
// ============================================

// CreatePasswordResetToken creates a new password reset token
func (pg *Postgres) CreatePasswordResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used)
		VALUES ($1, $2, $3, false)
	`

	_, err := pg.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("unable to create password reset token: %w", err)
	}

	log.Printf(" Created password reset token for user ID: %d", userID)
	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (pg *Postgres) GetPasswordResetToken(ctx context.Context, token string) (int, time.Time, bool, error) {
	query := `
		SELECT user_id, expires_at, used
		FROM password_reset_tokens
		WHERE token = $1
	`

	var userID int
	var expiresAt time.Time
	var used bool

	err := pg.db.QueryRow(ctx, query, token).Scan(&userID, &expiresAt, &used)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, time.Time{}, false, fmt.Errorf("token not found")
		}
		return 0, time.Time{}, false, fmt.Errorf("unable to get token: %w", err)
	}

	return userID, expiresAt, used, nil
}

// MarkPasswordResetTokenAsUsed marks a token as used
func (pg *Postgres) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE token = $1
	`

	result, err := pg.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("unable to mark token as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// DeleteExpiredPasswordResetTokens deletes expired tokens (cleanup)
func (pg *Postgres) DeleteExpiredPasswordResetTokens(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < CURRENT_TIMESTAMP OR used = true
	`

	result, err := pg.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to delete expired tokens: %w", err)
	}

	log.Printf(" Deleted %d expired/used password reset tokens", result.RowsAffected())
	return nil
}

// ============================================
// AUDIT LOG OPERATIONS
// ============================================

// CreateAuditLog creates a new audit log entry
func (pg *Postgres) CreateAuditLog(ctx context.Context, userID *int, action, ipAddress, userAgent string, success bool, failureReason string) error {
	query := `
		INSERT INTO audit_log (user_id, action, ip_address, user_agent, success, failure_reason)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := pg.db.Exec(ctx, query, userID, action, ipAddress, userAgent, success, failureReason)
	if err != nil {
		return fmt.Errorf("unable to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (pg *Postgres) GetAuditLogsByUser(ctx context.Context, userID int, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, action, ip_address, user_agent, success, failure_reason, created_at
		FROM audit_log
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := pg.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("unable to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.IPAddress,
			&log.UserAgent,
			&log.Success,
			&log.FailureReason,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// here we allocate how long the session would be
func (pg *Postgres) CreateSession(ctx context.Context, sessionID string, userID int, data string, expiresAt time.Time) error {
	query := `
		INSERT INTO sessions (id, user_id, data, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET data = $3, expires_at = $4
	`

	_, err := pg.db.Exec(ctx, query, sessionID, userID, data, expiresAt)
	if err != nil {
		return fmt.Errorf("unable to create session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by ID
func (pg *Postgres) GetSession(ctx context.Context, sessionID string) (int, string, time.Time, error) {
	query := `
		SELECT user_id, data, expires_at
		FROM sessions
		WHERE id = $1 AND expires_at > CURRENT_TIMESTAMP
	`

	var userID int
	var data string
	var expiresAt time.Time

	err := pg.db.QueryRow(ctx, query, sessionID).Scan(&userID, &data, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", time.Time{}, fmt.Errorf("session not found or expired")
		}
		return 0, "", time.Time{}, fmt.Errorf("unable to get session: %w", err)
	}

	return userID, data, expiresAt, nil
}

// DeleteSession deletes a session
func (pg *Postgres) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := pg.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("unable to delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions deletes expired sessions (cleanup)
func (pg *Postgres) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`

	result, err := pg.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to delete expired sessions: %w", err)
	}

	log.Printf("✅ Deleted %d expired sessions", result.RowsAffected())
	return nil
}

// ============================================
// QUERY HELPERS
// ============================================

// QueryRow executes a query that returns a single row
// Use for custom queries not covered by methods above
func (pg *Postgres) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return pg.db.QueryRow(ctx, sql, args...)
}

// Query executes a query that returns multiple rows
// Use for custom queries not covered by methods above
// take a look into this
func (pg *Postgres) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return pg.db.Query(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows (INSERT, UPDATE, DELETE)
// Use for custom queries not covered by methods above
func (pg *Postgres) Exec(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	result, err := pg.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// ============================================
// TRANSACTION HELPERS
// ============================================

// BeginTx starts a new transaction
func (pg *Postgres) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return pg.db.Begin(ctx)
}

// WithTransaction executes a function within a transaction
// Automatically handles commit/rollback
//
// Example usage:
//
//	err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
//	    _, err := tx.Exec(ctx, "INSERT INTO users ...")
//	    if err != nil { return err }
//	    _, err = tx.Exec(ctx, "INSERT INTO profiles ...")
//	    return err
//	})
func (pg *Postgres) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			tx.Rollback(ctx)
		}
	}()

	err = fn(tx)
	if err == nil {
		err = tx.Commit(ctx)
	}

	return err
}

// ============================================
// HEALTH CHECK & MONITORING
// ============================================

// HealthCheck returns database health status
func (pg *Postgres) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"database": "unknown",
		"pool":     map[string]interface{}{},
	}

	// Check connection
	if err := pg.db.Ping(ctx); err != nil {
		health["database"] = "unhealthy"
		health["error"] = err.Error()
		return health
	}

	health["database"] = "healthy"

	// Get pool stats
	stats := pg.db.Stat()
	health["pool"] = map[string]interface{}{
		"total_conns":      stats.TotalConns(),
		"acquired_conns":   stats.AcquiredConns(),
		"idle_conns":       stats.IdleConns(),
		"max_conns":        stats.MaxConns(),
		"acquire_count":    stats.AcquireCount(),
		"acquire_duration": stats.AcquireDuration().String(),
	}

	return health
}

// GetStats returns connection pool statistics
func (pg *Postgres) GetStats() *pgxpool.Stat {
	return pg.db.Stat()
}

// show a current overview of how the
func (pg *Postgres) LogStats() {
	stats := pg.db.Stat()
	log.Printf("📊 Database Pool Stats:")
	log.Printf("   Total connections: %d", stats.TotalConns())
	log.Printf("   Acquired connections: %d", stats.AcquiredConns())
	log.Printf("   Idle connections: %d", stats.IdleConns())
	log.Printf("   Max connections: %d", stats.MaxConns())
	log.Printf("   Acquire count: %d", stats.AcquireCount())
	log.Printf("   Acquire duration: %v", stats.AcquireDuration())
}

// once we move past the mvp phase then we move these over to the models users files and call for everytime we call the users in this files call the utils
// User represents a user in the database
type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Role          string    `json:"role"` // ✅ ADD THIS
	Provider      string    `json:"provider"`
	ProviderID    string    `json:"providerId"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AuditLog represents an audit log entry
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

/*
import (
	"context"
	"log"
	"os"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// i have not set a password once i do make a var called password and assgin it

// var (db *pgxpool.Pool)
func Init(ctx context.Context, connstring string, db *pgxpool.Pool) error {
	//Step 1: lets create a single connection pool
	db, err := pgxpool.New(ctx, connstring)
	if err != nil {
		 return fmt.Errorf("unable to parse connection string: %w", err)
	}
	//Step 2:  test if there is connection
	if err := db.Ping(ctx); err != nil{
		return fmt.Errorf("unable to ping database: %w", err)
	}
 // Step 3: close the database at the end
	defer db.Close()

	return nil
*/
