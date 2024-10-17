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
	"math"

	log "github.com/sirupsen/logrus"

	"github.com/BirenTechnology/go-brml/brml"
	"github.com/BirenTechnology/k8s-device-plugin/pkg/utils"
)

func Device2Graph(devices []string) (*utils.Graph, error) {
	res := &utils.Graph{}
	for _, v := range devices {
		diIndex, err := cardID2Index(v)
		if err != nil {
			return nil, err
		}

		cNode := &utils.Node{
			Name: cardIDFormat(diIndex),
		}
		res.AddNode(cNode)
		di, err := brml.HandleByNodeID(diIndex)
		if err != nil {
			return nil, err
		}
		for _, v2 := range devices {
			djIndex, err := cardID2Index(v2)
			if err != nil {
				return nil, err
			}
			dj, err := brml.HandleByNodeID(djIndex)
			if err != nil {
				return nil, err
			}
			ps, err := brml.P2PStatusV2(di, dj)
			if err != nil {
				return nil, err
			}

			res.AddEdge(cNode, &utils.Node{
				Name: cardIDFormat(djIndex),
			}, scoreEnlarge(int(ps.Type)))

		}
	}
	log.Infof("create topo for devices: %v; result: \n%s", devices, res.String())
	return res, nil
}

func scoreEnlarge(num int) int {
	return int(math.Pow(float64(num)+1, 2))
}

func Allocate(g utils.Graph, mustIncludeNodes []string, size int) []string {
	_, names := g.MaxValCount(size)
	if len(mustIncludeNodes) != 0 {
		log.Error("must include nodes not nil")
	}
	log.Infof("Select devices: %v from topo: %v", names, g.String())
	return names
}
