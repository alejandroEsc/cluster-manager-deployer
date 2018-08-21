package main



import (
	"github.com/spf13/viper"
	"github.com/alejandroEsc/cluster-manager-deployer/internal/options"
	flag "github.com/spf13/pflag"
)

const (
	envVarKubeconfigFileKey = "MANAGER_DEPLOYER_KUBECONFIG"

	keyKubeconfig = "kubeconfig-file"
)


func bindEnvVars() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MANAGER_DEPLOYER")
	viper.BindEnv(keyKubeconfig, envVarKubeconfigFileKey)
}


func initEnvDefaults() {
	viper.SetDefault(keyKubeconfig, "")
}


func bindCommonFlags(o *options.ManagerDeployerOptions, fs *flag.FlagSet) {
	fs.StringVarP(&o.KubeconfigFile, "kubeconfig", "k", viper.GetString(keyKubeconfig), "kubeconfig file path")
}