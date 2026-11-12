// Copyright (c) 2025 
//
// File: sparseSegmentTree.go
// Brief: Implementation of SparseSegmentTree.java from CSST paper (see csst.go)
//
//
// License: BSD-3-Clause

// CSST is based on
// Hünkar Can Tunç, Ameya Prashant Deshmukh, Berk Cirisci, Constantin Enea,
// and Andreas Pavlogiannis. 2024. CSSTs: A Dynamic Data Structure for Partial
// Orders in Concurrent Execution Analysis. In Proceedings of the 29th ACM
// International Conference on Architectural Support for Programming Languages
// and Operating Systems, Volume 3 (ASPLOS '24), Vol. 3. Association for
// Computing Machinery, New York, NY, USA, 223–238.
// https://doi.org/10.1145/3620666.3651358

package a_cssts

import (
	"math"
)

const blockSize = 128

type segmentTreeNode struct {
	start, end             int
	left, right            *segmentTreeNode
	min, pos               int
	activeStart, activeEnd int
	level                  int
	block                  *block
}

func newSegmentTreeNode1(start, end, level int) segmentTreeNode {
	return segmentTreeNode{
		start:       start,
		end:         end,
		left:        nil,
		right:       nil,
		min:         math.MaxInt,
		pos:         -1,
		activeStart: math.MinInt,
		activeEnd:   math.MinInt,
		block:       nil,
		level:       level,
	}
}

func newSegmentTreeNode2(start, end, pos, val, level int) segmentTreeNode {
	return segmentTreeNode{
		start:       start,
		end:         end,
		left:        nil,
		right:       nil,
		min:         val,
		pos:         pos,
		activeStart: pos,
		activeEnd:   pos,
		block:       nil,
		level:       level,
	}
}

type sparseSegmentTree struct {
	root     *segmentTreeNode
	maxLevel int
}

func newSparseSegmentTree1(length int) sparseSegmentTree {
	v := length / blockSize
	height := int(math.Log(float64(v)) / math.Log(2))

	height = max(1, height)

	segmentTreeNode := newSegmentTreeNode1(0, length-1, 0)

	return sparseSegmentTree{
		root:     &segmentTreeNode,
		maxLevel: height,
	}
}

// func newSparseSegmentTree2(other sparseSegmentTree) sparseSegmentTree {
// 	root := buildTree(other, other.root)
// 	return sparseSegmentTree{
// 		root:     root,
// 		maxLevel: other.maxLevel,
// 	}
// }

// func buildTree(other sparseSegmentTree, otherCurrent *segmentTreeNode) *segmentTreeNode {
// 	if otherCurrent == nil {
// 		return nil
// 	}

// 	thisCurrent := newSegmentTreeNode1(otherCurrent.start, otherCurrent.end, otherCurrent.level)
// 	thisCurrent.min = otherCurrent.min
// 	thisCurrent.pos = otherCurrent.pos

// 	if otherCurrent.left != nil {
// 		thisCurrent.left = buildTree(other, otherCurrent.left)
// 	}

// 	if otherCurrent.right != nil {
// 		thisCurrent.right = buildTree(other, otherCurrent.right)
// 	}

// 	return &thisCurrent
// }

func (this sparseSegmentTree) update1(i, val int) {
	this.update2(this.root, i, val)
}

func (this sparseSegmentTree) update2(root *segmentTreeNode, pos, val int) {
	if !(root == nil || root.pos == -1 || (root.pos >= root.start && root.pos <= root.end)) {
		panic("Invalid root in sparseSegmentTree::update2 (1)")
	}

	if root.block != nil {
		root.block.update1(pos, val)
		root.min = root.block.root.min
		root.pos = root.block.root.pos + root.block.offset
		if pos < root.activeStart {
			root.activeStart = pos
		}
		if pos > root.activeEnd {
			root.activeEnd = pos
		}
		return
	}

	switch root.pos {
	case pos:
		if val <= root.min {
			root.min = val
		} else if val > root.min {
			return
		} else {
			panic("Invalid root in sparseSegmentTree::update2 (2)")
		}
	case -1:
		root.min = val
		root.pos = pos
		root.activeStart = pos
		root.activeEnd = pos
		return
	default:
		if pos < root.activeStart {
			root.activeStart = pos
		}
		if pos > root.activeEnd {
			root.activeEnd = pos
		}

		if val < root.min || (val == root.min && pos > root.pos) {
			oldMin := root.min
			oldPos := root.pos
			root.min = val
			root.pos = pos
			val = oldMin
			pos = oldPos
		}

		mid := root.start + (root.end-root.start)/2
		if pos <= mid {
			if root.left == nil {
				root.left = this.createIntermediateNode(root.start, mid, pos, val, root.level+1)
			} else {
				this.update2(root.left, pos, val)
			}
		} else {
			if root.right == nil {
				root.right = this.createIntermediateNode(mid+1, root.end, pos, val, root.level+1)
			}
		}
	}
}

func (this *sparseSegmentTree) createIntermediateNode(start, end, pos, val, level int) *segmentTreeNode {
	if level >= this.maxLevel {
		transitionNode := newSegmentTreeNode2(start, end, pos, val, level)
		transitionNode.block = newBlock1(end-start+1, 0, start)
		transitionNode.block.update1(pos, val)
		return &transitionNode
	}
	if start > end {
		return nil
	}
	newNode := newSegmentTreeNode2(start, end, pos, val, level)
	return &newNode
}

func (this sparseSegmentTree) argMin1(x int) int {
	if this.root.min <= x {
		return this.argMin2(this.root, x)
	}
	return -1
}

func (this sparseSegmentTree) argMin2(root *segmentTreeNode, x int) int {
	if root.block != nil {
		return root.block.argMin1(x)
	}

	mid := root.start + (root.end-root.start)/2

	if root.start == root.end || root.min == x || (root.right == nil && root.left == nil) {
		return root.pos
	} else if root.right == nil && root.left != nil && root.pos >= root.left.end && root.min <= x {
		return root.pos
	} else if root.right != nil && root.pos >= root.right.end && root.min <= x {
		return root.pos
	} else if root.right != nil && root.right.min <= x {
		return max(root.pos, this.argMin2(root.right, x))
	} else if root.right != nil && root.right.min > x && root.min <= x && root.pos != root.right.pos && root.pos > mid {
		return root.pos
	} else {
		if root.left == nil || root.left.min > x {
			return root.pos
		}
		return max(root.pos, this.argMin2(root.left, x))
	}
}

func (this sparseSegmentTree) sumRange1(i int) int {
	return this.sumRange2(this.root, i)
}

func (this sparseSegmentTree) sumRange2(root *segmentTreeNode, index int) int {
	if root == nil || root.min == math.MaxInt || root.activeEnd < index {
		return math.MaxInt
	}

	if root.block != nil {
		return root.block.sumRange1(index, root.end)
	}

	if root.pos >= index {
		return root.min
	}

	l := this.sumRange2(root.left, index)
	r := this.sumRange2(root.right, index)
	return min(l, r)

}
