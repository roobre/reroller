package reroller

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"roob.re/reroller/registry"
	"strconv"
	"strings"
)

const rerollerAnnotation = "reroller.roob.re/reroll"

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
			log.Debugf("%s is not annotated, skipping", rollout.Name())
			continue
		}

		if !rollout.HasAlwaysPullPolicy() {
			log.Debugf("%s does not have pullPolicy == Always, skipping", rollout.Name())
			continue
		}

		statuses, err := rollout.ContainerStatuses()
		if err != nil {
			log.Errorf("error getting container statuses: %v", err)
			continue
		}

		if rr.hasUpdate(statuses) {
			log.Println("Restarting something")
			err := rollout.Restart()
			if err != nil {
				log.Errorf("error restarting %s: %v", rollout.Name(), err)
			}
		}
	}
}

func (rr *Reroller) deploymentRollouts() (rollouts []Rollout) {
	log.Debugf("Fetching deployments in %s", rr.Namespace)
	deployments, err := rr.K8S.AppsV1().Deployments(rr.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf(err.Error())
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
	log.Debugf("Fetching daemonSets in %s", rr.Namespace)
	daemonSets, err := rr.K8S.AppsV1().DaemonSets(rr.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, ds := range daemonSets.Items {
		if true {
			rollouts = append(rollouts, DaemonSetRollout(rr.K8S, &ds))
		}
	}

	return
}

func (rr *Reroller) shouldReroll(annotations map[string]string) bool {
	rawVal, found := annotations[rerollerAnnotation]
	var val bool

	if !found {
		val = rr.Unannotated
	} else {
		val, _ = strconv.ParseBool(rawVal)
	}

	return val
}

func (rr *Reroller) hasUpdate(statuses []v1.ContainerStatus) bool {
	for _, status := range statuses {
		imagePieces := strings.Split(status.ImageID, "@")
		if len(imagePieces) < 2 {
			log.Errorf("malformed imageID '%s', skipping upgrade check", status.ImageID)
			continue
		}
		digest := imagePieces[1]

		upstreamDigests, err := registry.ImageDigests(status.Image)
		if err != nil {
			log.Errorf("could not fetch latest digest for %s: %v", status.Image, err)
			continue
		}

		for _, ud := range upstreamDigests {
			if digest != ud {
				return true
			}
		}

		log.Debugf("no new digest found for %s", status.Image)
	}

	return false
}
