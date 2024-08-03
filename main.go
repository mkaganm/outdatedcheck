// mymodulelinter/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

type Module struct {
	Path    string
	Version string
	Update  *struct {
		Path    string
		Version string
	}
}

const (
	red   = "\033[31m"
	reset = "\033[0m"
)

type Config struct {
	ModulePrefix string `yaml:"module_prefix"`
}

func loadConfig() (*Config, error) {
	file, err := os.Open(".mymodulelintrc.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func run() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// Run the go list -u -m -json all command
	cmd := exec.Command("go", "list", "-u", "-m", "-json", "all")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running go list command: %v", err)
	}

	hasNewVersion := false

	// Create a tab writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)

	// Table headers
	fmt.Fprintln(w, "MODULE\tVERSION\tNEW VERSION\tSTATUS")

	// Decode the modules from JSON
	dec := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var mod Module
		if err := dec.Decode(&mod); err != nil {
			break
		}

		// Check only modules starting with the configured prefix
		if strings.HasPrefix(mod.Path, config.ModulePrefix) {
			if mod.Update != nil {
				fmt.Fprintf(w, "%s\t%s\t%s\tOutdated\n", mod.Path, mod.Version, mod.Update.Version)
				hasNewVersion = true
			} else {
				fmt.Fprintf(w, "%s\t%s\t-\tUp-to-date\n", mod.Path, mod.Version)
			}
		}
	}

	// Flush the table writer
	w.Flush()

	if hasNewVersion {
		fmt.Println(red + "There are outdated modules." + reset)
		return fmt.Errorf("there are outdated modules")
	} else {
		fmt.Println("All modules are up-to-date.")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
