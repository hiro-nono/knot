package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"knot-api/internal/config"
	"knot-api/internal/controller"
	infraai "knot-api/internal/infrastructure/ai"
	"knot-api/internal/infrastructure/auth/supabase"
	infradb "knot-api/internal/infrastructure/db"
	"knot-api/internal/infrastructure/http/cors"
	"knot-api/internal/infrastructure/http/csrf"
	"knot-api/internal/router"
	"knot-api/internal/usecase"
)

const (
	migrationsPath = "internal/infrastructure/db/migration"
	aiHTTPTimeout  = 60 * time.Second
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	conn, err := infradb.New(cfg.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := infradb.Migrate(conn, migrationsPath); err != nil {
		log.Fatal(err)
	}

	aiClient := infraai.NewClient(cfg.OrcaRouterBaseURL, cfg.OrcaRouterAPIKey, &http.Client{Timeout: aiHTTPTimeout})

	informationRepo := infradb.NewInformationRepository(conn)
	sourceRepo := infradb.NewSourceRepository(conn)
	optionRepo := infradb.NewOptionRepository(conn)
	preferenceRepo := infradb.NewPreferenceRepository(conn)
	recipientRepo := infradb.NewRecipientRepository(conn)
	responseRepo := infradb.NewResponseRepository(conn)
	accountRepo := infradb.NewAccountRepository(conn)
	userRepo := infradb.NewUserRepository(conn)
	membershipRepo := infradb.NewMembershipRepository(conn)
	membershipEventRepo := infradb.NewMembershipEventRepository(conn)
	removalRequestRepo := infradb.NewMembershipRemovalRequestRepository(conn)
	accountStatusEventRepo := infradb.NewAccountStatusEventRepository(conn)
	transactionManager := infradb.NewTransactionManager(conn)

	informationUsecase := usecase.NewInformationUsecase(
		aiClient,
		cfg.AIModel,
		informationRepo,
		sourceRepo,
		optionRepo,
		recipientRepo,
		transactionManager,
	)
	displayUsecase := usecase.NewDisplayUsecase(
		aiClient,
		cfg.AIModel,
		informationRepo,
		sourceRepo,
		optionRepo,
		preferenceRepo,
		recipientRepo,
	)
	responseUsecase := usecase.NewResponseUsecase(
		informationRepo,
		sourceRepo,
		optionRepo,
		recipientRepo,
		responseRepo,
		transactionManager,
	)
	accountUsecase := usecase.NewAccountUsecase(
		accountRepo,
		userRepo,
		membershipRepo,
		membershipEventRepo,
		accountStatusEventRepo,
		transactionManager,
	)
	membershipUsecase := usecase.NewMembershipUsecase(
		userRepo,
		membershipRepo,
		membershipEventRepo,
		removalRequestRepo,
		transactionManager,
	)

	verifier, err := supabase.NewVerifier(context.Background(), cfg.SupabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	authMiddleware := supabase.Middleware(verifier)
	optionalAuthMiddleware := supabase.OptionalMiddleware(verifier)

	corsMiddleware := cors.Middleware(cors.Config{AllowedOrigins: cfg.CORSAllowedOrigins})
	csrfCfg := csrf.Config{Secure: cfg.GinMode == gin.ReleaseMode}
	csrfMiddleware := csrf.Middleware(csrfCfg)
	csrfIssueHandler := csrf.IssueTokenHandler(csrfCfg)

	healthController := controller.NewHealthController()
	informationController := controller.NewInformationController(informationUsecase, accountUsecase)
	displayController := controller.NewDisplayController(displayUsecase, accountUsecase)
	responseController := controller.NewResponseController(responseUsecase, accountUsecase)
	accountController := controller.NewAccountController(accountUsecase)
	membershipController := controller.NewMembershipController(membershipUsecase, accountUsecase)

	handler := router.New(
		healthController,
		informationController,
		displayController,
		responseController,
		accountController,
		membershipController,
		authMiddleware,
		optionalAuthMiddleware,
		corsMiddleware,
		csrfMiddleware,
		csrfIssueHandler,
	)

	addr := ":" + cfg.Port
	log.Printf("starting server on %s", addr)
	if err := handler.Run(addr); err != nil {
		log.Fatal(err)
	}
}
