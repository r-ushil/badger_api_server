package server

import (
	"context"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"

	drill_submission_v1 "badger-api/gen/drill_submission/v1"
	"badger-api/gen/drill_submission/v1/drill_submissionv1connect"

	"badger-api/pkg/auth"
	"badger-api/pkg/common"
	"badger-api/pkg/drill"
	"badger-api/pkg/drill_submission"
	"badger-api/pkg/server"
)

type DrillSubmissionServer struct {
	ctx *server.ServerContext
}

func (s *DrillSubmissionServer) SubscribeToDrillSubmission(
	ctx context.Context,
	req *connect.Request[drill_submission_v1.SubscribeToDrillSubmissionRequest],
	stream *connect.ServerStream[drill_submission_v1.SubscribeToDrillSubmissionResponse],
) error {
	d, err := drill_submission.GetDrillSubmission(s.ctx, req.Msg.DrillSubmissionId)
	if err != nil {
		panic(err)
	}

	if d.ProcessingStatus != "Done" {
		drillId := common.WithDefault(d.DrillId, "Cover Drive")
		submissionId := d.SubmissionId
		userId := d.UserId

		score, advice1, advice2 := drill_submission.ProcessDrillSubmission(s.ctx, req.Msg.DrillSubmissionId, d.BucketUrl, drillId, userId, submissionId)
		res := &drill_submission_v1.SubscribeToDrillSubmissionResponse{
			DrillScore: score,
			Advice1:    advice1,
			Advice2:    advice2,
		}
		return stream.Send(res)
	} else {
		res := &drill_submission_v1.SubscribeToDrillSubmissionResponse{
			DrillScore: uint32(d.DrillScore),
			Advice1:    "",
			Advice2:    "",
		}
		return stream.Send(res)
	}
}

func (s *DrillSubmissionServer) GetDrillSubmission(
	ctx context.Context,
	req *connect.Request[drill_submission_v1.GetDrillSubmissionRequest],
) (*connect.Response[drill_submission_v1.GetDrillSubmissionResponse], error) {
	log.Println("Getting drill submission with ID:", req.Msg.DrillSubmissionId)

	d, err := drill_submission.GetDrillSubmission(s.ctx, req.Msg.DrillSubmissionId)

	if err != nil {
		// TODO: Handle properly
		log.Println(err)
		return nil, connect.NewError(connect.CodeUnimplemented, err)
	}
	timestampGoogleFormat := d.GetTimestampGoogleFormat()
	res := connect.NewResponse(&drill_submission_v1.GetDrillSubmissionResponse{
		DrillSubmission: &drill_submission_v1.DrillSubmission{
			DrillSubmissionId: d.GetId(),
			UserId:            d.GetUserId(),
			DrillId:           d.GetDrillId(),
			BucketUrl:         d.GetBucketUrl(),
			Timestamp:         &timestampGoogleFormat,
			ProcessingStatus:  d.GetProcessingStatus(),
			DrillScore:        d.GetDrillScore(),
		},
	})

	return res, nil
}

func (s *DrillSubmissionServer) GetUserScores(
	ctx context.Context,
	req *connect.Request[drill_submission_v1.GetUserScoresRequest],
) (*connect.Response[drill_submission_v1.GetUserScoresResponse], error) {
	log.Println("Getting user scores")

	coverDriveScore, katchetBoardScore := drill_submission.GetUserScores(s.ctx, req.Msg.UserId)

	res := connect.NewResponse(&drill_submission_v1.GetUserScoresResponse{
		CoverDriveScore:   coverDriveScore,
		KatchetBoardScore: katchetBoardScore,
	})

	return res, nil
}

func (s *DrillSubmissionServer) GetDrillSubmissions(
	ctx context.Context,
	req *connect.Request[drill_submission_v1.GetDrillSubmissionsRequest],
) (*connect.Response[drill_submission_v1.GetDrillSubmissionsResponse], error) {
	dss := drill_submission.GetDrillSubmissions(s.ctx)

	drill_submissions := make([]*drill_submission_v1.DrillSubmission, 0, len(dss))

	for _, d := range dss {
		timestampGoogleFormat := d.GetTimestampGoogleFormat()
		drill_submissions = append(drill_submissions, &drill_submission_v1.DrillSubmission{
			DrillSubmissionId: d.GetId(),
			UserId:            d.GetUserId(),
			DrillId:           d.GetDrillId(),
			BucketUrl:         d.GetBucketUrl(),
			Timestamp:         &timestampGoogleFormat,
			ProcessingStatus:  d.GetProcessingStatus(),
			DrillScore:        d.GetDrillScore(),
		})
	}

	res := connect.NewResponse(&drill_submission_v1.GetDrillSubmissionsResponse{
		DrillSubmissions: drill_submissions,
	})

	return res, nil
}

func InsertDrillSubmissionOfType(s *server.ServerContext, videoUrl string, userId string, drillName string) string {
	if drillName == "Katchet Board" {
		return drill.SubmitCatchingDrill(s, videoUrl, userId)
	} else {
		return drill.SubmitBattingDrill(s, videoUrl, userId)
	}
}

func (s *DrillSubmissionServer) InsertDrillSubmission(
	ctx context.Context,
	req *connect.Request[drill_submission_v1.InsertDrillSubmissionRequest],
) (*connect.Response[drill_submission_v1.InsertDrillSubmissionResponse], error) {
	authHeader := req.Header().Get("authorization")
	userId, err := auth.ParseAuthHeader(s.ctx, authHeader)

	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	submissionId := InsertDrillSubmissionOfType(s.ctx, req.Msg.DrillSubmission.BucketUrl, userId, req.Msg.DrillSubmission.DrillId)

	hex_id := drill_submission.InsertDrillSubmission(s.ctx, req.Msg.DrillSubmission, userId, submissionId)
	res := connect.NewResponse(&drill_submission_v1.InsertDrillSubmissionResponse{
		HexId: hex_id,
	})
	return res, nil
}

func (s *DrillSubmissionServer) GetUserDrillSubmissions(
	ctx context.Context,
	req *connect.Request[drill_submission_v1.GetUserDrillSubmissionsRequest],
) (*connect.Response[drill_submission_v1.GetUserDrillSubmissionsResponse], error) {
	dss := drill_submission.GetUserDrillSubmissions(s.ctx, req.Msg.UserId)

	drill_submissions := make([]*drill_submission_v1.DrillSubmission, 0, len(dss))

	for _, d := range dss {
		timestampGoogleFormat := d.GetTimestampGoogleFormat()
		drill_submissions = append(drill_submissions, &drill_submission_v1.DrillSubmission{
			DrillSubmissionId: d.GetId(),
			UserId:            d.GetUserId(),
			DrillId:           d.GetDrillId(),
			BucketUrl:         d.GetBucketUrl(),
			Timestamp:         &timestampGoogleFormat,
			ProcessingStatus:  d.GetProcessingStatus(),
			DrillScore:        d.GetDrillScore(),
		})
	}

	res := connect.NewResponse(&drill_submission_v1.GetUserDrillSubmissionsResponse{
		DrillSubmissions: drill_submissions,
	})

	return res, nil
}

func RegisterDrillSubmissionService(mux *http.ServeMux, ctx *server.ServerContext) {
	server := &DrillSubmissionServer{
		ctx,
	}

	path, handler := drill_submissionv1connect.NewDrillSubmissionServiceHandler(server)

	mux.Handle(path, handler)
}
