package helper

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"os"

	"gopkg.in/yaml.v2"
)

func PushToDocker(folderName string, name string) (string, string) {
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

	cmd := exec.Command(commandCog, task, cogTag, name)
	cmd.Dir = folderName
	if err := cmd.Run(); err != nil {
		fmt.Println("could not run command: ", err)
	}

	commandDocker := "docker"
	images := "images"
	login := "login"
	tag := "tag"
	push := "push"
	ip := "192.168.49.2:30001"
	userAuth := "-u"
	user := "admin"
	passAuth := "-p"
	password := "password"
	// docker images name
	cmd1, err := exec.Command(commandDocker, images, name).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput := string(cmd1)
	dockerDescript := strings.Split(dockerOutput, "\n")
	dockerDetail := strings.Fields(dockerDescript[1])
	dockerImageId := dockerDetail[2]

	fmt.Println(dockerImageId)

	// docker login 192.168.49.2:30001 -u 'admin' -p 'password'
	cmd2, err := exec.Command(commandDocker, login, ip, userAuth, user, passAuth, password).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput = string(cmd2)
	fmt.Println(dockerOutput)

	// docker tag {ImageID} 192.168.49.2:30001/{imageName}
	cmd3, err := exec.Command(commandDocker, tag, dockerImageId, ip+"/"+dockerImageId).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput = string(cmd3)
	fmt.Println(dockerOutput)

	// docker push 192.168.49.2:30001/{imageName}
	cmd4, err := exec.Command(commandDocker, push, ip+"/"+dockerImageId).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput = string(cmd4)
	fmt.Println(dockerOutput)
	return dockerImageId, ip + "/" + dockerImageId
}

func CreateDeploy(URL string, Name string) {
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
			Name: Name + "-service",
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
	err2 := ioutil.WriteFile("Deployment.yaml", DeployYaml, 0)

	if err2 != nil {

		log.Fatal(err2)
	}
}

func CreateService(Name string) {
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
			Name: Name + "-service",
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
	err2 := ioutil.WriteFile("Service.yaml", ServiceYaml, 0)

	if err2 != nil {

		log.Fatal(err2)
	}
}

func DeployKube(Name string) string {
	DeployToKubeCmd, err := exec.Command("kubectl", "apply", "-f", ".").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(DeployToKubeCmd)

	GetPredictUrlCmd, err := exec.Command("minikube", "service", "--url", Name+"-service").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(GetPredictUrlCmd)

	return string(GetPredictUrlCmd) + "/predictions"
}
