package build_event_server

import (
	"context"
	"io"
	"log"

	_ "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"

	bpb "proto"
	"proto/build_event_stream"
)

type BuildEventProtocolServer struct {
}

func NewBuildEventProtocolServer() (*BuildEventProtocolServer, error) {
	return &BuildEventProtocolServer{}, nil
}

func (s *BuildEventProtocolServer) chompBuildEvent(obe *bpb.OrderedBuildEvent) error {
	switch buildEvent := obe.Event.Event.(type) {
	case *bpb.BuildEvent_ComponentStreamFinished:
		log.Print("BuildTool: ComponentStreamFinished: ", buildEvent.ComponentStreamFinished)
	case *bpb.BuildEvent_BazelEvent:
		var bazelBuildEvent build_event_stream.BuildEvent
		if err := ptypes.UnmarshalAny(buildEvent.BazelEvent, &bazelBuildEvent); err != nil {
			return err
		}
		switch bazelBuildEvent.Payload.(type) {
		default:
			log.Printf("Payload: %+v", bazelBuildEvent.Payload)
		}
	default:
		log.Printf("Unknown event: %+v", buildEvent)
	}
	return nil
}

func (s *BuildEventProtocolServer) PublishLifecycleEvent(ctx context.Context, req *bpb.PublishLifecycleEventRequest) (*empty.Empty, error) {
	// We don't currently handle these events.
	return &empty.Empty{}, nil
}

// Handles Streaming BuildToolEvent
// From the bazel client: (Read more in BuildEventServiceUploader.java.)
// {@link BuildEventServiceUploaderCommands#OPEN_STREAM} is the first event and opens a
// bidi streaming RPC for sending build events and receiving ACKs.
// {@link BuildEventServiceUploaderCommands#SEND_REGULAR_BUILD_EVENT} sends a build event to
// the server. Sending of the Nth build event does
// does not wait for the ACK of the N-1th build event to have been received.
// {@link BuildEventServiceUploaderCommands#SEND_LAST_BUILD_EVENT} sends the last build event
// and half closes the RPC.
// {@link BuildEventServiceUploaderCommands#ACK_RECEIVED} is executed for every ACK from
// the server and checks that the ACKs are in the correct order.
// {@link BuildEventServiceUploaderCommands#STREAM_COMPLETE} checks that all build events
// have been sent and all ACKs have been received. If not it invokes a retry logic that may
// decide to re-send every build event for which an ACK has not been received. If so, it
// adds an OPEN_STREAM event.
func (s *BuildEventProtocolServer) PublishBuildToolEventStream(stream bpb.PublishBuildEvent_PublishBuildToolEventStreamServer) error {
	// Semantically, the protocol requires we ack events in order.
	// If we get an out of order event, we just bail out.
	lastReceived := int64(-1)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := s.chompBuildEvent(in.OrderedBuildEvent); err != nil {
			return err
		}

		if lastReceived == -1 {
			lastReceived = in.OrderedBuildEvent.SequenceNumber
		} else if lastReceived+1 == in.OrderedBuildEvent.SequenceNumber {
			lastReceived = in.OrderedBuildEvent.SequenceNumber
		} else {
			log.Printf("Got an out-of-order build event (expected %d, got %d), bailing...", lastReceived+1, in.OrderedBuildEvent.SequenceNumber)
			return io.EOF
		}
		rsp := &bpb.PublishBuildToolEventStreamResponse{
			StreamId:       in.OrderedBuildEvent.StreamId,
			SequenceNumber: lastReceived,
		}
		stream.Send(rsp)

	}
	return nil
}