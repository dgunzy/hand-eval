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

// Improved formatting for error messages
func formatDetailedHand(h TestHand, value HandValue) string {
	return fmt.Sprintf("\n  Description: %s\n  Cards: %s\n  Value: %d",
		h.desc, formatCards(h.cards), value)
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

		// Complex pair comparisons
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Diamonds},
			{card.Four, card.Clubs}, {card.Five, card.Spades}}), "Low pair with low kickers (2-2-3-4-5)"},

		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.King, card.Diamonds},
			{card.Queen, card.Clubs}, {card.Jack, card.Spades}}), "Low pair with high kickers (2-2-K-Q-J)"},

		{makeHand([]cardSpec{{card.King, card.Spades}, {card.King, card.Hearts}, {card.Two, card.Diamonds},
			{card.Three, card.Clubs}, {card.Four, card.Spades}}), "High pair with low kickers (K-K-2-3-4)"},

		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.Ace, card.Hearts}, {card.Two, card.Diamonds},
			{card.Three, card.Clubs}, {card.Four, card.Spades}}), "Ace pair with low kickers (A-A-2-3-4)"},

		// Edge cases with pairs and high cards
		{makeHand([]cardSpec{{card.Three, card.Spades}, {card.Three, card.Hearts}, {card.Four, card.Diamonds},
			{card.Five, card.Clubs}, {card.Six, card.Spades}}), "Medium low pair (3-3-4-5-6)"},

		{makeHand([]cardSpec{{card.Three, card.Spades}, {card.Three, card.Hearts}, {card.Ace, card.Diamonds},
			{card.King, card.Clubs}, {card.Queen, card.Spades}}), "Medium pair high kickers (3-3-A-K-Q)"},

		// Two pair variations
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Diamonds},
			{card.Three, card.Clubs}, {card.Four, card.Spades}}), "Low two pair (2-2-3-3-4)"},

		{makeHand([]cardSpec{{card.King, card.Spades}, {card.King, card.Hearts}, {card.Queen, card.Diamonds},
			{card.Queen, card.Clubs}, {card.Two, card.Spades}}), "High two pair low kicker (K-K-Q-Q-2)"},

		// Three of a kind variations
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Two, card.Diamonds},
			{card.Three, card.Clubs}, {card.Four, card.Spades}}), "Low trips (2-2-2-3-4)"},

		{makeHand([]cardSpec{{card.King, card.Spades}, {card.King, card.Hearts}, {card.King, card.Diamonds},
			{card.Two, card.Clubs}, {card.Three, card.Spades}}), "High trips low kickers (K-K-K-2-3)"},

		// Edge cases mixing pairs with straights/flushes
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Spades},
			{card.Four, card.Spades}, {card.Five, card.Spades}}), "Low pair with flush draw (2-2-3-4-5 three spades)"},

		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Three, card.Hearts},
			{card.Four, card.Hearts}, {card.Five, card.Hearts}}), "Low pair with better flush (2-2-3-4-5 four hearts)"},

		// Mixed penalty hands
		{makeHand([]cardSpec{{card.Two, card.Spades}, {card.Two, card.Hearts}, {card.Two, card.Diamonds},
			{card.Three, card.Hearts}, {card.Three, card.Spades}}), "Full house low (2-2-2-3-3)"},

		{makeHand([]cardSpec{{card.Ace, card.Spades}, {card.Ace, card.Hearts}, {card.Ace, card.Diamonds},
			{card.King, card.Hearts}, {card.King, card.Spades}}), "Full house high (A-A-A-K-K)"},
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

	runExtendedTestCases(t, rankedHands)
}
func runExtendedTestCases(t *testing.T, rankedHands []rankedHand) {
	t.Run("Pair Comparisons", func(t *testing.T) {
		var lowPairLowKickers, lowPairHighKickers, highPairLowKickers rankedHand

		for _, rh := range rankedHands {
			switch rh.hand.desc {
			case "Low pair with low kickers (2-2-3-4-5)":
				lowPairLowKickers = rh
			case "Low pair with high kickers (2-2-K-Q-J)":
				lowPairHighKickers = rh
			case "High pair with low kickers (K-K-2-3-4)":
				highPairLowKickers = rh
			}
		}

		// A pair of 2s with high kickers should be better than a pair of 2s with low kickers
		if lowPairHighKickers.value >= lowPairLowKickers.value {
			t.Errorf("Hand comparison error - better hand rated worse:"+
				"\nExpected better hand:%s"+
				"\nRated worse than:%s",
				formatDetailedHand(lowPairHighKickers.hand, lowPairHighKickers.value),
				formatDetailedHand(lowPairLowKickers.hand, lowPairLowKickers.value))
		}

		// A pair of Kings with low kickers should be worse than any pair of 2s
		if highPairLowKickers.value <= lowPairHighKickers.value {
			t.Errorf("Hand comparison error - worse hand rated better:"+
				"\nExpected worse hand:%s"+
				"\nRated better than:%s",
				formatDetailedHand(highPairLowKickers.hand, highPairLowKickers.value),
				formatDetailedHand(lowPairHighKickers.hand, lowPairHighKickers.value))
		}
	})

	t.Run("Two Pair Comparisons", func(t *testing.T) {
		var lowTwoPair, highTwoPairLowKicker rankedHand

		for _, rh := range rankedHands {
			switch rh.hand.desc {
			case "Low two pair (2-2-3-3-4)":
				lowTwoPair = rh
			case "High two pair low kicker (K-K-Q-Q-2)":
				highTwoPairLowKicker = rh
			}
		}

		// Any high two pair should be worse than any low two pair
		if highTwoPairLowKicker.value <= lowTwoPair.value {
			t.Errorf("Hand comparison error - worse hand rated better:"+
				"\nExpected worse hand:%s"+
				"\nRated better than:%s",
				formatDetailedHand(highTwoPairLowKicker.hand, highTwoPairLowKicker.value),
				formatDetailedHand(lowTwoPair.hand, lowTwoPair.value))
		}
	})

	t.Run("Mixed Penalty Comparisons", func(t *testing.T) {
		var lowFullHouse, highFullHouse rankedHand

		for _, rh := range rankedHands {
			switch rh.hand.desc {
			case "Full house low (2-2-2-3-3)":
				lowFullHouse = rh
			case "Full house high (A-A-A-K-K)":
				highFullHouse = rh
			}
		}

		// Any high full house should be worse than any low full house
		if highFullHouse.value <= lowFullHouse.value {
			t.Errorf("Hand comparison error - worse hand rated better:"+
				"\nExpected worse hand:%s"+
				"\nRated better than:%s",
				formatDetailedHand(highFullHouse.hand, highFullHouse.value),
				formatDetailedHand(lowFullHouse.hand, lowFullHouse.value))
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
