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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrap(t *testing.T) {
	g := Graph{}
	a, b, c, d := Node{"a"}, Node{"b"}, Node{"c"}, Node{"d"}

	g.AddNode(&a)
	g.AddNode(&b)
	g.AddNode(&c)
	g.AddNode(&d)

	g.AddEdge(&a, &b, 1)
	g.AddEdge(&a, &c, 2)
	g.AddEdge(&a, &d, 2)
	g.AddEdge(&b, &c, 2)
	g.AddEdge(&b, &d, 1)
	g.AddEdge(&c, &d, 2)

	g.String()
}

func TestSubset(t *testing.T) {
	cases := []struct {
		s   []string
		num int
		res [][]string
	}{
		{
			[]string{"a", "b", "c"},
			3,
			[][]string{
				{
					"a", "b", "c",
				},
			},
		},
		{
			[]string{"a", "b", "c"},
			2,
			[][]string{
				{
					"a", "b",
				},
				{
					"a", "c",
				},
				{
					"b", "c",
				},
			},
		},
	}

	for _, v := range cases {
		assert.Equal(t, v.res, subset(v.s, v.num))
	}
}

func TestBridgeVal(t *testing.T) {
	g1 := &Graph{
		nodes: []*Node{
			{
				"A",
			}, {
				"B",
			}, {
				"C",
			}, {
				"D",
			},
		},
		edges: map[string][]nodeWithVal{
			"A": {
				{
					node: &Node{
						Name: "A",
					},
					val: 100,
				},
				{
					node: &Node{
						Name: "B",
					},
					val: 10,
				},
				{
					node: &Node{
						Name: "C",
					},
					val: 3,
				},
				{
					node: &Node{
						Name: "D",
					},
					val: 11,
				},
			},
			"B": {
				{
					node: &Node{
						Name: "A",
					},
					val: 10,
				},
				{
					node: &Node{
						Name: "B",
					},
					val: 100,
				},
				{
					node: &Node{
						Name: "C",
					},
					val: 4,
				},
				{
					node: &Node{
						Name: "D",
					},
					val: 11,
				},
			},
			"C": {
				{
					node: &Node{
						Name: "A",
					},
					val: 3,
				},
				{
					node: &Node{
						Name: "B",
					},
					val: 4,
				},
				{
					node: &Node{
						Name: "C",
					},
					val: 100,
				},
				{
					node: &Node{
						Name: "D",
					},
					val: 11,
				},
			},
			"D": {
				{
					node: &Node{
						Name: "A",
					},
					val: 11,
				},
				{
					node: &Node{
						Name: "B",
					},
					val: 11,
				},
				{
					node: &Node{
						Name: "C",
					},
					val: 11,
				},
				{
					node: &Node{
						Name: "D",
					},
					val: 100,
				},
			},
		},
	}

	assert.Equal(t, 17, g1.bridgeVal([]string{"A", "B", "C"}))
	assert.Equal(t, 11, g1.bridgeVal([]string{"A", "D"}))
	assert.Equal(t, 25, g1.bridgeVal([]string{"A", "C", "D"}))
	assert.Equal(t, 50, g1.bridgeVal([]string{"A", "B", "C", "D"}))

	g2 := &Graph{
		nodes: []*Node{
			{
				"pe-system-0",
			},
			{
				"pe-system-1",
			},
		},
		edges: map[string][]nodeWithVal{
			"pe-system-0": {
				{
					node: &Node{
						Name: "pe-system-0",
					},
					val: 100,
				},
				{
					node: &Node{
						Name: "pe-system-1",
					},
					val: 5,
				},
			},
			"pe-system-1": {
				{
					node: &Node{
						Name: "pe-system-0",
					},
					val: 5,
				},
				{
					node: &Node{
						Name: "pe-system-1",
					},
					val: 100,
				},
			},
		},
	}

	assert.Equal(t, 1, g2.bridgeVal([]string{"pe-system-1"}))
}

