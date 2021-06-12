package ai

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/imega/snake-game/snake"
	"github.com/imega/snake-game/state"
	"github.com/nsf/termbox-go"
)

type Result struct {
	Neuronet neuronet
	Score    int
}

func New(p state.Parameters, ch chan state.SnakeGame, pad chan snake.KeyboardEvent, statCh chan state.Stat) error {
	var (
		population    []Result
		instance      int
		epoch         int
		maxEpochScore int
		bestBrain     Result

		maxInstance = p.MaxInstance

		n = neuronet{}

		lastState state.SnakeGame
		lastKey   termbox.Key
	)

	brain, err := loadBrain(p)
	if err != nil {
		return fmt.Errorf("failed to load brain, %s", err)
	}
	bestBrain = brain

	for st := range ch {
		var key termbox.Key

		if st.Snake.Steps > p.MaxSnakeSteps {
			pad <- snake.KeyboardEvent{EventType: snake.RETRY}

			instance++

			continue
		}

		if st.IsOver {
			if st.Score > p.MinScoreEpoch {
				population = append(population, Result{
					Neuronet: n,
					Score:    st.Score,
				})
			}
			if instance < maxInstance {
				pad <- snake.KeyboardEvent{EventType: snake.RETRY}
				n = mutate(bestBrain.Neuronet, p.MutationRate, p.MutationRange)

				instance++
			}

			if instance >= maxInstance {
				epoch++

				var best Result
				for _, p := range population {
					if best.Score < p.Score {
						best = p
					}
				}

				population = nil
				instance = 0
				maxEpochScore = 0

				if best.Score > 0 {
					bestBrain.Neuronet = crossoverBrain(bestBrain.Neuronet, best.Neuronet)
				}

				if bestBrain.Score < best.Score {
					bestBrain.Score = best.Score
					if err := saveBrain(p, bestBrain); err != nil {
						return fmt.Errorf("failed to save brain, %w", err)
					}
				}
			}

			if !p.Silent {
				statCh <- state.Stat{
					Epoch:         epoch,
					Instance:      instance,
					BestScore:     bestBrain.Score,
					MaxEpochScore: maxEpochScore,
				}
			}

			continue
		}

		if maxEpochScore < st.Score {
			maxEpochScore = st.Score
			if p.Silent {
				fmt.Printf(
					"Snake Game MaxScore: %d, epoch: %d, epochMaxScore: %d, inst: %d\n",
					bestBrain.Score,
					epoch,
					maxEpochScore,
					instance,
				)
			}
		}

		if lastState.Snake.Head.X == st.Snake.Head.X && lastState.Snake.Head.Y == st.Snake.Head.Y {
			continue
		}

		in := createInput(st)
		out := n.predict(in)

		var direction int
		var m float64
		for i, v := range out {
			if i == 0 || v > m {
				direction = i
				m = v
			}
		}

		switch direction {
		case 0:
			key = termbox.KeyArrowRight
		case 1:
			key = termbox.KeyArrowLeft
		case 2:
			key = termbox.KeyArrowUp
		case 3:
			key = termbox.KeyArrowDown
		}

		if lastKey != key {
			pad <- snake.KeyboardEvent{EventType: snake.MOVE, Key: key}
		}

		lastKey = key
	}

	return nil
}

type neuronet struct {
	WeightHidden1 [24][18]float64
	BiasHidden1   [18]float64
	WeightHidden2 [18][18]float64
	BiasHidden2   [18]float64
	WeightOutput  [18][4]float64
	BiasOut       [4]float64
}

func (n *neuronet) randFill() {
	rand.Seed(time.Now().UnixNano())

	for i := range n.WeightHidden1 {
		for j := range n.WeightHidden1[i] {
			n.WeightHidden1[i][j] = randFloat(-1, 1)
		}
	}

	for i := range n.BiasHidden1 {
		n.BiasHidden1[i] = randFloat(-1, 1)
	}

	for i := range n.WeightHidden2 {
		for j := range n.WeightHidden2[i] {
			n.WeightHidden2[i][j] = randFloat(-1, 1)
		}
	}

	for i := range n.BiasHidden2 {
		n.BiasHidden2[i] = randFloat(-1, 1)
	}

	for i := range n.WeightOutput {
		for j := range n.WeightOutput[i] {
			n.WeightOutput[i][j] = randFloat(-1, 1)
		}
	}

	for i := range n.BiasOut {
		n.BiasOut[i] = randFloat(-1, 1)
	}
}

