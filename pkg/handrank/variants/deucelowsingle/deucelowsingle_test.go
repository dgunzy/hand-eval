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

	hands := []TestHand{
		// Smooth hands (no penalties)
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Five, card.Clubs}, {card.Seven, card.Spades}}), "Perfect (2-3-4-5-7)"},

		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Five, card.Clubs}, {card.Eight, card.Spades}}), "Near perfect (2-3-4-5-8)"},

		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Six, card.Clubs}, {card.Seven, card.Spades}}), "Strong (2-3-4-6-7)"},

		{makeHand([]cardSpec{{card.Three, card.Spades}, {card.Five, card.Hearts}, {card.Seven, card.Diamonds},
			{card.Nine, card.Clubs}, {card.Jack, card.Spades}}), "Medium (3-5-7-9-J)"},

		// High card hands with Aces
		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.Three, card.Hearts}, {card.Seven, card.Diamonds},
			{card.Nine, card.Clubs}, {card.King, card.Spades}}), "Ace high (A-3-7-9-K)"},

		{makeHand([]cardSpec{{card.King, card.Spades}, {card.Three, card.Hearts}, {card.Seven, card.Diamonds},
			{card.Nine, card.Clubs}, {card.Jack, card.Spades}}), "King high (K-3-7-9-J)"},

		// Pairs
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Diamonds},
			{card.Four, card.Clubs}, {card.Five, card.Spades}}), "Low pair (2-2-3-4-5)"},

		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.Ace, card.Hearts}, {card.King, card.Diamonds},
			{card.Queen, card.Clubs}, {card.Jack, card.Spades}}), "Aces pair (A-A-K-Q-J)"},

		// Three of a kind
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Two, card.Diamonds},
			{card.Three, card.Clubs}, {card.Four, card.Spades}}), "Set of twos (2-2-2-3-4)"},

		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.Ace, card.Hearts}, {card.Ace, card.Diamonds},
			{card.King, card.Clubs}, {card.Queen, card.Spades}}), "Set of aces (A-A-A-K-Q)"},

		// Straights (worse than three of a kind)
		{makeHand([]cardSpec{{card.Three, card.Spades}, {card.Four, card.Hearts}, {card.Five, card.Diamonds},
			{card.Six, card.Clubs}, {card.Seven, card.Spades}}), "High straight (3-4-5-6-7)"},

		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Diamonds},
			{card.Four, card.Clubs}, {card.Five, card.Spades}}), "Wheel straight (A-2-3-4-5)"},

		// Flushes (worse than straights)
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Four, card.Spades}, {card.Six, card.Spades},
			{card.Eight, card.Spades}, {card.Ten, card.Spades}}), "Low flush (2-4-6-8-T spades)"},

		{makeHand([]cardSpec{{card.Ace, card.Hearts}, {card.King, card.Hearts}, {card.Queen, card.Hearts},
			{card.Jack, card.Hearts}, {card.Ten, card.Hearts}}), "High flush (A-K-Q-J-T hearts)"},

		// Straight flushes (worst)
		{makeHand([]cardSpec{{card.Two, card.Hearts}, {card.Three, card.Hearts}, {card.Four, card.Hearts},
			{card.Five, card.Hearts}, {card.Six, card.Hearts}}), "Low straight flush (2-3-4-5-6 hearts)"},

		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.King, card.Spades}, {card.Queen, card.Spades},
			{card.Jack, card.Spades}, {card.Ten, card.Spades}}), "Broadway straight flush (A-K-Q-J-T spades)"},
	}

	rankedHands := make([]rankedHand, len(hands))
	for i, h := range hands {
		rankedHands[i] = rankedHand{h, ht.Value(h.cards)}
	}

	sort.Slice(rankedHands, func(i, j int) bool {
		return rankedHands[i].value < rankedHands[j].value
	})

	fmt.Printf("\nHand Rankings (Best to Worst):\n")
	fmt.Printf("==============================\n")
	for i, rh := range rankedHands {
		fmt.Printf("%2d. %-50s Value: %d\n", i+1, rh.hand, rh.value)
	}

	runTestCases(t, rankedHands)
}

func runTestCases(t *testing.T, rankedHands []rankedHand) {
	t.Run("Hand Category Order", func(t *testing.T) {
		var (
			smoothHandValue    HandValue
			pairValue          HandValue
			tripsValue         HandValue
			straightValue      HandValue
			flushValue         HandValue
			straightFlushValue HandValue
		)

		for _, rh := range rankedHands {
			switch rh.hand.desc {
			case "Perfect (2-3-4-5-7)":
				smoothHandValue = rh.value
			case "Low pair (2-2-3-4-5)":
				pairValue = rh.value
			case "Set of twos (2-2-2-3-4)":
				tripsValue = rh.value
			case "High straight (3-4-5-6-7)":
				straightValue = rh.value
			case "Low flush (2-4-6-8-T spades)":
				flushValue = rh.value
			case "Low straight flush (2-3-4-5-6 hearts)":
				straightFlushValue = rh.value
			}
		}

		// Test category ordering
		if pairValue <= smoothHandValue {
			t.Errorf("Pair (%d) should be worse than smooth hand (%d)", pairValue, smoothHandValue)
		}
		if tripsValue <= pairValue {
			t.Errorf("Trips (%d) should be worse than pair (%d)", tripsValue, pairValue)
		}
		if straightValue <= tripsValue {
			t.Errorf("Straight (%d) should be worse than trips (%d)", straightValue, tripsValue)
		}
		if flushValue <= straightValue {
			t.Errorf("Flush (%d) should be worse than straight (%d)", flushValue, straightValue)
		}
		if straightFlushValue <= flushValue {
			t.Errorf("Straight flush (%d) should be worse than flush (%d)", straightFlushValue, flushValue)
		}
	})

	t.Run("Ace High Card Rankings", func(t *testing.T) {
		var aceHighValue, kingHighValue HandValue

		for _, rh := range rankedHands {
			switch rh.hand.desc {
			case "Ace high (A-3-7-9-K)":
				aceHighValue = rh.value
			case "King high (K-3-7-9-J)":
				kingHighValue = rh.value
			}
		}

		if aceHighValue <= kingHighValue {
			t.Errorf("Ace high (%d) should be worse than King high (%d)", aceHighValue, kingHighValue)
		}
	})

	t.Run("Pair Rankings", func(t *testing.T) {
		var lowPairValue, acePairValue HandValue

		for _, rh := range rankedHands {
			switch rh.hand.desc {
			case "Low pair (2-2-3-4-5)":
				lowPairValue = rh.value
			case "Aces pair (A-A-K-Q-J)":
				acePairValue = rh.value
			}
		}

		if acePairValue <= lowPairValue {
			t.Errorf("Pair of Aces (%d) should be worse than low pair (%d)", acePairValue, lowPairValue)
		}
	})

	t.Run("Perfect Low Hand", func(t *testing.T) {
		bestHand := rankedHands[0].hand
		if !containsRanks(bestHand.cards, []card.Rank{card.Two, card.Three, card.Four, card.Five, card.Seven}) {
			t.Errorf("Best hand should be 2-3-4-5-7, got %v", bestHand)
		}
	})
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
