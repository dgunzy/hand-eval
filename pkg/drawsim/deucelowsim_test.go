package drawsim

import (
	"fmt"
	"testing"

	"github.com/dgunzy/card/pkg/card"
)

const (
	pairPenalty          = uint64(1000000)
	twoPairPenalty       = uint64(2000000)
	tripsPenalty         = uint64(3000000)
	straightPenalty      = uint64(4000000)
	flushPenalty         = uint64(5000000)
	fullHousePenalty     = uint64(6000000)
	quadsPenalty         = uint64(7000000)
	straightFlushPenalty = uint64(8000000)
)

func TestDrawSimulator(t *testing.T) {
	t.Run("Basic Single Card Draw", func(t *testing.T) {
		kept := []card.Card{
			card.NewCard(card.Spades, card.Eight),
			card.NewCard(card.Hearts, card.Seven),
			card.NewCard(card.Diamonds, card.Six),
			card.NewCard(card.Clubs, card.Three),
		}
		dead := []card.Card{
			card.NewCard(card.Spades, card.King),
		}
		sim := NewSimulator(kept, dead, 1)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 1: 8763(draw1) Distribution ===\n")
		printDetailedDistribution(results)

		// Verify that best hand has a lower value than worst hand
		if results[0].HandValue >= results[len(results)-1].HandValue {
			t.Error("Best hand should have lower value than worst hand")
		}
	})

	t.Run("Wheel Draw", func(t *testing.T) {
		kept := []card.Card{
			card.NewCard(card.Spades, card.Two),
			card.NewCard(card.Hearts, card.Three),
			card.NewCard(card.Diamonds, card.Four),
			card.NewCard(card.Clubs, card.Five),
		}
		dead := []card.Card{
			card.NewCard(card.Spades, card.Seven),
		}
		sim := NewSimulator(kept, dead, 1)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 2: 2345(draw1) Distribution ===\n")
		fmt.Printf("Looking for seven draws, one seven blocked\n")
		printDetailedDistribution(results)

		// Verify straight penalties
		for _, result := range results {
			if isSequential(ranks(result.Hand)) && uint64(result.HandValue) < straightPenalty {
				t.Errorf("Found wheel straight without straight penalty: %v (Hash: %d)",
					formatHand(result.Hand), result.HandValue)
			}
		}
	})

	t.Run("Flush Draw Test", func(t *testing.T) {
		kept := []card.Card{
			card.NewCard(card.Spades, card.Three),
			card.NewCard(card.Spades, card.Four),
			card.NewCard(card.Spades, card.Six),
			card.NewCard(card.Spades, card.Eight),
		}
		dead := []card.Card{}
		sim := NewSimulator(kept, dead, 1)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 4: 3468s(draw1) Distribution ===\n")
		fmt.Printf("Testing flush draws - all cards spades\n")
		printDetailedDistribution(results)

		// Verify flush penalties
		for _, result := range results {
			spadeCount := 0
			for _, c := range result.Hand {
				if c.Suit() == card.Spades {
					spadeCount++
				}
			}
			if spadeCount == 5 && uint64(result.HandValue) < flushPenalty {
				t.Errorf("Found flush without flush penalty: %v (Hash: %d)",
					formatHand(result.Hand), result.HandValue)
			}
		}
	})

	t.Run("Straight Draw Test", func(t *testing.T) {
		kept := []card.Card{
			card.NewCard(card.Spades, card.Four),
			card.NewCard(card.Hearts, card.Five),
			card.NewCard(card.Diamonds, card.Six),
			card.NewCard(card.Clubs, card.Seven),
		}
		dead := []card.Card{}
		sim := NewSimulator(kept, dead, 1)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 5: 4567(draw1) Distribution ===\n")
		fmt.Printf("Testing straight draws\n")
		printDetailedDistribution(results)

		// Verify straight penalties
		for _, result := range results {
			if isSequential(ranks(result.Hand)) && uint64(result.HandValue) < straightPenalty {
				t.Errorf("Found straight without straight penalty: %v (Hash: %d)",
					formatHand(result.Hand), result.HandValue)
			}
		}
	})

	t.Run("Pairs Test", func(t *testing.T) {
		kept := []card.Card{
			card.NewCard(card.Spades, card.Three),
			card.NewCard(card.Hearts, card.Three),
			card.NewCard(card.Diamonds, card.Five),
		}
		dead := []card.Card{}
		sim := NewSimulator(kept, dead, 2)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 6: 335(draw2) Distribution ===\n")
		fmt.Printf("Testing pair evaluations\n")
		printDetailedDistribution(results)

		// Verify pair penalties
		for _, result := range results {
			if hasPair(ranks(result.Hand)) && uint64(result.HandValue) < pairPenalty {
				t.Errorf("Found pair without pair penalty: %v (Hash: %d)",
					formatHand(result.Hand), result.HandValue)
			}
		}
	})
}

// Helper functions

func ranks(hand []card.Card) []int {
	ranks := make([]int, len(hand))
	for i, c := range hand {
		ranks[i] = int(c.Rank())
	}
	return ranks
}

func hasPair(ranks []int) bool {
	count := make(map[int]int)
	for _, r := range ranks {
		count[r]++
		if count[r] > 1 {
			return true
		}
	}
	return false
}

func isSequential(ranks []int) bool {
	if len(ranks) < 5 {
		return false
	}
	sorted := make([]int, len(ranks))
	copy(sorted, ranks)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for i := 1; i < len(sorted); i++ {
		if sorted[i] != sorted[i-1]+1 {
			return false
		}
	}
	return true
}

// Existing helper functions remain the same
func printDetailedDistribution(results []SimulationResult) {
	if len(results) == 0 {
		return
	}

	// Header
	fmt.Printf("\nHand Rankings (Best to Worst)\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%-3s  %-25s  %-12s  %-10s\n", "Rank", "Hand", "Value", "Percentile")
	fmt.Printf("────────────────────────────────────────────────────\n")

	// Print each hand with ranking
	for i, result := range results {
		fmt.Printf("%-3d  %-25s  %-12d  %6.1f%%\n",
			i+1,
			formatHand(result.Hand),
			result.HandValue,
			result.Percentile)
	}

	// Summary statistics with clear separation
	fmt.Printf("\nSummary Statistics\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Best Hand:       %-25s  Value: %-d\n",
		formatHand(results[0].Hand),
		results[0].HandValue)
	fmt.Printf("Worst Hand:      %-25s  Value: %-d\n",
		formatHand(results[len(results)-1].Hand),
		results[len(results)-1].HandValue)

	// Percentile markers
	fmt.Printf("\nPercentile Distribution\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━\n")
	markers := []int{25, 50, 75}
	for _, p := range markers {
		idx := (p * len(results)) / 100
		if idx >= len(results) {
			idx = len(results) - 1
		}
		fmt.Printf("%-3d%%:           %-25s  Value: %-d\n",
			p,
			formatHand(results[idx].Hand),
			results[idx].HandValue)
	}
	fmt.Println()
}

func formatHand(hand []card.Card) string {
	result := ""
	for i, c := range hand {
		if i > 0 {
			result += " "
		}
		result += c.String()
	}
	return result
}
