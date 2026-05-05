package application

import (
	"context"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	mw "github.com/labstack/echo/v5/middleware"
	"github.com/servercurio/go-echo-starter/internal/config"
	"github.com/servercurio/go-echo-starter/internal/health"
	"github.com/servercurio/go-echo-starter/internal/logging"
	"github.com/servercurio/go-echo-starter/internal/router"
)

const defaultName = "appsvrd"
const defaultEnvPrefix = "APP"
const defaultConfigName = "appsvrd"
const defaultConfigPathElement = "appsvr"

type Application struct {
	Name              string
	ConfigFileName    string
	EnvVariablePrefix string

	config     *Config
	middleware []echo.MiddlewareFunc
	httpServer *echo.Echo
	tlsServer  *echo.Echo

	userHomeDirectory string
	certificate       *InMemoryCertificate
	modules           map[string]router.Module
	healthRegistry    *health.Registry

	ready atomic.Bool
}

// IsReady reports whether the application is currently accepting traffic. It
// returns true between the point Start has launched the server goroutines and
// the point shutdown is initiated.
func (app *Application) IsReady() bool {
	return app.ready.Load()
}

func NewApplication(cfg *Config) *Application {
	app := &Application{
		Name:              defaultName,
		ConfigFileName:    defaultConfigName,
		EnvVariablePrefix: defaultEnvPrefix,
		config:            cfg,
		healthRegistry:    health.NewRegistry(),
		httpServer:        echo.New(),
		tlsServer:         echo.New(),
		modules:           make(map[string]router.Module),
	}

	// Initialize the logging configuration early to avoid missing any critical logs
	loggingCfg := logging.NewConfigFromEnv(app.EnvVariablePrefix)
	logging.NotifyDaemonStartup(app.Name, loggingCfg)

	return app
}

// buildMiddleware assembles the global middleware stack from the loaded
// configuration. Called from Initialize() once Configure() has populated
// app.config — must run before configureHttpServer / configureTlsServer
// so the slice is ready when those methods call e.Use(app.middleware...).
//
// Lifted out of NewApplication because the Secure middleware needs the
// SecurityConfig values (HSTS, CSP, Referrer-Policy), which aren't
// available until config files + env vars have been resolved.
func (app *Application) buildMiddleware() {
	sec := app.config.Server.Security
	if sec == nil {
		sec = DefaultSecurityConfig()
	}

	app.middleware = []echo.MiddlewareFunc{
		mw.Recover(),
		mw.RequestID(),
		mw.GzipWithConfig(mw.GzipConfig{
			Level:     5,
			MinLength: 2 * 1024,
		}),
		logging.EchoMiddleware(),
		mw.SecureWithConfig(mw.SecureConfig{
			XSSProtection:         mw.DefaultSecureConfig.XSSProtection,
			ContentTypeNosniff:    mw.DefaultSecureConfig.ContentTypeNosniff,
			XFrameOptions:         mw.DefaultSecureConfig.XFrameOptions,
			HSTSMaxAge:            sec.HSTSMaxAge,
			HSTSExcludeSubdomains: sec.HSTSExcludeSubdomains,
			HSTSPreloadEnabled:    sec.HSTSPreloadEnabled,
			ContentSecurityPolicy: sec.ContentSecurityPolicy,
			ReferrerPolicy:        sec.ReferrerPolicy,
		}),
	}
}

func (app *Application) Configure() error {
	configLocations := configSearchPaths()

	logging.Daemon.
		Trace().
		Strs("paths", configLocations).
		Strs("fileNames", config.FileNameVariants(app.ConfigFileName)).
		Msg("searching for config files")

	// Load the server configuration from config files
	if err := config.FromPaths(app.config, app.ConfigFileName, configLocations...); err != nil {
		logging.Daemon.
			Warn().
			Err(err).
			Strs("paths", configLocations).
			Strs("fileNames", config.FileNameVariants(app.ConfigFileName)).
			Msg("error loading server config")
	}

	// Load the server configuration from environment variables
	app.config.FromEnv(app.EnvVariablePrefix)

	logging.NotifyDaemonLoggingStartup(app.config.Logging)
	logging.NotifyHttpLoggingStartup(app.config.Logging)

	NotifyHttpServerConfig(app.config.Server.Http)
	NotifyHttpsServerConfig(app.config.Server.Https)
	NotifyCorsConfig(app.config.Server.Cors)
	NotifySecurityConfig(app.config.Server.Security)
	NotifyProxySupportConfig(app.config.Proxy)
	NotifyDatabaseConfig(app.config.Database)
	NotifyOpenAPIConfig(app.config.OpenAPI)

	return nil
}

func (app *Application) Initialize() error {
	app.resolveHomeDirectory()

	app.buildMiddleware()

	if err := app.configureHttpServer(); err != nil {
		return err
	}

	if err := app.configureTlsServer(); err != nil {
		return err
	}

	if err := app.configureProxySupport(); err != nil {
		logging.Daemon.
			Warn().
			Err(err).
			Msg("invalid proxy support configuration")
	}

	if err := app.initializeDatabase(); err != nil {
		return err
	}

	app.registerHealthChecks()

	if err := app.initializeOpenAPI(); err != nil {
		return err
	}

	if err := app.initializeRouting(); err != nil {
		logging.Daemon.
			Warn().
			Err(err).
			Msg("failed to initialize routing")
	}

	return nil
}

func (app *Application) RegisterModule(m router.Module) error {
	if m == nil {
		return errorx.IllegalArgument.New("module argument must not be nil")
	}

	if _, exists := app.modules[m.Name()]; exists {
		logging.Daemon.
			Warn().
			Str("name", m.Name()).
			Str("id", m.Id()).
			Msg("application module already registered")
		return errorx.IllegalState.New("application module '%s' already registered", m.Name())
	}

	app.modules[m.Name()] = m
	return nil
}

func (app *Application) Start() (int, error) {
	signalCtx, signalCancel :=
		signal.NotifyContext(context.Background(), shutdownSignals...)
	defer signalCancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		app.startHttpServer(signalCtx)
	}()
	go func() {
		defer wg.Done()
		app.startTlsServer(signalCtx)
	}()

	app.ready.Store(true)

	<-signalCtx.Done()
	app.ready.Store(false)

	// Echo v5 shuts each server down internally when signalCtx cancels (using
	// the GracefulTimeout we set on StartConfig). Wait for the goroutines to
	// finish, but bound the wait so a stalled handler can't hang the process.
	maxShutdown := app.config.Server.Http.ShutdownTimeout
	if app.config.Server.Https != nil && app.config.Server.Https.ShutdownTimeout > maxShutdown {
		maxShutdown = app.config.Server.Https.ShutdownTimeout
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Daemon.Info().Msg("server goroutines shut down cleanly")
	case <-time.After(maxShutdown + 5*time.Second):
		logging.Daemon.Warn().
			Dur("timeout", maxShutdown+5*time.Second).
			Msg("server goroutines did not return within shutdown timeout")
	}

	app.shutdownDatabase()

	return 0, nil
}
