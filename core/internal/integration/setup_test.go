package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/budgets/core/internal/config"
	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/secrets"
	"github.com/budgets/core/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestSuite struct {
	Router      *gin.Engine
	DB          *pgxpool.Pool
	Encryptor   *encryption.Encryptor
	AuthToken   string
	TestUserID  string
}

func SetupTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	gin.SetMode(gin.TestMode)

	// Load config
	provider := secrets.GetProvider()
	cfg, err := config.Load(provider)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create encryptor
	enc, err := encryption.NewEncryptor(cfg.Auth.EncryptionKey.Value())
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	// Create server
	srv := server.New(cfg, db)

	// Generate test auth token
	testUserID := "test-user-123"
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.JWTSecret.Value())
	testUser := &domain.User{
		ExternalProviderID: testUserID,
		Email:              "test@example.com",
		DisplayName:        "Test User",
		AuthProvider:       "google",
	}
	authToken, err := authMiddleware.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate auth token: %v", err)
	}

	return &TestSuite{
		Router:     srv.Router(),
		DB:         db.Pool,
		Encryptor:  enc,
		AuthToken:  authToken,
		TestUserID: testUserID,
	}
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	t.Helper()
	if ts.DB != nil {
		ts.DB.Close()
	}
}

func (ts *TestSuite) AuthHeader() http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+ts.AuthToken)
	return header
}

func (ts *TestSuite) DoRequest(method, path string, body interface{}, headers http.Header) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	return w
}

func (ts *TestSuite) Get(path string) *httptest.ResponseRecorder {
	return ts.DoRequest(http.MethodGet, path, nil, ts.AuthHeader())
}

func (ts *TestSuite) Post(path string, body interface{}) *httptest.ResponseRecorder {
	return ts.DoRequest(http.MethodPost, path, body, ts.AuthHeader())
}

func (ts *TestSuite) Put(path string, body interface{}) *httptest.ResponseRecorder {
	return ts.DoRequest(http.MethodPut, path, body, ts.AuthHeader())
}

func (ts *TestSuite) Delete(path string) *httptest.ResponseRecorder {
	return ts.DoRequest(http.MethodDelete, path, nil, ts.AuthHeader())
}

func (ts *TestSuite) CleanupTestData(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// Clean up test data in reverse order of dependencies
	tables := []string{
		"actual_expenses",
		"expected_expenses",
		"budgets",
		"expense_categories",
		"user_participants",
		"participants",
		"budgeting_groups",
		"users",
	}

	for _, table := range tables {
		_, err := ts.DB.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE external_id IS NOT NULL", table))
		if err != nil {
			t.Logf("Warning: failed to clean up %s: %v", table, err)
		}
	}
}