func (n *neuronet) predict(input [24]float64) []float64 {
	var t1 [18]float64
	for i := range input {
		for j := range n.WeightHidden1[i] {
			if i == 0 {
				t1[j] += n.BiasHidden1[j]
			}
			t1[j] += input[i] * n.WeightHidden1[i][j]
		}
	}

	var h1 [18]float64
	for i := range t1 {
		h1[i] = ReLU(t1[i])
	}

	// Layer 2
	var t2 [18]float64
	for i := range h1 {
		for j := range n.WeightHidden2[i] {
			if i == 0 {
				t2[j] += n.BiasHidden2[j]
			}
			t2[j] += h1[i] * n.WeightHidden2[i][j]
		}
	}

	var h2 [18]float64
	for i := range t2 {
		h2[i] = ReLU(t2[i])
	}

	// OUT
	var t0 [4]float64
	for i := range h2 {
		for j := range n.WeightOutput[i] {
			if i == 0 {
				t0[j] += n.BiasOut[j]
			}
			t0[j] += h1[i] * n.WeightOutput[i][j]
		}
	}

	return SoftMax(t0[:])
}

func randFloat(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())

	return min + rand.Float64()*(max-min)
}

func ReLU(x float64) float64 {
	const (
		Overflow  = 1.0239999999999999e+03
		Underflow = -1.0740e+03
		NearZero  = 1.0 / (1 << 28) // 2**-28
	)

	switch {
	case math.IsNaN(x) || math.IsInf(x, 1):
		return x
	case math.IsInf(x, -1):
		return 0
	case x > Overflow:
		return math.Inf(1)
	case x < Underflow:
		return 0
	case -NearZero < x && x < NearZero:
		return 1 + x
	}

	if x > 0 {
		return x
	} else {
		return 0
	}
}

func SoftMax(x []float64) []float64 {
	var max float64 = x[0]
	for _, n := range x {
		max = math.Max(max, n)
	}

	a := make([]float64, len(x))

	var sum float64 = 0
	for i, n := range x {
		a[i] -= math.Exp(n - max)
		sum += a[i]
	}

	for i, n := range a {
		a[i] = n / sum
	}
	return a
}

func createInput(st state.SnakeGame) [24]float64 {
	in := [24]float64{}

	src := distanceHead2Arena(st)
	copy(in[:], src[:])

	src = distanceHead2Food(st)
	copy(in[8:], src[:])

	src = distanceHead2Body(st)
	copy(in[16:], src[:])

	return in
}

func distanceHead2Arena(st state.SnakeGame) [8]float64 {
	n := st.Arena.Height - st.Snake.Head.Y
	e := st.Arena.Width - st.Snake.Head.X
	s := st.Snake.Head.Y
	w := st.Snake.Head.X

	return distanceCardinalDirection(n, e, s, w)
}

func distanceHead2Food(st state.SnakeGame) [8]float64 {
	var n, e, s, w int
	if st.Snake.Head.X > st.Food.X {
		w = st.Snake.Head.X - st.Food.X
	} else {
		e = st.Food.X - st.Snake.Head.X
	}

	if st.Snake.Head.Y > st.Food.Y {
		s = st.Snake.Head.Y - st.Food.Y
	} else {
		n = st.Food.Y - st.Snake.Head.Y
	}

	return distanceCardinalDirection(n, e, s, w)
}

func distanceHead2Body(st state.SnakeGame) [8]float64 {
	var res [8]float64
	body := st.Snake.Body
	head := st.Snake.Head
	for i := range body {
		if head.X == body[i].X && head.Y < body[i].Y { // N ↑
			res[0] = float64(1) / float64(head.Y-body[i].Y)
		}

		if head.Y == body[i].Y && head.X < body[i].X { // E →
			res[1] = float64(1) / float64(body[i].X-head.X)
		}

		if head.X == body[i].X && head.Y < body[i].Y { // S ↓
			res[2] = float64(1) / float64(head.Y-body[i].Y)
		}

		if head.Y == body[i].Y && head.X > body[i].X { // W ←
			res[3] = float64(1) / float64(head.X-body[i].X)
		}

		if head.Y-body[i].Y == head.X-body[i].X && head.Y-body[i].Y > 0 { // SW ↙︎
			res[4] = float64(1) / float64(head.Y-body[i].Y)
		}

		if head.Y-body[i].Y == body[i].X-head.X && head.Y-body[i].Y > 0 { // SE ↘︎
			res[5] = float64(1) / float64(head.Y-body[i].Y)
		}

		if body[i].Y-head.Y == head.X-body[i].X && body[i].Y-head.Y > 0 { // NW ↖︎
			res[6] = float64(1) / float64(body[i].Y-head.Y)
		}

		if body[i].Y-head.Y == body[i].X-head.X && body[i].Y-head.Y > 0 { // NE ↗︎
			res[7] = float64(1) / float64(body[i].Y-head.Y)
		}
	}

	return res
}

