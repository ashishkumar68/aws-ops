package awsClient

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"log"
	"time"
)

var (
	INSTANCE_STATE_AVAILABLE = "available"
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

func (this *RDSSnapshotService) RenameInstance(
	currentInstance types.DBInstance,
	newInstanceName string) (*types.DBInstance, error) {
	modifyInstanceInput := rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: currentInstance.DBInstanceIdentifier,
		NewDBInstanceIdentifier: &newInstanceName,
		ApplyImmediately: true,
	}
	modifyInstanceOut, err := this.client.ModifyDBInstance(context.TODO(), &modifyInstanceInput)
	if err != nil {
		return nil, fmt.Errorf(
			"could not rename instance: %s due to error: %s",
			*currentInstance.DBInstanceIdentifier,
			err)
	}

	return modifyInstanceOut.DBInstance, nil
}

func (this *RDSSnapshotService) RunResnapMessageListener(rdsInstanceChan <-chan types.DBInstance) {
	log.Println("started resnap listener.")
	for rdsInstance := range rdsInstanceChan {
		oldInstanceName := *rdsInstance.DBInstanceIdentifier + "-old"
		oldInstance, err := this.RenameInstance(rdsInstance, oldInstanceName)
		if err != nil {
			log.Println("could not rename instance for resnap process, exiting..")
			continue
		}
		oldInstanceState := *oldInstance.DBInstanceStatus
		// Wait until the re-name changes get applied to the old instance.
		for oldInstanceState != INSTANCE_STATE_AVAILABLE {
			time.Sleep(30 * time.Second)
			instanceDetails, err := this.GetRDSInstanceDetails(*oldInstance.DBInstanceIdentifier)
			if err != nil {
				log.Println(fmt.Sprintf("could not fetch instance details:%s", *oldInstance.DBInstanceIdentifier))
				continue
			}
			oldInstanceState = *instanceDetails.DBInstanceStatus
			log.Println(fmt.Sprintf("Found renamed old instance state: %s", oldInstanceState))
		}
		// then run new instance launch
		log.Println(fmt.Sprintf("launching new instance: %s", *rdsInstance.DBInstanceIdentifier))
		resnapInstance, err := this.RunResnapForInstance(rdsInstance)
		if err != nil {
			log.Println(fmt.Sprintf("could not start resnap for instance: %s", *rdsInstance.DBInstanceIdentifier))
			continue
		}
		log.Println(fmt.Sprintf("sucessfully started launch for new instance: %s", *resnapInstance.DBInstanceIdentifier))
		log.Println(fmt.Sprintf("deleting old instance: %s", *oldInstance.DBInstanceIdentifier))
		deletedInstance, err := this.DeleteInstance(*oldInstance, false, false)
		if err != nil {
			log.Println(fmt.Sprintf("could not start delete for instance: %s", *oldInstance.DBInstanceIdentifier))
			continue
		}
		log.Println(fmt.Sprintf("Started delete for old instance: %s", *deletedInstance.DBInstanceIdentifier))
	}
}

func (this *RDSSnapshotService) RunResnapForInstance(resnapInstance types.DBInstance) (*types.DBInstance, error) {
	prodRdsName := "test-prod"
	prodInstance, err := this.GetRDSInstanceDetails(prodRdsName)
	if err != nil {
		return nil, fmt.Errorf("RDS instance with name:'%s'doesn't exist", prodRdsName)
	}
	lastSnapshot, err := this.GetLastRDSInstanceSnapshot(*prodInstance.DBInstanceIdentifier)
	if err != nil {
		return nil, fmt.Errorf("could not fetch last snapshot due to error: %s", err)
	}
	dbInstance, err := this.RestoreInstanceBySnapshot(*resnapInstance.DBInstanceIdentifier, lastSnapshot)
	if err != nil {
		return nil, fmt.Errorf("could not restore instance from snapshot due to error: %s", err)
	}
	dbInstance, err = this.ApplySecurityGroupToInstance(*dbInstance.DBInstanceIdentifier, []string{"default"})
	if err != nil {
		return nil, fmt.Errorf("could not apply security group to instance due to error: %s", err)
	}

	return dbInstance, err
}

func (this *RDSSnapshotService) DeleteInstance(
	instance types.DBInstance,
	deleteAutomatedBackups bool,
	skipFinalSnapshot bool) (*types.DBInstance, error) {
	deleteDBInstanceIn := rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: instance.DBInstanceIdentifier,
		DeleteAutomatedBackups: &deleteAutomatedBackups,
		SkipFinalSnapshot: skipFinalSnapshot,
	}
	deleteDBInstanceOut, err := this.client.DeleteDBInstance(context.TODO(), &deleteDBInstanceIn)
	if err != nil {
		return nil, fmt.Errorf(
			"could not delete rds instance: %s due to error: %s",
			*instance.DBInstanceIdentifier,
			err)
	}

	return deleteDBInstanceOut.DBInstance, nil
}