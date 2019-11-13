package common

// Common Entity types that can be used on any filters
type Entity int

const (
	Artifact Entity = iota
	Dataset
	Partition
	Tag
)

func (entity Entity) Name() string {
	switch entity {
	case Artifact:
		return "Artifact"
	case Dataset:
		return "Dataset"
	case Partition:
		return "Partition"
	case Tag:
		return "Tag"
	}

	return ""
}

// Supported operators that can be used on filters
type ComparisonOperator int

const (
	Equal ComparisonOperator = iota
	// Add more operators as needed, ie., gte, lte
)
