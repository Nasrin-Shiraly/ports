package cmd

import (
	"fmt"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"cmd": "ports",
	})
	cmd := NewCmdGrabber(logger)
	rootCmd.AddCommand(cmd)
}

type Ports []interface{}

type ServiceSettings map[interface{}]map[string]interface{}
type Compose struct {
	Services ServiceSettings `yaml:"services"`
}
type GrabberOptions struct {
	logger    *log.Entry
	Directory string
	Port      string
	validator func() error
}

func NewGrabberOptions() GrabberOptions {
	return GrabberOptions{}
}
func (o *GrabberOptions) Complete(logger *log.Entry) {
	var portRequested bool
	if o.Port != "" {
		portRequested = true
	}
	o.logger = logger.WithFields(log.Fields{
		"directory":      o.Directory,
		"port_requested": portRequested,
		"port":           o.Port,
	})
	o.validator = func() error {
		if _, err := os.Stat(o.Directory); os.IsNotExist(err) {
			return err
		}
		return nil
	}
}
func (o GrabberOptions) find(directory string) ([]string, error) {
	composeFilePaths := []string{}
	// TODO: don't panic on inaccessible path
	err := filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".yml" {
			if strings.HasPrefix(info.Name(), "docker-compose") {
				composeFilePaths = append(composeFilePaths, path)
			}
		}
		return nil
	})
	o.logger.Infof("%v", composeFilePaths)
	return composeFilePaths, err
}
func (o GrabberOptions) Run() error {
	if err := o.validator(); err != nil {
		o.logger.Errorf("%v", err)
		return err
	}
	// walk directory. Get all docker compose files. return list of file paths.
	ymlFiles, err := o.find(o.Directory)
	if err != nil {
		o.logger.Errorf("%v", err)
		return err
	}
	composeFiles := []Compose{}
	// for each file, yml to go struct
	for _, f := range ymlFiles {
		compose := Compose{}
		bytes, err := ioutil.ReadFile(f)
		if err != nil {
			o.logger.Errorf("%v", err)
			return err
		}
		err = yaml.Unmarshal(bytes, &compose)
		if err != nil {
			o.logger.Errorf("%v", err)
			return err
		}
		composeFiles = append(composeFiles, compose)
	}
	// get list of all current ports
	// TODO: there's a long syntax too. add support.
	// https://docs.docker.com/compose/compose-file/compose-file-v3/
	occupiedPortRanges := []string{}
	for _, c := range composeFiles {
		for service, settings := range c.Services {
			log.Infof("%v", service)
			if p, ok := settings["ports"]; ok {
				var ports, ok = p.([]interface{})
				if !ok {
					return fmt.Errorf("no ports")
				}
				for _, port := range ports {
					occupiedPortRanges = append(occupiedPortRanges, fmt.Sprintf("%v", port))
				}
			}
		}
	}
	o.logger.Infof("%v", occupiedPortRanges)

	// create a list of next ports --> covert
	// if -port is provided filter based on it.
	return nil
}
func NewCmdGrabber(logger *log.Entry) *cobra.Command {
	o := NewGrabberOptions()
	longDescription := fmt.Sprintf(
		"grab next available ips for docker compose files in directory" +
			"if -p is passed in, the next port for port that range is shown")

	cmd := &cobra.Command{
		Use:   "grab -d DIRECTORY_PATH -p PORT",
		Short: "grab next available ips",
		Long:  longDescription,
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(logger)
			if err := o.Run(); err != nil {
				logger.Errorf("%v", err)
				return
			}
		},
	}
	cmd.Flags().StringVar(&o.Directory, "directory", ".", "find docker compose files in this directory")
	cmd.Flags().StringVar(&o.Port, "port", "", "find open ports for this sequence")
	return cmd
}
