package session

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/alexedwards/scs/v2"
)

// SQLiteSessionStore is a custom implementation of scs.Store backed by our
// existing *sql.DB (modernc.org/sqlite). The off-the-shelf scs/sqlite3store
// depends on mattn/go-sqlite3 (CGo) and cannot be used here.
type SQLiteSessionStore struct {
	db          *sql.DB
	stopCleanup chan struct{}
}

// NewSQLiteSessionStore creates a session store and starts a background goroutine
// that purges expired sessions every cleanupInterval.
func NewSQLiteSessionStore(db *sql.DB, cleanupInterval time.Duration) *SQLiteSessionStore {
	st := &SQLiteSessionStore{
		db:          db,
		stopCleanup: make(chan struct{}),
	}
	go st.runCleanup(cleanupInterval)
	return st
}

// Find returns the session data for a token. ok is false when the session does
// not exist or has expired.
func (st *SQLiteSessionStore) Find(token string) ([]byte, bool, error) {
	row := st.db.QueryRow(
		`SELECT data FROM sessions WHERE token = ? AND expiry > ?`,
		token, float64(time.Now().UnixNano())/1e9,
	)
	var data []byte
	err := row.Scan(&data)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// Commit upserts session data with a Unix epoch expiry.
func (st *SQLiteSessionStore) Commit(token string, b []byte, expiry time.Time) error {
	_, err := st.db.Exec(
		`INSERT INTO sessions (token, data, expiry) VALUES (?, ?, ?)
		 ON CONFLICT(token) DO UPDATE SET data=excluded.data, expiry=excluded.expiry`,
		token, b, float64(expiry.UnixNano())/1e9,
	)
	return err
}

// Delete removes a session by token.
func (st *SQLiteSessionStore) Delete(token string) error {
	_, err := st.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// StopCleanup signals the background cleanup goroutine to exit.
func (st *SQLiteSessionStore) StopCleanup() {
	close(st.stopCleanup)
}

func (st *SQLiteSessionStore) runCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if _, err := st.db.Exec(
				`DELETE FROM sessions WHERE expiry < ?`,
				float64(time.Now().UnixNano())/1e9,
			); err != nil {
				log.Printf("session cleanup error: %v", err)
			}
		case <-st.stopCleanup:
			return
		}
	}
}

// NewManager constructs and configures an SCS session manager.
func NewManager(db *sql.DB, secret string, lifetime time.Duration) *scs.SessionManager {
	mgr := scs.New()
	mgr.Store = NewSQLiteSessionStore(db, 5*time.Minute)
	mgr.Lifetime = lifetime
	mgr.Cookie.Name = "kasku_session"
	mgr.Cookie.HttpOnly = true
	mgr.Cookie.SameSite = 1 // http.SameSiteStrictMode
	mgr.Cookie.Secure = false
	return mgr
}

// context key type to avoid collisions
type contextKey string

const sessionManagerKey contextKey = "sessionManager"

// WithManager stores the session manager in a context (convenience helper).
func WithManager(ctx context.Context, mgr *scs.SessionManager) context.Context {
	return context.WithValue(ctx, sessionManagerKey, mgr)
}
