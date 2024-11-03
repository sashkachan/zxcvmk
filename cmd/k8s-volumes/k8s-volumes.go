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
	"log/slog"
	"os"
	"time"
	"zxcvmk/pkg/config"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type K8sArguments struct {
	PvcSrc               string
	PvcDst               string
	Deployment           string
	DeploymentVolumeName string
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

	var deployment *v1.Deployment

	if k8sArguments.Deployment != "" {
		deployment, err = findPvcUseDeployment(k8sArguments, clientset)
		if err != nil {
			slog.Error("could not find deployment", "error", err)
			return
		}
		if deployment == nil {
			slog.Info("could not find deployment, will only migrate volume")
		} else {
			slog.Info("Deployment found", "deployment", deployment.Name, "spec", deployment.Spec)

			deployment, err = scaleDownDeployment(k8sArguments, deployment, clientset)
			if err != nil {
				slog.Error("could scale down deployment", "error", err)
				return
			}
		}
	}
	slog.Info("creating temporary pod")
	pod, err := createTemporaryPod(k8sArguments, clientset)
	time.Sleep(3 * time.Second)
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
	// defer cleanupPvc(k8sArguments, pvc.Name, clientset)

	phase, err := getPodStatusPhase(clientset, pod)
	slog.Debug("pod status phase retrieved", "phase", phase)
	if err != nil {
		slog.Error("cannot get pod status, timeout")
		cleanupPvc(k8sArguments, pvc.Name, clientset)
		return
	}
	if phase == corev1.PodRunning {
		slog.Debug("pod status is Running. continue.")
	}

	slog.Info("pod created", "pod", pod.Name, "mounts", pod.Spec.Containers[0].VolumeMounts)
	slog.Info("pvc created", "pvc", pvc.Name, "mounts", pvc.Spec.VolumeName)

	err = transferVolumeContents(clientset, pod)
	if err != nil {
		slog.Error("cannot transfer volume contents", "error", err)
		cleanupPvc(k8sArguments, pvc.Name, clientset)
		return
	}
	if deployment != nil {
		_, err = mountNewVolumesOnDeployment(k8sArguments, deployment, pvc, clientset)
		if err != nil {
			slog.Error("cannot restore deployment to previous state with the new volume", "error", err)
			cleanupPvc(k8sArguments, pvc.Name, clientset)
			return
		}
	}
	slog.Info("transfer complete")
}

func getPodStatusPhase(clientset *kubernetes.Clientset, pod *corev1.Pod) (corev1.PodPhase, error) {
	for attempts := 0; attempts < 20; attempts++ {
		podStatus, err := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("cannot get pod status phase: %w", err)
		}

		if podStatus.Status.Phase == corev1.PodRunning {
			return podStatus.Status.Phase, nil
		}

		time.Sleep(5 * time.Second)
	}

	return "", fmt.Errorf("pod did not reach Running state within expected time")
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

	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	parameterCodec := runtime.NewParameterCodec(scheme)
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.ObjectMeta.Name).
		Namespace(pod.ObjectMeta.Namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Command: command,
		Stdout:  true,
		Stderr:  true,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("cannot cmd src to dst: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Stdin:  nil,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("cannot cmd src to dst: %w", err)
	}

	slog.Debug("cmd output", "output", stdout.String())
	return nil
}

func transferVolumeContents(clientset *kubernetes.Clientset, pod *corev1.Pod) error {
	err := runCmdOnAPod(clientset, pod, []string{"apk", "add", "--no-cache", "rsync"})
	if err != nil {
		return err
	}
	return runCmdOnAPod(clientset, pod, []string{"rsync", "-a", "--progress", "/source/", "/destination"})
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
	deployments, err := clientset.AppsV1().Deployments(k8sArgs.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get deployment %s: %w", k8sArgs.PvcSrc, err)
	}
	for _, deployment := range deployments.Items {
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == k8sArgs.PvcSrc {
				return &deployment, nil
			}
		}
	}
	return nil, nil
}

func scaleDownDeployment(k8sArgs K8sArguments, d *v1.Deployment, clientset *kubernetes.Clientset) (*v1.Deployment, error) {
	newReplicasCount := int32(0)
	d.Spec.Replicas = &newReplicasCount
	slog.Info("scaling deployment to 0", "deployment", d.Name, "namespace", d.Namespace)
	d, err := clientset.AppsV1().Deployments(k8sArgs.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot scale down deployment: %w", err)
	}
	return d, nil
}

func mountNewVolumesOnDeployment(k8sArgs K8sArguments, d *v1.Deployment, pvc *corev1.PersistentVolumeClaim, clientset *kubernetes.Clientset) (*v1.Deployment, error) {
	d, err := clientset.AppsV1().Deployments(k8sArgs.Namespace).Get(context.TODO(), d.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get deployment: %w", err)
	}
	pvcFound := false
	for k, v := range d.Spec.Template.Spec.Volumes {
		if v.Name == k8sArgs.DeploymentVolumeName {
			d.Spec.Template.Spec.Volumes[k].VolumeSource = corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			}
			pvcFound = true
			slog.Debug("updating pvc source", "pvcName", pvc.Name)
		}
	}
	if !pvcFound {
		return nil, fmt.Errorf("cannot update deployment, volume not found: %s", pvc.Name)
	}

	newReplicasCount := int32(1)
	d.Spec.Replicas = &newReplicasCount
	slog.Info("scaling deployment to 0", "deployment", d.Name, "namespace", d.Namespace)
	d, err = clientset.AppsV1().Deployments(k8sArgs.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot update deployment, update failed: %w", err)
	}
	return d, nil
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
			Name:      k8sArgs.PvcDst,
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
							ClaimName: k8sArgs.PvcSrc,
						},
					},
				},
				{
					Name: "destination",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: k8sArgs.PvcDst,
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
