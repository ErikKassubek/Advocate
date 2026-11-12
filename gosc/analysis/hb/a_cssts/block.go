// Copyright (c) 2025 
//
// File: block.go
// Brief: Implementation of Block.java from CSST paper (sea csst.go)
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
	"gosc/utils/types"
	"math"
)

type segmentTreeNodeBlock struct {
	start, end  int
	left, right *segmentTreeNodeBlock
	min, pos    int
	block       []int
	isLeaf      bool
}

func newSegmentTreeNodeBlock(start, end int) segmentTreeNodeBlock {
	return segmentTreeNodeBlock{
		start:  start,
		end:    end,
		left:   nil,
		right:  nil,
		min:    math.MaxInt,
		pos:    -1,
		isLeaf: false,
	}
}

type block struct {
	offset   int
	root     *segmentTreeNodeBlock
	maxLevel int
	length   int
}

func newBlock1(length, maxLevel, offset int) *block {
	return &block{
		offset:   offset,
		root:     buildTreeBlock2(0, length-1, 0, maxLevel),
		maxLevel: maxLevel,
		length:   length,
	}
}

// func newBlock2(nums []int, maxLevel int) *block {
// 	return &block{
// 		root:     buildTreeBlock3(nums, 0, len(nums)-1, 0, maxLevel),
// 		maxLevel: maxLevel,
// 	}
// }

// func newBlock3(other *block) *block {
// 	return &block{
// 		root:     buildTreeBlock1(other, other.root),
// 		maxLevel: other.maxLevel,
// 	}
// }

// func buildTreeBlock1(other *block, otherCurrent *segmentTreeNodeBlock) *segmentTreeNodeBlock {
// 	if otherCurrent == nil {
// 		return nil
// 	}
// 	thisCurrent := newSegmentTreeNodeBlock(otherCurrent.start, otherCurrent.end)
// 	thisCurrent.min = otherCurrent.min
// 	thisCurrent.isLeaf = otherCurrent.isLeaf
// 	thisCurrent.pos = otherCurrent.pos
// 	if otherCurrent.block != nil {
// 		copy(thisCurrent.block, otherCurrent.block)
// 	}
// 	if otherCurrent.left != nil {
// 		thisCurrent.left = buildTreeBlock1(other, otherCurrent.left)
// 	}
// 	if otherCurrent.right != nil {
// 		thisCurrent.right = buildTreeBlock1(other, otherCurrent.right)
// 	}
// 	return &thisCurrent
// }

func buildTreeBlock2(start, end, level, maxLevel int) *segmentTreeNodeBlock {
	if start > end {
		return nil
	}
	ret := newSegmentTreeNodeBlock(start, end)

	if level >= maxLevel {
		ret.min = math.MaxInt
		ret.pos = -1
		ret.isLeaf = true
		return &ret
	} else if start == end {
		ret.min = math.MaxInt
		ret.pos = start
		ret.isLeaf = true
	} else {
		mid := start + (end-start)/2
		ret.left = buildTreeBlock2(start, mid, level+1, maxLevel)
		ret.right = buildTreeBlock2(mid+1, end, level+1, maxLevel)
		ret.min = min(ret.left.min, ret.right.min)
		if ret.left.min < ret.right.min {
			ret.pos = ret.left.pos
		} else {
			ret.pos = ret.right.pos
		}
	}
	return &ret
}

// func buildTreeBlock3(nums []int, start, end, level, maxLevel int) *segmentTreeNodeBlock {
// 	if start > end {
// 		return nil
// 	}

// 	ret := newSegmentTreeNodeBlock(start, end)

// 	if level >= maxLevel {
// 		ret.block = types.CopyOfRange(nums, start, end+1)
// 		minPos := getBlockMin(&ret, start, end)
// 		ret.min = minPos.X
// 		ret.pos = minPos.Y
// 		ret.isLeaf = true
// 		return &ret
// 	}
// 	if start == end {
// 		ret.min = nums[start]
// 		ret.pos = start
// 		ret.block = types.CopyOfRange(nums, start, end+1)
// 		ret.isLeaf = true
// 	} else {
// 		mid := start + (end-start)/2
// 		ret.left = buildTreeBlock3(nums, start, mid, level+1, maxLevel)
// 		ret.right = buildTreeBlock3(nums, mid+1, end, level+1, maxLevel)
// 		ret.min = min(ret.left.min, ret.right.min)
// 		if ret.left.min < ret.right.min {
// 			ret.pos = ret.left.pos
// 		} else {
// 			ret.pos = ret.right.pos
// 		}
// 	}
// 	return &ret
// }

func getBlockMin(node *segmentTreeNodeBlock, start, end int) types.Pair[int, int] {
	min := math.MaxInt
	pos := -1

	if node.block != nil {
		for i := end - node.start; i >= start-node.start; i-- {
			v := node.block[i]
			if v < min {
				min = v
				pos = i
			}
		}
		if pos == -1 {
			pos = end
		} else {
			pos = node.start + pos
		}
	}
	return types.NewPair(min, pos)
}

func getBlockArgMin(node *segmentTreeNodeBlock, x int) int {
	for i := len(node.block) - 1; i >= 0; i-- {
		v := node.block[i]
		if v <= x {
			return node.start + i
		}
	}
	return -1
}

func (this block) update1(i, val int) {
	i -= this.offset
	this.update2(this.root, i, val)
}

func (this block) update2(root *segmentTreeNodeBlock, pos, val int) {
	if root.isLeaf {
		if root.block == nil {
			root.block = make([]int, root.end+1-root.start)
			types.Fill(root.block, math.MaxInt)
		}
		if val > root.block[pos-root.start] {
			return
		}

		root.block[pos-root.start] = val

		if val < root.min {
			root.min = val
			root.pos = pos
		}
		if val == root.min && pos > root.pos {
			root.pos = pos
		}
		if root.pos == pos {
			minPos := getBlockMin(root, root.start, root.end)
			root.min = minPos.X
			root.pos = minPos.Y
		}
	} else {
		mid := root.start + (root.end-root.start)/2
		if pos <= mid {
			this.update2(root.left, pos, val)
		} else {
			this.update2(root.right, pos, val)
		}
		if root.left.min < root.right.min {
			root.min = root.left.min
			root.pos = root.left.pos
		} else {
			root.min = root.right.min
			root.pos = root.right.pos
		}
	}
}

func (this block) argMin1(x int) int {
	if this.root.min <= x {
		return this.argMin2(this.root, x) + this.offset
	}
	return -1
}

func (this block) argMin2(root *segmentTreeNodeBlock, x int) int {
	if root.isLeaf {
		return getBlockArgMin(root, x)
	}

	if root.right.min <= x {
		return this.argMin2(root.right, x)
	}

	return this.argMin2(root.left, x)
}

func (this block) sumRange1(i, j int) int {
	i -= this.offset
	j -= this.offset
	return this.sumRange2(this.root, i, j)
}

func (this block) sumRange2(root *segmentTreeNodeBlock, start, end int) int {
	if root.end == end && root.start == start {
		return root.min
	} else if root.isLeaf {
		if root.pos >= start && root.end <= end {
			return root.min
		}
		return getBlockMin(root, start, end).X
	} else {
		mid := root.start + (root.end-root.start)/2
		if end <= mid {
			return this.sumRange2(root.left, start, end)
		} else if start >= mid+1 {
			return this.sumRange2(root.right, start, end)
		} else {
			l := this.sumRange2(root.left, start, mid)
			r := this.sumRange2(root.right, mid+1, end)
			return min(l, r)
		}
	}
}
