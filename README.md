# Datacatalog
Data Catalog is a service for indexing parameterized, strongly-typed data artifacts across revisions. It is a stateless
GO GRPC service. DataCatalog utilizes the [GORM ORM library](https://github.com/jinzhu/gorm) to abstract it's backend
relational database and offloads the data contents associated with artifacts to a separate store that supports s3/azure/gcs (configured through storage config).

## Flyte usage
Flyte is integrated with DataCatalog to index some of it's executions onto DataCatalog.
More information can be found here: https://lyft.github.io/flyte/user/features/task_cache.html

## Development
`make compile` - compiles the service and produces runnable executable
`make install` - installs the necessary dependencies for DataCatalog
`make generate_idl` - generates the Protobuf IDL code
`make lint` - runs GO linter
`make test_unit` - runs the unit tests
