package conf

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// API defines config settings for the REST API server.
type API struct {
	General     *General     `koanf:"general"`
	Development *Development `koanf:"development"`
	Server      *Server      `koanf:"server"`
}

func DefaultAPIConfig() *API {
	return &API{
		General:     DefaultGeneralConfig(),
		Development: DefaultDevelopmentConfig(),
		Server:      DefaultServerConfig(),
	}
}

type General struct {
	// Log level affects the entire application's log level.
	LogLevel string `koanf:"log_level"`

	// should be in format: "https://recurse.proxmox.com:8006/api2/json"
	// omitting the api route will cause requests to fail with 501 errors that translate to 404 errors.
	ProxmoxURL         string `koanf:"proxmox_url"`
	ProxmoxTokenID     string `koanf:"proxmox_token_id"`
	ProxmoxTokenSecret string `koanf:"proxmox_token_secret"`

	// Connect to proxmox using TLS.
	PromoxUseTLS bool `koanf:"proxmox_use_tls"`
}

func DefaultGeneralConfig() *General {
	return &General{
		LogLevel:     "debug",
		PromoxUseTLS: false,
	}
}

type Development struct {
	PrettyLogging bool `koanf:"pretty_logging"`
	BypassAuth    bool `koanf:"bypass_auth"`

	// Instead of having to recompile the static files into the binary during development for every change
	// instead uses another implementation of the fileserver to easily serve files from local disk.
	LoadFrontendFilesFromDisk bool `koanf:"load_frontend_files_from_disk"`
}

func DefaultDevelopmentConfig() *Development {
	return &Development{
		PrettyLogging:             true,
		BypassAuth:                true,
		LoadFrontendFilesFromDisk: true,
	}
}

// Server represents lower level HTTP server settings.
type Server struct {
	// URL for the server to bind to. Ex: localhost:8080
	Host string `koanf:"host"`

	// How long the service should wait on in-progress connections before hard closing everything out.
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
}

// DefaultServerConfig returns a pre-populated configuration struct that is used as the base for super imposing user configuration
// settings.
func DefaultServerConfig() *Server {
	return &Server{
		Host:            "0.0.0.0:8080",
		ShutdownTimeout: mustParseDuration("15s"),
	}
}

// Get the final configuration for the server.
// This involves correctly finding and ordering different possible paths for the configuration file:
//
//  1. The function is intended to be called with paths gleaned from the -config flag in the cli.
//  2. If the user does not use the -config path of the path does not exist,
//     then we default to a few hard coded config path locations.
//  3. Then try to see if the user has set an envvar for the config file, which overrides
//     all previous config file paths.
//  4. Finally, whatever configuration file path is found first is the processed.
//
// Whether or not we use the configuration file we then search the environment for all environment variables:
//   - Environment variables are loaded after the config file and therefore overwrite any conflicting keys.
//   - All configuration that goes into a configuration file can also be used as an environment variable.
func InitAPIConfig(userDefinedPath string, loadDefaults bool) (*API, error) {
	var config *API

	// First we initiate the default values for the config.
	if loadDefaults {
		config = DefaultAPIConfig()
	}

	possibleConfigPaths := []string{userDefinedPath, "/etc/rc3/rc3.hcl"}

	path := searchFilePaths(possibleConfigPaths...)

	// envVars top all other entries so if its not empty we just insert it over the current path
	// regardless of if we found one.
	envPath := os.Getenv("RC3_CONFIG_PATH")
	if envPath != "" {
		path = envPath
	}

	configParser := koanf.New(".")

	if path != "" {
		err := configParser.Load(file.Provider(path), toml.Parser())
		if err != nil {
			return nil, err
		}
	}

	err := configParser.Load(env.Provider("RC3_", "__", func(s string) string {
		newStr := strings.TrimPrefix(s, "RC3_")
		newStr = strings.ToLower(newStr)
		return newStr
	}), nil)
	if err != nil {
		return nil, err
	}

	err = configParser.Unmarshal("", &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetAPIEnvVars() []string {
	api := API{
		General:     &General{},
		Development: &Development{},
		Server:      &Server{},
	}
	fields := structs.Fields(api)

	vars := getEnvVarsFromStruct("RC3_", fields)
	sort.Strings(vars)
	return vars
}
