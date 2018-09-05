package main

import (
	"fmt"
	"os"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/alejandroEsc/cluster-manager-deployer/internal/options"
	"github.com/alejandroEsc/cluster-manager-deployer/internal/apiserver"

)

const (

	// launched apiserver aggregate in default namespace, cannot change at the moment.
	nameSpace = "default"
	// name of the aggregate apiserver
	name      = "clusterapi"
)

func deployClusterAPICmd() *cobra.Command {
	dco := &options.DeployClusterAPIOptions{}
	cmd := &cobra.Command{
		Use:   "cluster-api",
		Short: "Deploy the cluster-api stack to an existing k8s cluster.",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {

			if err := runDeployClusterAPICmd(dco); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

		},
	}
	fs := cmd.Flags()

	bindCommonFlags(&dco.ManagerDeployerOptions, fs)
	return cmd
}

func runDeployClusterAPICmd(o *options.DeployClusterAPIOptions) error {
	var kubeconfig string

	kubeconfigPath := o.KubeconfigFile

	logger.Infof("kubeconfig file: %s", kubeconfigPath)

	if kubeconfigPath != "" {
		if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
			return fmt.Errorf("file at %s does not exist", kubeconfigPath)
		}

		b, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return err
		}

		kubeconfig = string(b)

		if kubeconfig == "" {
			return fmt.Errorf("valid kubeconfig file is required")
		}

	} else {
		return fmt.Errorf("valid kubeconfig file path is required")
	}

	client, err := apiserver.NewClusterClient(kubeconfig, nameSpace)
	if err != nil {
		return fmt.Errorf("unable to create a kubernetes client object: %v", err)
	}

	yaml, err := apiserver.GetApiServerYaml(name, nameSpace)
	if err != nil {
		return fmt.Errorf("unable to generate apiserver yaml: %v", err)
	}

	err = client.Apply(yaml)
	if err != nil {
		return fmt.Errorf("unable to apply apiserver yaml: %v", err)
	}

	return nil
}
