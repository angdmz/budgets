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
	"github.com/jackc/pgx/v5/pgxpool"
)

// Dependencies holds all external dependencies needed by the server.
type Dependencies struct {
	Encryptor          *encryption.Encryptor
	GroupHandler       *handler.GroupHandler
	CategoryHandler    *handler.CategoryHandler
	BudgetHandler      *handler.BudgetHandler
	ExpenseHandler     *handler.ExpenseHandler
	PreferenceHandler  *handler.PreferenceHandler
	CurrencyHandler    *handler.CurrencyHandler
	InvitationHandler  *handler.InvitationHandler
	UserResolver       middleware.UserResolver
}

// BuildDependencies creates the Dependencies struct with all handlers and middleware.
func BuildDependencies(pool *pgxpool.Pool, enc *encryption.Encryptor) Dependencies {
	userRepo := repository.NewUserRepository()
	prefRepo := repository.NewUserPreferenceRepository()
	preferenceService := service.NewPreferenceService(pool, prefRepo)

	exchangeProvider := currency.NewStubExchangeRateProvider()
	exchangeCache := currency.NewInMemoryCache()
	marketplace := currency.NewCurrencyMarketplace(exchangeProvider, exchangeCache)

	userResolver := middleware.NewUserResolver(pool, userRepo.GetOrCreateByProvider)

	return Dependencies{
		Encryptor:          enc,
		GroupHandler:       handler.NewGroupHandler(pool),
		CategoryHandler:    handler.NewCategoryHandler(pool),
		BudgetHandler:      handler.NewBudgetHandler(pool),
		ExpenseHandler:     handler.NewExpenseHandler(pool, enc),
		PreferenceHandler:  handler.NewPreferenceHandler(preferenceService),
		CurrencyHandler:    handler.NewCurrencyHandler(marketplace),
		InvitationHandler:  handler.NewInvitationHandler(pool),
		UserResolver:       userResolver,
	}
}

type Server struct {
	router        *gin.Engine
	config        *config.Config
	db            *database.DB
	deps          Dependencies
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

func New(cfg *config.Config, db *database.DB, deps Dependencies, opts ...Option) *Server {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	s := &Server{
		router: router,
		config: cfg,
		db:     db,
		deps:   deps,
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
	groupHandler := s.deps.GroupHandler
	categoryHandler := s.deps.CategoryHandler
	budgetHandler := s.deps.BudgetHandler
	expenseHandler := s.deps.ExpenseHandler
	preferenceHandler := s.deps.PreferenceHandler
	currencyHandler := s.deps.CurrencyHandler
	invitationHandler := s.deps.InvitationHandler
	userResolver := s.deps.UserResolver

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

		public := api.Group("")
		{
			public.GET("/invitations/token/:token", invitationHandler.GetInvitationByToken)
		}

		protected := api.Group("")
		protected.Use(s.authenticator.RequireAuth())
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
			protected.POST("/budgets/:budget_id/expected-expenses", expenseHandler.CreateExpectedExpense)
			protected.GET("/budgets/:budget_id/expected-expenses", expenseHandler.GetExpectedExpenses)
			protected.POST("/budgets/:budget_id/actual-expenses", expenseHandler.CreateActualExpense)
			protected.GET("/budgets/:budget_id/actual-expenses", expenseHandler.GetActualExpenses)
			protected.GET("/budgets/:budget_id", budgetHandler.GetBudget)
			protected.PUT("/budgets/:budget_id", budgetHandler.UpdateBudget)
			protected.DELETE("/budgets/:budget_id", budgetHandler.DeleteBudget)

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

			// Invitations
			protected.POST("/groups/:id/invitations", invitationHandler.CreateInvitation)
			protected.GET("/groups/:id/invitations", invitationHandler.ListInvitations)
			protected.DELETE("/invitations/:id", invitationHandler.RevokeInvitation)
			protected.POST("/invitations/token/:token/accept", invitationHandler.AcceptInvitation)
		}
	}
}
