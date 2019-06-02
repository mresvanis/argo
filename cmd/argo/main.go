package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mresvanis/argo/pkg/registry"
	"github.com/urfave/cli"
)

const Version = "0.1.0"

// populated at build-time with -ldflags
var VersionSuffix string

func main() {
	app := cli.NewApp()
	app.Name = "argo"
	app.Usage = "A simple JSON log forwarder to ES"
	app.HideVersion = false
	app.Version = Version
	if VersionSuffix != "" {
		app.Version = Version + "-" + VersionSuffix[:7]
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.json",
			Usage: "Load configuration from `FILE`",
		},
		cli.StringFlag{
			Name:  "registry, r",
			Value: "argo.db",
			Usage: "Use the specified bolt db `FILE`",
		},
	}
	app.Action = func(c *cli.Context) error {
		cfg, err := parseConfigFromCli(c)
		if err != nil {
			return err
		}
		reg, err := loadRegistryFromCli(c)
		if err != nil {
			return err
		}
		return StartProcess(cfg, reg)
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// StartProcess spawns the output and inputs given a config.
func StartProcess(cfg *Config, reg registry.Registrar) error {
	var wg sync.WaitGroup

	out := startOutput(cfg, &wg)
	fis := startInputs(cfg, out, reg, &wg)

	handleIntTermSignals(out, fis, reg, &wg)

	wg.Wait()
	log.Printf("process terminated")
	return nil
}

func parseConfigFromCli(c *cli.Context) (*Config, error) {
	f, err := os.Open(c.String("config"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse configuration; %s", err)
	}
	cfg, err := ParseConfig(f)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadRegistryFromCli(c *cli.Context) (registry.Registrar, error) {
	reg := registry.NewRegistry(c.String("registry"))

	err := reg.Open()
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func getUniquePaths(paths []string) []string {
	u := make([]string, 0, len(paths))
	m := make(map[string]bool)

	for _, path := range paths {
		if _, ok := m[path]; !ok {
			m[path] = true
			u = append(u, path)
		}
	}

	return u
}

func handleIntTermSignals(out Output, fis []Input, reg registry.Registrar, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		gracefulTerm := make(chan os.Signal, 1)
		signal.Notify(gracefulTerm, syscall.SIGINT, syscall.SIGTERM)
		sig := <-gracefulTerm
		log.Printf("process notified, %+v", sig)

		for _, fi := range fis {
			fi.Stop()
		}

		out.Stop()
	}()
}

func startInputs(cfg *Config, out Output, reg registry.Registrar, wg *sync.WaitGroup) []Input {
	paths := getUniquePaths(cfg.Paths)

	fis := make([]Input, 0, len(paths))
	for _, path := range paths {
		fi := NewFileInput(cfg, path, reg)

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			fi.Start(out.Input(), out.Subscribe(path))
		}(path)

		fis = append(fis, fi)
	}

	return fis
}

func startOutput(cfg *Config, wg *sync.WaitGroup) Output {
	out := NewEsOutput(cfg)

	wg.Add(1)
	go func() {
		defer wg.Done()
		out.Start()
	}()

	return out
}
