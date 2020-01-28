module github.com/lyft/datacatalog

go 1.13

require (
	github.com/Selvatico/go-mocket v1.0.7
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/gorm v1.9.11
	github.com/lib/pq v1.2.0
	github.com/lyft/flyteidl v0.17.0
	github.com/lyft/flytestdlib v0.3.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	google.golang.org/grpc v1.26.0
)

// Pin the version of client-go to something that's compatible with katrogan's fork of api and apimachinery
// Type the following
//   replace k8s.io/client-go => k8s.io/client-go kubernetes-1.16.2
// and it will be replaced with the 'sha' variant of the version

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
