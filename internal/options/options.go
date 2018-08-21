package options


type ManagerDeployerOptions struct {
	KubeconfigFile string
}


type DeployClusterAPIOptions struct {
	ManagerDeployerOptions
}
