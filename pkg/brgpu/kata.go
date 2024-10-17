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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	log "github.com/sirupsen/logrus"
)

const (
	BirenVendorID = "1ee0"
	basePath      = "/sys/bus/pci/devices"
)

type VFDeviceInfo struct {
	DeviceID     string
	IOMMUGroup   string
	Addr         string
	ResourceName string
}

func (v VFDeviceInfo) deviceEndpoint() string {
	return "/dev/vfio/" + v.IOMMUGroup
}

type PFDeviceInfo struct {
	Addr    string
	VFs     []VFDeviceInfo
	VFCount int
}

type PFDeviceInfoList []PFDeviceInfo

func (p PFDeviceInfoList) ResourceNames() []string {
	res := []string{}
	names := map[string]int{}
	for _, v := range p {
		names[v.VFs[0].ResourceName] += 1
	}

	for k, _ := range names {
		res = append(res, k)
	}
	return res

}

func (p PFDeviceInfoList) FilterByName(resourceName string) PFDeviceInfoList {
	res := PFDeviceInfoList{}
	for _, v := range p {
		if v.VFs[0].ResourceName == resourceName {
			res = append(res, v)
		}
	}
	return res
}

func (p PFDeviceInfoList) Contain(pciaddr string) bool {
	for _, v := range p {
		for _, val := range v.VFs {
			if val.Addr == pciaddr {
				return true
			}
		}
	}
	return false
}

func (p PFDeviceInfoList) getResourceByCardId(cardId string) string {
	for _, vs := range p {
		for _, v := range vs.VFs {
			if v.DeviceID == cardId {
				return v.ResourceName
			}
		}
	}
	return "gpu"
}

func (bgm *brGPUManager) kataManager() {
	info, err := vfDeviceDiscover()
	if err != nil {
		log.Errorf("kata device discover failed %v", err)
		bgm.Stop <- true
	}
	l := Lister{
		ResUpdateChan:    make(chan dpm.PluginNameList),
		Heartbeat:        make(chan bool),
		PFDeviceInfoList: info,
		Runtime:          string(RuntimeKata),
	}
	manager := dpm.NewManager(&l)
	go func() {
		l.ResUpdateChan <- info.ResourceNames()
	}()

	err = bgm.generateCdiConfigFile(RuntimeKata)
	if err != nil {
		log.Errorf("kata generate cdi config failed %v", err)
		bgm.Stop <- true
	}

	manager.Run()
}

func readIDFromFile(basePath string, deviceAddress string, property string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(basePath, deviceAddress, property))
	if err != nil {
		log.Errorf("Could not read %s for device %s: %s", property, deviceAddress, err)
		return "", err
	}
	id := strings.Trim(string(data[2:]), "\n")
	return id, nil
}

func readLink(basePath string, deviceAddress string, link string) (string, error) {
	path, err := os.Readlink(filepath.Join(basePath, deviceAddress, link))
	if err != nil {
		log.Errorf("Could not read link %s for device %s: %s", link, deviceAddress, err)
		return "", err
	}
	_, file := filepath.Split(path)
	return file, nil
}

func readNumFromFile(basePath string, deviceAddress string, property string) (int, error) {
	data, err := ioutil.ReadFile(filepath.Join(basePath, deviceAddress, property))
	if err != nil {
		log.Errorf("Could not read %s for device %s: %s", property, deviceAddress, err)
		return 0, err
	}

	s := strings.Trim(string(data), "\n")
	return strconv.Atoi(s)
}

func vfDeviceDiscover() (PFDeviceInfoList, error) {
	pdl := PFDeviceInfoList{}
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		vendorID, err := readIDFromFile(basePath, info.Name(), "vendor")
		if err != nil {
			log.Error("Could not get vendor ID for device ", info.Name())
			return nil
		}

		if vendorID == BirenVendorID {
			log.Infof("Birentech device %s", info.Name())
			driver, err := readLink(basePath, info.Name(), "driver")
			if err != nil {
				log.Error("Could not get driver for device, Skip it.", info.Name())
				return nil
			}
			if driver == "BEV_HYPER_DRIVER" {
				_, err := readLink(basePath, info.Name(), "physfn")
				if os.IsNotExist(err) {
					vfNum, err := readNumFromFile(basePath, info.Name(), "sriov_numvfs")
					if err != nil {
						return err
					}
					vfs := []VFDeviceInfo{}
					for i := 0; i < vfNum; i++ {
						vfLink := fmt.Sprintf("virtfn%d", i)
						vfAddr, err := readLink(basePath, info.Name(), vfLink)
						if err != nil {
							return err
						}

						iommuGroup, err := readLink(basePath, vfAddr, "iommu_group")
						if err != nil {
							return err
						}

						deviceID, err := readIDFromFile(basePath, vfAddr, "device")
						if err != nil {
							return err
						}

						vfs = append(vfs, VFDeviceInfo{
							DeviceID:   deviceID,
							IOMMUGroup: iommuGroup,
							Addr:       vfAddr,
							ResourceName: func() string {
								if vfNum == 1 {
									return "gpu"
								}
								return fmt.Sprintf("1-%d-gpu", vfNum)
							}(),
						})
					}
					pdl = append(pdl, PFDeviceInfo{
						Addr:    info.Name(),
						VFCount: vfNum,
						VFs:     vfs,
					})
				}
			}
			if driver == "vfio-pci" {
				iommuGroup, err := readLink(basePath, info.Name(), "iommu_group")
				if err != nil {
					return err
				}
				deviceID, err := readIDFromFile(basePath, info.Name(), "device")
				if err != nil {
					return err
				}
				if !pdl.Contain(info.Name()) {
					pdl = append(pdl, PFDeviceInfo{
						Addr:    info.Name(),
						VFCount: 1,
						VFs: []VFDeviceInfo{
							{
								DeviceID:     deviceID,
								IOMMUGroup:   iommuGroup,
								Addr:         info.Name(),
								ResourceName: "gpu",
							},
						},
					})
				}
			}
		}

		return nil
	})
	return pdl, err
}
