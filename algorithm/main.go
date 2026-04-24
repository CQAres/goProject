package main

import (
	"fmt"
)

// 二维前缀和
func SumMatrixRegion(matrix [][]int, row1 int, col1 int, row2 int, col2 int) int {
	matSum := matrix[row2][col2]
	var right, left, up int
	if row1 != 0 {
		right = matrix[row1-1][col2]
	}
	if col1 != 0 {
		left = matrix[row2][col1-1]
	}
	if row1 != 0 && col1 != 0 {
		up = matrix[row1-1][col1-1]
	}
	return matSum - left - right + up
}

// 差分
func isZeroArray(nums []int, queries [][]int) bool {
	nNum := len(nums)
	coverNum := make([]int, nNum)
	nQuery := len(queries)
	for i := range nQuery {
		left := queries[i][0]
		right := queries[i][1]
		if left < 0 || right >= nNum || left > right {
			return false
		}
		coverNum[left] += 1
		if right+1 < nNum {
			coverNum[right+1] -= 1
		}
	}
	prefixSum := 0
	for i, val := range coverNum {
		prefixSum += val
		if prefixSum < nums[i] {
			return false
		}
	}
	return true
}

func minKBitFlips(nums []int, k int) int {
	n := len(nums)
	diff := make([]int, n+k)
	prefix := 0
	res := 0 // 翻转次数
	for i := range nums {
		prefix += diff[i]

		current := nums[i]
		if prefix%2 == 1 {
			current ^= 1 // 翻转
		}

		if current == 0 {
			// 需要翻转
			if i+k > n {
				return -1
			}
			prefix++
			diff[i+k]--
			res++
		}
	}
	return res
}

func main() {
	s := "gopher"
	fmt.Printf("Hello and welcome, %s!\n", s)
}
