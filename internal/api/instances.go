package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *APIContext) instancesRouter() RouteEntry {
	router := func(router chi.Router) {
		router.Get("/", api.getInstances)
	}

	return RouteEntry{
		Pattern: "/instances",
		Router:  router,
	}
}

type InstanceType string

const (
	InstanceTypeContainer InstanceType = "container"
	InstanceTypeVM        InstanceType = "vm"
)

type InstanceSize string

const (
	InstanceSizeSmall  InstanceSize = "small"
	InstanceSizeMedium InstanceSize = "medium"
	InstanceSizeLarge  InstanceSize = "large"
)

type Instance struct {
	ID       uint64       `json:"id"`
	Kind     InstanceType `json:"kind"`
	Size     InstanceSize `json:"size"`
	Name     string       `json:"name"`
	Node     string       `json:"node"`
	Status   string       `json:"status"`
	Uptime   uint64       `json:"uptime"`
	Recurser string       `json:"recurser"`
}

type GetInstancesResponse struct {
	Instances []Instance `json:"instances"`
}

func (api *APIContext) getInstances(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	returnedInstances := []Instance{}

	nodes, err := api.Client.Nodes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not query for nodes while attempting to get instances: %v", err))
		return
	}

	for _, nodeStatus := range nodes {
		nodeName := nodeStatus.Node

		node, err := api.Client.Node(ctx, nodeName)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("could not query for specific node while attempting to get instances: %v", err))
			return
		}

		containers, err := node.Containers(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("could not query for containers while attempting to get instances: %v", err))
			return
		}

		for _, container := range containers {
			newInstance := Instance{
				ID:       uint64(container.VMID),
				Kind:     InstanceTypeContainer,
				Size:     "", // We need to store a tag somewhere to repeat size.
				Name:     container.Name,
				Node:     container.Node,
				Status:   container.Status,
				Uptime:   container.Uptime,
				Recurser: "", // We need to store a tag somewhere to identify recurser.
			}

			returnedInstances = append(returnedInstances, newInstance)
		}

		vms, err := node.VirtualMachines(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("could not query for vms while attempting to get instances: %v", err))
			return
		}

		for _, vm := range vms {
			newInstance := Instance{
				ID:       uint64(vm.VMID),
				Kind:     InstanceTypeVM,
				Size:     "", // We need to store a tag somewhere to repeat size.
				Name:     vm.Name,
				Node:     vm.Node,
				Status:   vm.Status,
				Uptime:   vm.Uptime,
				Recurser: "", // We need to store a tag somewhere to identify recurser.
			}

			returnedInstances = append(returnedInstances, newInstance)
		}

	}

	writeResponse(w, http.StatusOK, GetInstancesResponse{
		Instances: returnedInstances,
	})
}
