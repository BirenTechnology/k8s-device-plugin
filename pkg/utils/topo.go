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
package utils

import (
	"fmt"
	"sort"
	"strings"
)

type Node struct {
	Name string
}

type nodeWithVal struct {
	val  int
	node *Node
}

type Graph struct {
	nodes []*Node
	edges map[string][]nodeWithVal
}

func (g *Graph) AddEdge(u, v *Node, val int) {
	if g.edges == nil {
		g.edges = make(map[string][]nodeWithVal)
	}
	if g.edgeExist(u, v) {
		return
	}
	g.edges[u.Name] = append(g.edges[u.Name], nodeWithVal{
		val:  val,
		node: v,
	})
	g.edges[v.Name] = append(g.edges[v.Name], nodeWithVal{
		val:  val,
		node: u,
	})
}

func (g *Graph) AddNode(n *Node) {
	g.nodes = append(g.nodes, n)
}

func (g *Graph) DeleteNode(n *Node) *Graph {
	newGraph := &Graph{
		nodes: []*Node{},
		edges: map[string][]nodeWithVal{},
	}
	for _, v := range g.nodes {
		if v.Name != n.Name {
			newGraph.nodes = append(newGraph.nodes, v)
		}
	}
	for k, nodes := range g.edges {
		newVals := []nodeWithVal{}
		for _, v := range nodes {
			if v.node.Name != n.Name {
				newVals = append(newVals, v)
			}
		}
		newGraph.edges[k] = newVals
	}
	return newGraph
}

func (g *Graph) DeleteNodes(nodes []*Node) *Graph {
	tg := g
	for _, n := range nodes {
		tg = tg.DeleteNode(n)
	}
	return tg
}

func (g *Graph) SelectNodes(nodes []*Node) *Graph {
	nodeNames := []string{}
	for _, v := range nodes {
		nodeNames = append(nodeNames, v.Name)
	}
	delNodes := []*Node{}
	for _, v := range g.nodes {
		if !exist(v.Name, nodeNames) {
			delNodes = append(delNodes, v)
		}
	}
	return g.DeleteNodes(delNodes)
}

func (g *Graph) String() string {
	str := ""
	for k, iNode := range g.nodes {
		str += iNode.String() + " -> "
		nexts := g.edges[iNode.Name]
		sort.SliceStable(nexts, func(i, j int) bool {
			si := strings.Split(nexts[i].node.Name, "-")
			sj := strings.Split(nexts[j].node.Name, "-")
			return si[len(si)-1] < sj[len(sj)-1]
		})
		for _, next := range nexts {
			str += next.node.String() + fmt.Sprintf("(%d) ", next.val)
		}
		if k != len(g.nodes)-1 {
			str += "\n"
		}
	}
	return str
}

func (n *Node) String() string {
	return n.Name
}

func (g *Graph) edgeExist(u, v *Node) bool {
	if u.Name == v.Name {
		return true
	}
	val, ok := g.edges[u.Name]
	if ok {
		for _, n := range val {
			if n.node.Name == v.Name {
				return true
			}
		}
	}
	return false
}

func (g *Graph) MaxValCount(x int) (int, []string) {
	if x < 1 {
		return 0, nil
	}
	if x > len(g.nodes) {
		return 0, nil
	}
	if len(g.nodes) == 1 {
		return 10, []string{g.nodes[0].Name}
	}
	allNodes := []string{}

	res := 0
	resSet := []string{}
	for _, n := range g.nodes {
		allNodes = append(allNodes, n.Name)
	}
	for _, ss := range subset(allNodes, x) {
		num := g.bridgeVal(ss)
		if num > res {
			res = num
			resSet = ss
		}
	}
	return res, resSet
}

func subset(g []string, x int) [][]string {
	res := [][]string{}
	var dfs func(index int, list []string)

	dfs = func(index int, list []string) {
		if len(list) == x {
			tmp := make([]string, len(list))
			copy(tmp, list)
			res = append(res, tmp)
		}
		for i := index; i < len(g); i++ {
			list = append(list, g[i])
			dfs(i+1, list)
			list = list[:len(list)-1]
		}
	}
	dfs(0, []string{})
	return res
}

func (g *Graph) bridgeVal(subset []string) int {
	if len(subset) <= 1 {
		return 1
	}
	val := 0
	for i := 0; i < len(subset); i++ {
		for _, n := range g.edges[subset[i]] {
			if exist(n.node.Name, subset) {
				if n.node.Name != subset[i] {
					val += n.val
				}
			}
		}
	}
	return val / 2
}

func exist(s string, ss []string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
