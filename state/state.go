package state

type SnakeGame struct {
	Arena  Arena
	Snake  Snake
	Food   Coord
	IsOver bool
	Score  int
}

type Arena struct {
	Width  int
	Height int
}

type Snake struct {
	Head  Coord
	Body  []Coord
	Steps int
}

type Coord struct {
	X int
	Y int
}

type Stat struct {
	Epoch         int
	Instance      int
	BestScore     int
	MaxEpochScore int
	Err           string
}

type Parameters struct {
	Speed          int
	MaxInstance    int
	MutationRate   float64
	MutationRange  float64
	MaxSnakeSteps  int
	MinScoreEpoch  int
	PrefixFilename string
	Silent         bool
	Human          bool
	BrainFilename  string
	CreateBrain    bool
}
