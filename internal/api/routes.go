package api

import (
	_ "github.com/17HIERARCH70/SocialManager/docs"
	"github.com/17HIERARCH70/SocialManager/internal/api/middleware"
	"github.com/17HIERARCH70/SocialManager/internal/handlers/authHandlers"
	"github.com/17HIERARCH70/SocialManager/internal/handlers/emailHandlers"
	authService2 "github.com/17HIERARCH70/SocialManager/internal/services/authService"
	emailService2 "github.com/17HIERARCH70/SocialManager/internal/services/emailService"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/swaggo/swag"
)

// SetupRoutes sets up the application's routes
func (a *App) SetupRoutes() {
	router := mux.NewRouter()

	authService, _ := authService2.NewAuthService(a.psql, a.cfg, a.log)
	authHandler := authHandlers.NewAuthHandler(authService)

	emailService, _ := emailService2.NewEmailService(a.psql, a.cfg, authService, a.log)
	emailHandler := emailHandlers.NewEmailHandler(emailService, a.log)

	// Swagger route
	router.PathPrefix("/api/swagger/").Handler(httpSwagger.WrapHandler)

	// Authentication routes
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("/google_login", authHandler.GoogleLoginHandler).Methods("GET")
	authRouter.HandleFunc("/google_callback", authHandler.GoogleCallbackHandler).Methods("GET")

	protectedRouter := router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.JWTAuthMiddleware)

	// Email routes
	protectedRouter.HandleFunc("/emails/user/{user_id:[0-9]+}", emailHandler.GetEmailsByUserIDHandler).Methods("GET")
	protectedRouter.HandleFunc("/emails", emailHandler.GetAllEmailsHandler).Methods("GET")
	protectedRouter.HandleFunc("/emails/user", emailHandler.GetUserIDByEmailHandler).Methods("GET")
	protectedRouter.HandleFunc("/emails/{email_id}", emailHandler.DeleteEmailByIDHandler).Methods("DELETE")
	protectedRouter.HandleFunc("/emails/user/{user_id:[0-9]+}", emailHandler.DeleteAllEmailsByUserIDHandler).Methods("DELETE")

	a.router = router
}
