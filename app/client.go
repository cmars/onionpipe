package app

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/cmars/onionpipe/secrets"
)

// ListClientKeys implements the `client ls` command.
func ListClientKeys(ctx *cli.Context) error {
	if ctx.Args().Present() {
		return cli.ShowSubcommandHelp(ctx)
	}
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	clients := sec.ClientsPublic()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(&clients)
}

// NewClientKey implements the `client new` command.
func NewClientKey(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if name == "" {
		return fmt.Errorf("missing client name")
	}
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	clients := sec.ClientsPublic()
	if _, ok := clients[name]; ok {
		return fmt.Errorf("client %q already exists", name)
	}
	_, err = sec.EnsureClientKey(name)
	if err != nil {
		return err
	}
	err = sec.WriteFile()
	if err != nil {
		return err
	}
	clients = sec.ClientsPublic()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]secrets.ClientPublic{
		name: clients[name],
	})
}

// ShowPrivateClientKey implements the `client show-private` command.
func ShowPrivateClientKey(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if name == "" {
		return fmt.Errorf("missing client name")
	}
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	clients := sec.ClientsPublic()
	if _, ok := clients[name]; !ok {
		return fmt.Errorf("client %q not found", name)
	}
	keyPair, err := sec.EnsureClientKey(name)
	if err != nil {
		return err
	}
	fmt.Println(strings.ToLower(
		base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(keyPair.Private[:])))
	return nil
}

// RemoveClientKey implements the `client rm` command.
func RemoveClientKey(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if name == "" {
		return fmt.Errorf("missing client name")
	}
	sec, err := openSecrets(ctx)
	if err != nil {
		return err
	}
	err = sec.RemoveClientKey(name)
	if err != nil {
		return err
	}
	return sec.WriteFile()
}
