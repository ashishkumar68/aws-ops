package controller

import (
	"fmt"
	"github.com/ashishkumar68/aws-ops/awsClient"
	"io"
	"log"
	"net/http"
)

func ResnapRDSByName(res http.ResponseWriter, req *http.Request) {
	rdsName := req.URL.Query().Get("rdsName")
	if rdsName == "" {
		_,_ = io.WriteString(res, "empty RDS name was found.")
		return
	}
	snapshotService := awsClient.NewRDSSnapshotService()
	prodRdsName := "test-prod"
	currentRdsInstance, err := snapshotService.GetRDSInstanceDetails(prodRdsName)
	if err != nil {
		msg := fmt.Sprintf("RDS Instance with name:'%s'doesn't exist.", prodRdsName)
		log.Println(msg)
		_, _ = io.WriteString(res, msg)
		return
	}
	lastSnapshot, err := snapshotService.GetLastRDSInstanceSnapshot(*currentRdsInstance.DBInstanceIdentifier)
	if err != nil {
		log.Println(err)
		_,_ =io.WriteString(res, fmt.Sprintf("could not fetch last snapshot due to error: %s", err))
		return
	}
	dbInstance, err := snapshotService.RestoreInstanceBySnapshot(rdsName, lastSnapshot)
	if err != nil {
		log.Println(err)
		_,_ =io.WriteString(res, fmt.Sprintf("could not restore instance from snapshot due to error: %s", err))
		return
	}
	dbInstance, err = snapshotService.ApplySecurityGroupToInstance(*dbInstance.DBInstanceIdentifier, []string{"default"})
	if err != nil {
		log.Println(err)
		_,_ = io.WriteString(res, fmt.Sprintf("could not apply security group to instance due to error: %s", err))
		return
	}
	_,_ =io.WriteString(res, fmt.Sprintf("started rds:%s resnap", *dbInstance.DBInstanceIdentifier))
}