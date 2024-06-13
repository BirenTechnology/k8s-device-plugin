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
//
package brgpu

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BirenTechnology/go-brml/brml"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	log "github.com/sirupsen/logrus"
)

type Instance struct {
	UUID         string
	Memory       int
	ResourceName string
	CardID       string
}

type DevicesInfo struct {
	PhysicalNum int
	Instances   []Instance
	SVICount    int
}

type DevicesInfoList []DevicesInfo

func (d DevicesInfoList) ResourceNames() []string {
	res := []string{}
	names := map[string]int{}
	for _, v := range d {
		names[v.Instances[0].ResourceName] += 1
	}

	for k, _ := range names {
		res = append(res, k)
	}
	return res
}

func (d DevicesInfoList) FilterByName(resourceName string) DevicesInfoList {
	res := DevicesInfoList{}
	for _, v := range d {
		if v.Instances[0].ResourceName == resourceName {
			res = append(res, v)
		}
	}
	return res
}

func (d DevicesInfoList) AllCardIDs() []string {
	res := []string{}
	for _, v := range d {
		for _, i := range v.Instances {
			res = append(res, i.CardID)
		}
	}
	return res
}

func (d DevicesInfoList) getResourceByCardId(cardId string) string {
	for _, vs := range d {
		for _, v := range vs.Instances {
			if v.CardID == cardId {
				return v.ResourceName
			}
		}
	}
	return "gpu"
}

func (bgm *brGPUManager) runcManager(pulse int, mountAllDev bool, mountDriDevice bool) {
	err := brml.Init()
	if err != nil {
		log.Error(err)
		bgm.Stop <- true
	}
	defer brml.Shutdown()

	info, err := deviceDiscover()
	if err != nil {
		log.Error(err)
		bgm.Stop <- true
	}
	l := Lister{
		ResUpdateChan:   make(chan dpm.PluginNameList),
		Heartbeat:       make(chan bool),
		MountAllDevice:  mountAllDev,
		MountDriDevice:  mountDriDevice,
		DevicesInfoList: info,
		Runtime:         string(RuntimeRunc),
		MountHostPath:   MountHostPath,
	}

	manager := dpm.NewManager(&l)
	if pulse > 0 {
		go func() {
			for {
				time.Sleep(time.Second * time.Duration(pulse))
				_, err = brml.DeviceCount()
				if err != nil {
					log.Errorf("Can't find device from host")
					bgm.Stop <- true
				}
				l.Heartbeat <- true
			}
		}()
	}

	go func() {
		var path = "/sys/class/biren"
		if _, err := os.Stat(path); err == nil {
			l.ResUpdateChan <- info.ResourceNames()
		}
	}()

	err = bgm.generateCdiConfigFile(RuntimeRunc)
	if err != nil {
		log.Error(err)
		bgm.Stop <- true
	}

	manager.Run()
}

func deviceDiscover() (DevicesInfoList, error) {
	dis := DevicesInfoList{}
	physicalNum, err := brml.DeviceCount()
	if err != nil {
		return nil, err
	}

	for i := 0; i < physicalNum; i++ {
		device, err := brml.HandleByNodeID(i)
		if err != nil {
			return nil, err
		}
		sviCount, err := brml.GetSviMode(device)
		if err != nil {
			return nil, err
		}

		phyUUID, err := brml.DeviceUUID(device)

		if err != nil {
			return nil, err
		}

		phyUUID = strings.TrimSpace(phyUUID)

		switch sviCount {
		case 0, 1:
			memInfo, err := brml.MemoryInfo(device)
			if err != nil {
				return nil, err
			}

			id, err := brml.GetGPUNodeIds(device)
			if err != nil {
				return nil, err
			}

			dis = append(dis, DevicesInfo{
				PhysicalNum: i,
				Instances: []Instance{{
					UUID:         phyUUID,
					Memory:       int(memInfo.Total),
					ResourceName: "gpu",
					CardID:       cardIDFormat(id),
				}},
				SVICount: 1,
			})
		case 2, 4:
			di := DevicesInfo{
				PhysicalNum: i,
				Instances:   []Instance{},
				SVICount:    sviCount,
			}
			for j := 0; j < sviCount; j++ {
				ins, err := brml.GetGPUInstanceByID(device, uint32(j))
				if err != nil {
					return nil, err
				}

				mem, err := brml.MemoryInfo(ins)
				if err != nil {
					return nil, err
				}

				id, err := brml.GetGPUNodeIds(ins)
				if err != nil {
					return nil, err
				}

				di.Instances = append(di.Instances, Instance{
					UUID:         fmt.Sprintf("%s-instance-%d", phyUUID, j),
					Memory:       int(mem.Total),
					ResourceName: fmt.Sprintf("1-%d-gpu", sviCount),
					CardID:       cardIDFormat(id),
				})
			}
			dis = append(dis, di)
		}
	}
	return dis, nil
}

func cardIDFormat(i int) string {
	return fmt.Sprintf("card_%d", i)
}

func cardID2Index(s string) (int, error) {
	if !strings.HasPrefix(s, "card_") {
		return 0, errors.New(fmt.Sprintf("card handler %s is not valid format", s))
	}
	s = strings.ReplaceAll(s, "card_", "")
	return strconv.Atoi(s)
}
