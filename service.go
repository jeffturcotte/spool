package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type Package struct {
	Path           string
	Package        string
	Services       []*Service
	ServicesByName map[string]*Service
}

type Service struct {
	Package         *Package
	Name            string
	Build           string
	Image           string
	Volumes         []string
	PublishAllPorts bool
	Link            []string
}

type ServiceContainer struct {
	Package string
	Env     string
	Service string
	UID     string
}

func NewPackage(configPath string) (p Package, err error) {
	// read the spool config file
	blob, err := ioutil.ReadFile(configPath)
	if err != nil {
		return
	}

	err = json.Unmarshal(blob, &p)
	if err != nil {
		return
	}

	if p.Path == "" {
		p.Path = path.Dir(configPath)

	}

	// this initialization is silly
	p.ServicesByName = make(map[string]*Service)

	for _, service := range p.Services {
		p.ServicesByName[service.Name] = service
		service.Package = &p
	}

	return
}

func (service *Service) BuildImage(client *docker.Client, out io.Writer, env string, uid string) (string, error) {
	tarPath := service.Package.Path + "/" + service.Build

	if strings.Index(service.Build, "/") == 0 {
		tarPath = service.Build
	}

	// generate tar command
	cmd := exec.Command("tar", "-C", tarPath, "-c", ".")
	tar, err := cmd.StdoutPipe()

	if err != nil {
		return "", err
	}

	// start tar process/stream
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	// todo: this needs to change location
	imageName := service.getImageName(env, uid)

	// build the image
	err = client.BuildImage(docker.BuildImageOptions{
		Name:                imageName,
		ForceRmTmpContainer: true,
		InputStream:         tar,
		OutputStream:        out,
	})

	if err != nil {
		return "", err
	}

	// wait for the tar stream to finish
	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	return imageName, nil
}

func (service *Service) PullImage(client *docker.Client, out io.Writer) (string, error) {
	image := strings.SplitN(service.Image, ":", 2)

	err := client.PullImage(
		docker.PullImageOptions{
			Repository:   image[0],
			Tag:          image[1],
			OutputStream: out,
		},
		// keep auth empty for now assume public docker hub if you are using an image
		docker.AuthConfiguration{},
	)

	return service.Image, err
}

func (service *Service) RunContainer(client *docker.Client, image string, env string, uid string) (*docker.Container, error) {
	current := []*docker.Container{}
	volumes := []string{}
	links := []string{}

	currentServiceContainers, err := service.ListContainers(client, env)
	current = append(current, currentServiceContainers...)

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: service.getContainerName(env, uid),
		Config: &docker.Config{
			Image: image,
		},
	})

	if err != nil {
		return nil, err
	}

	// handle volumes
	if len(current) > 0 {
		volumes = append(volumes, current[0].Name)
	}

	// handle links
	for _, name := range service.Link {
		serviceContainers, err := service.Package.ServicesByName[name].ListContainers(client, env)
		if err != nil {
			return nil, err
		}
		if len(serviceContainers) > 0 {
			links = append(links, fmt.Sprintf("%s:%s", serviceContainers[0].Name, name))
		}
	}

	hostConfig := docker.HostConfig{
		VolumesFrom: volumes,
		Links:       links,
	}

	err = client.StartContainer(container.ID, &hostConfig)

	if err != nil {
		return nil, err
	}

	container, err = client.InspectContainer(container.ID)

	return container, err
}

func (p *Package) ListContainers(client *docker.Client, env string) ([]*docker.Container, error) {
	var containers []*docker.Container

	apiContainers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	for _, apiContainer := range apiContainers {
		container, _ := client.InspectContainer(apiContainer.ID)
		serviceContainer := containerToServiceInfo(container)
		if serviceContainer.Package == p.Package && serviceContainer.Env == env {
			containers = append(containers, container)
		}

	}

	return containers, err
}

func (service *Service) ListContainers(client *docker.Client, env string) ([]*docker.Container, error) {
	var containers []*docker.Container

	apiContainers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	for _, apiContainer := range apiContainers {
		container, _ := client.InspectContainer(apiContainer.ID)
		serviceInfo := containerToServiceInfo(container)

		if serviceInfo.Package == service.Package.Package &&
			serviceInfo.Service == service.Name &&
			serviceInfo.Env == env {
			containers = append(containers, container)
		}

	}

	return containers, err
}

func (s *Service) getImageName(env string, uid string) string {
	return fmt.Sprintf("%s/%s/%s:%s", s.Package.Package, env, s.Name, uid)
}

func (s *Service) getContainerName(env string, uid string) string {
	return fmt.Sprintf("%s-%s-%s-%s", s.Package.Package, env, s.Name, uid)
}

func containerToServiceInfo(container *docker.Container) ServiceContainer {
	pattern := "^/(.*)-(.*)-(.*)-(.*)$"
	s := ServiceContainer{}

	// todo: there must be a better way to do this

	match, _ := regexp.MatchString(pattern, container.Name)

	if match {
		r, _ := regexp.Compile(pattern)
		result := r.FindAllStringSubmatch(container.Name, -1)

		s.Package = result[0][1]
		s.Env = result[0][2]
		s.Service = result[0][3]
		s.UID = result[0][4]
	}

	return s
}