func distanceCardinalDirection(n, e, s, w int) [8]float64 {
	var res [8]float64

	if n > 0 { // N ↑
		res[0] = float64(1) / float64(n)
	}

	if e > 0 { // E →
		res[1] = float64(1) / float64(e)
	}

	if s > 0 { // S ↓
		res[2] = float64(1) / float64(s)
	}

	if w > 0 { // W ←
		res[3] = float64(1) / float64(w)
	}

	if s > 0 && w > 0 { // SW ↙︎
		res[4] = float64(1 / math.Sqrt(float64(s^2+w^2)))
	}

	if s > 0 && e > 0 { // SE ↘︎
		res[5] = float64(1 / math.Sqrt(float64(s^2+e^2)))
	}

	if n > 0 && w > 0 { // NW ↖︎
		res[6] = float64(1 / math.Sqrt(float64(n^2+w^2)))
	}

	if n > 0 && e > 0 { // NE ↗︎
		res[7] = float64(1 / math.Sqrt(float64(n^2+e^2)))
	}

	return res
}

func mutate(n neuronet, mutationRate, mutationRange float64) neuronet {
	for i := range n.WeightHidden1 {
		for j := range n.WeightHidden1[i] {
			n.WeightHidden1[i][j] = mutateNeurone(n.WeightHidden1[i][j], mutationRate, mutationRange)
		}
	}

	for i := range n.BiasHidden1 {
		n.BiasHidden1[i] = mutateNeurone(n.BiasHidden1[i], mutationRate, mutationRange)
	}

	for i := range n.WeightHidden2 {
		for j := range n.WeightHidden2[i] {
			n.WeightHidden2[i][j] = mutateNeurone(n.WeightHidden2[i][j], mutationRate, mutationRange)
		}
	}

	for i := range n.BiasHidden2 {
		n.BiasHidden2[i] = mutateNeurone(n.BiasHidden2[i], mutationRate, mutationRange)
	}

	for i := range n.WeightOutput {
		for j := range n.WeightOutput[i] {
			n.WeightOutput[i][j] = mutateNeurone(n.WeightOutput[i][j], mutationRate, mutationRange)
		}
	}

	for i := range n.BiasOut {
		n.BiasOut[i] = mutateNeurone(n.BiasOut[i], mutationRate, mutationRange)
	}

	return n
}

func mutateNeurone(weight, mutationRate, mutationRange float64) float64 {
	if randFloat(0, 1) <= mutationRate {
		weight += randFloat(mutationRange*-1, mutationRange)
	}

	return weight
}

func saveBrain(p state.Parameters, bestBrain Result) error {
	b, err := json.Marshal(bestBrain.Neuronet)
	if err != nil {
		return fmt.Errorf("failed to unmarshal, %s", err)
	}

	prefix := ""
	if p.PrefixFilename != "" {
		prefix = p.PrefixFilename + "-"
	}

	f, err := os.Create(prefix + "brain-" + strconv.Itoa(bestBrain.Score) + ".json")
	if err != nil {
		return fmt.Errorf("failed to save file, %s", err)
	}

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("failed to write to file, %s", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file, %s", err)
	}

	return nil
}

func loadBrain(p state.Parameters) (Result, error) {
	var n neuronet

	nameWithScore := fileNameWithoutExtension(p.BrainFilename)

	pos := strings.LastIndexByte(nameWithScore, '-')
	if pos == -1 {
		return Result{}, fmt.Errorf("failed to parse filename. Need format <prefix>brain-<score>.json")
	}

	score := nameWithScore[pos:]

	b, err := ioutil.ReadFile(p.BrainFilename)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read from file, %s", err)
	}

	if err := json.Unmarshal(b, &n); err != nil {
		return Result{}, fmt.Errorf("failed to marshal, %s", err)
	}

	s, err := strconv.Atoi(score)
	if err != nil {
		return Result{}, fmt.Errorf("failed to convert int from string, %s", err)
	}

	return Result{
		Neuronet: n,
		Score:    s,
	}, nil
}

func fileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}

	return fileName
}

func crossoverBrain(n1, n2 neuronet) neuronet {
	next := n1

	for i := range n2.WeightHidden1 {
		if randFloat(0, 1) <= 0.5 {
			next.WeightHidden1[i] = n2.WeightHidden1[i]
		}
	}

	for i := range n2.BiasHidden1 {
		if randFloat(0, 1) <= 0.5 {
			next.BiasHidden1[i] = n2.BiasHidden1[i]
		}
	}

	for i := range n2.WeightHidden2 {
		if randFloat(0, 1) <= 0.5 {
			next.WeightHidden2[i] = n2.WeightHidden2[i]
		}
	}

	for i := range n2.BiasHidden2 {
		if randFloat(0, 1) <= 0.5 {
			next.BiasHidden2[i] = n2.BiasHidden2[i]
		}
	}

	for i := range n2.WeightOutput {
		if randFloat(0, 1) <= 0.5 {
			next.WeightOutput[i] = n2.WeightOutput[i]
		}
	}

	for i := range n2.BiasOut {
		if randFloat(0, 1) <= 0.5 {
			next.BiasOut[i] = n2.BiasOut[i]
		}
	}

	return next
}

func CreateBrain(p state.Parameters) error {
	n := neuronet{}
	n.randFill()

	r := Result{Neuronet: n}

	if err := saveBrain(p, r); err != nil {
		return fmt.Errorf("failed to save brain, %w", err)
	}

	return nil
}
