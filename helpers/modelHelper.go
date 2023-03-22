package helper

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"

	v1d "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func PushToDocker(folderName string, name string, modelVersion string) string {
	libRegEx, e := regexp.Compile("cog.yaml")
	if e != nil {
		log.Fatal(e)
	}

	e = filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		if err == nil && libRegEx.MatchString(info.Name()) {
			println(info.Name(), path)
			// pathModel = path
			println(folderName)
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}

	commandCog := "cog"
	task := "build"
	cogTag := "-t"

	pathDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(pathDir)
	name = strings.ToLower(name)
	cmd := exec.Command(commandCog, task, cogTag, name)
	cmd.Dir = folderName
	if err := cmd.Run(); err != nil {
		fmt.Println("could not run command: ", err)
	}

	commandDocker := "docker"
	images := "images"
	// docker images name
	cmd1, err := exec.Command(commandDocker, images, name).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput := string(cmd1)
	dockerDescript := strings.Split(dockerOutput, "\n")
	dockerDetail := strings.Fields(dockerDescript[1])
	dockerImageID := dockerDetail[2]
	dockerName := dockerDetail[0]
	fmt.Println(dockerImageID)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	authConfig := types.AuthConfig{
		Username:      "admin",
		Password:      "Harbor12345",
		ServerAddress: "core.harbor.domain",
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	imageTag := "core.harbor.domain/library/" + dockerName + ":" + modelVersion

	err = cli.ImageTag(ctx, dockerImageID, imageTag)
	fmt.Print(err)
	pusher, err := cli.ImagePush(ctx, imageTag, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		panic(err)
	}
	defer pusher.Close()

	type ErrorMessage struct {
		Error string
	}
	var errorMessage ErrorMessage
	buffIOReader := bufio.NewReader(pusher)

	for {
		streamBytes, err := buffIOReader.ReadBytes('\n')
		fmt.Printf("%s", streamBytes)
		if err == io.EOF {
			break
		}
		json.Unmarshal(streamBytes, &errorMessage)
		if errorMessage.Error != "" {
			panic(errorMessage.Error)
		}
	}
	return imageTag
}

func CreateDeploy1(URL string, Name string, ServiceName string) {
	type MetadataStruct struct {
		Name string `yaml:"name"`
	}

	type MatchLabelsStruct struct {
		App string `yaml:"app"`
	}

	// ==============================================
	type SelectorStruct struct {
		MatchLabels MatchLabelsStruct `yaml:"matchLabels"`
	}

	type LabelsStruct struct {
		App string `yaml:"app"`
	}

	type TemplateMetadataStruct struct {
		Labels LabelsStruct `yaml:"labels"`
	}

	// ==============================================

	type PortsStruct struct {
		ContainerPort int `yaml:"containerPort"`
	}

	type ContainersStruct struct {
		Name  string `yaml:"name"`
		Image string `yaml:"image"`
		Ports []PortsStruct
	}

	// ==============================================

	type ImagePullSecretsStruct struct {
		Name string `yaml:"name"`
	}

	type TemplateSpecStruct struct {
		Containers       []ContainersStruct       `yaml:"containers"`
		ImagePullSecrets []ImagePullSecretsStruct `yaml:"imagePullSecrets"`
	}

	type TemplateStruct struct {
		Metadata TemplateMetadataStruct `yaml:"metadata"`
		Spec     TemplateSpecStruct     `yaml:"spec"`
	}

	type SpecStruct struct {
		Replicas int            `yaml:"replicas"`
		Selector SelectorStruct `yaml:"selector"`
		Template TemplateStruct `yaml:"template"`
	}

	type Deploy struct {
		Kind       string         `yaml:"kind"`
		ApiVersion string         `yaml:"apiVersion"`
		Metadata   MetadataStruct `yaml:"metadata"`
		Spec       SpecStruct     `yaml:"spec"`
	}
	DeployStruct := Deploy{
		Kind:       "Deployment",
		ApiVersion: "apps/v1",
		Metadata: MetadataStruct{
			Name: ServiceName + "-service",
		},
		Spec: SpecStruct{
			Replicas: 1,
			Selector: SelectorStruct{
				MatchLabels: MatchLabelsStruct{
					App: Name,
				},
			},
			Template: TemplateStruct{
				Metadata: TemplateMetadataStruct{
					Labels: LabelsStruct{
						App: Name,
					},
				},
				Spec: TemplateSpecStruct{
					Containers: []ContainersStruct{{
						Name:  Name,
						Image: URL,
						Ports: []PortsStruct{{
							ContainerPort: 5000,
						}},
					}},
					ImagePullSecrets: []ImagePullSecretsStruct{
						{
							Name: "regcred",
						},
					},
				},
			},
		},
	}

	DeployYaml, err := yaml.Marshal(&DeployStruct)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	err2 := ioutil.WriteFile("Deployment.yaml", DeployYaml, 0777)

	if err2 != nil {

		log.Fatal(err2)
	}
}

