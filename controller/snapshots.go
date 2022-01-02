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
	rdsInstance, err := snapshotService.GetRDSInstanceDetails(rdsName)
	if err != nil {
		msg := fmt.Sprintf("RDS Instance with name:'%s'doesn't exist.", rdsName)
		log.Println(msg)
		_, _ = io.WriteString(res, msg)
		return
	}
	if *rdsInstance.DBInstanceStatus != awsClient.INSTANCE_STATE_AVAILABLE {
		msg := fmt.Sprintf("RDS Instance:'%s' is not in '%s' state.", rdsName, awsClient.INSTANCE_STATE_AVAILABLE)
		log.Println(msg)
		_, _ = io.WriteString(res, msg)
		return
	}
	awsClient.PublishNewResnapInstanceMessage(rdsInstance)
	io.WriteString(res, fmt.Sprintf("initiated resnap for instance: %s", rdsName))
}