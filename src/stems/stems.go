package stems

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"github.com/connoraubry/losers_circle/src/stems/graph"
)

type Stems struct {
	Levels []*StemLevel
	Loops  map[int][]int
}

type StemLevel struct {
	//int is start idx
	Starts    map[int]*StemStart
	FilledOut bool
}

type BitmaskStruct struct {
	Bitmask uint32

	FrontLoc BitmaskLocation
	EndLoc   BitmaskLocation
}

type BitmaskLocation struct {
	Start      int
	End        int
	StemLen    int
	BitmaskIdx int
}

func (bs *BitmaskStruct) ToString() string {
	mask := strconv.FormatInt(int64(bs.Bitmask), 2)
	frontString := bs.FrontLoc.ToString()
	backString := bs.EndLoc.ToString()

	return fmt.Sprintf("Mask: %v\n%v%v", mask, frontString, backString)
}

func (bs *BitmaskStruct) Print() {
	fmt.Printf(bs.ToString())
}

func (bl *BitmaskLocation) ToString() string {
	return fmt.Sprintf("Loc: %v %v %v %v\n",
		bl.Start, bl.End, bl.StemLen, bl.BitmaskIdx)
}

func (bl *BitmaskLocation) Print() {
	fmt.Printf(bl.ToString())
}

func New() *Stems {
	s := &Stems{}
	s.Loops = make(map[int][]int)

	s.Levels = make([]*StemLevel, 33)
	for i := 0; i < 33; i++ {
		s.Levels[i] = NewStemLevel()
	}

	return s
}

func (s *Stems) Print() {
	for idx, level := range s.Levels {
		fmt.Printf("Level %d\n", idx)
		level.Print()
	}
}

func (s *Stems) PrintLatest() {
	lastIdx := -1
	for idx, l := range s.Levels {
		if l.FilledOut {
			lastIdx = idx
		}
	}
	if lastIdx == -1 {
		slog.Warn("Cannot print latest if no fields filled out!")
	}
	fmt.Printf("Level: %v\n", lastIdx)
	s.Levels[lastIdx].Print()
}

func (sl *StemLevel) Print() {
	for idx, starts := range sl.Starts {
		fmt.Printf("  start: %v\n", idx)
		starts.Print()
	}
}

// one for each node in level (32, start of stem)
type StemStart struct {
	//int is end idx
	Ends map[int]*StemEnds
}

func (ss *StemStart) Print() {
	for eidx, ends := range ss.Ends {
		fmt.Printf("    ends: %v\n", eidx)
		ends.Print()
	}
}

// one for each stem end for each start
type StemEnds struct {
	Bitmasks []BitmaskStruct
}

func (se *StemEnds) Print() {
	for _, mask := range se.Bitmasks {
		mask.Print()
	}
}

func NewStemLevel() *StemLevel {
	return &StemLevel{
		Starts: make(map[int]*StemStart),
	}
}

func NewStemStart() *StemStart {
	return &StemStart{
		Ends: make(map[int]*StemEnds),
	}
}

func NewStemEnd() *StemEnds {
	return &StemEnds{}
}

func (s *Stems) Init(g *graph.Graph) {
	level := NewStemLevel()

	for _, node := range g.Nodes {
		start := node.ID

		level.Starts[start] = NewStemStart()

		for _, end := range node.Outgoing {

			bms := BitmaskStruct{
				Bitmask: IdToBitmask(end),
			}
			v, ok := level.Starts[start].Ends[end]
			if !ok {
				level.Starts[start].Ends[end] = NewStemEnd()
				v = level.Starts[start].Ends[end]
			}
			v.Add(bms)
		}
	}
	level.FilledOut = true
	s.Levels[2] = level
}

