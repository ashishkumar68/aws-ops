//+build wireinject

package awsClient

import (
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/google/wire"
)


func NewRDSClient() *rds.Client {
	return rds.NewFromConfig(NewConfig())
}

func NewRDSSnapshotService() *RDSSnapshotService {
	wire.Build(NewRDSClient, GetNewRDSSnapshotService)
	return &RDSSnapshotService{}
}