package deucelowsingle

import (
	"fmt"
	"sort"
	"testing"

	"github.com/dgunzy/card/pkg/card"
)

// TestHand is a helper struct for testing
type TestHand struct {
	cards []card.Card
	desc  string
}
type rankedHand struct {
	hand  TestHand
	value HandValue
}

func (h TestHand) String() string {
	return fmt.Sprintf("%s: %v", h.desc, formatCards(h.cards))
}

// formatCards returns a readable string of cards
func formatCards(cards []card.Card) string {
	result := ""
	for i, c := range cards {
		if i > 0 {
			result += " "
		}
		result += c.String()
	}
	return result
}

func TestHandRankings(t *testing.T) {
	ht := NewHashTable()

	// Create test hands
	hands := []TestHand{
		// Best possible hand
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Five, card.Clubs}, {card.Seven, card.Spades}}), "Best hand (2-3-4-5-7)"},

		// Near-best hands
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Five, card.Clubs}, {card.Eight, card.Spades}}), "Almost best (2-3-4-5-8)"},
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Six, card.Clubs}, {card.Seven, card.Spades}}), "Near best (2-3-4-6-7)"},

		// Medium hands
		{makeHand([]cardSpec{{card.Three, card.Spades}, {card.Five, card.Hearts}, {card.Seven, card.Diamonds},
			{card.Nine, card.Clubs}, {card.Jack, card.Spades}}), "Medium (3-5-7-9-J)"},
		{makeHand([]cardSpec{{card.Four, card.Spades}, {card.Six, card.Hearts}, {card.Eight, card.Diamonds},
			{card.Ten, card.Clubs}, {card.Queen, card.Spades}}), "Medium (4-6-8-T-Q)"},

		// Pairs
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Diamonds},
			{card.Four, card.Clubs}, {card.Five, card.Spades}}), "One pair (2-2-3-4-5)"},
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Two, card.Diamonds},
			{card.Three, card.Clubs}, {card.Four, card.Spades}}), "Three of a kind (2-2-2-3-4)"},

		// Straights
		{makeHand([]cardSpec{{card.Two, card.Hearts}, {card.Three, card.Hearts}, {card.Four, card.Hearts},
			{card.Five, card.Hearts}, {card.Six, card.Hearts}}), "Straight flush (2-3-4-5-6)"},
		{makeHand([]cardSpec{{card.Three, card.Spades}, {card.Four, card.Hearts}, {card.Five, card.Diamonds},
			{card.Six, card.Clubs}, {card.Seven, card.Spades}}), "Straight (3-4-5-6-7)"},

		// Flushes
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Four, card.Spades}, {card.Six, card.Spades},
			{card.Eight, card.Spades}, {card.Ten, card.Spades}}), "Flush (2-4-6-8-T)"},

		// Bad hands
		{makeHand([]cardSpec{{card.Eight, card.Spades}, {card.Nine, card.Hearts}, {card.Ten, card.Diamonds},
			{card.Jack, card.Clubs}, {card.King, card.Spades}}), "High cards (8-9-T-J-K)"},
		{makeHand([]cardSpec{{card.Ten, card.Spades}, {card.Jack, card.Hearts}, {card.Queen, card.Diamonds},
			{card.King, card.Clubs}, {card.Ace, card.Spades}}), "Worst non-penalty (T-J-Q-K-A)"},
	}

	rankedHands := make([]rankedHand, len(hands))
	for i, h := range hands {
		rankedHands[i] = rankedHand{h, ht.Value(h.cards)}
	}

	sort.Slice(rankedHands, func(i, j int) bool {
		return rankedHands[i].value < rankedHands[j].value
	})

	// Print sorted hands and verify rankings
	fmt.Printf("\nHand Rankings (Best to Worst):\n")
	fmt.Printf("==============================\n")
	for i, rh := range rankedHands {
		fmt.Printf("%2d. %-50s Value: %d\n", i+1, rh.hand, rh.value)
	}

	// Test cases...
	runTestCases(t, rankedHands)
}

type cardSpec struct {
	rank card.Rank
	suit card.Suit
}

func makeHand(specs []cardSpec) []card.Card {
	cards := make([]card.Card, len(specs))
	for i, spec := range specs {
		cards[i] = card.NewCard(spec.suit, spec.rank)
	}
	return cards
}

