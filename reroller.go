package reroller

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"strconv"
)

const annotation = "reroller.roob.re/reroll"

type Reroller struct {
	K8S         *kubernetes.Clientset
	Namespace   string
	Unannotated bool
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, err
	}

	if kubeconfig == "" {
		return nil, errors.New("could not get config from env and kubeconfig path is empty: " + err.Error())
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err == nil {
		return config, err
	}

	return nil, err
}

func New() (*Reroller, error) {
	return NewWithKubeconfig("")
}

func NewWithKubeconfig(kubeconfig string) (*Reroller, error) {
	rr := &Reroller{}
	// creates the in-cluster config
	config, err := restConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	rr.K8S = clientset
	return rr, nil
}

func (rr *Reroller) Run() {
	var rollouts []Rollout
	rollouts = append(rollouts, rr.deploymentRollouts()...)
	rollouts = append(rollouts, rr.daemonSetRollouts()...)

	for _, rollout := range rollouts {
		if !rr.shouldReroll(rollout.Annotations()) {
			log.Printf("%s is not annotated, skipping", rollout.Name())
			continue
		}

		if !rollout.HasAlwaysPullPolicy() {
			log.Printf("%s does not have pullPolicy == Always, skipping", rollout.Name())
			continue
		}

		statuses, err := rollout.ContainerStatuses()
		if err != nil {
			log.Printf("error getting container statuses: %v", err)
			continue
		}

		if rr.hasUpdate(statuses) {
			log.Println("Restarting something")
			//rollout.Restart()
		}
	}
}

func (rr *Reroller) deploymentRollouts() (rollouts []Rollout) {
	deployments, err := rr.K8S.AppsV1().Deployments(rr.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, depl := range deployments.Items {
		if depl.Status.AvailableReplicas > 0 {
			rollouts = append(rollouts, DeploymentRollout(rr.K8S, &depl))
		}
	}

	return
}

func (rr *Reroller) daemonSetRollouts() (rollouts []Rollout) {
	deployments, err := rr.K8S.AppsV1().DaemonSets(rr.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, ds := range deployments.Items {
		if true {
			rollouts = append(rollouts, DaemonSetRollout(rr.K8S, &ds))
		}
	}

	return
}

func (rr *Reroller) shouldReroll(annotations map[string]string) bool {
	rawVal, found := annotations[annotation]
	var val bool

	if !found {
		val = rr.Unannotated
	} else {
		val, _ = strconv.ParseBool(rawVal)
	}

	return val
}

// TODO
func (rr *Reroller) hasUpdate(statuses []v1.ContainerStatus) bool {
	for _, status := range statuses {
		fmt.Printf("Checking update for %s:%s\n", status.Image, status.ImageID)
	}

	return false
}
