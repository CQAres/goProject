package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 二维前缀和
func MatrixSum(matrix [][]int) [][]int {
	rows := len(matrix)
	cols := len(matrix[0])
	ans := make([][]int, rows)
	for i := range ans {
		ans[i] = make([]int, cols)
	}
	for i := range rows {
		for j := range cols {
			ans[i][j] = matrix[i][j]
			if i > 0 {
				ans[i][j] += ans[i-1][j]
			}
			if j > 0 {
				ans[i][j] += ans[i][j-1]
			}
			if i > 0 && j > 0 {
				ans[i][j] -= ans[i-1][j-1]
			}
		}
	}
	return ans
}

// 二维前缀和查询
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
// 数组 [1, 3, 3, 5, 8]
// 对其补上第0位做差分可以得到数组 [1, 2, 0, 2, 3]
// 差分数组的前缀和就等于原数组, 即 [1, 3, 3, 5, 8]
// 对原数组某个区间进行加减操作,只需对其差分数组的端点做加减操作
// 即 [1, 2+3, 0, 2, 3-3] -> [1, 5, 0, 2, 0] -> [1, 6, 6, 8, 8]
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

// 二维差分
func rangeAddQueries(n int, queries [][]int) [][]int {
	matrix := make([][]int, n)
	for i := range matrix {
		matrix[i] = make([]int, n)
	}

	for i := range queries {
		x1 := queries[i][0]
		y1 := queries[i][1]
		x2 := queries[i][2]
		y2 := queries[i][3]

		matrix[x1][y1] += 1
		if y2+1 < n {
			matrix[x1][y2+1] -= 1
		}
		if x2+1 < n {
			matrix[x2+1][y1] -= 1
		}
		if x2+1 < n && y2+1 < n {
			matrix[x2+1][y2+1] += 1
		}
	}

	for i := range n {
		for j := range n {
			if i > 0 {
				matrix[i][j] += matrix[i-1][j]
			}
			if j > 0 {
				matrix[i][j] += matrix[i][j-1]
			}
			if i > 0 && j > 0 {
				matrix[i][j] -= matrix[i-1][j-1]
			}
		}
	}
	return matrix
}

func buildArray(target []int, n int) []string {
	res := make([]string, 0)
	i := 0
	j := 1
	for i < len(target) {
		val := target[i]
		if val == j {
			res = append(res, "Push")
			i++
			j++
		} else {
			diff := val - j
			for k := 0; k < diff; k++ {
				res = append(res, "Push")
				res = append(res, "Pop")
			}
			j += diff
		}
	}
	return res
}

func simplifyPath(path string) string {
	tokens := strings.Split(path, "/")
	res := make([]string, 0)
	for i := range tokens {
		ch := tokens[i]
		switch ch {
		case "":
			continue
		case ".":
			continue
		case "..":
			if len(res) > 0 {
				res = res[:len(res)-1]
			}
		default:
			res = append(res, ch)
		}
	}
	if len(res) == 0 {
		return "/"
	}
	return "/" + strings.Join(res, "/")
}

func exclusiveTime(n int, logs []string) []int {
	res := make([]int, n)
	stack := make([]int, 0)
	prevTime := 0
	for i := range logs {
		log := logs[i]
		tokens := strings.Split(log, ":")
		id, _ := strconv.Atoi(tokens[0])
		action := tokens[1]
		time, _ := strconv.Atoi(tokens[2])
		if action == "start" {
			if len(stack) > 0 {
				res[stack[len(stack)-1]] += time - prevTime
			}
			stack = append(stack, id)
			prevTime = time
		} else {
			res[stack[len(stack)-1]] += time - prevTime + 1
			stack = stack[:len(stack)-1]
			prevTime = time + 1
		}
	}
	return res
}

// 压缩栈
func removeSubstring(s string, k int) string {

	type pair struct {
		ch  byte
		num int
	}

	stack := make([]pair, 0)
	for i := range s {
		if len(stack) > 0 && stack[len(stack)-1].ch == s[i] {
			stack[len(stack)-1].num++
		} else {
			stack = append(stack, pair{s[i], 1})
		}

		if s[i] == ')' && len(stack) >= 2 {
			n := len(stack)
			rightCnt := stack[n-1].num
			leftCnt := stack[n-2].num
			if leftCnt >= k && rightCnt == k {
				stack = stack[:n-1]
				stack[len(stack)-1].num -= k
				if stack[len(stack)-1].num == 0 {
					stack = stack[:len(stack)-1]
				}
			}
		}
	}

	res := make([]byte, 0)
	for i := range stack {
		for j := 0; j < stack[i].num; j++ {
			res = append(res, stack[i].ch)
		}
	}
	return string(res)
}

func removeOuterParentheses(s string) string {
	var res []byte
	count := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			if count > 0 {
				res = append(res, s[i])
			}
			count++
		} else {
			count--
			if count > 0 {
				res = append(res, s[i])
			}
		}
	}
	return string(res)
}

func scoreOfParentheses(s string) int {
	st := make([]int, 0)
	st = append(st, 0)
	for _, c := range s {
		if c == '(' {
			st = append(st, 0)
		} else {
			v := st[len(st)-1]
			st = st[:len(st)-1]
			st[len(st)-1] += max(2*v, 1)
		}
	}
	return st[0]
}

func main() {
	s := "gopher"
	fmt.Printf("Hello and welcome, %s!\n", s)

}
