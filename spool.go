package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

var startGreen = "\033[0;32m"
var startYello = "\033[0;33m"
var startCyan = "\033[0;36m"
var resetText = "\033[0m"

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func green(s string) string {
	return "\033[0;32m" + s + resetText
}

func yellow(s string) string {
	return "\033[1;33m" + s + resetText
}

func cyan(s string) string {
	return startCyan + s + resetText
}

func getEnv(name, d string) string {
	env := os.Getenv(name)

	if env != "" {
		return env
	}

	return d
}

var cwd string
var configFile string
var dockerHost string
var verbose bool

func main() {
	cwd, _ = os.Getwd()
	host := getEnv("DOCKER_HOST", "unix:///var/run/docker.sock")

	var rootCmd = &cobra.Command{
		Use:   "spool",
		Short: "A docker container deployment tool",
	}

	var cmdUp = &cobra.Command{
		Use:   "up [env]",
		Short: "Builds and starts an environment",
		Run:   up,
	}

	var cmdStop = &cobra.Command{
		Use:   "stop [env]",
		Short: "Stops an environment",
		Run:   stop,
	}

	var cmdDestroy = &cobra.Command{
		Use:   "destroy [env]",
		Short: "Destroys an environment",
		Run:   destroy,
	}

	var cmdInspect = &cobra.Command{
		Use:   "inspect [env]",
		Short: "Inspects an environment",
		Run:   inspect,
	}

	rootCmd.AddCommand(cmdUp, cmdStop, cmdDestroy, cmdInspect)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", cwd+"/spool.json",
		"The full path of a config file")

	rootCmd.PersistentFlags().StringVar(&dockerHost, "host", host,
		"The address of the docker daemon")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Print verbose output")

	rootCmd.Execute()
}

func up(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Not enough arguments")
	}

	env := args[0]
	writer := ioutil.Discard

	if verbose {
		writer = os.Stdout
	}

	p, err := NewPackage(configFile)
	assert(err)
	client, err := docker.NewClient(dockerHost)
	assert(err)

	// generate uid
	// noob todo: there must be a cleaner way to do this
	t := []byte(time.Now().String())
	md5Bytes := md5.Sum(t)
	uid := fmt.Sprintf("%x", md5Bytes)[0:6]

	// define vars
	var image string
	var currentContainers []*docker.Container
	var startedContainers []*docker.Container

	currentContainers, err = p.ListContainers(client, env)
	assert(err)

	fmt.Println(yellow("Deploying " + env))

	// pull or build image
	for _, service := range p.Services {
		assert(err)

		if service.Build != "" {
			fmt.Println("Building image for " + green(service.Name))
			fmt.Print(startCyan)
			image, err = service.BuildImage(client, writer, env, uid)
			fmt.Print(resetText)
			assert(err)
		} else if service.Image != "" {
			fmt.Println("Pulling image for " + green(service.Name))
			fmt.Print(startCyan)
			image, err = service.PullImage(client, writer)
			fmt.Print(resetText)
			assert(err)
		}

		fmt.Println("Starting " + green(service.Name))
		fmt.Print(startCyan)
		container, err := service.RunContainer(client, image, env, uid)
		fmt.Print(resetText)

		startedContainers = append(startedContainers, container)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Cleaning up")

	for _, container := range currentContainers {
		client.StopContainer(container.ID, 3)
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
	}

	fmt.Println(yellow("Done!"))
}

func stop(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Not enough arguments")
	}

	env := args[0]

	p, err := NewPackage(configFile)
	assert(err)

	client, err := docker.NewClient(dockerHost)
	assert(err)

	fmt.Println(yellow("Stopping " + env))

	containers, err := p.ListContainers(client, env)
	assert(err)

	for _, container := range containers {
		client.StopContainer(container.ID, 3)
	}

	fmt.Println(yellow("Done!"))
}

func inspect(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Not enough arguments")
	}

	env := args[0]

	p, err := NewPackage(configFile)
	assert(err)

	client, err := docker.NewClient(dockerHost)
	assert(err)

	containers, err := p.ListContainers(client, env)
	assert(err)

	for _, container := range containers {
		c, err := client.InspectContainer(container.ID)
		assert(err)

		network := c.NetworkSettings
		ipAddress := network.IPAddress
		ports := network.Ports

		for port, _ := range ports {
			line := fmt.Sprintf(
				"%s %s %s",
				c.Name,
				ipAddress,
				port,
			)
			fmt.Println(line)
		}
	}
}

func destroy(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Not enough arguments")
	}

	env := args[0]

	p, err := NewPackage(configFile)
	assert(err)

	client, err := docker.NewClient(dockerHost)
	assert(err)

	fmt.Println(yellow("Destroying " + env))

	containers, err := p.ListContainers(client, env)
	assert(err)

	for _, container := range containers {
		client.StopContainer(container.ID, 3)
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
	}

	fmt.Println(yellow("Done!"))
}
