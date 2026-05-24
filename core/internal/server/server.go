package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/budgets/core/docs"
	"github.com/budgets/core/internal/config"
	"github.com/budgets/core/internal/currency"
	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/handler"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/repository"
	"github.com/budgets/core/internal/service"
)

type Dependencies struct {
	// Repositories
	UserRepo            repository.UserRepository
	GroupRepo           repository.BudgetingGroupRepository
	ParticipantRepo     repository.ParticipantRepository
	UserParticipantRepo repository.UserParticipantRepository
	CategoryRepo        repository.ExpenseCategoryRepository
	BudgetRepo          repository.BudgetRepository
	ExpectedExpenseRepo repository.ExpectedExpenseRepository
	ActualExpenseRepo   repository.ActualExpenseRepository
	PrefRepo            repository.UserPreferenceRepository

	// Services
	GroupService      service.GroupService
	CategoryService   service.CategoryService
	BudgetService     service.BudgetService
	ExpenseService    service.ExpenseService
	PreferenceService service.PreferenceService

	// Currency
	Marketplace *currency.CurrencyMarketplace
}

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

func New(cfg *config.Config, db *database.DB, deps *Dependencies, opts ...Option) *Server {
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

	s.setupRoutes(deps)
	return s
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

// BuildDependencies constructs all dependencies needed by the server.
// This function should be called before creating a new server instance.
func BuildDependencies(db *database.DB, encryptor *encryption.Encryptor) *Dependencies {
	userRepo := repository.NewUserRepository()
	groupRepo := repository.NewBudgetingGroupRepository()
	participantRepo := repository.NewParticipantRepository()
	userParticipantRepo := repository.NewUserParticipantRepository()
	categoryRepo := repository.NewExpenseCategoryRepository()
	budgetRepo := repository.NewBudgetRepository()
	expectedExpenseRepo := repository.NewExpectedExpenseRepository()
	actualExpenseRepo := repository.NewActualExpenseRepository()
	prefRepo := repository.NewUserPreferenceRepository()

	groupService := service.NewGroupService(db.Pool, groupRepo, participantRepo, userParticipantRepo)
	categoryService := service.NewCategoryService(db.Pool, categoryRepo, groupRepo, userParticipantRepo)
	budgetService := service.NewBudgetService(db.Pool, budgetRepo, groupRepo, userParticipantRepo)
	expenseService := service.NewExpenseService(db.Pool, encryptor, expectedExpenseRepo, actualExpenseRepo, budgetRepo, categoryRepo, userParticipantRepo)
	preferenceService := service.NewPreferenceService(db.Pool, prefRepo)

	exchangeProvider := currency.NewStubExchangeRateProvider()
	exchangeCache := currency.NewInMemoryCache()
	marketplace := currency.NewCurrencyMarketplace(exchangeProvider, exchangeCache)

	return &Dependencies{
		UserRepo:            userRepo,
		GroupRepo:           groupRepo,
		ParticipantRepo:     participantRepo,
		UserParticipantRepo: userParticipantRepo,
		CategoryRepo:        categoryRepo,
		BudgetRepo:          budgetRepo,
		ExpectedExpenseRepo: expectedExpenseRepo,
		ActualExpenseRepo:   actualExpenseRepo,
		PrefRepo:            prefRepo,
		GroupService:        groupService,
		CategoryService:     categoryService,
		BudgetService:       budgetService,
		ExpenseService:      expenseService,
		PreferenceService:   preferenceService,
		Marketplace:         marketplace,
	}
}

func (s *Server) setupRoutes(deps *Dependencies) {

	authMiddleware := s.authenticator
	userResolver := middleware.NewUserResolver(s.db.Pool, deps.UserRepo.GetOrCreateByProvider)

	groupHandler := handler.NewGroupHandler(deps.GroupService)
	categoryHandler := handler.NewCategoryHandler(deps.CategoryService)
	budgetHandler := handler.NewBudgetHandler(deps.BudgetService)
	expenseHandler := handler.NewExpenseHandler(deps.ExpenseService)
	preferenceHandler := handler.NewPreferenceHandler(deps.PreferenceService)
	currencyHandler := handler.NewCurrencyHandler(deps.Marketplace)

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
