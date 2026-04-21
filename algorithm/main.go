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

func maxAliveYear(birth []int, death []int) int {
	diff := 1900
	dissVec := make([]int, 102)
	length := len(birth)
	ansCnt, ansIndex, peopleCnt := 0, 0, 0
	for i := 0; i < length; i++ {
		bir := birth[i]
		dea := death[i]
		dissVec[bir-diff] += 1
		dissVec[dea-diff+1] -= 1
	}
	for i := 0; i < len(dissVec); i++ {
		peopleCnt += dissVec[i]
		if peopleCnt > ansCnt {
			ansCnt = peopleCnt
			ansIndex = i
		}
	}
	return ansIndex + diff
}

func countTestedDevices(batteryPercentages []int) int {
	dec := 0
	for _, battery := range batteryPercentages {
		if battery > dec {
			dec++
		}
	}
	return dec
}

func carPooling(trips [][]int, capacity int) bool {
	CarTrips := make([]int, 1002)
	for i := 0; i < len(trips); i++ {
		numPassengers := trips[i][0]
		from := trips[i][1]
		to := trips[i][2]
		CarTrips[from] += numPassengers
		CarTrips[to] -= numPassengers
	}
	peopleCnt := 0
	for i := range CarTrips {
		peopleCnt += CarTrips[i]
		if peopleCnt > capacity {
			return false
		}
	}
	return true
}

func corpFlightBookings(bookings [][]int, n int) []int {
	flightList := make([]int, n+1)
	for i := range bookings {
		first := bookings[i][0]
		last := bookings[i][1]
		peopleCnt := bookings[i][2]

		flightList[first-1] += peopleCnt
		flightList[last] -= peopleCnt
	}
	ans := make([]int, n)
	peopleCnt := 0
	for i := range ans {
		peopleCnt += flightList[i]
		ans[i] = peopleCnt
	}
	return ans
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
