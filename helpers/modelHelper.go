package helper

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	v1d "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var Username string = "admin"
var Password string = "Harbor12345"
var Registry string = "core.harbor.domain"

func PushToDocker(folderName string, name string, modelVersion string) string {

	USERNAME, exists := os.LookupEnv("USERNAME")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	PASSWORD, exists := os.LookupEnv("PASSWORD")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	REGISTRY, exists := os.LookupEnv("REGISTRY")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	libRegEx, e := regexp.Compile("cog.yaml")
	if e != nil {
		log.Fatal(e)
	}

	e = filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		if err == nil && libRegEx.MatchString(info.Name()) {
			fmt.Println("Info Path "+info.Name(), path)
			// pathModel = path
			fmt.Println("Folder Name " + folderName)
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}

	pathDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(pathDir)
	name = strings.ToLower(name)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	fmt.Println("Start Running COG")
	cmd := exec.Command("cog", "build", "-t", name)
	cmd.Dir = pathDir + "/" + folderName

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running command COG:", err)
	} else {
		fmt.Println("Command output:", string(output))
	}

	fmt.Println("Get Image Name")
	// docker images name
	cmd1, err := exec.Command("docker", "images", name).Output()
	// cmd1, err := exec.Command("ls").Output()
	if err != nil {
		fmt.Println(err)
	}
	dockerOutput := string(cmd1)
	fmt.Println(dockerOutput)
	dockerDescript := strings.Split(dockerOutput, "\n")
	dockerDetail := strings.Fields(dockerDescript[1])
	dockerImageID := dockerDetail[2]
	dockerName := dockerDetail[0]
	fmt.Println(dockerImageID)

	authConfig := types.AuthConfig{
		Username:      USERNAME,
		Password:      PASSWORD,
		ServerAddress: REGISTRY,
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	imageTag := REGISTRY + "/library/" + dockerName + ":" + modelVersion

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
	defer cli.Close()
	return imageTag
}

func CreateService(name string, serviceName string) error {
	// Use the current context in kubeconfig
	config, err := GetKubeConfig()
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
			Name:      serviceName + "-service",
			Namespace: "streaming",
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
	_, err = clientset.CoreV1().Services("streaming").Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func CreateDeployment(name, serviceName, imageURL string) error {
	// Use the current context in kubeconfig
	config, err := GetKubeConfig()
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
			Name:      serviceName + "-service",
			Namespace: "streaming",
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
	_, err = clientset.AppsV1().Deployments("streaming").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func DeleteDeploymentAndService(deploymentName string, serviceName string) error {
	// Use the current context in kubeconfig
	config, err := GetKubeConfig()
	if err != nil {
		return err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Delete the Deployment
	err = clientset.AppsV1().Deployments("streaming").Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Delete the Service
	err = clientset.CoreV1().Services("streaming").Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func GetServiceURL(name string) (string, string, error) {
	return "http://" + name + "-service:5000", "http://" + name + "-service:5000" + "/predictions", nil
}

func GetKubeConfig() (*rest.Config, error) {
	// homeDir := os.Getenv("HOME")
	// config, err := clientcmd.BuildConfigFromFlags("", homeDir+"/.kube/config")
	// if err != nil {
	// 	return nil, err
	// }
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return config, nil
}