func TestMaxValCount(t *testing.T) {
	a := `
  A        B       C        D 
A x   single  multiple multiple 
B single   x multiple   multiple 
C multiple   multiple x multiple 
D multiple   multiple multiple x 
`
	var score int
	g1 := string2Graph(a)
	score, _ = g1.MaxValCount(2)
	assert.Equal(t, 10, score)

	score, _ = g1.MaxValCount(3)
	assert.Equal(t, 28, score)

	b := `
  A           B         C         D         E        F        G        H 
A x           single    multiple  multiple  node     node     node     node 
B single      x         multiple  multiple  node     node     node     node
C multiple    multiple  x         multiple  node     node     node     node
D multiple    multiple  multiple  x         node     node     node     node   
E node        node      node      node      x        single   multiple multiple
F node        node      node      node      single   x        multiple multiple
G node        node      node      node      multiple multiple x        multiple
H node        node      node      node      multiple multiple multiple x 
`

	g2 := string2Graph(b)
	score, _ = g2.MaxValCount(3)
	assert.Equal(t, 28, score)

	score, _ = g2.MaxValCount(4)
	assert.Equal(t, 55, score)

	score, _ = g2.MaxValCount(5)
	assert.Equal(t, 79, score)

	score, _ = g2.MaxValCount(6)
	assert.Equal(t, 113, score)

	score, _ = g2.MaxValCount(7)
	assert.Equal(t, 155, score)

	score, _ = g2.MaxValCount(8)
	assert.Equal(t, 206, score)
}

func string2Graph(s string) *Graph {
	splitFn := func(c rune) bool {
		return c == ' '
	}
	valMapping := map[string]int{
		"x":        999,
		"multiple": 9,
		"single":   10,
		"node":     6,
		"n":        4,
		"p":        9,
	}
	g := &Graph{
		nodes: []*Node{},
		edges: map[string][]nodeWithVal{},
	}
	lines := strings.Split(s, "\n")
	nodeNames := strings.FieldsFunc(lines[1], splitFn)
	for k, v := range lines {
		if len(v) == 0 {
			continue
		}
		if k == 1 {
			for _, name := range nodeNames {
				g.nodes = append(g.nodes, &Node{
					Name: name,
				})
			}
			continue
		}
		columes := strings.FieldsFunc(v, splitFn)
		var tempEdges = []nodeWithVal{}
		for k2, c := range columes {
			if k2 != 0 {
				tempEdges = append(tempEdges, nodeWithVal{
					val: valMapping[c],
					node: &Node{
						Name: nodeNames[k2-1],
					},
				})
			}
		}
		g.edges[nodeNames[k-2]] = tempEdges
	}
	return g
}

func TestScore(t *testing.T) {
	s := `
  A B C D E F G H
A x p p p n n p n
B p x p p n n n n 
C p p x p p n n n
D p p p x n p n n
E n n p n x p p p
F n n n p n x p p
G p n n n p p x p
H n p n n p p p x
`
	g1 := string2Graph(s)
	score, _ := g1.MaxValCount(4)
	assert.Equal(t, 54, score)

	num := g1.bridgeVal([]string{"A", "C", "D", "G"})
	assert.Equal(t, 44, num)
	n2 := g1.bridgeVal([]string{"A", "B", "C", "D"})
	assert.Equal(t, 54, n2)

	n3 := g1.bridgeVal([]string{"A", "B"})
	assert.Equal(t, 9, n3)

}

func TestDeleteNodes(t *testing.T) {
	s := `
  A B C D E F G H
A x p p p n n p n
B p x p p n n n n 
C p p x p p n n n
D p p p x n p n n
E n n p n x p p p
F n n n p n x p p
G p n n n p p x p
H n p n n p p p x
`

	g1 := string2Graph(s)
	g2 := g1.DeleteNodes([]*Node{
		{Name: "A"},
		{Name: "F"},
		{Name: "G"},
	})
	assert.Equal(t, 5, len(g2.nodes))
	assert.Equal(t, 8, len(g1.nodes))
}

func TestSelectNodes(t *testing.T) {
	s := `
  A B C D E F G H
A x p p p n n p n
B p x p p n n n n 
C p p x p p n n n
D p p p x n p n n
E n n p n x p p p
F n n n p n x p p
G p n n n p p x p
H n p n n p p p x
`
	g1 := string2Graph(s)
	g2 := g1.SelectNodes([]*Node{
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
		{Name: "D"},
	})
	assert.Equal(t, 4, len(g2.nodes))

	g3 := g1.SelectNodes([]*Node{
		{Name: "A"},
		{Name: "E"},
		{Name: "F"},
		{Name: "D"},
	})

	assert.Equal(t, 4, len(g3.nodes))
}
