type Direction int

const (
	North Direction = iota
	South
	East
	West
	Up
	Down
	Out
	In
)

// Helper for JSON marshaling/unmarshaling
func (d Direction) String() string {
	return [...]string{"north", "south", "east", "west", "up", "down", "out", "in"}[d]
}

func DirectionFromString(s string) Direction {
	// mapping logic here
}