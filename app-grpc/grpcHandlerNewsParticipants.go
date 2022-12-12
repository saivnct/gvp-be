package appgrpc

import (
	"context"
	"errors"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/static"
	"gbb.go/gvp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func (sv *XVPGRPCService) GetNewsParticipants(ctx context.Context, req *grpcXVPPb.GetNewsParticipantsRequest) (*grpcXVPPb.GetNewsParticipantsResponse, error) {
	_, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("GetNewsParticipants - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	page := req.GetPage()
	pageSize := req.GetPageSize()

	if page < 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = static.Pagination_Default_PageSize
	}
	if pageSize > static.Pagination_Max_PageSize {
		pageSize = static.Pagination_Max_PageSize
	}

	filter := primitive.M{}
	totalItem, err := dao.GetNewsParticipantDAO().CountDocuments(ctx, filter)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	sort := primitive.M{"newsCount": -1}
	listParticipants, err := dao.GetNewsParticipantDAO().FetchListParticipants(ctx, page, pageSize, filter, sort)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	participants := []string{}
	for _, participant := range listParticipants {
		participants = append(participants, participant.Participant)
	}

	return &grpcXVPPb.GetNewsParticipantsResponse{
		Page:         page,
		PageSize:     pageSize,
		TotalItem:    totalItem,
		Participants: participants,
	}, nil
}

func CreateAndIncreaseParticipantNewsCount(participants []string) {
	if len(participants) == 0 {
		return
	}

	log.Println("CreateAndIncreaseParticipantNewsCount", participants)
	for _, participant := range participants {
		newsParticipant, err := dao.GetNewsParticipantDAO().FindByParticipant(context.Background(), participant)
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			newsParticipant = &model.NewsParticipant{
				Participant: participant,
				NewsCount:   0,
				SearchCount: 0,
				CreatedAt:   utils.UTCNowMilli(),
			}
			newsParticipant, err = dao.GetNewsParticipantDAO().Save(context.Background(), newsParticipant)
			if err != nil {
				log.Println(err)
			}
		}
	}
	_, err := dao.GetNewsParticipantDAO().IncreaseListNewsCount(context.Background(), participants)
	if err != nil {
		log.Println(err)
	}
}

func CreateAndUpdateParticipantNewsCount(deleteParticipants []string, addParticipants []string) {
	log.Println("CreateAndUpdateParticipantNewsCount", deleteParticipants, addParticipants)

	if len(addParticipants) > 0 {
		for _, participant := range addParticipants {
			newsParticipant, err := dao.GetNewsParticipantDAO().FindByParticipant(context.Background(), participant)
			if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
				newsParticipant = &model.NewsParticipant{
					Participant: participant,
					NewsCount:   0,
					SearchCount: 0,
					CreatedAt:   utils.UTCNowMilli(),
				}
				newsParticipant, err = dao.GetNewsParticipantDAO().Save(context.Background(), newsParticipant)
				if err != nil {
					log.Println(err)
				}
			}
		}
		_, err := dao.GetNewsParticipantDAO().IncreaseListNewsCount(context.Background(), addParticipants)
		if err != nil {
			log.Println(err)
		}
	}

	if len(deleteParticipants) > 0 {
		_, err := dao.GetNewsParticipantDAO().DecreaseListNewsCount(context.Background(), deleteParticipants)
		if err != nil {
			log.Println(err)
		}
	}

}

func IncreaseParticipantSearchCount(participants []string) {
	if len(participants) > 0 {
		_, err := dao.GetNewsParticipantDAO().IncreaseListSearchCount(context.Background(), participants)

		if err != nil {
			log.Println(err)
		}
	}
}
