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
	"context"
	"fmt"
	"os"
	"zxcvmk/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"log/slog"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sArguments struct {
	Pvc        string
	Deployment string
	Namespace  string
	DryRun     bool
}

func Replant(cfg *config.Config, backupArguments K8sArguments) {
	// todo: add support for KUBECONFIG

	clientset, err := getClientSet()
	if err != nil {
		slog.Error("cannot get clientset: {err}", "err", err)
		return
	}

	d, err := findPvcUseDeployment(backupArguments.Pvc, backupArguments.Namespace, clientset)
	if err != nil {
		slog.Error("could not find deployment: %s", err)
	}
	slog.Debug("Deployment found", "deployment", d)
}

func getClientSet() (*kubernetes.Clientset, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get userdir: %w", err)
	}
	kubeconfiglocation := homedir + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfiglocation)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot get clientset: %w", err)
	}
	return clientset, nil
}

func findPvcUseDeployment(pvcName, namespace string, clientset *kubernetes.Clientset) (*v1.Deployment, error) {
	slog.Info("Namespace", "ns", namespace)
	// pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot get pvc %s: %w", pvcName, err)
	// }

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get deployment %s: %w", pvcName, err)
	}
	for _, deployment := range deployments.Items {
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				return &deployment, nil
			}
		}
	}
	return nil, nil
}

//func getDeploymentsinNamespace(namespace string, clientset *kubernetes.Clientset) []Deployment

// find what deployment is using pvc
