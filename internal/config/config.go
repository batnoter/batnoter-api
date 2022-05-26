package config

// App represents configuration properties specific to the application.
type App struct {
	SecretKey string
	ClientURL string
}

// Database represents configuration properties required to connect to a database.
// The url config optional.
// But if url is set then the values of host, port, dbname, username, password, driver-name
// will be overridden with their respective values from url string.
type Database struct {
	Host       string
	Port       string
	DBName     string
	Username   string
	Password   string
	DriverName string
	SSLMode    string
	URL        string
	Debug      bool
}

// HTTPServer represents configuration properties required for starting http server.
type HTTPServer struct {
	Host  string
	Port  string
	Debug bool
}

// OAuth2 represents configuration grouped by the oauth2 provider.
type OAuth2 struct {
	Github Github
}

// Github represents configuration properties required consume github oauth2 api.
type Github struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Config represents all the application configurations grouped as per their category.
type Config struct {
	App        App
	Database   Database
	HTTPServer HTTPServer
	OAuth2     OAuth2
}
