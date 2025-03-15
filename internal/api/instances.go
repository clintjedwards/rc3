package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/luthermonson/go-proxmox"
)

func (api *APIContext) instancesRouter() RouteEntry {
	router := func(router chi.Router) {
		router.Get("/", api.getInstances)
		router.Post("/", api.createInstance)
		router.Delete("/{id}", api.deleteInstance)
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

func (it *InstanceType) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	switch InstanceType(str) {
	case InstanceTypeContainer, InstanceTypeVM:
		*it = InstanceType(str)
		return nil
	default:
		return fmt.Errorf("invalid InstanceType: %s", str)
	}
}

type InstanceSize string

const (
	InstanceSizeSmall  InstanceSize = "small"
	InstanceSizeMedium InstanceSize = "medium"
	InstanceSizeLarge  InstanceSize = "large"
)

func (api *APIContext) getContainerOptions(size InstanceSize) ([]proxmox.ContainerOption, error) {
	// Common settings
	options := []proxmox.ContainerOption{
		{Name: "arch", Value: "amd64"},
		{Name: "onboot", Value: 1}, // Start on boot
		{Name: "ostype", Value: "ubuntu"},
		{Name: "unprivileged", Value: true},
		{Name: "features", Value: "nesting=1"},
		{Name: "ostemplate", Value: api.ProxmoxConfig.OSTemplate},
		{Name: "net0", Value: "name=eth0,bridge=vmbr0,firewall=0,ip=dhcp"},
		{Name: "rootfs", Value: fmt.Sprintf("%s,size=60", api.ProxmoxConfig.InstanceStorage)}, // 60 GB of storage
	}

	switch size {
	case InstanceSizeSmall:
		options = append(options,
			proxmox.ContainerOption{Name: "cores", Value: 2},
			proxmox.ContainerOption{Name: "cpulimit", Value: 2},
			proxmox.ContainerOption{Name: "memory", Value: "2048"}, // 2 GB of memory
		)
		return options, nil
	case InstanceSizeMedium:
		options = append(options,
			proxmox.ContainerOption{Name: "cores", Value: 2},
			proxmox.ContainerOption{Name: "cpulimit", Value: 2},
			proxmox.ContainerOption{Name: "memory", Value: "4096"}, // 4 GB of memory
		)
		return options, nil
	case InstanceSizeLarge:
		options = append(options,
			proxmox.ContainerOption{Name: "cores", Value: 4},
			proxmox.ContainerOption{Name: "cpulimit", Value: 4},
			proxmox.ContainerOption{Name: "memory", Value: "8192"}, // 8 GB of memory
		)
		return options, nil
	default:
		return nil, fmt.Errorf("invalid instance size")
	}
}

// Add tags to a container option list.
func createTagsContainerOption(tags ...string) proxmox.ContainerOption {
	tagStr := ""

	for _, tag := range tags {
		tagStr += tag + ","
	}

	tagStr = strings.TrimSuffix(tagStr, ",")

	return proxmox.ContainerOption{Name: "tags", Value: tagStr}
}

func (is *InstanceSize) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	switch InstanceSize(str) {
	case InstanceSizeSmall, InstanceSizeMedium, InstanceSizeLarge:
		*is = InstanceSize(str)
		return nil
	default:
		return fmt.Errorf("invalid InstanceSize: %s", str)
	}
}

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

type CreateInstanceRequest struct {
	Size         InstanceSize `json:"size"`
	InstanceType InstanceType `json:"type"`
}

type CreateInstanceResponse struct{}

func (api *APIContext) createInstance(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var request CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	nodes, err := api.Client.Nodes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not query for nodes while attempting to get instances: %v", err))
		return
	}

	if len(nodes) == 0 {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("received no proxmox nodes while attempting to create instance: %v", err))
		return

	}

	// We default to the first node that we find since that will work for the single proxmox instance we have.
	// If we ever expand the proxmox cluster then we'll have to change this logic.
	targetNodeName := nodes[0].Node

	node, err := api.Client.Node(ctx, targetNodeName)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not get target node %s: %v", targetNodeName, err))
		return
	}

	// We need to query the cluster in order to figure out what the next VMID should be.
	cluster, err := api.Client.Cluster(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not get cluster: %v", err))
		return
	}

	// Proxmox gives us an endpoint we can hit to get the next sequential ID for the next instance.
	nextID, err := cluster.NextID(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not get next id for instance from cluster: %v", err))
		return
	}

	switch request.InstanceType {
	case InstanceTypeContainer:
		containerOptions, err := api.getContainerOptions(request.Size)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("could not get container settings: %v", err))
			return
		}

		createTagsContainerOption(
			fmt.Sprintf("size=%s", request.Size),
			fmt.Sprintf("recurser=%s", "TODO"),
		)

		_, err = node.NewContainer(ctx, nextID, containerOptions...)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("could not create new container: %v", err))
			return
		}

		writeResponse(w, http.StatusCreated, CreateInstanceResponse{})
		return
	case InstanceTypeVM:
		writeError(w, http.StatusNotImplemented, "VMs are not currently supported")
		return
	default:
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("invalid instance type: %v", err))
		return
	}
}

func (api *APIContext) deleteInstance(w http.ResponseWriter, r *http.Request) {
	// TODO(): We need to check permissions here. Does the current user own the instance?

	ctx := context.Background()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeError(w, http.StatusInternalServerError, "received empty identifier in path")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not ")
		return
	}

	nodes, err := api.Client.Nodes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not query for nodes while attempting to get instances: %v", err))
		return
	}

	if len(nodes) == 0 {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("received no proxmox nodes while attempting to create instance", err))
		return

	}

	// We default to the first node that we find since that will work for the single proxmox instance we have.
	// If we ever expand the proxmox cluster then we'll have to change this logic.
	targetNodeName := nodes[0].Node

	node, err := api.Client.Node(ctx, targetNodeName)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("could not get target node %s: %v", targetNodeName, err))
		return
	}

	// TODO(): WE need to go get the details of the instance here to see who owns it.

	_ = node
	_ = id
	// node.Container(ctx, strconv.Atoi(id))
}
