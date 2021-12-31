package awsClient

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"log"
)

func GetNewRDSSnapshotService(rdsClient *rds.Client) *RDSSnapshotService {
	return &RDSSnapshotService{client: rdsClient}
}

type RDSSnapshotService struct {
	client *rds.Client
}

func (this *RDSSnapshotService) GetRDSInstanceDetails(instanceName string) (types.DBInstance, error) {

	describeDbInstanceInput := rds.DescribeDBInstancesInput{DBInstanceIdentifier: &instanceName}
	describeDbInstanceOut, err := this.client.DescribeDBInstances(context.TODO(), &describeDbInstanceInput)
	if err != nil {
		return types.DBInstance{}, fmt.Errorf("could not get details about instance: %s", instanceName)
	}
	if len(describeDbInstanceOut.DBInstances) == 0 {
		return types.DBInstance{}, fmt.Errorf("no matching instance was found by name: %s", instanceName)
	}

	return describeDbInstanceOut.DBInstances[0], nil
}

func (this *RDSSnapshotService) GetRDSSnapshotsByInstanceName(instanceName string) ([]types.DBSnapshot, error) {
	// Fetching Snapshots details for RDS
	describeInput := rds.DescribeDBSnapshotsInput{DBInstanceIdentifier: &instanceName}
	snapshots, err := this.client.DescribeDBSnapshots(context.TODO(), &describeInput)
	if err != nil {
		return nil, fmt.Errorf("could not describe snapshots for instance: %s", instanceName)
	}

	return snapshots.DBSnapshots, nil
}

func (this *RDSSnapshotService) GetLastRDSInstanceSnapshot(instanceName string) (types.DBSnapshot, error) {
	var lastSnapshot types.DBSnapshot
	rdsSnapshots, err := this.GetRDSSnapshotsByInstanceName(instanceName)
	if err != nil {
		log.Println(err)
		return lastSnapshot, err
	}
	if len(rdsSnapshots) == 0 {
		return lastSnapshot, fmt.Errorf("no snapshots exist for instance at this time")
	}
	for _, snapshot := range rdsSnapshots {
		if lastSnapshot.DBSnapshotArn == nil {
			lastSnapshot = snapshot
			continue
		}
		if snapshot.OriginalSnapshotCreateTime.After(*lastSnapshot.OriginalSnapshotCreateTime) {
			lastSnapshot = snapshot
		}
	}

	return lastSnapshot, nil
}

func (this *RDSSnapshotService) RestoreInstanceBySnapshot(
	instanceName string,
	snapshot types.DBSnapshot) (*types.DBInstance, error) {

	restoreDBInstanceIn := rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceIdentifier: &instanceName,
		DBSnapshotIdentifier: snapshot.DBSnapshotIdentifier,
	}
	restoreInstance, err := this.client.RestoreDBInstanceFromDBSnapshot(context.TODO(), &restoreDBInstanceIn)
	if err != nil {
		return nil, fmt.Errorf("could not restore instance due to error: %s", err)
	}

	return restoreInstance.DBInstance, nil
}

func (this *RDSSnapshotService) ApplySecurityGroupToInstance(
	instanceName string,
	securityGroups []string) (*types.DBInstance, error) {

	modifyInstanceInput := rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: &instanceName,
		DBSecurityGroups: securityGroups,
		ApplyImmediately: false,
	}
	modifyInstanceOut, err := this.client.ModifyDBInstance(context.TODO(), &modifyInstanceInput)
	if err != nil {
		return nil, fmt.Errorf("could not apply security group to instance due to error: %s", err)
	}

	return modifyInstanceOut.DBInstance, nil
}