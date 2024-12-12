package drawsim

import (
	"fmt"
	"testing"

	"github.com/dgunzy/card/pkg/card"
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
	})

	t.Run("Wheel Draw", func(t *testing.T) {
		// Drawing one to a 2345
		kept := []card.Card{
			card.NewCard(card.Spades, card.Two),
			card.NewCard(card.Hearts, card.Three),
			card.NewCard(card.Diamonds, card.Four),
			card.NewCard(card.Clubs, card.Five),
		}
		dead := []card.Card{
			card.NewCard(card.Spades, card.Seven), // Block one seven
		}
		sim := NewSimulator(kept, dead, 1)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 2: 2345(draw1) Distribution ===\n")
		fmt.Printf("Looking for seven draws, one seven blocked\n")
		printDetailedDistribution(results)
	})

	t.Run("Drawing Three Cards", func(t *testing.T) {
		// Drawing three to a 32
		kept := []card.Card{
			card.NewCard(card.Spades, card.Three),
			card.NewCard(card.Hearts, card.Two),
		}
		dead := []card.Card{
			card.NewCard(card.Spades, card.Four),
			card.NewCard(card.Hearts, card.Five),
		}
		sim := NewSimulator(kept, dead, 3)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 3: 32(draw3) Distribution ===\n")
		fmt.Printf("Drawing three with 4♠ 5♥ dead\n")
		printDetailedDistribution(results)
	})

	t.Run("Flush Draw Test", func(t *testing.T) {
		// Drawing one with possible flush
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
			// Count spades
			spadeCount := 0
			for _, c := range result.Hand {
				if c.Suit() == card.Spades {
					spadeCount++
				}
			}
			if spadeCount == 5 && result.HandValue < 200000 { // flushPenalty
				t.Errorf("Found flush without flush penalty: %v (Hash: %d)",
					formatHand(result.Hand), result.HandValue)
			}
		}
	})

	t.Run("Straight Draw Test", func(t *testing.T) {
		// Drawing one with possible straight
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
			ranks := make([]int, 5)
			for i, c := range result.Hand {
				ranks[i] = int(c.Rank())
			}
			if isSequential(ranks) && result.HandValue < 5000000 { // straightPenalty
				t.Errorf("Found straight without straight penalty: %v (Hash: %d)",
					formatHand(result.Hand), result.HandValue)
			}
		}
	})

	t.Run("Many Dead Cards", func(t *testing.T) {
		// Drawing two with many dead cards
		kept := []card.Card{
			card.NewCard(card.Spades, card.Three),
			card.NewCard(card.Hearts, card.Four),
			card.NewCard(card.Diamonds, card.Five),
		}
		dead := []card.Card{
			card.NewCard(card.Spades, card.Two),
			card.NewCard(card.Hearts, card.Two),
			card.NewCard(card.Diamonds, card.Two),
			card.NewCard(card.Clubs, card.Two), // All twos dead
			card.NewCard(card.Spades, card.Six),
			card.NewCard(card.Hearts, card.Six),
			card.NewCard(card.Diamonds, card.Six), // Most sixes dead
		}
		sim := NewSimulator(kept, dead, 2)
		results := sim.RunSimulation(20)

		fmt.Printf("\n=== Test 6: 345(draw2) Distribution ===\n")
		fmt.Printf("Drawing with all 2s and most 6s dead\n")
		printDetailedDistribution(results)
	})
}

// Helper function to check if ranks form a sequential straight
func isSequential(ranks []int) bool {
	// First sort the ranks
	for i := 0; i < len(ranks)-1; i++ {
		for j := i + 1; j < len(ranks); j++ {
			if ranks[i] > ranks[j] {
				ranks[i], ranks[j] = ranks[j], ranks[i]
			}
		}
	}

	// Check if they're sequential
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
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

	fmt.Printf("\nHand Rankings (Best to Worst):\n")
	fmt.Printf("===============================\n")
	fmt.Printf("%-25s %-15s %-10s\n", "Hand", "Hash Value", "Percentile")
	fmt.Printf("-----------------------------------------------\n")

	for _, result := range results {
		fmt.Printf("%-25s %d\t%.1f%%\n",
			formatHand(result.Hand),
			result.HandValue,
			result.Percentile)
	}

	fmt.Printf("\nKey Statistics:\n")
	fmt.Printf("Best hand:  %v (Hash: %d)\n", formatHand(results[0].Hand), results[0].HandValue)
	fmt.Printf("Worst hand: %v (Hash: %d)\n",
		formatHand(results[len(results)-1].Hand),
		results[len(results)-1].HandValue)

	markers := []int{25, 50, 75}
	for _, p := range markers {
		idx := (p * len(results)) / 100
		if idx >= len(results) {
			idx = len(results) - 1
		}
		fmt.Printf("%dth percentile: %v (Hash: %d)\n",
			p,
			formatHand(results[idx].Hand),
			results[idx].HandValue)
	}
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
