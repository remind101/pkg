package profiling

import (
	"io/ioutil"
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
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_CONTENT")
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if creds != "" {
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
			return errors.Wrap(err, "could not write $GOOGLE_APPLICATION_CREDENTIALS_CONTENT to "+credsPath)
		}
	}

	cfg := profiler.Config{
		Service:        os.Getenv("EMPIRE_APPNAME") + "." + os.Getenv("EMPIRE_PROCESS"),
		ServiceVersion: os.Getenv("EMPIRE_RELEASE"),
		ProjectID:      projectID,
	}

	return profiler.Start(cfg)
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
