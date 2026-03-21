package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/budgets/core/internal/config"
	"github.com/budgets/core/internal/currency"
	_ "github.com/budgets/core/docs"
	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/handler"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/repository"
	"github.com/budgets/core/internal/service"
)

type Server struct {
	router        *gin.Engine
	config        *config.Config
	db            *database.DB
	authenticator middleware.Authenticator
}

// Option is a functional option for configuring the server.
type Option func(*Server)

// WithAuthenticator sets a custom authenticator (useful for testing).
func WithAuthenticator(auth middleware.Authenticator) Option {
	return func(s *Server) {
		s.authenticator = auth
	}
}

func New(cfg *config.Config, db *database.DB, opts ...Option) *Server {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	s := &Server{
		router: router,
		config: cfg,
		db:     db,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Default to Auth0 middleware if no authenticator was provided
	if s.authenticator == nil {
		s.authenticator = middleware.NewAuth0Middleware(cfg.Auth.Auth0Domain, cfg.Auth.Auth0Audience)
	}

	s.setupRoutes()
	return s
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) setupRoutes() {
	encryptor, err := encryption.NewEncryptor(s.config.Auth.EncryptionKey.Value())
	if err != nil {
		panic("failed to initialize encryptor: " + err.Error())
	}

	userRepo := repository.NewUserRepository()
	groupRepo := repository.NewBudgetingGroupRepository()
	participantRepo := repository.NewParticipantRepository()
	userParticipantRepo := repository.NewUserParticipantRepository()
	categoryRepo := repository.NewExpenseCategoryRepository()
	budgetRepo := repository.NewBudgetRepository()
	expectedExpenseRepo := repository.NewExpectedExpenseRepository()
	actualExpenseRepo := repository.NewActualExpenseRepository()
	prefRepo := repository.NewUserPreferenceRepository()

	groupService := service.NewGroupService(s.db.Pool, groupRepo, participantRepo, userParticipantRepo)
	categoryService := service.NewCategoryService(s.db.Pool, categoryRepo, groupRepo, userParticipantRepo)
	budgetService := service.NewBudgetService(s.db.Pool, budgetRepo, groupRepo, userParticipantRepo)
	expenseService := service.NewExpenseService(s.db.Pool, encryptor, expectedExpenseRepo, actualExpenseRepo, budgetRepo, categoryRepo, userParticipantRepo)
	preferenceService := service.NewPreferenceService(s.db.Pool, prefRepo)

	exchangeProvider := currency.NewStubExchangeRateProvider()
	exchangeCache := currency.NewInMemoryCache()
	marketplace := currency.NewCurrencyMarketplace(exchangeProvider, exchangeCache)

	authMiddleware := s.authenticator
	userResolver := middleware.NewUserResolver(s.db.Pool, userRepo.GetOrCreateByProvider)

	groupHandler := handler.NewGroupHandler(groupService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	preferenceHandler := handler.NewPreferenceHandler(preferenceService)
	currencyHandler := handler.NewCurrencyHandler(marketplace)

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Inject config into all requests for error handling
	s.router.Use(middleware.InjectConfig(s.config))

	api := s.router.Group("/api/v1")
	{
		// Auth0 handles authentication via frontend
		// No backend auth routes needed - just validate JWTs

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		protected.Use(userResolver.ResolveUser())
		{
			// Groups
			protected.POST("/groups", groupHandler.CreateGroup)
			protected.GET("/groups", groupHandler.GetGroups)
			protected.POST("/groups/:id/categories", categoryHandler.CreateCategory)
			protected.GET("/groups/:id/categories", categoryHandler.GetCategories)
			protected.POST("/groups/:id/budgets", budgetHandler.CreateBudget)
			protected.GET("/groups/:id/budgets", budgetHandler.GetBudgets)
			protected.GET("/groups/:id", groupHandler.GetGroup)
			protected.PUT("/groups/:id", groupHandler.UpdateGroup)
			protected.DELETE("/groups/:id", groupHandler.DeleteGroup)

			// Categories
			protected.PUT("/categories/:id", categoryHandler.UpdateCategory)
			protected.DELETE("/categories/:id", categoryHandler.DeleteCategory)

			// Budgets
			protected.POST("/budgets/:id/expected-expenses", expenseHandler.CreateExpectedExpense)
			protected.GET("/budgets/:id/expected-expenses", expenseHandler.GetExpectedExpenses)
			protected.POST("/budgets/:id/actual-expenses", expenseHandler.CreateActualExpense)
			protected.GET("/budgets/:id/actual-expenses", expenseHandler.GetActualExpenses)
			protected.GET("/budgets/:id", budgetHandler.GetBudget)
			protected.PUT("/budgets/:id", budgetHandler.UpdateBudget)
			protected.DELETE("/budgets/:id", budgetHandler.DeleteBudget)

			// Expected Expenses
			protected.GET("/expected-expenses/:id", expenseHandler.GetExpectedExpense)
			protected.PUT("/expected-expenses/:id", expenseHandler.UpdateExpectedExpense)
			protected.DELETE("/expected-expenses/:id", expenseHandler.DeleteExpectedExpense)

			// Actual Expenses
			protected.GET("/actual-expenses/:id", expenseHandler.GetActualExpense)
			protected.PUT("/actual-expenses/:id", expenseHandler.UpdateActualExpense)
			protected.DELETE("/actual-expenses/:id", expenseHandler.DeleteActualExpense)

			// User Preferences
			protected.GET("/preferences", preferenceHandler.GetPreferences)
			protected.PUT("/preferences", preferenceHandler.UpdatePreferences)

			// Currency
			protected.POST("/currency/convert", currencyHandler.Convert)
			protected.GET("/currency/rates", currencyHandler.GetExchangeRates)
		}
	}
}
