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

	authMiddleware := middleware.NewAuthMiddleware(s.config.Auth.JWTSecret.Value())

	groupHandler := handler.NewGroupHandler(groupService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	authHandler := handler.NewAuthHandler(
		s.config.Auth.GoogleClientID,
		s.config.Auth.GoogleClientSecret.Value(),
		"http://localhost:8080/api/v1/auth/google/callback",
		authMiddleware,
	)

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Inject config into all requests for error handling
	s.router.Use(middleware.InjectConfig(s.config))

	api := s.router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/google/login", authHandler.GoogleLogin)
			auth.GET("/google/callback", authHandler.GoogleCallback)
			auth.GET("/me", authMiddleware.RequireAuth(), authHandler.GetCurrentUser)
		}

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			groups := protected.Group("/groups")
			{
				groups.POST("", groupHandler.CreateGroup)
				groups.GET("", groupHandler.GetGroups)
				groups.GET("/:id", groupHandler.GetGroup)
				groups.PUT("/:id", groupHandler.UpdateGroup)
				groups.DELETE("/:id", groupHandler.DeleteGroup)

				groups.POST("/:group_id/categories", categoryHandler.CreateCategory)
				groups.GET("/:group_id/categories", categoryHandler.GetCategories)

				groups.POST("/:group_id/budgets", budgetHandler.CreateBudget)
				groups.GET("/:group_id/budgets", budgetHandler.GetBudgets)
			}

			categories := protected.Group("/categories")
			{
				categories.PUT("/:id", categoryHandler.UpdateCategory)
				categories.DELETE("/:id", categoryHandler.DeleteCategory)
			}

			budgets := protected.Group("/budgets")
			{
				budgets.GET("/:id", budgetHandler.GetBudget)
				budgets.PUT("/:id", budgetHandler.UpdateBudget)
				budgets.DELETE("/:id", budgetHandler.DeleteBudget)

				budgets.POST("/:budget_id/expected-expenses", expenseHandler.CreateExpectedExpense)
				budgets.GET("/:budget_id/expected-expenses", expenseHandler.GetExpectedExpenses)

				budgets.POST("/:budget_id/actual-expenses", expenseHandler.CreateActualExpense)
				budgets.GET("/:budget_id/actual-expenses", expenseHandler.GetActualExpenses)
			}

			expectedExpenses := protected.Group("/expected-expenses")
			{
				expectedExpenses.PUT("/:id", expenseHandler.UpdateExpectedExpense)
				expectedExpenses.DELETE("/:id", expenseHandler.DeleteExpectedExpense)
			}

			actualExpenses := protected.Group("/actual-expenses")
			{
				actualExpenses.PUT("/:id", expenseHandler.UpdateActualExpense)
				actualExpenses.DELETE("/:id", expenseHandler.DeleteActualExpense)
			}
		}
	}
}
