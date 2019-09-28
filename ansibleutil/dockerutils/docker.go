/*
Copyright (c) 2019 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Main used to spawn a bunch of containers for dev purpose in the local environment.

The main idea i deploying vagrant or full blow VM to validate ansible playbooks or terraform scenario.
A developer can instantiate locally all containers

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package dockerutils

import (
	"context"
	"fmt"

	"io"
	"log"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Function create a docker container docker in local environment.
// Docker must be installed in a system
func CreateNewContainer(image string, localPort int) (string, error) {

	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	// pull image
	out, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed pull image %s %s", image, err)
	}

	// show progress
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return "", fmt.Errorf("failed io image %s %s", image, err)
	}

	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: strconv.Itoa(localPort),
	}
	containerPort, err := nat.NewPort("tcp", "22")
	if err != nil {
		return "", fmt.Errorf("failed create container port %s", err)
	}

	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	// create containers with ssh port exposed
	cont, err := cli.ContainerCreate(context.Background(),
		&container.Config{
			Image:        image,
			ExposedPorts: nat.PortSet{"22/tcp": struct{}{}},
		},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, "")

	if err != nil {
		return "", fmt.Errorf("failed create container error: %s", err)
	}

	// start container
	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed start container error: %s", err)
	}

	log.Printf("Container %s is started\n", cont.ID)
	return cont.ID, nil
}

// Function stops all containers
func StopContainers() error {

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		fmt.Print("Stopping container ", container.ID[:10], "... ")
		if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
			panic(err)
		}
		fmt.Println("Success")
	}

	return nil
}
