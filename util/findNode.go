package util

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

var (
	UnknownPod = errors.New("unknown pod")
	Client     *kubernetes.Clientset
	Node       string
)

func BuildClient() {
	config, _ := rest.InClusterConfig()
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorf("Failed to create a Kubernetes API Client: %v", err)
	}
	Client = client
	logrus.Info("Kubernetes API Client built successfully!")
}

func FindKubernetesNode() error {
	host, _ := os.Hostname()
	pod, err := Client.CoreV1().Pods("del").Get(context.TODO(), host, v1.GetOptions{})
	if err != nil {
		logrus.Errorf("Failed fetching pod info: %v", err)
		return UnknownPod
	}
	Node = pod.Spec.NodeName
	return nil
}
