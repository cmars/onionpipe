package main

import (
	"embed"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/mitchellh/go-homedir"
)

//go:embed etc/tor/torrc
var Etc embed.FS

var name = flag.String("name", "default", "service name")

type Config struct {
	Network        string
	HostAddress    string
	HiddenServices []HiddenService
}

type HiddenService struct {
	Name  string
	Ports []int
}

func main() {
	flag.Parse()

	portStr := flag.Arg(0)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("invalid port %s: %v", portStr, err)
	}

	torrcTmpl, err := template.New("torrc").ParseFS(Etc, "etc/tor/torrc")
	if err != nil {
		log.Fatalf("failed to parse torrc template: %v", err)
	}

	cfg := &Config{
		Network:     "host",
		HostAddress: "127.0.0.1",
		HiddenServices: []HiddenService{{
			Name:  *name,
			Ports: []int{port},
		}},
	}

	torrcFile, err := ioutil.TempFile("", "torrc")
	if err != nil {
		log.Fatal(err)
	}
	torrcWr := io.MultiWriter(torrcFile, os.Stdout)
	err = torrcTmpl.Execute(torrcWr, cfg)
	if err != nil {
		log.Fatalf("failed to render torrc: %v", err)
	}

	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("failed to find home directory: %v", err)
	}
	ogDir := filepath.Join(homeDir, ".local/share/oniongrok")
	err = os.MkdirAll(ogDir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(
		"docker", "run",
		"--rm",
		"--network="+cfg.Network,
		"-v", torrcFile.Name()+":/etc/tor/torrc",
		"-v", ogDir+":/var/lib/tor",
		"cmars/oniongrok-tor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start tor: %v (is docker installed?)", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
