package httpservice

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/batnoter/batnoter-api/internal/applicationconfig"
	"github.com/sirupsen/logrus"
)

// Run starts the http server.
func Run(applicationconfig *applicationconfig.ApplicationConfig) error {
	gin.SetMode(gin.ReleaseMode)
	if applicationconfig.Config.HTTPServer.Debug {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()
	router.UseRawPath = true

	clientBaseURL := baseURL(applicationconfig.Config.App.ClientURL)
	router.Use(cors.New(corsConfig(clientBaseURL)))
	logrus.Infof("allowing cors for %s", clientBaseURL)

	noteHandler := NewNoteHandler(applicationconfig.GithubService, applicationconfig.UserService)
	loginHandler := NewLoginHandler(applicationconfig.AuthService, applicationconfig.GithubService, applicationconfig.UserService, applicationconfig.Config.App.ClientURL)
	userHandler := NewUserHandler(applicationconfig.UserService)
	preferenceHandler := NewPreferenceHandler(applicationconfig.PreferenceService, applicationconfig.GithubService, applicationconfig.UserService)
	authMiddleware := NewMiddleware(applicationconfig.AuthService)

	v1 := router.Group("api/v1")
	v1.GET("/user/me", authMiddleware.AuthorizeToken(), userHandler.Profile)
	v1.GET("/user/preference/repo", authMiddleware.AuthorizeToken(), preferenceHandler.GetRepos)
	v1.POST("/user/preference/repo", authMiddleware.AuthorizeToken(), preferenceHandler.SaveDefaultRepo)
	v1.POST("/user/preference/auto/repo", authMiddleware.AuthorizeToken(), preferenceHandler.AutoSetupRepo)

	v1.GET("/search/notes", authMiddleware.AuthorizeToken(), noteHandler.SearchNotes)  // search notes (provide filters using query-params)
	v1.GET("/tree/notes", authMiddleware.AuthorizeToken(), noteHandler.GetNotesTree)   // get complete notes repo tree
	v1.GET("/notes", authMiddleware.AuthorizeToken(), noteHandler.GetAllNotes)         // get all notes from path (provide filters using query-params)
	v1.GET("/notes/:path", authMiddleware.AuthorizeToken(), noteHandler.GetNote)       // get single note
	v1.POST("/notes/:path", authMiddleware.AuthorizeToken(), noteHandler.SaveNote)     // create/update single note
	v1.DELETE("/notes/:path", authMiddleware.AuthorizeToken(), noteHandler.DeleteNote) // delete single note

	v1.GET("/auth/token", loginHandler.TokenPayload)
	v1.GET("/oauth2/login/github", loginHandler.GithubLogin)
	v1.GET("/oauth2/github/callback", loginHandler.GithubOAuth2Callback)

	address := net.JoinHostPort(applicationconfig.Config.HTTPServer.Host, applicationconfig.Config.HTTPServer.Port)
	server := http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}
	return server.ListenAndServe()
}

func corsConfig(clientBaseURL string) cors.Config {
	return cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == clientBaseURL
		},
		MaxAge: 12 * time.Hour,
	}
}

func baseURL(clientURL string) string {
	if clientURL == "" {
		logrus.Fatal("client url is not configured")
	}
	u, err := url.Parse(clientURL)
	if err != nil {
		logrus.WithField("client-url", clientURL).Fatal("invalid client url")
	}
	return fmt.Sprintf(`%s://%s`, u.Scheme, u.Host)
}
