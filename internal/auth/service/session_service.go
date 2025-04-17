package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

// SessionService handles session operations
type SessionService interface {
	GetStore() sessions.Store
	Get(ctx context.Context, key string) (interface{}, error)
	Set(r *http.Request, w http.ResponseWriter, key string, value interface{}) error
	Delete(r *http.Request, w http.ResponseWriter, key string) error
}

type sessionService struct {
	config *types.Config
	repo   repository.Repository
	logger *zap.Logger
	store  sessions.Store
}

// NewSessionService creates a new session service
func NewSessionService(cfg *types.Config, repo repository.Repository, logger *zap.Logger) SessionService {
	// Create a secure random key for sessions
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		logger.Fatal("failed to generate session key", zap.Error(err))
	}

	store := sessions.NewCookieStore(sessionKey)
	store.Options = &sessions.Options{
		Path:     cfg.Cookie.Path,
		Domain:   cfg.Cookie.Domain,
		MaxAge:   int(cfg.JWT.RefreshTokenTTL.Seconds()),
		Secure:   cfg.Cookie.Secure,
		HttpOnly: true,
	}

	return &sessionService{
		config: cfg,
		repo:   repo,
		logger: logger,
		store:  store,
	}
}

// GetStore returns the session store
func (s *sessionService) GetStore() sessions.Store {
	return s.store
}

// Get retrieves a value from the session
func (s *sessionService) Get(ctx context.Context, key string) (interface{}, error) {
	session, err := s.repo.GetSession(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal(session.Value, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session value: %w", err)
	}

	return value, nil
}

// Set stores a value in the session
func (s *sessionService) Set(r *http.Request, w http.ResponseWriter, key string, value interface{}) error {
	// Store in database
	expiresAt := time.Now().Add(s.config.JWT.RefreshTokenTTL)
	if err := s.repo.StoreSession(r.Context(), key, value, expiresAt); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Also store in cookie session for Goth compatibility
	session, err := s.store.Get(r, StateSessionName)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	session.Values[key] = value
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// Delete removes a value from the session
func (s *sessionService) Delete(r *http.Request, w http.ResponseWriter, key string) error {
	// Delete from database
	if err := s.repo.DeleteSession(r.Context(), key); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Also delete from cookie session
	session, err := s.store.Get(r, StateSessionName)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	delete(session.Values, key)
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}
