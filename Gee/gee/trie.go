package gee

import "strings"

type node struct {
	pattern  string
	part     string
	children []*node
	// wildcard 通配
	isWild bool
}
type nodeWithWeight struct {
	n *node
	w int
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part {
			return child
		}
	}
	return nil
}

func (n *node) matchChildren(part string) []*nodeWithWeight {
	nodes := make([]*nodeWithWeight, 0)
	for _, child := range n.children {
		if child.part == part {
			// 完全匹配加权
			nodes = append(nodes, &nodeWithWeight{n: child, w: 2})
		}
		if child.isWild {
			nodes = append(nodes, &nodeWithWeight{n: child, w: 1})
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)

	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *nodeWithWeight {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return &nodeWithWeight{n: n}
	}

	part := parts[height]
	children := n.matchChildren(part)
	res := make([]*nodeWithWeight, 0)
	for _, child := range children {
		findNode := child.n.search(parts, height+1)
		if findNode != nil {
			res = append(res, &nodeWithWeight{n: findNode.n, w: child.w + findNode.w})
		}
	}
	if len(res) == 0 {
		return nil
	}

	maxWeight := 0
	maxIndex := 0
	for i, item := range res {
		if item.w > maxWeight {
			maxWeight = item.w
			maxIndex = i
		}
	}
	return res[maxIndex]
}
