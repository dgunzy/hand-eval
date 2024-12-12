package deucelowsingle

import (
	"sort"

	"github.com/dgunzy/card/pkg/card"
)

const (
	// Penalties in ascending order of badness
	pairPenalty          = uint64(1000000)
	twoPairPenalty       = uint64(2000000)
	tripsPenalty         = uint64(3000000)
	straightPenalty      = uint64(4000000)
	flushPenalty         = uint64(5000000)
	fullHousePenalty     = uint64(6000000)
	quadsPenalty         = uint64(7000000)
	straightFlushPenalty = uint64(8000000)

	handSize = 5
)

type HandValue uint64

type HashTable struct {
	flushTable    []HandValue
	nonFlushTable []HandValue
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
		if remaining == 0 {
			index := encodeQuinary(ranks)
			value := calculateNonFlushValue(ranks)
			ht.nonFlushTable[index] = value
			return
		}

		if pos >= 13 {
			return
		}

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

// getRankBinary converts 5 cards of the same suit to 13-bit binary
func getRankBinary(cards []card.Card) uint16 {
	var binary uint16
	for _, c := range cards {
		binary |= 1 << c.Rank()
	}
	return binary
}

// getRankQuinary converts 5 cards to 13-bit quinary (base-5)
func getRankQuinary(cards []card.Card) uint32 {
	counts := make([]uint8, 13)
	for _, c := range cards {
		counts[c.Rank()]++
	}
	return encodeQuinary(counts)
}

// isSequential checks if a sorted slice of ranks forms a straight
func isSequential(ranks []int) bool {
	if len(ranks) < handSize {
		return false
	}
	// Since ranks are now sorted high to low, adjust comparison
	for i := 1; i < len(ranks); i++ {
		if ranks[i-1] != ranks[i]+1 {
			return false
		}
	}
	return true
}

// getHandPattern returns penalties for pairs/trips/etc and card ranks sorted high to low
func getHandPattern(ranks []uint8) (uint64, []int) {
	var penalty uint64
	rankCounts := make(map[int]int)
	rankList := make([]int, 0)

	// Count ranks and build rank list
	for i, count := range ranks {
		if count > 0 {
			rankCounts[i] = int(count)
			for j := uint8(0); j < count; j++ {
				rankList = append(rankList, i)
			}
		}
	}

	// Sort ranks high to low for 2-7 comparison
	sort.Sort(sort.Reverse(sort.IntSlice(rankList)))

	// Count pairs, trips, etc
	pairs := 0
	trips := 0
	quads := 0

	for _, count := range rankCounts {
		switch count {
		case 2:
			pairs++
		case 3:
			trips++
		case 4:
			quads++
		}
	}

	// Apply penalties in order
	if quads > 0 {
		penalty = quadsPenalty
	} else if trips > 0 && pairs > 0 {
		penalty = fullHousePenalty
	} else if trips > 0 {
		penalty = tripsPenalty
	} else if pairs == 2 {
		penalty = twoPairPenalty
	} else if pairs == 1 {
		penalty = pairPenalty
	}

	// Check for straight
	if isSequential(rankList) {
		if penalty == 0 {
			penalty = straightPenalty
		} else if straightPenalty > penalty {
			penalty = straightPenalty
		}
	}

	return penalty, rankList
}

func calculateNonFlushValue(ranks []uint8) HandValue {
	penalty, rankList := getHandPattern(ranks)

	// Calculate base value from card ranks
	var value uint64 = penalty
	multiplier := uint64(1)

	// Add values for each card, highest first
	for _, rank := range rankList {
		adjustedRank := rank
		if rank == 0 { // Ace
			adjustedRank = 13
		}
		value += uint64(adjustedRank) * multiplier
		multiplier *= 14 // Use 14 to ensure unique values
	}

	return HandValue(value)
}

func calculateFlushValue(binary uint16) HandValue {
	ranks := make([]uint8, 13)
	for i := uint(0); i < 13; i++ {
		if binary&(1<<i) != 0 {
			ranks[i] = 1
		}
	}

	_, rankList := getHandPattern(ranks)

	// For flushes, always add flush penalty
	penalty := flushPenalty

	// If it's also a straight, make it a straight flush
	if isSequential(rankList) {
		penalty = straightFlushPenalty
	}

	var value uint64 = penalty
	multiplier := uint64(1)

	// Add values for each card, highest first
	for _, rank := range rankList {
		adjustedRank := rank
		if rank == 0 { // Ace
			adjustedRank = 13
		}
		value += uint64(adjustedRank) * multiplier
		multiplier *= 14
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
