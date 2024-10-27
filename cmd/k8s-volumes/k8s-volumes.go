package k8svolumes

// Add new command:
// - k8s volume-replant

// I have a volume on k8s, I need to stop the deployment / pod that uses it
// then, I need to spin up a pod and attach the volume to it
// I need to run kubectl cp command to get the volume contents
// I need to create a new volume using longhorn
// I need to copy the contents of the volume in the new location.
// Attach the old depoyment to the new volume

import (
	"log"
	"zxcvmk/pkg/config"
	"k8s.io/client-go/tools/clientcmd"
	k8sapi "k8s.io/client-go/tools/clientcmd/api"
)

type K8sArguments struct {
	Pvc        string
	Deployment string
	Namespace  string
	DryRun     bool
}

func makeClient() *k8sapi.Config {
	kubeconfiglocation := "~/.kube/config"
	kubeconfig, err := clientcmd.LoadFromFile(kubeconfiglocation)
	if err != nil {
		log.Fatalf("Can't load config from: %s", kubeconfiglocation)
	}
	return kubeconfig
}

func Replant(cfg *config.Config, backupArguments K8sArguments) {
	log.Println("Can't replant. Not implemented")
	log.Printf("%#v", backupArguments)

}

// find what deployment is using vpc
