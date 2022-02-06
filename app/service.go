package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cmars/oniongrok/secrets"
	"github.com/urfave/cli/v2"
)

// ListServiceKeys implements the `service ls` command.
func ListServiceKeys(ctx *cli.Context) error {
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	services := sec.ServicesPublic()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(&services)
}

// NewServiceKey implements the `service new` command.
func NewServiceKey(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if name == "" {
		return fmt.Errorf("missing service name")
	}
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	services := sec.ServicesPublic()
	if _, ok := services[name]; ok {
		return fmt.Errorf("service %q already exists", name)
	}
	_, err = sec.EnsureServiceKey(name)
	if err != nil {
		return err
	}
	err = sec.WriteFile()
	if err != nil {
		return err
	}
	services = sec.ServicesPublic()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]secrets.ServicePublic{
		name: services[name],
	})
}

// RemoveServiceKey implements the `service rm` command.
func RemoveServiceKey(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if name == "" {
		return fmt.Errorf("missing service name")
	}
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	err = sec.RemoveServiceKey(name)
	if err != nil {
		return err
	}
	return sec.WriteFile()
}
