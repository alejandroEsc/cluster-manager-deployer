package main

import (
	"github.com/spf13/cobra"

	"github.com/juju/loggo"

	"github.com/alejandroEsc/cluster-manager-deployer/pkg/util"
	"fmt"
	"os"
	"github.com/alejandroEsc/cluster-manager-deployer/internal/options"
	"github.com/spf13/viper"
)




var (
	logLevel = "UNSPECIFIED"
	logger = util.GetModuleLogger("cmd.manager-deployer", loggo.UNSPECIFIED)
	rootOptions = &options.ManagerDeployerOptions{}
)


// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "manager-deployer",
	Short: "Deploys tools to create cluster-api manager to your existing cluster",
	Run: func(cmd *cobra.Command, args []string) {
		level, isValid := loggo.ParseLevel(logLevel)
		if isValid {
			logger.SetLogLevel(level)
		}

		cmd.Help()
	},
}


func init() {
	// init viper defaults
	initEnvDefaults()

	// root flags
	RootCmd.PersistentFlags().StringVarP(&rootOptions.KubeconfigFile, "kubeconfig", "k", viper.GetString(keyKubeconfig), "kubeconfig file path")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "verbose", "v", "UNSPECIFIED", "log level")

	// bind environment vars
	bindEnvVars()

	// add commands
	addCommands()
}


func addCommands() {
	RootCmd.AddCommand(deployCmd())
}


// Execute performs root command task.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func logError(err error) {
	if err != nil {
		logger.Errorf(err.Error())
	}
}

