package cmd

import (
	"github.com/Nasrin-Shiraly/ports/pkg/composeFile"
	"github.com/Nasrin-Shiraly/ports/pkg/directory"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
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

type GrabberOptions struct {
	logger           *log.Entry
	DirectoryHandler directory.IHandler
	ComposeHandler   composeFile.IHandler
	Directory        string
	Port             string
}

func NewGrabberOptions() GrabberOptions {
	return GrabberOptions{}
}
func (o *GrabberOptions) Complete(logger *log.Entry) {
	o.logger = logger.WithFields(log.Fields{
		"directory": o.Directory,
		"port":      o.Port,
	})
}
func NewCmdGrabber(logger *log.Entry) *cobra.Command {
	o := NewGrabberOptions()
	longDescription := `grab next available ports for docker compose files in --directory
		if --port is passed in, the next port for port that range is shown
		--port also supports regex like search.
		for example --port 80xx will return next available four digit long ports that start with 80`
	cmd := &cobra.Command{
		Use:   "grab --directory DIRECTORY_PATH --port PORT",
		Short: "grab next available ips",
		Long:  longDescription,
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(logger)
			o.Run()
		},
	}
	cmd.Flags().StringVar(&o.Directory, "directory", ".", "find docker compose files in this directory")
	cmd.Flags().StringVar(&o.Port, "port", "", "find open ports for this sequence")
	return cmd
}

func (o GrabberOptions) Run() {

	o.DirectoryHandler = directory.NewDirectoryHandler(o.Directory)
	o.ComposeHandler = composeFile.NewPortHandler(o.Port)
	// walk directory. Get all docker compose files. return list of file paths.
	composeFiles, err := o.DirectoryHandler.Find()
	if err != nil {
		o.logger.Errorf("%v", err)
		return
	}
	nextPorts, err := o.ComposeHandler.Ports(composeFiles)
	if err != nil {
		o.logger.Errorf("%v", err)
		return
	}
	o.logger.Infof("%v", nextPorts)
	return
}
