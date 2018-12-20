package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	jobName := "job-name"
	command := []string{
		"python",
		"manage.py",
		"migrate",
	}
	tag := "imagetag"
	imageName := fmt.Sprintf("image_url:%s", tag)
	var job batchv1.Job
	jobContainer := corev1.Container{
		Name:    jobName,
		Image:   imageName,
		Command: command,
		EnvFrom: []corev1.EnvFromSource{
			corev1.EnvFromSource{
				Prefix: "",
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-name",
					},
				},
			},
			corev1.EnvFromSource{
				Prefix: "",
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "secret-name",
					},
				},
			},
		},
	}

	job = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						jobContainer,
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	jobs, err := clientset.BatchV1().Jobs("default").Create(&job)
	if err != nil {
		log.Printf("%+v", err)
	}
	err = clientset.BatchV1().Jobs("default").Delete(jobName, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("%+v", err)
	}
	fmt.Print(jobs)

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
