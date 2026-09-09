package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Mabarik667f/fsserver/internal/api/handler"
	"github.com/Mabarik667f/fsserver/internal/closer"
	"github.com/Mabarik667f/fsserver/internal/infrastructure"
	"github.com/Mabarik667f/fsserver/internal/infrastructure/cache"
	"github.com/Mabarik667f/fsserver/internal/infrastructure/db"
	"github.com/Mabarik667f/fsserver/internal/infrastructure/security"
	"github.com/Mabarik667f/fsserver/internal/integration"
	"github.com/Mabarik667f/fsserver/internal/integration/filestorage"
	"github.com/Mabarik667f/fsserver/internal/repository"
	"github.com/Mabarik667f/fsserver/internal/service"
	"github.com/Mabarik667f/fsserver/pkg/config"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/jackc/pgx/v5/pgxpool"
)

type diContainer struct {
	api http.Handler
	db  *pgxpool.Pool

	hasher         *security.ArgonHasher
	sessionManager *scs.SessionManager

	storage integration.FileStorage
	cache   infrastructure.Cache

	validator *validator.Validate
	decoder   *schema.Decoder

	userRepo repository.UserRepository
	docRepo  repository.DocRepository

	userService service.UserService
	docService  service.DocService

	userHandler handler.UserHandler
	docHandler  handler.DocHandler
}

func newDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) DB() *pgxpool.Pool {
	if d.db == nil {
		pool, err := db.NewPool(config.AppConfig().DatabaseDSN())
		if err != nil {
			slog.Error("failed to connect DB", "err", err)
			os.Exit(1)
		}

		closer.Add("DB", func(ctx context.Context) error {
			pool.Close()
			return nil
		})

		d.db = pool
	}

	return d.db
}

func (d *diContainer) Cache() infrastructure.Cache {
	if d.cache == nil {
		instance := cache.NewInMemoryCache()
		d.cache = instance
		closer.Add("Cache", func(ctx context.Context) error {
			instance.StartCleanup(
				ctx,
				time.Duration(config.AppConfig().CacheParams.CleanupDurationMinutes)*time.Minute,
			)
			return nil
		})
	}

	return d.cache
}

func (d *diContainer) Storage() integration.FileStorage {
	if d.storage == nil {
		d.storage = filestorage.NewStorage("./data")
	}

	return d.storage
}

func (d *diContainer) SessionManager() *scs.SessionManager {
	if d.sessionManager == nil {
		d.sessionManager = scs.New()
		d.sessionManager.Lifetime = 24 * time.Hour
		d.sessionManager.Store = memstore.New()
	}
	return d.sessionManager
}

func (d *diContainer) Hasher() *security.ArgonHasher {
	if d.hasher == nil {
		d.hasher = security.NewArgonHasher(
			config.AppConfig().ArgonParams.TimeCost,
			config.AppConfig().ArgonParams.MemoryCost,
			config.AppConfig().ArgonParams.KeyLength,
			config.AppConfig().ArgonParams.Threads,
		)
	}

	return d.hasher
}

func (d *diContainer) Validator() *validator.Validate {
	if d.validator == nil {
		d.validator = validator.New()
	}

	return d.validator
}

func (d *diContainer) Decoder() *schema.Decoder {
	if d.decoder == nil {
		d.decoder = schema.NewDecoder()
	}

	return d.decoder
}

func (d *diContainer) UserRepo() repository.UserRepository {
	if d.userRepo == nil {
		d.userRepo = repository.NewUserRepo(d.DB())
	}

	return d.userRepo
}

func (d *diContainer) DocRepo() repository.DocRepository {
	if d.docRepo == nil {
		d.docRepo = repository.NewDocRepo(d.DB())
	}

	return d.docRepo
}

func (d *diContainer) UserService() service.UserService {
	if d.userService == nil {
		d.userService = service.NewUserService(
			d.UserRepo(),
			d.Hasher(),
			config.AppConfig().Service.AdminToken,
		)
	}

	return d.userService
}

func (d *diContainer) DocService() service.DocService {
	if d.docService == nil {
		d.docService = service.NewDocService(d.UserRepo(), d.DocRepo(), d.Storage())
	}

	return d.docService
}

func (d *diContainer) UserHandler() handler.UserHandler {
	if d.userHandler == nil {
		d.userHandler = handler.NewUserHandler(d.UserService(), d.SessionManager(), d.Validator())
	}
	return d.userHandler
}

func (d *diContainer) DocHandler() handler.DocHandler {
	if d.docHandler == nil {
		d.docHandler = handler.NewDocHandler(
			d.DocService(),
			d.Validator(),
			d.Decoder(),
			d.SessionManager(),
		)
	}
	return d.docHandler
}

func (d *diContainer) API() http.Handler {
	if d.api == nil {
		d.api = handler.API(
			d.SessionManager(),
			d.UserHandler(),
			d.DocHandler(),
		)
	}

	return d.api
}