func CreateService1(Name string, ServiceName string) {
	type LabelsStruct struct {
		App string `yaml:"app"`
	}

	type MetadataStruct struct {
		Name   string       `yaml:"name"`
		Labels LabelsStruct `yaml:"labels"`
	}

	type PortsStruct struct {
		Port       int    `yaml:"port"`
		TargetPort int    `yaml:"targetPort"`
		Protocol   string `yaml:"protocol"`
	}

	type SelectorStruct struct {
		App string `yaml:"app"`
	}

	type SpecStruct struct {
		Type     string `yaml:"type"`
		Ports    []PortsStruct
		Selector SelectorStruct `yaml:"selector"`
	}

	type Service struct {
		ApiVersion string         `yaml:"apiVersion"`
		Kind       string         `yaml:"kind"`
		Metadata   MetadataStruct `yaml:"metadata"`
		Spec       SpecStruct     `yaml:"spec"`
	}

	ServiceStruct := Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: MetadataStruct{
			Name: ServiceName + "-service",
			Labels: LabelsStruct{
				App: Name,
			},
		},
		Spec: SpecStruct{
			Type: "LoadBalancer",
			Ports: []PortsStruct{
				{
					Port:       5000,
					TargetPort: 5000,
					Protocol:   "TCP",
				},
			},
			Selector: SelectorStruct{
				App: Name,
			},
		},
	}

	ServiceYaml, err := yaml.Marshal(&ServiceStruct)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	err2 := ioutil.WriteFile("Service.yaml", ServiceYaml, 0777)

	if err2 != nil {

		log.Fatal(err2)
	}
}

func DeployKube(Name string) (string, string) {
	DeployToKubeCmd, err := exec.Command("kubectl", "apply", "-f", ".").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(DeployToKubeCmd)

	GetPredictUrlCmd, err := exec.Command("minikube", "service", "--url", Name).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(GetPredictUrlCmd)
	PodURL := strings.Trim(string(GetPredictUrlCmd), "\n")
	return PodURL, PodURL + "/predictions"
}

func CreateService(name string, serviceName string) error {
	// Use the current context in kubeconfig
	homeDir := os.Getenv("HOME")
	config, err := clientcmd.BuildConfigFromFlags("", homeDir+"/.kube/config")
	if err != nil {
		return err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Create the Service object
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName + "-service",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []v1.ServicePort{
				{
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
					Protocol:   "TCP",
				},
			},
			Selector: map[string]string{
				"app": name,
			},
		},
	}

	// Create the Service
	_, err = clientset.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func CreateDeployment(name, serviceName, imageURL string) error {
	// Use the current context in kubeconfig
	homeDir := os.Getenv("HOME")
	config, err := clientcmd.BuildConfigFromFlags("", homeDir+"/.kube/config")
	if err != nil {
		return err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Define the Deployment object
	deployment := &v1d.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName + "-service",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1d.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name,
							Image: imageURL,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 5000,
								},
							},
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "regcred",
						},
					},
				},
			},
		},
	}

	// Create the Deployment
	_, err = clientset.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func DeleteDeploymentAndService(deploymentName string, serviceName string) error {
	// Use the current context in kubeconfig
	homeDir := os.Getenv("HOME")
	config, err := clientcmd.BuildConfigFromFlags("", homeDir+"/.kube/config")
	if err != nil {
		return err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Delete the Deployment
	err = clientset.AppsV1().Deployments("default").Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Delete the Service
	err = clientset.CoreV1().Services("default").Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func GetServiceURL(name string) (string, string) {
	// Use the current context in kubeconfig
	GetPredictUrlCmd, err := exec.Command("minikube", "service", "--url", name+"-service").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(GetPredictUrlCmd)
	PodURL := strings.Trim(string(GetPredictUrlCmd), "\n")
	return PodURL, PodURL + "/predictions"
}
