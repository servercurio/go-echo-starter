package config

type EnvironmentSource interface {
	FromEnv(prefix string)
}
