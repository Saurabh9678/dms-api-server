package config

type DBConfig struct {
	URL string
}

func LoadDBConfig() DBConfig {
	return DBConfig{
		URL: getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/dms?sslmode=disable"),
	}
}
