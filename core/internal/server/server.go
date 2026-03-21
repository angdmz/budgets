package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/budgets/core/internal/config"
	_ "github.com/budgets/core/docs"
	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/handler"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/repository"
	"github.com/budgets/core/internal/service"
)

type Server struct {
	router *gin.Engine
	config *config.Config
	db     *database.DB
}

func New(cfg *config.Config, db *database.DB) *Server {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	s := &Server{
		router: router,
		config: cfg,
		db:     db,
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

	groupRepo := repository.NewBudgetingGroupRepository()
	participantRepo := repository.NewParticipantRepository()
	categoryRepo := repository.NewExpenseCategoryRepository()
	budgetRepo := repository.NewBudgetRepository()
	expectedExpenseRepo := repository.NewExpectedExpenseRepository()
	actualExpenseRepo := repository.NewActualExpenseRepository()

	groupService := service.NewGroupService(s.db.Pool, groupRepo, participantRepo)
	categoryService := service.NewCategoryService(s.db.Pool, categoryRepo, groupRepo, participantRepo)
	budgetService := service.NewBudgetService(s.db.Pool, budgetRepo, groupRepo, participantRepo)
	expenseService := service.NewExpenseService(s.db.Pool, encryptor, expectedExpenseRepo, actualExpenseRepo, budgetRepo, categoryRepo, participantRepo)

	authMiddleware := middleware.NewAuth0Middleware(s.config.Auth.Auth0Domain, s.config.Auth.Auth0Audience)

	groupHandler := handler.NewGroupHandler(groupService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	expenseHandler := handler.NewExpenseHandler(expenseService)

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
			protected.PUT("/expected-expenses/:id", expenseHandler.UpdateExpectedExpense)
			protected.DELETE("/expected-expenses/:id", expenseHandler.DeleteExpectedExpense)

			// Actual Expenses
			protected.PUT("/actual-expenses/:id", expenseHandler.UpdateActualExpense)
			protected.DELETE("/actual-expenses/:id", expenseHandler.DeleteActualExpense)
		}
	}
}
