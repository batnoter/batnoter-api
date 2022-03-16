package applicationconfig

import (
	"github.com/vivekweb2013/batnoter/internal/auth"
	"github.com/vivekweb2013/batnoter/internal/config"
	"github.com/vivekweb2013/batnoter/internal/note"
	"gorm.io/gorm"
)

type ApplicationConfig struct {
	Config      config.Config
	DB          *gorm.DB
	AuthService auth.Service
	NoteService note.Service
}

func NewApplicationConfig(config config.Config, db *gorm.DB) *ApplicationConfig {
	noteRepo := note.NewRepository(db)
	noteService := note.NewService(noteRepo)
	authService := auth.NewService(auth.TokenConfig{
		SecretKey: config.App.SecretKey,
		Issuer:    "https://batnoter.com",
	})
	return &ApplicationConfig{
		Config:      config,
		DB:          db,
		AuthService: authService,
		NoteService: noteService,
	}
}
