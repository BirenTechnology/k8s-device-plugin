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
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/BirenTechnology/go-brml/brml"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

const (
	deviceBasePath = "/dev/biren"
	cdiConfigPath  = "/etc/cdi/biren.yaml"
	// 最新版 0.6.0 containerd不兼容
	cdiVersion = "0.5.0"
)

var (
	CdiFeature         bool
	OverwriteCdiConfig bool
)

func cdiSPec(runtime ContainerRuntime) ([]*cdi.Spec, error) {
	switch runtime {
	case RuntimeRunc:
		return runcCDI()
	case RuntimeKata:
		return kataCDI()
	}

	return nil, nil
}

func runcCDI() ([]*cdi.Spec, error) {
	info, err := deviceDiscover()
	if err != nil {
		log.Errorf("deviceDiscover error: %v", err)
		return nil, err
	}

	specs := make([]*cdi.Spec, 0)

	resourceInstances := make(map[string][]Instance)

	for _, v := range info {
		for _, ins := range v.Instances {
			if _, ok := resourceInstances[ins.ResourceName]; !ok {
				resourceInstances[ins.ResourceName] = []Instance{}
			}
			resourceInstances[ins.ResourceName] = append(resourceInstances[ins.ResourceName], ins)
		}
	}

	for k, vs := range resourceInstances {
		spec := genSpec(k, MountHostPath)
		for _, v := range vs {
			spec.Devices = append(spec.Devices, cdi.Device{
				Name:        v.CardID,
				Annotations: map[string]string{},
				ContainerEdits: cdi.ContainerEdits{
					Env: []string{},
					DeviceNodes: []*cdi.DeviceNode{
						{
							Path:        path.Join(deviceBasePath, v.CardID),
							HostPath:    path.Join(deviceBasePath, v.CardID),
							Type:        "c",
							Permissions: "rw",
						},
					},
					Hooks:  []*cdi.Hook{},
					Mounts: []*cdi.Mount{},
				},
			})
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

func kataCDI() ([]*cdi.Spec, error) {
	info, err := vfDeviceDiscover()
	if err != nil {
		log.Errorf("vfDeviceDiscover error: %v", err)
		return nil, err
	}

	specs := make([]*cdi.Spec, 0)
	resourceVFDeviceInfos := make(map[string][]VFDeviceInfo)
	for _, v := range info {
		for _, vf := range v.VFs {
			if _, ok := resourceVFDeviceInfos[vf.ResourceName]; !ok {
				resourceVFDeviceInfos[vf.ResourceName] = []VFDeviceInfo{}
			}
			resourceVFDeviceInfos[vf.ResourceName] = append(resourceVFDeviceInfos[vf.ResourceName], vf)
		}
	}

	for k, vs := range resourceVFDeviceInfos {
		spec := genSpec(k, MountHostPath)
		for _, v := range vs {
			spec.Devices = append(spec.Devices, cdi.Device{
				Name:        v.deviceEndpoint(),
				Annotations: map[string]string{},
				ContainerEdits: cdi.ContainerEdits{
					Env: []string{},
					DeviceNodes: []*cdi.DeviceNode{{
						Path:        path.Join(deviceBasePath, v.DeviceID),
						HostPath:    path.Join(deviceBasePath, v.DeviceID),
						Type:        "char",
						Permissions: "c",
					}},
					Hooks:  []*cdi.Hook{},
					Mounts: []*cdi.Mount{},
				},
			})
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func genSpec(resource string, mountHostPath bool) *cdi.Spec {
	spec := &cdi.Spec{
		Version:     cdiVersion,
		Kind:        fmt.Sprintf("%s/%s", vendor, resource),
		Annotations: map[string]string{},
		Devices:     []cdi.Device{},
		ContainerEdits: cdi.ContainerEdits{
			Env:         []string{},
			DeviceNodes: []*cdi.DeviceNode{},
			Hooks:       []*cdi.Hook{},
			Mounts:      []*cdi.Mount{},
		},
	}
	if mountHostPath {
		cdiMounts := []*cdi.Mount{}
		brmlVersion, _ := brml.BRMLVersion()
		mountPaths := map[string]func(string) string{
			"/usr/lib/libbiren-ml.so":                              defaultMountPathFunc,
			"/usr/lib/libbiren-ml.so.1":                            defaultMountPathFunc,
			"/usr/bin/brsmi":                                       defaultMountPathFunc,
			fmt.Sprintf("/usr/lib/libbiren-ml.so.%s", brmlVersion): defaultMountPathFunc,
		}

		for h, c := range mountPaths {
			if _, err := os.Stat(defaultMountPathFunc(h)); err == nil {
				m := &cdi.Mount{
					HostPath:      h,
					ContainerPath: c(h),
					Options:       []string{"ro", "nosuid", "nodev", "bind"},
				}
				cdiMounts = append(cdiMounts, m)
			}
		}
		spec.ContainerEdits.Mounts = append(spec.ContainerEdits.Mounts, cdiMounts...)
	}
	return spec
}

func generateConfigCdiFile(runtime ContainerRuntime) error {
	if !CdiFeature {
		log.Info("cdi feature isn't open")
		return nil
	}
	exists, err := PathExists(cdiConfigPath)
	if err != nil {
		log.Error(err)
		return err
	}
	// 如果文件存在并且不需要覆盖写 直接返回
	if exists && !OverwriteCdiConfig {
		log.Infof("file already exists and no need to rewrite")
		return nil
	}

	specs, err := cdiSPec(runtime)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(cdiConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Errorf("open file failed:%v \n", err)
		return err
	}
	defer file.Close()

	for _, spec := range specs {
		bs, err := json.Marshal(spec)
		if err != nil {
			return err
		}

		bs, err = yaml.JSONToYAML(bs)
		if err != nil {
			return err
		}

		// 写入分隔符
		if _, err := file.WriteString("---\n"); err != nil {
			log.Errorf("Error writing separator: %v \n", err)
			return err
		}

		_, err = file.Write(bs)
		if err != nil {
			return err
		}
	}

	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
