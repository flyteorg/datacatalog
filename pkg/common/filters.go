package common

// Common constants and types for Filtering
const (
	DefaultPageOffset = uint32(0)
	MaxPageLimit      = uint32(50)
)

// Common Entity types that can be used on any filters
type Entity string

const (
	Artifact  Entity = "Artifact"
	Dataset   Entity = "Dataset"
	Partition Entity = "Partition"
	Tag       Entity = "Tag"
)

// Supported operators that can be used on filters
type ComparisonOperator int

const (
	Equal ComparisonOperator = iota
	IsNull
	// Add more operators as needed, ie., gte, lte
)
