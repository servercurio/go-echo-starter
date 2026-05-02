package application

import (
	"context"
	"os/signal"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	mw "github.com/labstack/echo/v5/middleware"
	"github.com/servercurio/go-echo-starter/internal/config"
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
}

func NewApplication(cfg *Config) *Application {
	app := &Application{
		Name:              defaultName,
		ConfigFileName:    defaultConfigName,
		EnvVariablePrefix: defaultEnvPrefix,
		config:            cfg,
		middleware: []echo.MiddlewareFunc{
			mw.Recover(),
			mw.RequestID(),
			mw.GzipWithConfig(mw.GzipConfig{
				Level:     5,
				MinLength: 2 * 1024,
			}),
			logging.EchoMiddleware(),
			mw.CORS("*"),
			//mw.CSRF(),
			mw.Secure(),
		},
		httpServer: echo.New(),
		tlsServer:  echo.New(),
		modules:    make(map[string]router.Module),
	}

	// Initialize the logging configuration early to avoid missing any critical logs
	loggingCfg := logging.NewConfigFromEnv(app.EnvVariablePrefix)
	logging.NotifyDaemonStartup(app.Name, loggingCfg)

	return app
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
	NotifyProxySupportConfig(app.config.Proxy)

	return nil
}

func (app *Application) Initialize() error {
	app.resolveHomeDirectory()

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

	go app.startHttpServer()
	go app.startTlsServer()

	<-signalCtx.Done()
	app.shutdownHttpServer()
	app.shutdownTlsServer()

	return 0, nil
}
