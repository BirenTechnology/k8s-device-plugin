// Copyright 2024 Shanghai Biren Technology Co., Ltd.
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package brgpu

import (
	"os"
	"sync"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type ContainerRuntime string

const (
	RuntimeKata ContainerRuntime = "kata"
	RuntimeRunc ContainerRuntime = "runc"
)

const (
	vendor = "birentech.com"
)

type Lister struct {
	ResUpdateChan    chan dpm.PluginNameList
	Heartbeat        chan bool
	MountAllDevice   bool
	MountDriDevice   bool
	DevicesInfoList  DevicesInfoList
	PFDeviceInfoList PFDeviceInfoList
	Runtime          string
	MountHostPath    bool
}

func (l *Lister) GetResourceNamespace() string {
	return vendor
}

func (l *Lister) NewPlugin(resourceLastName string) dpm.PluginInterface {
	return &Plugin{
		Runtime:        l.Runtime,
		PFDevices:      l.PFDeviceInfoList.FilterByName(resourceLastName),
		BRGPUs:         l.DevicesInfoList.FilterByName(resourceLastName),
		Heartbeat:      l.Heartbeat,
		MountAllDevice: l.MountAllDevice,
		MountDriDevice: l.MountDriDevice,
		MountHostPath:  l.MountHostPath,
	}
}
func (l *Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	for {
		select {
		case newResourcesList := <-l.ResUpdateChan: // New resources found
			pluginListCh <- newResourcesList
		case <-pluginListCh: // Stop message received
			// Stop resourceUpdateCh
			return
		}
	}
}

type MountPath struct {
	HostPath      string
	ContainerPath string
}

// GPUConfig stores the settings used to configure the GPUs on a node.
type GPUConfig struct {
	GPUPartitionSize string
}

type brGPUManager struct {
	devDirectory   string
	defaultDevices []string
	devices        map[string]pluginapi.Device
	grpcServer     *grpc.Server
	socket         string
	Stop           chan bool
	devicesMutex   sync.Mutex
	gpuConfig      GPUConfig
	Health         chan pluginapi.Device

	// 生成 cdi config
	generateCdiConfigFile func(runtime ContainerRuntime) error
}

func NewBrGPUManager(devDirectory string, gpuConfig GPUConfig) *brGPUManager {
	return &brGPUManager{
		devDirectory:          devDirectory,
		devices:               make(map[string]pluginapi.Device),
		Stop:                  make(chan bool),
		gpuConfig:             gpuConfig,
		Health:                make(chan pluginapi.Device),
		generateCdiConfigFile: generateConfigCdiFile,
	}
}

func (bgm *brGPUManager) ListDevices() map[string]pluginapi.Device {
	if bgm.gpuConfig.GPUPartitionSize == "" {
		return bgm.devices
	}
	return map[string]pluginapi.Device{}
}

func (bgm *brGPUManager) Serve(pulse int, mountAllDev bool, mountDriDevice bool, runtime string) {
	log.Info("Container runtime: ", runtime)
	go func() {
		<-bgm.Stop
		close(bgm.Stop)
		os.Exit(1)
	}()

	switch runtime {
	case string(RuntimeKata):
		bgm.kataManager()
	case string(RuntimeRunc):
		bgm.runcManager(pulse, mountAllDev, mountDriDevice)
	}
	log.Error("Can't find any manager for runtime ", runtime)
}
