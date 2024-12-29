package sudoku

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"sudoku_go/global"
)

func valid(a [4][4]int) bool {
	for n := 0; n < 4; n++ {
		column := make(map[int]bool)
		for i := 0; i < 4; i++ {
			column[a[i][n]] = true
		}
		if len(column) != 4 {
			return false
		}
	}

	blocks := [4][4]int{
		{a[0][0], a[0][1], a[1][0], a[1][1]},
		{a[0][2], a[0][3], a[1][2], a[1][3]},
		{a[2][0], a[2][1], a[3][0], a[3][1]},
		{a[2][2], a[2][3], a[3][2], a[3][3]},
	}

	for _, block := range blocks {
		blockMap := make(map[int]bool)
		for _, val := range block {
			blockMap[val] = true
		}
		if len(blockMap) != 4 {
			return false
		}
	}

	return true
}

func permute(arr []int) [][]int {
	var result [][]int
	var permuteFunc func(int)
	permuteFunc = func(i int) {
		if i == len(arr)-1 {
			result = append(result, append([]int(nil), arr...))
		}
		for j := i; j < len(arr); j++ {
			arr[i], arr[j] = arr[j], arr[i]
			permuteFunc(i + 1)
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	permuteFunc(0)
	return result
}

func solutions() [][4][4]int {
	digits := []int{1, 2, 3, 4}
	rows := permute(digits)
	var grids [][4][4]int
	for _, r1 := range rows {
		for _, r2 := range rows {
			for _, r3 := range rows {
				for _, r4 := range rows {
					grid := [4][4]int{
						{r1[0], r1[1], r1[2], r1[3]},
						{r2[0], r2[1], r2[2], r2[3]},
						{r3[0], r3[1], r3[2], r3[3]},
						{r4[0], r4[1], r4[2], r4[3]},
					}
					if valid(grid) {
						grids = append(grids, grid)
					}
				}
			}
		}
	}
	return grids
}

func combinations(arr []int, n int) [][]int {
	var result [][]int
	var combineFunc func(start int, combo []int)
	combineFunc = func(start int, combo []int) {
		if len(combo) == n {
			result = append(result, append([]int(nil), combo...))
			return
		}
		for i := start; i < len(arr); i++ {
			combineFunc(i+1, append(combo, arr[i]))
		}
	}
	combineFunc(0, []int{})
	return result
}

func validPuzzles() [][16]int {
	sols := solutions()
	clueIndices := combinations([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 4)
	var puzzles [][16]int
	for _, sol := range sols {
		flattened := [16]int{}
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				flattened[i*4+j] = sol[i][j]
			}
		}
		for _, clue := range clueIndices {
			clueSet := make(map[int]bool)
			for _, idx := range clue {
				clueSet[flattened[idx]] = true
			}
			if len(clueSet) >= 3 {
				puzzle := [16]int{}
				for i, val := range flattened {
					if contains(clue, i) {
						puzzle[i] = val
					} else {
						puzzle[i] = 0
					}
				}
				puzzles = append(puzzles, puzzle)
			}
		}
	}
	return puzzles
}

func contains(arr []int, val int) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

func allUnique() [][4][4]int {
	puzzles := validPuzzles()
	puzzleMap := make(map[[16]int]int)
	for _, p := range puzzles {
		puzzleMap[p]++
	}
	var uniquePuzzles [][4][4]int
	for p, count := range puzzleMap {
		if count == 1 {
			grid := [4][4]int{}
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					grid[i][j] = p[i*4+j]
				}
			}
			uniquePuzzles = append(uniquePuzzles, grid)
		}
	}
	return uniquePuzzles
}

func AllPuzzle() {
	solutionsData := make(map[string]int)
	for n, p := range allUnique() {
		flattened := [16]int{}
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				flattened[i*4+j] = p[i][j]
			}
		}
		pfStr := ""
		for _, val := range flattened {
			pfStr += fmt.Sprintf("%d", val)
		}
		solutionsData[pfStr] = n
	}
	global.AllPuzz = &solutionsData

	yamlFile, err := os.Create("sudo_dict_puzzle.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer yamlFile.Close()

	encoder := yaml.NewEncoder(yamlFile)
	defer encoder.Close()

	err = encoder.Encode(solutionsData)
	if err != nil {
		fmt.Println(err)
	}

}

func GenByteMap() {
	solutionsData := make(map[int]string)
	solutionsList := solutions()

	for n, p := range solutionsList {
		pf := ""
		for _, row := range p {
			for _, val := range row {
				pf += strconv.Itoa(val)
			}
		}
		solutionsData[n] = pf
	}
	global.ByteMap = &solutionsData

	yamlFile, err := os.Create("sudo_dict_pos.yaml")
	negYamlFile, err := os.Create("sudo_dict_neg.yaml")
	if err != nil {
		fmt.Println("Error creating YAML file:", err)
		return
	}
	defer yamlFile.Close()
	defer negYamlFile.Close()

	encoder := yaml.NewEncoder(yamlFile)
	defer encoder.Close()

	err = encoder.Encode(solutionsData)
	if err != nil {
		fmt.Println("Error encoding YAML data:", err)
	}

	negSolutionsData := make(map[string]int)
	for n, p := range solutionsList {
		pf := ""
		for _, row := range p {
			for _, val := range row {
				pf += strconv.Itoa(val)
			}
		}
		negSolutionsData[pf] = n
	}
	global.StrByte = &negSolutionsData

	negEncoder := yaml.NewEncoder(negYamlFile)
	defer negEncoder.Close()

	err = negEncoder.Encode(negSolutionsData)
	if err != nil {
		fmt.Println("Error encoding YAML data:", err)
	}
}
