package config

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/armadasystems/skydio-svc/skydio-client/pkg/ratelimiter"
	"github.com/spf13/viper"
)

type SkydioConfig struct {
	Endpoint   string
	APIVersion string
}

type DataBaseConfig struct {
	ScyllaDBHosts []string
}

type GoogleApiConfig struct {
	APIToken string
}

type AssetSvcConfig struct {
	Endpoint        string
	CacheTTL        int
	RefreshInterval int
}

type AppConfig struct {
	SkydioConfig      *SkydioConfig
	DataBaseConfig    *DataBaseConfig
	AssetSvcConfig    *AssetSvcConfig
	PostGresConfig    *PostGresConfig
	GoogleApiConfig   *GoogleApiConfig
	KeyVaultName      string
	RateLimiterConfig ratelimiter.RateLimiterConfig
}

type PostGresConfig struct {
	PostgresUrl string
	DBName      string
	DBUserName  string
	DBPassword  string
	DBPort      int
	DisableSSL  bool
}
type envConfig struct {
	SkydioEndpoint      string `mapstructure:"SKYDIO_ENDPOINT"`
	SkydioAPIVersion    string `mapstructure:"SKYDIO_APIVERSION"`
	ScyllaDBHosts       string `mapstructure:"DB_HOSTS"`
	TerminusEndpoint    string `mapstructure:"TERMINUS_ENDPOINT"`
	AssetSvcEndpoint    string `mapstructure:"ASSET_SVC_ENDPOINT"`
	AssetSvcTTL         int    `mapstructure:"ASSET_SVC_TTL"`
	RefreshInterval     int    `mapstructure:"ASSET_SVC_REFRESH_INTERVAL"`
	GoogleAPIToken      string `mapstructure:"GOOGLE_API_TOKEN"`
	PostgresUrl         string `mapstructure:"POSTGRES_URL"`
	DBPort              int    `mapstructure:"POSTGRES_PORT"`
	DBName              string `mapstructure:"POSTGRES_DB_NAME"`
	DBUserName          string `mapstructure:"POSTGRES_USER_NAME"`
	DBPassword          string `mapstructure:"POSTGRES_PASSWORD"`
	DBDisableSSL        bool   `mapstructure:"POSTGRES_SSL_DISABLE"`
	KeyVaultName        string `mapstructure:"KEY_VAULT_NAME"`
	EnableRateLimiter   bool   `mapstructure:"ENABLE_RATE_LIMITER"`
	RateLimiterPercent  int    `mapstructure:"RATE_LIMITER_USAGE"`
	VEHICLES_LIMIT      int    `mapstructure:"VEHICLES_LIMIT"`
	FLIGHTS_LIMIT       int    `mapstructure:"FLIGHTS_LIMIT"`
	TELEMETRY_LIMIT     int    `mapstructure:"TELEMETRY_LIMIT"`
	MEDIA_LIMIT         int    `mapstructure:"MEDIA_LIMIT"`
	MEDIA_BY_ID_LIMIT   int    `mapstructure:"MEDIA_BY_ID_LIMIT"`
	FLIGHT_BY_ID_LIMIT  int    `mapstructure:"FLIGHT_BY_ID_LIMIT"`
	THUMBNAIL_LIMIT     int    `mapstructure:"THUMBNAIL_LIMIT"`
	DOWNLOAD_LIMIT      int    `mapstructure:"DOWNLOAD_LIMIT"`
	WEBHOOK_LIMIT       int    `mapstructure:"WEBHOOK_LIMIT"`
	MAX_429_RETRY_LIMIT int    `mapstructure:"MAX_429_RETRY_LIMIT"`
}

var configMutex sync.Mutex

var config *AppConfig

// LoadAppConfig loads the application configuration from the specified path
func LoadAppConfig(configPath string) *AppConfig {

	viper.AddConfigPath(configPath)
	//Refer this issue reported by viper community: https://github.com/spf13/viper/issues/792, we should do it for any unmarshal/set operations while using viper
	configMutex.Lock()
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Failed to open env file at %s Err: %v", configPath, err)
	}

	envData := &envConfig{}
	err := viper.Unmarshal(&envData)
	defer configMutex.Unlock()
	if err != nil {
		log.Fatalf("Failed to unmarshal config: %s", err)
	}
	scyllaDBHosts := strings.Split(envData.ScyllaDBHosts, ",")

	if len(scyllaDBHosts) == 0 {
		log.Fatalf("DB_HOSTS cannot be empty")
	}

	// Update StaticConfig
	config = &AppConfig{
		SkydioConfig: &SkydioConfig{
			Endpoint:   envData.SkydioEndpoint,
			APIVersion: envData.SkydioAPIVersion,
		},

		DataBaseConfig: &DataBaseConfig{
			ScyllaDBHosts: scyllaDBHosts,
		},

		AssetSvcConfig: &AssetSvcConfig{
			Endpoint:        envData.AssetSvcEndpoint,
			CacheTTL:        envData.AssetSvcTTL,
			RefreshInterval: envData.RefreshInterval,
		},
		GoogleApiConfig: &GoogleApiConfig{
			APIToken: envData.GoogleAPIToken,
		},
		PostGresConfig: &PostGresConfig{
			PostgresUrl: envData.PostgresUrl,
			DBPort:      envData.DBPort,
			DBName:      envData.DBName,
			DBUserName:  envData.DBUserName,
			DBPassword:  envData.DBPassword,
			DisableSSL:  envData.DBDisableSSL,
		},
		RateLimiterConfig: ratelimiter.RateLimiterConfig{
			IS_ENABLE:           envData.EnableRateLimiter,
			USAGE_PERCENTAGE:    envData.RateLimiterPercent,
			VEHICLES_LIMIT:      envData.VEHICLES_LIMIT,
			FLIGHTS_LIMIT:       envData.FLIGHTS_LIMIT,
			TELEMETRY_LIMIT:     envData.TELEMETRY_LIMIT,
			MEDIA_LIMIT:         envData.MEDIA_LIMIT,
			MEDIA_BY_ID_LIMIT:   envData.MEDIA_BY_ID_LIMIT,
			FLIGHT_BY_ID_LIMIT:  envData.FLIGHT_BY_ID_LIMIT,
			THUMBNAIL_LIMIT:     envData.THUMBNAIL_LIMIT,
			DOWNLOAD_LIMIT:      envData.DOWNLOAD_LIMIT,
			WEBHOOK_LIMIT:       envData.WEBHOOK_LIMIT,
			MAX_429_RETRY_LIMIT: envData.MAX_429_RETRY_LIMIT,
		},
		KeyVaultName: envData.KeyVaultName,
	}

	return config

}

func GetDBConnectionString() string {
	var ssl_string string = "require"
	if config == nil {
		LoadAppConfig("../../cmd/app")
	}
	if config.PostGresConfig.DisableSSL {
		ssl_string = "disable"
	}
	//prepare database connection string
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.PostGresConfig.PostgresUrl, config.PostGresConfig.DBPort, config.PostGresConfig.DBName, config.PostGresConfig.DBUserName, config.PostGresConfig.DBPassword, ssl_string)
}
