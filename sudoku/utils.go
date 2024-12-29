package sudoku

import (
	"log"
	"math/rand"
	"strconv"
	"sudoku_go/global"
	"time"
)

// 组合生成器
func comb(n, k int) [][]int {
	var result [][]int
	var comb []int
	var backtrack func(start int)
	backtrack = func(start int) {
		if len(comb) == k {
			result = append(result, append([]int(nil), comb...))
			return
		}
		for i := start; i < n; i++ {
			comb = append(comb, i)
			backtrack(i + 1)
			comb = comb[:len(comb)-1]
		}
	}
	backtrack(0)
	return result
}

// 生成所有可能的 4 个线索的数独谜题
func generate4CluePuzzles(sudoku [4][4]int) [][16]int {
	var puzzles [][16]int
	clues := comb(16, 4) // 4 = number of given clues

	for _, clue := range clues {
		puzzle := [4][4]int{}
		for _, c := range clue {
			puzzle[c/4][c%4] = sudoku[c/4][c%4]
		}
		if len(uniqueValues(puzzle)) >= 3 {
			var pz [16]int
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					pz[i*4+j] = puzzle[i][j]
				}
			}
			puzzles = append(puzzles, pz)
		}
	}

	return puzzles
}

// 提取数独中的唯一值
func uniqueValues(sudoku [4][4]int) map[int]struct{} {
	values := make(map[int]struct{})
	for _, row := range sudoku {
		for _, val := range row {
			if val != 0 {
				values[val] = struct{}{}
			}
		}
	}
	return values
}

//func parseSudoDict(filename string) map[int]string {
//	data, err := ioutil.ReadFile(filename)
//	if err != nil {
//		log.Fatalf("Failed to read file: %v", err)
//	}
//
//	var sudoDict map[int]string
//	err = yaml.Unmarshal(data, &sudoDict)
//	if err != nil {
//		log.Fatalf("Failed to parse YAML: %v", err)
//	}
//
//	return sudoDict
//}

func byteToSudoku(byteValue byte, sudoDict map[int]string) [4][4]int {
	key := int(byteValue)
	sudokuStr, exists := sudoDict[key]
	if !exists {
		log.Fatalf("Key %d not found in sudo_dict", key)
	}

	var sudoku [4][4]int
	for i := 0; i < 16; i++ {
		val, _ := strconv.Atoi(string(sudokuStr[i]))
		sudoku[i/4][i%4] = val
	}

	return sudoku
}

// 检查数独板上的指定位置是否可以放置数字num
func isValid(board [4][4]int, row, col, num int) bool {
	// 检查行
	for x := 0; x < 4; x++ {
		if board[row][x] == num {
			return false
		}
	}

	// 检查列
	for x := 0; x < 4; x++ {
		if board[x][col] == num {
			return false
		}
	}

	// 检查 2x2 子方块
	startRow := row - row%2
	startCol := col - col%2
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if board[startRow+i][startCol+j] == num {
				return false
			}
		}
	}

	return true
}

// 检查数独是否有多个解
func CheckMultipleSolution(board [4][4]int) bool {
	count := 0
	return solveSudokuWithCount(board, &count)
}

// 辅助函数：解决数独并计数解的数量
func solveSudokuWithCount(board [4][4]int, count *int) bool {
	find := false
	var row, col int
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if board[i][j] == 0 {
				row = i
				col = j
				find = true
				break
			}
		}
		if find {
			break
		}
	}

	if !find {
		*count++
		return *count > 1
	}

	for num := 1; num <= 4; num++ {
		if isValid(board, row, col, num) {
			board[row][col] = num
			if solveSudokuWithCount(board, count) {
				return true
			}
			board[row][col] = 0
		}
	}

	return false
}

// 解决数独问题
// SolveSudoku solves a 4x4 Sudoku puzzle using backtracking with optimizations.
func SolveSudoku(board [4][4]int) ([4][4]int, bool) {
	var rows, cols [4][5]bool
	var blocks [4][5]bool

	// Initialize constraint tracking
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if board[i][j] != 0 {
				num := board[i][j]
				rows[i][num] = true
				cols[j][num] = true
				blocks[getBlock(i, j)][num] = true
			}
		}
	}

	// Recursive function to solve the board
	var backtrack func(int, int) bool
	backtrack = func(row, col int) bool {
		if row == 4 {
			return true
		}

		nextRow, nextCol := row, col+1
		if nextCol == 4 {
			nextRow, nextCol = row+1, 0
		}

		if board[row][col] != 0 {
			return backtrack(nextRow, nextCol)
		}

		for num := 1; num <= 4; num++ {
			block := getBlock(row, col)
			if !rows[row][num] && !cols[col][num] && !blocks[block][num] {
				// Place the number
				board[row][col] = num
				rows[row][num] = true
				cols[col][num] = true
				blocks[block][num] = true

				// Recursively solve the rest
				if backtrack(nextRow, nextCol) {
					return true
				}

				// Undo the placement
				board[row][col] = 0
				rows[row][num] = false
				cols[col][num] = false
				blocks[block][num] = false
			}
		}
		return false
	}

	if backtrack(0, 0) {
		return board, true
	} else {
		return board, false
	}
}

