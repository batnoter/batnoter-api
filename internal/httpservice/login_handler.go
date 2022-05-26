package httpservice

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/batnoter/batnoter-api/internal/auth"
	"github.com/batnoter/batnoter-api/internal/github"
	"github.com/batnoter/batnoter-api/internal/user"
	gh "github.com/google/go-github/v43/github"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LoginHandler represents http handler for serving user login actions.
type LoginHandler struct {
	authService   auth.Service
	githubService github.Service
	userService   user.Service
	clientURL     string
}

// NewLoginHandler creates and returns a new login handler.
func NewLoginHandler(authService auth.Service, githubService github.Service, userService user.Service, clientURL string) *LoginHandler {
	return &LoginHandler{
		authService:   authService,
		githubService: githubService,
		userService:   userService,
		clientURL:     clientURL,
	}
}

// GithubLogin initiates oauth2 login flow with github provider.
func (l *LoginHandler) GithubLogin(c *gin.Context) {
	state := uuid.NewString()
	c.SetCookie("state", state, 600, "/", "", true, true)

	url := l.githubService.GetAuthCodeURL(state)

	// trigger authorization code grant flow
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GithubOAuth2Callback processes github oauth2 callback.
// It validates the state, fetch token and user from github, stores the user to db, generates app token.
// The app token will be sent as token cookie with a redirect to client url.
func (l *LoginHandler) GithubOAuth2Callback(c *gin.Context) {
	logrus.Info("github oauth2 callback started")
	state, _ := c.Cookie("state")
	stateFromCallback := c.Query("state")
	code := c.Query("code")

	if stateFromCallback != state {
		logrus.Error("invalid oauth state")
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=invalid-state")
		return
	}

	githubToken, err := l.githubService.GetToken(c, code)
	if err != nil {
		logrus.Errorf("auth code exchange for token failed: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=auth-code-exchange-failure")
		return
	}

	githubUser, err := l.githubService.GetUser(c, githubToken)
	if err != nil {
		logrus.Errorf("retrieving user from github failed: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=user-retrieval-failure")
		return
	}

	// get user from db if exists
	dbUser, err := l.userService.GetByEmail(*githubUser.Email)
	if err != nil {
		logrus.Errorf("retrieving user from db using email failed: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=internal-error")
		return
	}
	githubTokenJSON, err := json.Marshal(githubToken)
	if err != nil {
		logrus.Errorf("converting github token to json failed: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=internal-error")
		return
	}
	mapUserAttributes(&dbUser, string(githubTokenJSON), githubUser)

	// create/update the user record
	userID, err := l.userService.Save(dbUser)
	if err != nil {
		logrus.Errorf("saving user to db failed: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=internal-error")
		return
	}

	appToken, err := l.authService.GenerateToken(userID)
	if err != nil {
		logrus.Errorf("token generation failed: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, l.clientURL+"/login?success=false&error=internal-error")
		return
	}

	// set the token cookie
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("token", appToken, 60, "/", c.Request.URL.Hostname(), true, true)

	// redirect to client
	c.Redirect(http.StatusFound, l.clientURL+"/login?success=true")

	logrus.Info("github oauth2 callback finished")
}

// TokenPayload reads the token from request cookie and sends it as response payload.
// The idea is to use header based jwt auth (instead of cookie based auth) to avoid any security issues.
func (l *LoginHandler) TokenPayload(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// delete the token cookie
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("token", "", 0, "/", c.Request.URL.Hostname(), true, true)

	// send token as response payload
	c.String(http.StatusOK, token)
}

func mapUserAttributes(dbUser *user.User, ghToken string, githubUser gh.User) {
	dbUser.GithubToken = ghToken
	dbUser.Email = githubUser.GetEmail()
	dbUser.Name = githubUser.GetName()
	dbUser.Location = githubUser.GetLocation()
	dbUser.AvatarURL = githubUser.GetAvatarURL()
	dbUser.GithubID = githubUser.GetID()
	dbUser.GithubUsername = githubUser.GetLogin()
}
