package sudoku_go

import (
	"log"
	"strconv"
	"sudoku_go/global"
	"sudoku_go/sudoku"
)

type cipher struct {
	// 设置位掩码
	SBcode uint8
}

// 编码原数据
func (cipher *cipher) Encode(bs []byte) (sixTimeByte []byte) {
	bsLen := len(bs)
	bufLarge := make([]byte, 0, bsLen*6) // 初始化 bufLarge 的容量为 bsLen*6
	for i := 0; i < bsLen; i++ {
		sixBytes := sudoku.FlattenSudoTo6Bytes(sudoku.ByteToRandSudoku(bs[i]), cipher.SBcode)
		bufLarge = append(bufLarge, sixBytes[:]...) // 使用 ... 将 [6]byte 转换为 []byte 并追加
	}
	return bufLarge
}

// 解码原数据
func (cipher *cipher) Decode(sixTimeByte []byte) (bs []byte) {
	// 函数调用时能确保sixTimeByte长度为6的倍数
	bsLen := len(sixTimeByte) / 6
	for i := 0; i < bsLen; i++ {
		flattenSudo := sudoku.UnflattenSudoFrom6Bytes([6]byte(sixTimeByte[i*6:i*6+6]), cipher.SBcode)
		flattenSudoStr := ""
		for i := 0; i < 16; i++ {
			flattenSudoStr += strconv.Itoa(flattenSudo[i])
		}
		// 从global.AllPuzz中，找是否有键flattenSudoStr
		if _, ok := (*global.AllPuzz)[flattenSudoStr]; !ok {
			log.Print("There are multiple solutions")
			bs = []byte(sudoku.ErrNotUnique)
		} else {
			var board [4][4]int
			for j := 0; j < 16; j++ {
				board[j/4][j%4] = flattenSudo[j]
			}
			s, _ := sudoku.SolveSudoku(board)
			e := sudoku.SudokuToByte(s)
			bs = append(bs, e)
		}
	}
	return
}
