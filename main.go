package main

import (
	"fmt"
)

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

func maximumPopulation(logs [][]int) int {
	cover := make([]int, 102)
	for _, val := range logs {
		birth := val[0] - 1950
		death := val[1] - 1950
		cover[birth] += 1
		cover[death] -= 1
	}

	maxPopulation := 0
	maxIndex := 0
	ans := 0
	for i, val := range cover {
		maxPopulation += val
		if maxPopulation > ans {
			ans = maxPopulation
			maxIndex = i
		}
	}
	return maxIndex + 1950
}

func main() {
	s := "gopher"
	fmt.Printf("Hello and welcome, %s!\n", s)

	m := [][]int{
		{1950, 1961},
		{1960, 1971},
		{1970, 1981},
	}
	fmt.Print(maximumPopulation(m))
}
