package app

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/cmars/oniongrok/secrets"
)

var forwardFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debug log output",
	},
	&cli.BoolFlag{
		Name:  "anonymous",
		Usage: "publish anonymous hidden services",
		Value: true,
	},
	&cli.PathFlag{
		Name:  "secrets",
		Usage: "path where service and client secrets are stored",
	},
}

func defaultSecretsPath() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Printf("failed to locate home directory: %v", err)
		return ""
	}
	return filepath.Join(home, ".local", "share", "oniongrok", "secrets.json")
}

// App returns a new oniongrok command line app.
func App() *cli.App {
	return &cli.App{
		Name:   "oniongrok",
		Usage:  "forward services through Tor; .onion addresses for anything",
		Flags:  forwardFlags,
		Action: Forward,
		Commands: []*cli.Command{{
			Name:    "forward",
			Aliases: []string{"fwd"},
			Usage:   "forward socket address through Tor network",
			Flags:   forwardFlags,
			Action:  Forward,
		}, {
			Name:  "service",
			Usage: "manage onion services",
			Subcommands: []*cli.Command{{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list onion service keys",
				Action:  ListServiceKeys,
			}, {
				Name:    "new",
				Aliases: []string{"add", "create"},
				Usage:   "new onion service",
				Action:  NewServiceKey,
			}, {
				Name:    "remove",
				Aliases: []string{"rm", "delete", "del"},
				Usage:   "remove onion service",
				Action:  RemoveServiceKey,
			}},
			Action: ListServiceKeys,
		}},
	}
}

const startTorTimeout = time.Minute * 3

func secretsPath(path string, anonymous bool) string {
	if !anonymous {
		return strings.TrimSuffix(path, filepath.Ext(path)) + ".not-anonymous" + filepath.Ext(path)
	}
	return path
}

func openSecrets(ctx *cli.Context) (*secrets.Secrets, error) {
	secPath := ctx.Path("secrets")
	if secPath == "" {
		secPath = defaultSecretsPath()
	}
	return secrets.ReadFile(secretsPath(secPath, ctx.Bool("anonymous")))
}
