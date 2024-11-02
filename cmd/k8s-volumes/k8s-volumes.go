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
	"bytes"
	"context"
	"fmt"
	"os"
	"time"
	"zxcvmk/pkg/config"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"log/slog"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type K8sArguments struct {
	Pvc                  string
	Deployment           string
	Namespace            string
	DestVolumeSize       string
	DestStorageClassName string
	DryRun               bool
}

func Replant(cfg *config.Config, k8sArguments K8sArguments) {
	// todo: add support for KUBECONFIG

	clientset, err := getClientSet()
	if err != nil {
		slog.Error("cannot get clientset: {err}", "err", err)
		return
	}

	d, err := findPvcUseDeployment(k8sArguments, clientset)
	if err != nil {
		slog.Error("could not find deployment", "error", err)
		return
	}
	slog.Debug("Deployment found", "deployment", d.Name, "replicas", d.Spec.Replicas)

	scaleDownDeployment(k8sArguments, d, clientset)

	slog.Info("creating temporary pod")
	pod, err := createTemporaryPod(k8sArguments, clientset)

	if err != nil {
		slog.Error("could not create pod", "error", err)
		return
	}
	defer cleanupPod(k8sArguments, pod.Name, clientset)

	pvc, err := createTargetPvc(k8sArguments, clientset)
	if err != nil {
		slog.Error("could not create pvc", "error", err)
		return
	}
	defer cleanupPvc(k8sArguments, pvc.Name, clientset)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	done := make(chan struct{})
	go func() {
		err := waitTillRunning(ctx, clientset, pod)
		if err != nil {
			slog.Error("cannot wait till running", "error", err)
		}
		close(done)
	}()

	select {
	case <-done:
		phase, err := getPodStatusPhase(clientset, pod)
		if err != nil {
			slog.Error("cannot get pod status, timeout")
			return
		}
		if phase == corev1.PodRunning {
			slog.Debug("pod status is Running. continue.")
		}
	case <-ctx.Done():
		slog.Error("cannot wait till running, timeout")
		return
	}

	slog.Info("pod created", "pod", pod.Name, "mounts", pod.Spec.Containers[0].VolumeMounts)
	slog.Info("pvc created", "pvc", pvc.Name, "mounts", pvc.Spec.VolumeName)

	err = transferVolumeContents(clientset, pod)
	if err != nil {
		slog.Error("cannot transfer volume contents", "error", err)

	}
	time.Sleep(time.Duration(10) * time.Second)

}

func getPodStatusPhase(clientset *kubernetes.Clientset, pod *corev1.Pod) (corev1.PodPhase, error) {
	podStatus, err := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).Get(context.TODO(), pod.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("cannot rsync src to dst: %w", err)
	}

	return podStatus.Status.Phase, nil

}

func runCmdOnAPod(clientset *kubernetes.Clientset, pod *corev1.Pod, command []string) error {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return fmt.Errorf("cannot cmd src to dst: %w", err)
	}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.ObjectMeta.Name).
		Namespace(pod.ObjectMeta.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Command: command,
		Stdout:  true,
		Stderr:  true,
	}, runtime.NewParameterCodec(&runtime.Scheme{}))

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("cannot cmd src to dst: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("cannot cmd src to dst: %w", err)
	}

	slog.Debug("cmd output", "output", string(stdout.Bytes()))
	return nil
}

func transferVolumeContents(clientset *kubernetes.Clientset, pod *corev1.Pod) error {
	err := runCmdOnAPod(clientset, pod, []string{"apk", "add", "--no-cache", "rsync"})
	if err != nil {
		return err
	}
	return runCmdOnAPod(clientset, pod, []string{"rsync", "-a", "--progress", "/source", "/destination"})
}

