package asciitree

import (
	"regexp"
	"slices"
	"strings"
)

// NOTE: This function take the data with the assumption that the indentation size is 1 space each
func GenerateTree(data string) string {
	// Convert into matrix
	treeMap := [][]string{}
	rs := ""

	lines := strings.Split(data, "\n")
	lines = lines[:len(lines)-1]
	for _, line := range lines {
		// Split the string, but keep the non whitespace part as a whole string
		indentSize := regexp.MustCompile(`\S`).FindStringIndex(line)[0]
		if indentSize == 0 {
			treeMap = append(treeMap, []string{line})
		} else {
			treeMap = append(treeMap, append(strings.Split(line[:indentSize], ""), line[indentSize:]))
		}
	}

	// Start adding tree visualization
	for i := 0; i < len(treeMap); i++ {
		// Find the range of a branch
		branchRange := slices.IndexFunc(treeMap[i+1:], func(line []string) bool {
			return len(line) <= len(treeMap[i])
		})
		if branchRange == -1 {
			branchRange = len(treeMap[i:]) - 1
		}

		// Start drawing from the end of range to the beginning
		startDraw := false

		for j := i + branchRange; j >= i+1; j-- {
			if len(treeMap[j])-len(treeMap[i]) != 1 {
				if startDraw {
					treeMap[j][len(treeMap[i])-1] = "│  "
				}
			} else {
				if startDraw {
					treeMap[j][len(treeMap[i])-1] = "├──"
					continue
				}
				treeMap[j][len(treeMap[i])-1] = "└──"
				startDraw = true
			}

		}
	}

	// Render the final tree
	for i := range treeMap {
		for j := range treeMap[i] {
			if treeMap[i][j] == " " {
				rs += "   "
			} else {
				rs += treeMap[i][j]
			}
		}

		rs += "\n"
	}

	return rs
}
