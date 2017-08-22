package common

import (
	"strconv"
	"strings"
)

// Version provides utility methods for comparing versions.
type Version string

func (v Version) compareTo(other Version) int {
	currentDash := strings.SplitN(string(v), "-", 2)
	otherDash := strings.SplitN(string(other), "-", 2)
	currTab := strings.Split(currentDash[0], ".")
	otherTab := strings.Split(otherDash[0], ".")

	max := len(currTab)
	if len(otherTab) > max {
		max = len(otherTab)
	}
	for i := 0; i < max; i++ {
		var currInt, otherInt int

		if len(currTab) > i {
			currInt, _ = strconv.Atoi(currTab[i])
		}
		if len(otherTab) > i {
			otherInt, _ = strconv.Atoi(otherTab[i])
		}
		if currInt > otherInt {
			return 1
		}
		if otherInt > currInt {
			return -1
		}
	}

	if len(otherDash) > len(currentDash) {
		return -1
	} else if len(otherDash) < len(currentDash) {
		return 1
	}

	if len(otherDash) > 1 {
		if otherDash[1] == currentDash[1] {
			return 0
		}

		nonNumbers := func(r rune) bool {
			if !(r > 0 && r < 9) {
				return false
			}
			return true
		}

		otherPost := strings.TrimRightFunc(otherDash[1], nonNumbers)
		currentPost := strings.TrimRightFunc(currentDash[1], nonNumbers)

		currInt, _ := strconv.Atoi(currentPost)
		otherInt, _ := strconv.Atoi(otherPost)
		if currInt > otherInt {
			return 1
		}
		if otherInt > currInt {
			return -1
		}

		// compare by rune number
		max := len(currentDash[1])
		if len(otherDash[1]) > max {
			max = len(otherDash[1])
		}
		for i := 0; i < max; i++ {
			if otherDash[1][i] != currentDash[1][i] {
				if otherDash[1][i] > currentDash[1][i] {
					return -1
				}
				return 1
			}
		}

	}
	return 0
}

// LessThan checks if a version is less than another
func (v Version) LessThan(other Version) bool {
	return v.compareTo(other) == -1
}

// LessThanOrEqualTo checks if a version is less than or equal to another
func (v Version) LessThanOrEqualTo(other Version) bool {
	return v.compareTo(other) <= 0
}

// GreaterThan checks if a version is greater than another
func (v Version) GreaterThan(other Version) bool {
	return v.compareTo(other) == 1
}

// GreaterThanOrEqualTo checks if a version is greater than or equal to another
func (v Version) GreaterThanOrEqualTo(other Version) bool {
	return v.compareTo(other) >= 0
}

// Equal checks if a version is equal to another
func (v Version) Equal(other Version) bool {
	return v.compareTo(other) == 0
}
