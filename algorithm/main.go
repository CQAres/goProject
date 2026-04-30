package main

import (
	"fmt"
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

func clearStars(s string) string {
	ans := make([]byte, 0)
	charSet := make([]int, 26)

	for i := range s {
		if s[i] == '*' {
			if len(ans) == 0 {
				continue
			}

			var minChar byte
			for j := range charSet {
				if charSet[j] > 0 {
					minChar = byte(j) + 'a'
					charSet[j]--
					break
				}
			}

			stack := make([]byte, 0)
			for j := len(ans) - 1; j >= 0; j-- {
				if ans[j] == minChar {
					ans = ans[:j]
					break
				} else {
					stack = append(stack, ans[j])
				}
			}

			for j := len(stack) - 1; j >= 0; j-- {
				ans = append(ans, stack[j])
			}
		} else {
			ans = append(ans, s[i])
			charSet[s[i]-'a']++
		}
	}
	return string(ans)
}

func main() {
	s := "gopher"
	fmt.Printf("Hello and welcome, %s!\n", s)

	fmt.Println(simplifyPath("/.../a/../b/c/../d/./"))
}
