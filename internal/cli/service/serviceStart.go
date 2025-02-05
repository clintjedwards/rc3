package service

import (
	"os"
	"strings"

	"github.com/clintjedwards/rc3/internal/api"
	"github.com/clintjedwards/rc3/internal/cli/global"
	"github.com/clintjedwards/rc3/internal/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cmdServiceStart = &cobra.Command{
	Use:   "start",
	Short: "Start the RC3 REST API service",
	Long: `Start the RC3 REST API service. Running this command attempts to start the long running service. This
commadn will block and only gracefully stop on SIGINT or SIGTERM signals.

### List of Environment Variables:

` + strings.Join(conf.GetAPIEnvVars(), "\n"),
	RunE: serverStart,
}

func init() {
	CmdService.AddCommand(cmdServiceStart)
}

func serverStart(cmd *cobra.Command, _ []string) error {
	global.CLIContext.Fmt.Finish()

	conf, err := conf.InitAPIConfig("", true)
	if err != nil {
		log.Fatal().Err(err).Msg("error in config initialization")
	}

	setupLogging(conf.General.LogLevel, conf.Development.PrettyLogging)
	api.StartAPIServer(conf)

	return nil
}

func setupLogging(loglevel string, pretty bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	zerolog.SetGlobalLevel(parseLogLevel(loglevel))
	if pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func parseLogLevel(loglevel string) zerolog.Level {
	switch loglevel {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		log.Error().Msgf("loglevel %s not recognized; defaulting to debug", loglevel)
		return zerolog.DebugLevel
	}
}
