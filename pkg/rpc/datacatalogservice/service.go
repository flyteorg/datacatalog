package datacatalogservice

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/lyft/datacatalog/pkg/manager/impl"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"
	"github.com/lyft/datacatalog/pkg/repositories/config"
	"github.com/lyft/datacatalog/pkg/runtime"
	catalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/profutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
)

type serviceMetrics struct {
	createDatasetResponseTime  labeled.StopWatch
	getDatasetResponseTime     labeled.StopWatch
	createArtifactResponseTime labeled.StopWatch
	getArtifactResponseTime    labeled.StopWatch
	addTagResponseTime         labeled.StopWatch
}

type DataCatalogService struct {
	DatasetManager  interfaces.DatasetManager
	ArtifactManager interfaces.ArtifactManager
	TagManager      interfaces.TagManager
	serviceMetrics  serviceMetrics
}

func (s *DataCatalogService) CreateDataset(ctx context.Context, request *catalog.CreateDatasetRequest) (*catalog.CreateDatasetResponse, error) {
	timer := s.serviceMetrics.createDatasetResponseTime.Start(ctx)
	defer timer.Stop()
	return s.DatasetManager.CreateDataset(ctx, *request)
}

func (s *DataCatalogService) CreateArtifact(ctx context.Context, request *catalog.CreateArtifactRequest) (*catalog.CreateArtifactResponse, error) {
	timer := s.serviceMetrics.createArtifactResponseTime.Start(ctx)
	defer timer.Stop()
	return s.ArtifactManager.CreateArtifact(ctx, *request)
}

func (s *DataCatalogService) GetDataset(ctx context.Context, request *catalog.GetDatasetRequest) (*catalog.GetDatasetResponse, error) {
	timer := s.serviceMetrics.getDatasetResponseTime.Start(ctx)
	defer timer.Stop()
	return s.DatasetManager.GetDataset(ctx, *request)
}

func (s *DataCatalogService) GetArtifact(ctx context.Context, request *catalog.GetArtifactRequest) (*catalog.GetArtifactResponse, error) {
	timer := s.serviceMetrics.getArtifactResponseTime.Start(ctx)
	defer timer.Stop()
	return s.ArtifactManager.GetArtifact(ctx, *request)
}

func (s *DataCatalogService) AddTag(ctx context.Context, request *catalog.AddTagRequest) (*catalog.AddTagResponse, error) {
	timer := s.serviceMetrics.addTagResponseTime.Start(ctx)
	defer timer.Stop()
	return s.TagManager.AddTag(ctx, *request)
}

func NewDataCatalogService() *DataCatalogService {
	dataCatalogName := "datacatalog"
	catalogScope := promutils.NewScope(dataCatalogName).NewSubScope("service")
	ctx := contextutils.WithAppName(context.Background(), dataCatalogName)

	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)

	defer func() {
		if err := recover(); err != nil {
			catalogScope.MustNewCounter("initialization_panic",
				"panics encountered initializating the datacatalog service").Inc()
			logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	storeConfig := storage.GetConfig()
	dataStorageClient, err := storage.NewDataStore(storeConfig, catalogScope.NewSubScope("storage"))
	if err != nil {
		logger.Errorf(ctx, "Failed to create DataStore %v, err %v", storeConfig, err)
		panic(err)
	}
	logger.Infof(ctx, "Created data storage.")

	configProvider := runtime.NewConfigurationProvider()
	baseStorageReference := dataStorageClient.GetBaseContainerFQN(ctx)
	dataCatalogConfig := configProvider.ApplicationConfiguration().GetDataCatalogConfig()
	storagePrefix, err := dataStorageClient.ConstructReference(ctx, baseStorageReference, dataCatalogConfig.StoragePrefix)
	if err != nil {
		logger.Errorf(ctx, "Failed to create prefix %v, err %v", dataCatalogConfig.StoragePrefix, err)
		panic(err)
	}

	dbConfigValues := configProvider.ApplicationConfiguration().GetDbConfig()
	dbConfig := config.DbConfig{
		Host:         dbConfigValues.Host,
		Port:         dbConfigValues.Port,
		DbName:       dbConfigValues.DbName,
		User:         dbConfigValues.User,
		Password:     dbConfigValues.Password,
		ExtraOptions: dbConfigValues.ExtraOptions,
	}
	repos := repositories.GetRepository(repositories.POSTGRES, dbConfig, catalogScope)
	logger.Infof(ctx, "Created DB connection.")

	// Serve profiling endpoint.
	go func() {
		err := profutils.StartProfilingServerWithDefaultHandlers(
			context.Background(), dataCatalogConfig.ProfilerPort, nil)
		if err != nil {
			logger.Panicf(context.Background(), "Failed to Start profiling and Metrics server. Error, %v", err)
		}
	}()

	return &DataCatalogService{
		DatasetManager:  impl.NewDatasetManager(repos, dataStorageClient, catalogScope.NewSubScope("dataset")),
		ArtifactManager: impl.NewArtifactManager(repos, dataStorageClient, storagePrefix, catalogScope.NewSubScope("artifact")),
		TagManager:      impl.NewTagManager(repos, dataStorageClient, catalogScope.NewSubScope("tag")),
		serviceMetrics: serviceMetrics{
			createDatasetResponseTime:  labeled.NewStopWatch("create_dataset_duration", "The duration of the create artifact calls.", time.Millisecond, catalogScope, labeled.EmitUnlabeledMetric),
			getDatasetResponseTime:     labeled.NewStopWatch("get_dataset_duration", "The duration of the get artifact calls.", time.Millisecond, catalogScope, labeled.EmitUnlabeledMetric),
			createArtifactResponseTime: labeled.NewStopWatch("create_artifact_duration", "The duration of the get artifact calls.", time.Millisecond, catalogScope, labeled.EmitUnlabeledMetric),
			getArtifactResponseTime:    labeled.NewStopWatch("get_artifact_duration", "The duration of the get artifact calls.", time.Millisecond, catalogScope, labeled.EmitUnlabeledMetric),
			addTagResponseTime:         labeled.NewStopWatch("add_tag_duration", "The duration of the get artifact calls.", time.Millisecond, catalogScope, labeled.EmitUnlabeledMetric),
		},
	}
}