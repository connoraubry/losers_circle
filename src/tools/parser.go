package tools

import (
	"strconv"
)

type Parser struct {
	ByteMap map[byte][]GameResult
}
type GameResult struct {
	Selected bool
	HomeWon  bool
}

func NewParser() *Parser {
	p := &Parser{}

	byteMap := make(map[byte][]GameResult)

	for i := 0; i < 16; i++ {

		res := []GameResult{
			{i&8 > 0, i&4 > 0},
			{i&2 > 0, i&1 > 0},
		}
		key := strconv.FormatInt(int64(i), 16)
		byteKey := byte(key[0])

		byteMap[byteKey] = res
	}
	p.ByteMap = byteMap
	return p
}

func (p Parser) ParseWeek(week []byte) []GameResult {

	var res []GameResult
	for _, b := range week {
		res = append(res, p.ByteMap[b]...)
	}

	return res
}