func runTestCases(t *testing.T, rankedHands []rankedHand) {
	// Helper function to find value of first hand matching a predicate
	findHandValue := func(predicate func([]card.Card) bool) HandValue {
		for _, rh := range rankedHands {
			if predicate(rh.hand.cards) {
				return rh.value
			}
		}
		return HandValue(^uint64(0))
	}

	// Helper for single card hands (no pairs/straights/flushes)
	isSingleCards := func(cards []card.Card) bool {
		return !hasPair(cards) && !hasStraight(cards) && !hasFlush(cards)
	}

	// Helper for pairs (exactly one pair)
	isOnePair := func(cards []card.Card) bool {
		counts := make(map[card.Rank]int)
		for _, c := range cards {
			counts[c.Rank()]++
		}
		pairs := 0
		for _, count := range counts {
			if count == 2 {
				pairs++
			} else if count > 2 {
				return false
			}
		}
		return pairs == 1
	}

	// Helper for three of a kind
	isTrips := func(cards []card.Card) bool {
		counts := make(map[card.Rank]int)
		for _, c := range cards {
			counts[c.Rank()]++
			if counts[c.Rank()] == 3 {
				return true
			}
		}
		return false
	}

	t.Run("Best Hand Should Be 2-3-4-5-7", func(t *testing.T) {
		bestHand := rankedHands[0].hand
		if !containsRanks(bestHand.cards, []card.Rank{card.Two, card.Three, card.Four, card.Five, card.Seven}) {
			t.Errorf("Expected best hand to be 2-3-4-5-7, got %v", bestHand)
		}
	})

	t.Run("Verify Basic Hand Category Rankings", func(t *testing.T) {
		singleCardsValue := findHandValue(isSingleCards)
		pairValue := findHandValue(isOnePair)
		tripsValue := findHandValue(isTrips)
		straightValue := findHandValue(func(cards []card.Card) bool {
			return hasStraight(cards) && !hasFlush(cards)
		})
		flushValue := findHandValue(func(cards []card.Card) bool {
			return hasFlush(cards) && !hasStraight(cards)
		})
		straightFlushValue := findHandValue(func(cards []card.Card) bool {
			return hasFlush(cards) && hasStraight(cards)
		})

		// Verify ordering
		if pairValue <= singleCardsValue {
			t.Error("Pair should be worse than single cards")
		}
		if tripsValue <= pairValue {
			t.Error("Trips should be worse than pair")
		}
		if straightValue <= tripsValue {
			t.Error("Straight should be worse than trips")
		}
		if flushValue <= straightValue {
			t.Error("Flush should be worse than straight")
		}
		if straightFlushValue <= flushValue {
			t.Error("Straight flush should be worse than flush")
		}
	})

	t.Run("Verify Single Card Hand Rankings", func(t *testing.T) {
		var best, middle, worst HandValue
		for _, rh := range rankedHands {
			if isSingleCards(rh.hand.cards) {
				// Find hands by description
				if rh.hand.desc == "Best hand (2-3-4-5-7)" {
					best = rh.value
				} else if rh.hand.desc == "Medium (3-5-7-9-J)" {
					middle = rh.value
				} else if rh.hand.desc == "Worst non-penalty (T-J-Q-K-A)" {
					worst = rh.value
				}
			}
		}

		if best >= middle {
			t.Error("2-3-4-5-7 should be better than 3-5-7-9-J")
		}
		if middle >= worst {
			t.Error("3-5-7-9-J should be better than T-J-Q-K-A")
		}
	})

	t.Run("Verify Pair Rankings", func(t *testing.T) {
		var pairValue, tripsValue HandValue
		for _, rh := range rankedHands {
			if rh.hand.desc == "One pair (2-2-3-4-5)" {
				pairValue = rh.value
			} else if rh.hand.desc == "Three of a kind (2-2-2-3-4)" {
				tripsValue = rh.value
			}
		}

		if tripsValue <= pairValue {
			t.Error("Three of a kind should be worse than one pair")
		}
	})

	t.Run("Verify Straight vs Flush Rankings", func(t *testing.T) {
		straightValue := findHandValue(func(cards []card.Card) bool {
			return hasStraight(cards) && !hasFlush(cards)
		})
		flushValue := findHandValue(func(cards []card.Card) bool {
			return hasFlush(cards) && !hasStraight(cards)
		})
		straightFlushValue := findHandValue(func(cards []card.Card) bool {
			return hasFlush(cards) && hasStraight(cards)
		})

		if straightFlushValue <= flushValue || straightFlushValue <= straightValue {
			t.Error("Straight flush should be worse than both straight and flush")
		}
	})
}

func containsRanks(cards []card.Card, ranks []card.Rank) bool {
	if len(cards) != len(ranks) {
		return false
	}
	cardRanks := make(map[card.Rank]bool)
	for _, c := range cards {
		cardRanks[c.Rank()] = true
	}
	for _, r := range ranks {
		if !cardRanks[r] {
			return false
		}
	}
	return true
}

// Helper functions remain the same
func hasPair(cards []card.Card) bool {
	counts := make(map[card.Rank]int)
	for _, c := range cards {
		counts[c.Rank()]++
		if counts[c.Rank()] > 1 {
			return true
		}
	}
	return false
}

func hasFlush(cards []card.Card) bool {
	if len(cards) == 0 {
		return false
	}
	suit := cards[0].Suit()
	for _, c := range cards[1:] {
		if c.Suit() != suit {
			return false
		}
	}
	return true
}

func hasStraight(cards []card.Card) bool {
	if len(cards) < 5 {
		return false
	}
	ranks := make([]int, len(cards))
	for i, c := range cards {
		ranks[i] = int(c.Rank())
	}
	sort.Ints(ranks)
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
			return false
		}
	}
	return true
}
