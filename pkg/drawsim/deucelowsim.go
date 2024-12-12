package drawsim

import (
	"math/rand"
	"sort"

	"github.com/dgunzy/card/pkg/card"
	"github.com/dgunzy/hand-eval/pkg/handrank/variants/deucelowsingle"
)

// SimulationResult represents a single draw result
type SimulationResult struct {
	Hand       []card.Card
	HandValue  deucelowsingle.HandValue
	Percentile float64
}

type DrawSimulator struct {
	keptCards []card.Card
	deadCards []card.Card
	drawCount int
	handEval  *deucelowsingle.HashTable
	results   []SimulationResult
}

func NewSimulator(kept []card.Card, dead []card.Card, drawCount int) *DrawSimulator {
	return &DrawSimulator{
		keptCards: kept,
		deadCards: dead,
		drawCount: drawCount,
		handEval:  deucelowsingle.NewHashTable(),
	}
}

func (ds *DrawSimulator) RunSimulation(n int) []SimulationResult {
	ds.results = make([]SimulationResult, 0, n)
	usedCards := make(map[card.Card]bool)

	// Mark kept and dead cards as used
	for _, c := range ds.keptCards {
		usedCards[c] = true
	}
	for _, c := range ds.deadCards {
		usedCards[c] = true
	}

	// Create deck without used cards
	availableCards := make([]card.Card, 0, 52-len(usedCards))
	for i := 0; i < 52; i++ {
		c := card.Card(i)
		if !usedCards[c] {
			availableCards = append(availableCards, c)
		}
	}

	// Run simulations
	for i := 0; i < n; i++ {
		drawnHand := ds.simulateSingleDraw(availableCards)

		// Only include valid 5-card hands
		if len(drawnHand) == 5 {
			value := ds.handEval.Value(drawnHand)
			ds.results = append(ds.results, SimulationResult{
				Hand:      drawnHand,
				HandValue: value,
			})
		}
	}

	// Sort results purely by HandValue (lower is better in 2-7)
	sort.Slice(ds.results, func(i, j int) bool {
		return ds.results[i].HandValue < ds.results[j].HandValue
	})

	// Calculate percentiles after sorting
	totalHands := float64(len(ds.results))
	for i := range ds.results {
		ds.results[i].Percentile = (float64(i) + 1) / totalHands * 100
	}

	return ds.results
}

func (ds *DrawSimulator) simulateSingleDraw(availableCards []card.Card) []card.Card {
	// Create copy of available cards to shuffle
	drawPool := make([]card.Card, len(availableCards))
	copy(drawPool, availableCards)

	// Fisher-Yates shuffle
	for i := len(drawPool) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		drawPool[i], drawPool[j] = drawPool[j], drawPool[i]
	}

	// Combine kept cards with drawn cards
	result := make([]card.Card, len(ds.keptCards)+ds.drawCount)
	copy(result, ds.keptCards)
	copy(result[len(ds.keptCards):], drawPool[:ds.drawCount])

	return result
}

func (ds *DrawSimulator) GetResults() []SimulationResult {
	return ds.results
}
