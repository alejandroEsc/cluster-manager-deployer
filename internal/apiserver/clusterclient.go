package apiserver

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tcmd "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	"sigs.k8s.io/cluster-api/pkg/clientcmd"
	"sigs.k8s.io/cluster-api/pkg/util"

	"github.com/golang/glog"
)

const (
	retryIntervalKubectlApply   = 10 * time.Second
	retryIntervalResourceReady  = 10 * time.Second
	retryIntervalResourceDelete = 10 * time.Second
	timeoutKubectlApply         = 15 * time.Minute
	timeoutResourceReady        = 15 * time.Minute
	timeoutMachineReady         = 30 * time.Minute
	timeoutResourceDelete       = 15 * time.Minute
)

type clusterClient struct {
	clientSet       clientset.Interface
	kubeconfigFile  string
	configOverrides tcmd.ConfigOverrides
	closeFn         func() error
	namespace       string
}

// NewClusterClient creates and returns the address of a clusterClient, the kubeconfig argument is expected to be the string represenattion
// of a valid kubeconfig.
func NewClusterClient(kubeconfig string, namespace string) (*clusterClient, error) {
	f, err := createTempFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	defer ifErrRemove(&err, f)
	c, err := NewClusterClientFromDefaultSearchPath(namespace, f, clientcmd.NewConfigOverrides())
	if err != nil {
		return nil, err
	}
	c.closeFn = c.removeKubeconfigFile
	return c, nil
}

func (c *clusterClient) removeKubeconfigFile() error {
	return os.Remove(c.kubeconfigFile)
}

// NewClusterClientFromDefaultSearchPath creates and returns the address of a clusterClient, the kubeconfigFile argument is expected to be the path to a
// valid kubeconfig file.
func NewClusterClientFromDefaultSearchPath(namespace string, kubeconfigFile string, overrides tcmd.ConfigOverrides) (*clusterClient, error) {
	c, err := clientcmd.NewClusterApiClientForDefaultSearchPath(kubeconfigFile, overrides)
	if err != nil {
		return nil, err
	}

	return &clusterClient{
		kubeconfigFile:  kubeconfigFile,
		clientSet:       c,
		configOverrides: overrides,
		namespace:namespace,
	}, nil
}

// Frees resources associated with the cluster client
func (c *clusterClient) Close() error {
	if c.closeFn != nil {
		return c.closeFn()
	}
	return nil
}

func (c *clusterClient) Delete(manifest string) error {
	return c.kubectlDelete(manifest)
}

func (c *clusterClient) Apply(manifest string) error {
	return c.waitForKubectlApply(manifest)
}

func (c *clusterClient) kubectlDelete(manifest string) error {
	return c.kubectlManifestCmd("delete", manifest)
}

func (c *clusterClient) kubectlApply(manifest string) error {
	return c.kubectlManifestCmd("apply", manifest)
}

func (c *clusterClient) kubectlManifestCmd(commandName, manifest string) error {
	cmd := exec.Command("kubectl", c.buildKubectlArgs(commandName)...)
	cmd.Stdin = strings.NewReader(manifest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't kubectl apply: %v, output: %s", err, string(out))
	}
	return nil
}

func (c *clusterClient) buildKubectlArgs(commandName string) []string {
	args := []string{commandName}
	if c.kubeconfigFile != "" {
		args = append(args, "--kubeconfig", c.kubeconfigFile)
	}
	if c.configOverrides.Context.Cluster != "" {
		args = append(args, "--cluster", c.configOverrides.Context.Cluster)
	}
	if c.configOverrides.Context.Namespace != "" {
		args = append(args, "--namespace", c.configOverrides.Context.Namespace)
	}
	if c.configOverrides.Context.AuthInfo != "" {
		args = append(args, "--user", c.configOverrides.Context.AuthInfo)
	}
	return append(args, "-f", "-")
}

func (c *clusterClient) waitForKubectlApply(manifest string) error {
	err := util.PollImmediate(retryIntervalKubectlApply, timeoutKubectlApply, func() (bool, error) {
		glog.V(2).Infof("Waiting for kubectl apply...")
		err := c.kubectlApply(manifest)
		if err != nil {
			if strings.Contains(err.Error(), "refused") {
				// Connection was refused, probably because the API server is not ready yet.
				glog.V(4).Infof("Waiting for kubectl apply... server not yet available: %v", err)
				return false, nil
			}
			if strings.Contains(err.Error(), "unable to recognize") {
				glog.V(4).Infof("Waiting for kubectl apply... api not yet available: %v", err)
				return false, nil
			}
			if strings.Contains(err.Error(), "namespaces \"default\" not found") {
				glog.V(4).Infof("Waiting for kubectl apply... default namespace not yet available: %v", err)
				return false, nil
			}
			return false, err
		}

		return true, nil
	})

	return err
}

func waitForClusterResourceReady(namespace string, cs clientset.Interface) error {
	deadline := time.Now().Add(timeoutResourceReady)
	err := util.PollImmediate(retryIntervalResourceReady, timeoutResourceReady, func() (bool, error) {
		glog.V(2).Info("Waiting for Cluster v1alpha resources to become available...")
		_, err := cs.Discovery().ServerResourcesForGroupVersion("cluster.k8s.io/v1alpha1")
		if err == nil {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}
	timeout := time.Until(deadline)
	return util.PollImmediate(retryIntervalResourceReady, timeout, func() (bool, error) {
		glog.V(2).Info("Waiting for Cluster v1alpha resources to be listable...")
		_, err := cs.ClusterV1alpha1().Clusters(namespace).List(metav1.ListOptions{})
		if err == nil {
			return true, nil
		}
		return false, nil
	})
}


func createTempFile(contents string) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer ifErrRemove(&err, f.Name())
	if err = f.Close(); err != nil {
		return "", err
	}
	err = ioutil.WriteFile(f.Name(), []byte(contents), 0644)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func ifErrRemove(pErr *error, path string) {
	if *pErr != nil {
		if err := os.Remove(path); err != nil {
			glog.Warningf("Error removing file '%v': %v", err)
		}
	}
}