func (s *Stems) ProcessNextLevel(startLen, endLen int) {
	// logrus.WithFields(logrus.Fields{"start": startLen,
	// 	"end": endLen, "target": startLen + endLen - 1}).Info("Processing level")

	level := NewStemLevel()

	startLevel := s.Levels[startLen]
	endLevel := s.Levels[endLen]
	targetLevel := startLen + endLen - 1

	count := 0

	for startIdx, startObj := range startLevel.Starts {
		level.Starts[startIdx] = NewStemStart()

		//want to match 'midpoint' (end of start and start of end)
		for midIdx, midObj := range startObj.Ends {
			for endIdx, endObj := range endLevel.Starts[midIdx].Ends {
				newEnd := NewStemEnd()

				for firstMaskIdx, firstMaskStruct := range midObj.Bitmasks {
					for secondMaskIdx, secondMaskStruct := range endObj.Bitmasks {
						//bitmask does not include start

						if firstMaskStruct.Bitmask&secondMaskStruct.Bitmask > 0 {
							continue
						}

						combine := firstMaskStruct.Bitmask | secondMaskStruct.Bitmask

						//if start in path
						if combine&IdToBitmask(startIdx) > 0 {
							// fmt.Println("Start in path")
							if startIdx == endIdx {
								if len(s.Loops[startIdx]) > targetLevel {
									continue
								}

								FrontLoc := BitmaskLocation{
									Start:      startIdx,
									End:        midIdx,
									StemLen:    startLen,
									BitmaskIdx: firstMaskIdx,
								}
								EndLoc := BitmaskLocation{
									Start:      midIdx,
									End:        endIdx,
									StemLen:    endLen,
									BitmaskIdx: secondMaskIdx,
								}
								s.SetLoop(startIdx, FrontLoc, EndLoc)
							}
							continue
						}

						//they can be combined, start not a part of it

						bs := BitmaskStruct{
							Bitmask: combine,
							FrontLoc: BitmaskLocation{
								Start:      startIdx,
								End:        midIdx,
								StemLen:    startLen,
								BitmaskIdx: firstMaskIdx,
							},
							EndLoc: BitmaskLocation{
								Start:      midIdx,
								End:        endIdx,
								StemLen:    endLen,
								BitmaskIdx: secondMaskIdx,
							},
						}
						newEnd.Add(bs)
					}
				}
				if len(newEnd.Bitmasks) > 0 {
					if level.Starts[startIdx].Ends[endIdx] != nil {
						old := level.Starts[startIdx].Ends[endIdx].Bitmasks
						level.Starts[startIdx].Ends[endIdx].Bitmasks = append(old, newEnd.Bitmasks...)
					} else {
						level.Starts[startIdx].Ends[endIdx] = newEnd
					}
					count += len(newEnd.Bitmasks)
				}
			}
		}
	}
	level.FilledOut = true
	s.Levels[targetLevel] = level
	// logrus.WithFields(logrus.Fields{"count": count,
	// 	"target": startLen + endLen - 1}).Info("Level processed")
}

func (end *StemEnds) Add(bs BitmaskStruct) {
	for _, entry := range end.Bitmasks {
		if bs.Bitmask == entry.Bitmask {
			return
		}
	}
	end.Bitmasks = append(end.Bitmasks, bs)
}

func (s *Stems) SetLoop(startIdx int, front, back BitmaskLocation) {
	frontLoop := s.buildLoop(front)
	backLoop := s.buildLoop(back)
	loop := append(frontLoop, backLoop...)

	for _, elem := range loop {
		if len(s.Loops[elem]) < len(loop) {
			s.Loops[elem] = loop
		}
	}

}

func (s *Stems) buildLoop(Loc BitmaskLocation) []int {

	stack := []BitmaskLocation{Loc}
	var res []int

	for len(stack) > 0 {
		curr := stack[0]
		stack = stack[1:]

		if curr.StemLen == 2 {
			res = append(res, curr.Start)
			continue
		}

		//whew
		mask := s.Levels[curr.StemLen].Starts[curr.Start].Ends[curr.End].Bitmasks[curr.BitmaskIdx]
		stack = append(stack, mask.FrontLoc, mask.EndLoc)

	}
	return res

}

func IdToBitmask(id int) uint32 {
	if id > 32 || id < 1 {
		log.Fatal("invalid id")
		return 0
	}
	return 1 << (id - 1)
}
