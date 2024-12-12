package deucelowsingle

import (
	"sort"

	"github.com/dgunzy/card/pkg/card"
)

const (
	// Size constants for hash tables

	flushPenalty    = uint64(10000000)
	straightPenalty = uint64(5000000)

	// Number of cards in a hand
	handSize = 5
)

// HandValue represents the strength of a hand (lower is better in 2-7)
type HandValue uint64 // Changed to uint64 for larger range

// HashTable contains pre-computed values for all possible hands
type HashTable struct {
	flushTable    []HandValue // 13-bit binary -> value
	nonFlushTable []HandValue // 13-bit quinary -> value
}

// Initialize the lookup tables
func NewHashTable() *HashTable {
	ht := &HashTable{
		flushTable:    make([]HandValue, 8192),  // 2^13 possible flush combinations
		nonFlushTable: make([]HandValue, 49205), // All possible 5-card combinations
	}

	// Initialize all values to maximum (worst possible hand)
	for i := range ht.flushTable {
		ht.flushTable[i] = HandValue(^uint64(0))
	}
	for i := range ht.nonFlushTable {
		ht.nonFlushTable[i] = HandValue(^uint64(0))
	}

	ht.initializeFlushTable()
	ht.initializeNonFlushTable()
	return ht
}

// getRankBinary converts 5 cards of the same suit to 13-bit binary
func getRankBinary(cards []card.Card) uint16 {
	var binary uint16
	for _, c := range cards {
		binary |= 1 << c.Rank() // Simplified uint cast
	}
	return binary
}

// getRankQuinary converts 5 cards to 13-bit quinary (base-5)
// Each position can have 0-4 cards, requiring base-5 representation
func getRankQuinary(cards []card.Card) uint32 {
	counts := make([]uint8, 13)
	for _, c := range cards {
		counts[c.Rank()]++
	}
	return encodeQuinary(counts)
}

// Value returns the pre-computed value for a hand
func (ht *HashTable) Value(cards []card.Card) HandValue {
	if len(cards) != handSize {
		return HandValue(^uint64(0)) // Return max value for invalid hands
	}

	// Fast flush check using XOR
	suit := cards[0].Suit()
	isFlush := true
	for _, c := range cards[1:] {
		if c.Suit() != suit {
			isFlush = false
			break
		}
	}

	if isFlush {
		return ht.flushTable[getRankBinary(cards)]
	}
	return ht.nonFlushTable[getRankQuinary(cards)]
}

func (ht *HashTable) initializeFlushTable() {
	// Initialize all values to maximum
	for i := range ht.flushTable {
		ht.flushTable[i] = HandValue(^uint64(0))
	}

	var generateFlushCombinations func(pos, count int, binary uint16)
	generateFlushCombinations = func(pos, count int, binary uint16) {
		if count == handSize {
			value := calculateFlushValue(binary)
			ht.flushTable[binary] = value
			return
		}

		if pos >= 13 || count+(13-pos) < handSize {
			return
		}

		generateFlushCombinations(pos+1, count, binary)
		generateFlushCombinations(pos+1, count+1, binary|(1<<uint(pos)))
	}

	generateFlushCombinations(0, 0, 0)
}

func (ht *HashTable) initializeNonFlushTable() {
	var generateCombinations func(pos int, remaining int, ranks []uint8)
	generateCombinations = func(pos int, remaining int, ranks []uint8) {
		// If we've placed all cards, evaluate the hand
		if remaining == 0 {
			index := encodeQuinary(ranks)
			value := calculateNonFlushValue(ranks)
			ht.nonFlushTable[index] = value
			return
		}

		// If we can't place all remaining cards, return
		if pos >= 13 {
			return
		}

		// Try placing 0 to min(4, remaining) cards at current position
		maxCards := min(4, remaining)
		for count := uint8(0); count <= uint8(maxCards); count++ {
			ranks[pos] = count
			generateCombinations(pos+1, remaining-int(count), ranks)
			ranks[pos] = 0
		}
	}

	ranks := make([]uint8, 13)
	generateCombinations(0, 5, ranks)
}

// isSequential checks if a sorted slice of ranks forms a straight
// isSequential checks if a sorted slice of ranks forms a straight
func isSequential(ranks []int) bool {
	if len(ranks) < handSize {
		return false
	}
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
			return false
		}
	}
	return true
}

// calculateFlushValue returns a value for a flush hand (higher value = worse hand)
func calculateFlushValue(binary uint16) HandValue {
	ranks := make([]int, 0, handSize)
	for i := uint(0); i < 13; i++ {
		if binary&(1<<i) != 0 {
			ranks = append(ranks, int(i))
		}
	}

	// Sort ranks in ascending order
	sort.Ints(ranks)

	// Calculate base value
	var value uint64
	multiplier := uint64(1)

	// Convert ranks so Ace is worst (12) and Two is best (1)
	for _, rank := range ranks {
		adjustedRank := (rank+11)%13 + 1
		value += uint64(adjustedRank) * multiplier
		multiplier *= 13
	}

	// Add penalties
	value += flushPenalty
	if isSequential(ranks) {
		value += straightPenalty
	}

	return HandValue(value)
}

// calculateNonFlushValue returns a value for a non-flush hand
func calculateNonFlushValue(ranks []uint8) HandValue {
	cardRanks := make([]int, 0, handSize)
	for i, count := range ranks {
		for j := uint8(0); j < count; j++ {
			cardRanks = append(cardRanks, i)
		}
	}

	// Sort ranks in ascending order (2 is best)
	sort.Ints(cardRanks)

	// Calculate base value - lower cards are better
	var value uint64
	multiplier := uint64(1)

	// First, heavily penalize pairs/trips/etc
	pairPenalty := uint64(1000000)
	for i := 0; i < len(cardRanks)-1; i++ {
		if cardRanks[i] == cardRanks[i+1] {
			value += pairPenalty
			pairPenalty *= 2 // Double penalty for each additional matching card
		}
	}

	// Then add card values (Ace=12 is worst, Two=1 is best)
	for _, rank := range cardRanks {
		// Convert rank so Ace is worst (12) and Two is best (1)
		adjustedRank := (rank+11)%13 + 1
		value += uint64(adjustedRank) * multiplier
		multiplier *= 13
	}

	// Add straight penalty if needed
	if isSequential(cardRanks) {
		value += straightPenalty
	}

	return HandValue(value)
}

// encodeQuinary converts rank counts to a quinary number
func encodeQuinary(ranks []uint8) uint32 {
	// We need to convert the ranks into a unique index within our table size
	// The pattern here is that we're mapping all possible 5-card combinations
	// where each position (rank) can have 0-4 cards

	// Count total cards to validate
	total := uint8(0)
	for _, count := range ranks {
		total += count
	}
	if total != 5 {
		return 0 // Invalid hand
	}

	// Calculate unique index using combinatorial number system
	var index uint32
	remainingCards := uint32(5)
	multiplier := uint32(1)

	for i := 0; i < 13 && remainingCards > 0; i++ {
		count := uint32(ranks[i])
		if count > 0 {
			// For each position, we multiply by combinations of remaining positions
			for j := uint32(0); j < count; j++ {
				index += multiplier * uint32(i)
				multiplier *= 13
			}
			remainingCards -= count
		}
	}

	return index % 49205 // Ensure we stay within table bounds
}

// Helper function for min value
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