func cleanupPod(args K8sArguments, podName string, clientset *kubernetes.Clientset) {
	clientset.CoreV1().Pods(args.Namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
}

func cleanupPvc(args K8sArguments, pvcName string, clientset *kubernetes.Clientset) {
	clientset.CoreV1().PersistentVolumeClaims(args.Namespace).Delete(context.TODO(), pvcName, metav1.DeleteOptions{})
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

func findPvcUseDeployment(k8sArgs K8sArguments, clientset *kubernetes.Clientset) (*v1.Deployment, error) {
	slog.Info("Namespace", "ns", k8sArgs.Namespace)
	// pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot get pvc %s: %w", k8sArgs.Pvc, err)
	// }

	deployments, err := clientset.AppsV1().Deployments(k8sArgs.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get deployment %s: %w", k8sArgs.Pvc, err)
	}
	for _, deployment := range deployments.Items {
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == k8sArgs.Pvc {
				return &deployment, nil
			}
		}
	}
	return nil, nil
}

func scaleDownDeployment(k8sArgs K8sArguments, d *v1.Deployment, clientset *kubernetes.Clientset) {
	newReplicasCount := int32(0)
	d.Spec.Replicas = &newReplicasCount
	slog.Info("scaling deployment to 0", "deployment", d.Name, "namespace", d.Namespace)
	clientset.AppsV1().Deployments(k8sArgs.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
}

func createTargetPvc(k8sArgs K8sArguments, clientset *kubernetes.Clientset) (*corev1.PersistentVolumeClaim, error) {
	storageclassname := k8sArgs.DestStorageClassName
	storageQuantity := resource.MustParse(k8sArgs.DestVolumeSize)

	// Initialize a ResourceList with the desired memory request
	resourceList := corev1.ResourceList{
		corev1.ResourceStorage: storageQuantity,
	}
	newpvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sArgs.Pvc + "-v2",
			Namespace: k8sArgs.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageclassname,
			Resources: corev1.VolumeResourceRequirements{
				Requests: resourceList,
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
		},
	}

	pvccr, err := clientset.CoreV1().PersistentVolumeClaims(k8sArgs.Namespace).Create(context.TODO(), &newpvc, metav1.CreateOptions{})
	slog.Info("created new pvc", "pvc", "spec", pvccr.Name, pvccr.Spec)
	if err != nil {
		return nil, fmt.Errorf("cannot create target pvc: %w", err)
	}
	return pvccr, nil
}

func createTemporaryPod(k8sArgs K8sArguments, clientset *kubernetes.Clientset) (*corev1.Pod, error) {
	tempPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "temppod",
			Namespace: k8sArgs.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "transfer-container",
					Image:   "alpine:latest",
					Command: []string{"/bin/sh", "-c", "while :; do sleep 2073600; done"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "source",
							MountPath: "/source",
						},
						{
							Name:      "destination",
							MountPath: "/destination",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "source",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: k8sArgs.Pvc,
						},
					},
				},
				{
					Name: "destination",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: k8sArgs.Pvc + "-v2",
						},
					},
				},
			},
		},
	}
	newpod, err := clientset.CoreV1().Pods(k8sArgs.Namespace).Create(context.TODO(), &tempPod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot create pod: %w", err)
	}
	return newpod, nil
}

func waitTillRunning(ctx context.Context, clientset *kubernetes.Clientset, pod *corev1.Pod) error {
	watch, err := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).Watch(ctx, metav1.ListOptions{})

	if err != nil {
		return fmt.Errorf("cannot wait till running: %w", err)
	}
	defer watch.Stop()

	for {
		select {
		case event, ok := <-watch.ResultChan():
			if !ok {
				return fmt.Errorf("cannot wait till running: %w", err)
			}
			p, ok := event.Object.(*corev1.Pod)
			if !ok {
				slog.Error("unexpected type", "type", event.Type)
				continue
			}

			if p.Status.Phase == corev1.PodRunning {
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//func getDeploymentsinNamespace(namespace string, clientset *kubernetes.Clientset) []Deployment

// find what deployment is using pvc
