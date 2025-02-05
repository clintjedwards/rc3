package conf

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/structs"
)

func mustParseDuration(duration string) time.Duration {
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatalf("could not parse duration %q; %v", duration, err)
	}

	return parsedDuration
}

// searchFilePaths will search each path given in order for a file
//
//	and return the first path that exists.
func searchFilePaths(paths ...string) string {
	for _, path := range paths {
		if path == "" {
			continue
		}

		stat, err := os.Stat(path)

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if stat.IsDir() {
			continue
		}
		return path

	}

	return ""
}

func possibleConfigPaths(homeDir, flagPath string) []string {
	return []string{
		flagPath,
		fmt.Sprintf("%s/%s", homeDir, ".rc3.toml"),
		fmt.Sprintf("%s/%s/%s", homeDir, ".config", "rc3.toml"),
	}
}

// This function helps us generate always up to date env vars for the help and documentation. To do that it leverages
// reflection and an empty struct to reconstruct what those keys should be.
//
// If you get a panic pointing to this area check the higher level function GetAPIEnvVars() to make sure the struct
// doesn't have any uninitialized struct pointers.
func getEnvVarsFromStruct(prefix string, fields []*structs.Field) []string {
	output := []string{}

	for _, field := range fields {
		tag := field.Tag("koanf")
		if field.Kind() == reflect.Pointer {
			output = append(output, getEnvVarsFromStruct(strings.ToUpper(prefix+tag+"__"), field.Fields())...)
			continue
		}

		output = append(output, strings.ToUpper(prefix+tag))
	}

	return output
}
