package operation

import (
	"github.com/buildbuddy-io/buildbuddy/server/remote_cache/digest"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc/codes"

	repb "github.com/buildbuddy-io/buildbuddy/proto/remote_execution"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	gstatus "google.golang.org/grpc/status"
)

func Assemble(stage repb.ExecutionStage_Value, d *digest.InstanceNameDigest, status error, r *repb.ActionResult) (string, *longrunning.Operation, error) {
	name := digest.DownloadResourceName(d.Digest, d.GetInstanceName())
	metadata, err := ptypes.MarshalAny(&repb.ExecuteOperationMetadata{
		Stage:        stage,
		ActionDigest: d.Digest,
	})
	if err != nil {
		return name, nil, err
	}
	operation := &longrunning.Operation{
		Name:     name,
		Metadata: metadata,
	}
	if r != nil {
		c := gstatus.Code(status)
		er := &repb.ExecuteResponse{
			Status: &statuspb.Status{Code: int32(c)},
		}
		if c == codes.OK {
			er.Result = r
		} else {
			er.Message = status.Error()
		}
		result, err := ptypes.MarshalAny(er)
		if err != nil {
			return name, nil, err
		}
		operation.Result = &longrunning.Operation_Response{Response: result}
	}
	if stage == repb.ExecutionStage_COMPLETED {
		operation.Done = true
	}
	return name, operation, nil
}

func AssembleFailed(stage repb.ExecutionStage_Value, d *digest.InstanceNameDigest, status error) (*longrunning.Operation, error) {
	emptyActionResult := &repb.ActionResult{}
	_, op, err := Assemble(stage, d, status, emptyActionResult)
	return op, err
}
