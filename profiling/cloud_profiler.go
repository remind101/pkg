package profiling

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/profiler"
	"github.com/pkg/errors"
	"github.com/urfave/cli" // formerly known as github.com/codegangsta/cli
)

type GoogleProfilerFlag struct{}

// Apply starts the Google Cloud profiling agent.
//
// This allows us to use Google's nifty profiling UI to view live production
// profiling. https://cloud.google.com/profiler
func (f GoogleProfilerFlag) Set(projectID string) error {
	// Get all of the empire configs from the environment, and make sure they're
	// set.
	empire_appname := os.Getenv("EMPIRE_APPNAME")
	if empire_appname == "" {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: missing/blank required env var EMPIRE_APPNAME",
				),
			),
		)

		return nil
	}

	empire_process := os.Getenv("EMPIRE_PROCESS")
	if empire_process == "" {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: missing/blank required env var EMPIRE_PROCESS",
				),
			),
		)

		return nil
	}

	empire_release := os.Getenv("EMPIRE_RELEASE")
	if empire_release == "" {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: missing/blank required env var EMPIRE_RELEASE",
				),
			),
		)

		return nil
	}

	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_CONTENT")
	if creds == "" {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: missing/blank required env var GOOGLE_APPLICATION_CREDENTIALS_CONTENT",
				),
			),
		)

		return nil
	}

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: missing/blank required env var GOOGLE_APPLICATION_CREDENTIALS",
				),
			),
		)

		return nil
	}

	// kludge alert! Our 12-factor way of deploying applications can't pass
	// configuration to programs except through environment variables, but
	// the GCP SDK expects the credentials to be in a file at the path of
	// $GOOGLE_APPLICATION_CREDENTIALS. So, we put the actual credentials
	// in a different environment variable, and write them to a file.
	//
	// If $GOOGLE_APPLICATION_CREDENTIALS_CONTENT isn't set then we do
	// nothing, since we might be running in an environment where the
	// credentials might be discovered through other means.
	if err := ioutil.WriteFile(credsPath, []byte(creds), 0600); err != nil {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: could not write $GOOGLE_APPLICATION_CREDENTIALS_CONTENT to %s: %s",
					credsPath,
					err,
				),
			),
		)

		return nil
	}

	cfg := profiler.Config{
		Service:        fmt.Sprintf("%s.%s", empire_appname, empire_process),
		ServiceVersion: empire_release,
		ProjectID:      projectID,
	}

	err := profiler.Start(cfg)
	if err != nil {
		log.Printf(
			"%+v",
			errors.WithStack(
				fmt.Errorf(
					"pkg/profiling: GoogleProfilerFlag.Set: error starting profiler: %s",
					err,
				),
			),
		)
	}

	return nil
}

func (f GoogleProfilerFlag) String() string {
	return ""
}

// NewCliFlag returns a flag that will enable Cloud Profiler
func NewCliFlag() cli.Flag {
	return cli.GenericFlag{
		Name:   "google-profiler-project",
		Value:  GoogleProfilerFlag{},
		Usage:  "The Google Project ID for submitting Cloud Profiler data",
		EnvVar: "GOOGLE_PROFILER_PROJECT",
	}
}
