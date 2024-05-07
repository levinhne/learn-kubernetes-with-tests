package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	deploymentRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	podObject := &apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hellok8s",
			Namespace: apiv1.NamespaceDefault,
			Labels: map[string]string{
				"ok": "1",
			},
		},
		Spec: apiv1.PodSpec{
			InitContainers: []apiv1.Container{
				{
					Name:  "hellok8s-init-container",
					Image: "levinhne/hellok8s:v1",
					Command: []string{
						"env",
					},
				},
			},
			Containers: []apiv1.Container{
				{
					Name:  "hellok8s-container",
					Image: "levinhne/hellok8s:v1",
					Env: []apiv1.EnvVar{
						{
							Name:  "port",
							Value: "3000",
						},
					},
					Lifecycle: &apiv1.Lifecycle{
						PostStart: &apiv1.LifecycleHandler{Exec: &apiv1.ExecAction{Command: []string{"echo", "PostStart"}}},
					},
				},
			},
		},
	}
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name": "hellok8s",
				"lables": map[string]string{
					"ok": "1",
				},
			},
			"spec": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"name":  "hellok8s-container",
						"image": "levinhne/hellok8s:v1",
					},
				},
			},
		},
	}
	pod.Object["metadata"] = podObject.GetObjectMeta()
	pod.Object["spec"] = podObject.Spec
	_, err = client.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	prompt()
	list, err := client.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "ok=2",
	})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(d.GetName())
	}
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
