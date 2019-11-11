package common

import (
	"github.com/lyft/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"
)

type Entity int

const (
	Artifact Entity = iota
	Dataset
	Partition
	Tag
)

type ComparisonOperator int

const (
	Equal ComparisonOperator = iota
	// Add more operators as needed, ie., gte, lte
)

func GetUnsupportedFilterExpressionErr(operator ComparisonOperator) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, "unsupported filter expression operator: %s",
		operator)
}

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