// getBlock determines the block index for a given cell.
func getBlock(row, col int) int {
	return (row/2)*2 + (col / 2)
}

func ByteToRandSudoku(index byte) [16]int {
	rand.Seed(time.Now().UnixNano())
	l := len((*global.ByteList)[int(index)])
	arrSudoStr := (*global.ByteList)[int(index)][rand.Intn(l)]
	// arrSudoStr转为16位数组
	var arrSudo [16]int
	for i := 0; i < 16; i++ {
		val, _ := strconv.Atoi(string(arrSudoStr[i]))
		arrSudo[i] = val
	}
	return arrSudo
}

func ByteToSudokuList(index byte) []string {
	sudoDict := *global.ByteMap
	sudoku := byteToSudoku(index, sudoDict)
	puzzles := generate4CluePuzzles(sudoku)
	var AllSudoStr []string
	for _, puzzle := range puzzles {
		// reshape puzzle
		var board [4][4]int
		for i := 0; i < 16; i++ {
			board[i/4][i%4] = puzzle[i]
		}

		// 检查是否有多个解
		if CheckMultipleSolution(board) {
			continue
		} else {
			// 唯一解
			pf := ""
			for _, row := range board {
				for _, val := range row {
					pf += strconv.Itoa(val)
				}
			}
			AllSudoStr = append(AllSudoStr, pf)
		}
	}
	return AllSudoStr
}

func FlattenSudoTo6Bytes(sudoku [16]int, sbCode uint8) (encode [6]byte) {
	var one_positions [16]int
	if sbCode == 0x01 {
		for i := 0; i < 16; i++ {
			one_positions[i] = 0
			if sudoku[i] == 1 {
				one_positions[i] = 1
			}
			if sudoku[i] != 0 {
				sudoku[i] -= 1
			}
		}
	}
	if sbCode == 0x00 {
		for i := 0; i < 16; i++ {
			one_positions[i] = 1
			if sudoku[i] == 1 {
				one_positions[i] = 0
			}
			if sudoku[i] != 0 {
				sudoku[i] -= 1
			}
		}
	}

	//  Convert the sudoku array to a 32-bit binary number
	// each 2 bits represent a number in the sudoku array
	// 00 -> 0, 01 -> 1, 10 -> 2, 11 -> 3

	var binary32 [4]byte
	for i := 0; i < 16; i++ {
		binary32[i/4] |= byte(sudoku[i] << uint((3-i%4)*2))
	}

	//  Convert the one_positions array to a 16-bit binary number

	var binary16 [2]byte
	for i := 0; i < 16; i++ {
		binary16[i/8] |= byte(one_positions[i] << uint(7-i%8))
	}

	// combine the two binary numbers to a 48-bit binary number
	for i := 0; i < 6; i++ {
		if i < 4 {
			encode[i] = binary32[i]
		} else {
			encode[i] = binary16[i-4]
		}
	}
	return
}

// 逆向上面这个函数，从6byte生成[16]int
func UnflattenSudoFrom6Bytes(encode [6]byte, sbCode uint8) (sudoku [16]int) {
	var binary32 [4]byte
	var binary16 [2]byte
	for i := 0; i < 6; i++ {
		if i < 4 {
			binary32[i] = encode[i]
		} else {
			binary16[i-4] = encode[i]
		}
	}

	//  Convert the 48-bit binary number to two 32-bit binary numbers
	// 32-bit binary number to a sudoku array
	// each 2 bits represent a number in the sudoku array
	// 00 -> 0, 01 -> 1, 10 -> 2, 11 -> 3
	for i := 0; i < 16; i++ {
		sudoku[i] = int((binary32[i/4] >> uint((3-i%4)*2)) & 0x03)
	}

	//  Convert the 16-bit binary number to a one_positions array
	var onePositions [16]int
	for i := 0; i < 16; i++ {
		onePositions[i] = int((binary16[i/8] >> uint(7-i%8)) & 0x01)
	}

	if sbCode == 0x01 {
		for i := 0; i < 16; i++ {
			if sudoku[i] != 0 {
				sudoku[i] += 1
			}
			if onePositions[i] == 1 {
				sudoku[i] = 1
			}
		}
	}
	if sbCode == 0x00 {
		for i := 0; i < 16; i++ {
			if sudoku[i] != 0 {
				sudoku[i] += 1
			}
			if onePositions[i] == 0 {
				sudoku[i] = 1
			}
		}
	}

	return
}

func SudokuToByte(sudoku [4][4]int) byte {
	var byteValue byte
	var sudoDictNeg map[string]int
	sudoDictNeg = *global.StrByte
	pf := ""
	for _, row := range sudoku {
		for _, val := range row {
			pf += strconv.Itoa(val)
		}
	}
	byteValue = byte(sudoDictNeg[pf])
	return byteValue

}
