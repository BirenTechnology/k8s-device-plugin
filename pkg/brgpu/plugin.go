// Copyright 2024 Shanghai Biren Technology Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package brgpu

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BirenTechnology/go-brml/brml"
	"github.com/BirenTechnology/k8s-device-plugin/pkg/utils"

	log "github.com/sirupsen/logrus"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var (
	MountHostPath bool
)

const (
	allocatedDeviceEnv = "BR_PHY_CARDS"
)

func healthCheck() bool {
	return true
}

// int8Slice wraps an []int8 with more functions.
type int8Slice []int8

// String turns a nil terminated int8Slice into a string
func (s int8Slice) String() string {
	var b []byte
	for _, c := range s {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}

// uintPtr returns a *uint from a uint32
func uintPtr(c uint32) *uint {
	i := uint(c)
	return &i
}

type Plugin struct {
	PFDevices      PFDeviceInfoList
	BRGPUs         DevicesInfoList
	Runtime        string
	Heartbeat      chan bool
	resourceName   string
	MountAllDevice bool
	MountDriDevice bool
	MountHostPath  bool
	TopoGraph      *utils.Graph
}

func (p *Plugin) gpuExist(id string) (bool, error) {
	for _, v := range p.BRGPUs.AllCardIDs() {
		if id == v {
			return true, nil
		}
	}
	return false, nil
}

func (p *Plugin) Start() error {
	return nil
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	options := &pluginapi.DevicePluginOptions{
		GetPreferredAllocationAvailable: true,
	}
	log.Infof("Start Plugin With Options %v", options)
	return options, nil
}

func (p *Plugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *Plugin) GetPreferredAllocation(ctx context.Context, r *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	if p.TopoGraph == nil {
		return &pluginapi.PreferredAllocationResponse{}, nil
	}
	nodes := []*utils.Node{}
	for _, v := range r.ContainerRequests[0].AvailableDeviceIDs {
		nodes = append(nodes, &utils.Node{
			Name: v,
		})
	}

	newGraph := p.TopoGraph.SelectNodes(nodes)

	devices := Allocate(*newGraph, r.ContainerRequests[0].MustIncludeDeviceIDs, int(r.ContainerRequests[0].AllocationSize))

	return &pluginapi.PreferredAllocationResponse{
		ContainerResponses: []*pluginapi.ContainerPreferredAllocationResponse{
			{
				DeviceIDs: devices,
			},
		},
	}, nil
}

func (d *Plugin) GetNumaNode(idx int) (bool, int, error) {
	dev, err := brml.HandleByIndex(idx)
	if err != nil {
		log.Errorf("parse device id index %v fail %v", idx, err)
		return false, 0, err
	}
	pcie, err := brml.DevicePciInfo(dev)
	if err != nil {
		log.Errorf("get device index %v %v pcie info err %v", idx, d, err)
		return false, 0, err
	}

	// Discard leading zeros.
	busID := strings.ToLower(strings.TrimPrefix(int8Slice(pcie.BusId[:]).String(), "0000"))
	b, err := os.ReadFile(fmt.Sprintf("/sys/bus/pci/devices/%s/numa_node", busID))
	if err != nil {
		log.Errorf("read bus file id %v fail %v ", busID, err)
		return false, 0, nil
	}

	node, err := strconv.Atoi(string(bytes.TrimSpace(b)))
	if err != nil {
		return false, 0, fmt.Errorf("eror parsing value for NUMA node: %v", err)
	}

	if node < 0 {
		return false, 0, nil
	}

	return true, node, nil
}

func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs := []*pluginapi.Device{}
	if p.Runtime == string(RuntimeRunc) {
		devIDs := []string{}
		for _, v := range p.BRGPUs {
			for _, ins := range v.Instances {
				dev := &pluginapi.Device{
					ID:     ins.CardID,
					Health: pluginapi.Healthy,
				}

				hasNum, numa, err := p.GetNumaNode(v.PhysicalNum)
				if err != nil {
					log.Errorf("get numa node %v err %v", v.PhysicalNum, err)
				}

				if hasNum {
					log.Infof("dev %v topology numa %v", v.PhysicalNum, numa)
					dev.Topology = &pluginapi.TopologyInfo{
						Nodes: []*pluginapi.NUMANode{
							{
								ID: int64(numa),
							},
						},
					}
				}
				devs = append(devs, dev)
				devIDs = append(devIDs, ins.CardID)
			}

		}
		tg, err := Device2Graph(devIDs)
		if err != nil {
			log.Errorf("Generate gpu %v topo error %v", devIDs, err)
		}
		p.TopoGraph = tg
	}
	if p.Runtime == string(RuntimeKata) {
		for _, v := range p.PFDevices {
			for _, vf := range v.VFs {
				dev := &pluginapi.Device{
					ID:     vf.deviceEndpoint(),
					Health: pluginapi.Healthy,
				}
				devs = append(devs, dev)
			}
		}
	}

	s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

	select {}
}

func (p *Plugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}
	for _, req := range r.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{}
		if CdiFeature {
			for _, id := range req.DevicesIDs {
				response.CDIDevices = append(response.CDIDevices, &pluginapi.CDIDevice{
					Name: fmt.Sprintf("%s/%s=%s", vendor, p.getResourceByCardId(ContainerRuntime(p.Runtime), id), id),
				})
			}
			responses.ContainerResponses = append(responses.ContainerResponses, &response)
			continue
		}
		if p.MountHostPath {
			response.Mounts = append(response.Mounts, podMounts()...)
		}
		if p.Runtime == string(RuntimeRunc) {
			if p.MountDriDevice {
				driDevs, err := drmDevices()
				if err != nil {
					return nil, err
				}
				response.Devices = append(response.Devices, driDevs...)
			}

			for _, id := range req.DevicesIDs {
				exist, err := p.gpuExist(id)
				if err != nil {
					return nil, err
				}
				if !exist {
					log.Errorf("Invalid allocation request for %s: unknown device %s", p.resourceName, id)
					return nil, fmt.Errorf("invalid allocation request for %s: unknown device %s", p.resourceName, id)
				}

				devpath := fmt.Sprintf("/dev/biren/%s", id)
				dev := pluginapi.DeviceSpec{
					HostPath:      devpath,
					ContainerPath: devpath,
					Permissions:   "rw",
				}

				response.Devices = append(response.Devices, &dev)
				log.Infof("Allocate device %s successfully", id)
			}
		}
		if p.Runtime == string(RuntimeKata) {
			for _, id := range req.DevicesIDs {
				dev := pluginapi.DeviceSpec{
					HostPath:      id,
					ContainerPath: id,
					Permissions:   "rw",
				}
				response.Devices = append(response.Devices, &dev)
				log.Infof("Allocate device %s successfully", id)
			}
		}
		response.Envs = map[string]string{
			allocatedDeviceEnv: strings.Join(req.DevicesIDs, ","),
		}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}
	//log.Info(responses.ContainerResponses)
	return &responses, nil
}

func defaultMountPathFunc(h string) string {
	return strings.Replace(h, "/usr/", "/opt/birentech/", -1)
}

func podMounts() []*pluginapi.Mount {
	mounts := []*pluginapi.Mount{}
	brmlVersion, _ := brml.BRMLVersion()
	mountPaths := map[string]func(string) string{
		"/usr/lib/libbiren-ml.so":                              defaultMountPathFunc,
		"/usr/lib/libbiren-ml.so.1":                            defaultMountPathFunc,
		"/usr/bin/brsmi":                                       defaultMountPathFunc,
		fmt.Sprintf("/usr/lib/libbiren-ml.so.%s", brmlVersion): defaultMountPathFunc,
	}

	for h, c := range mountPaths {
		if _, err := os.Stat(defaultMountPathFunc(h)); err == nil {
			m := &pluginapi.Mount{
				HostPath:      h,
				ContainerPath: c(h),
				ReadOnly:      true,
			}
			mounts = append(mounts, m)
		}
	}
	return mounts
}

func drmDevices() ([]*pluginapi.DeviceSpec, error) {
	res := []*pluginapi.DeviceSpec{}
	path := "/dev/dri/renderD128"
	res = append(res, &pluginapi.DeviceSpec{
		ContainerPath: path,
		HostPath:      path,
		Permissions:   "rw",
	})
	return res, nil
}

func allDevices() ([]*pluginapi.DeviceSpec, error) {
	res := []*pluginapi.DeviceSpec{}
	c, err := brml.DeviceCount()
	if err != nil {
		return nil, err
	}
	for i := 0; i < c; i++ {
		path := fmt.Sprintf("/dev/biren/card_%d", i)
		res = append(res, &pluginapi.DeviceSpec{
			ContainerPath: path,
			HostPath:      path,
			Permissions:   "rw",
		})
	}
	return res, nil
}

func (p *Plugin) getResourceByCardId(runtime ContainerRuntime, id string) string {
	switch runtime {
	case RuntimeRunc:
		return p.BRGPUs.getResourceByCardId(id)
	case RuntimeKata:
		return p.PFDevices.getResourceByCardId(id)
	}
	return "gpu"
}
