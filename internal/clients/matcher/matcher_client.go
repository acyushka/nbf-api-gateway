package matcher

import (
	"context"

	dto "api-gateway/internal/ports/handlers/matcher_handler"

	matcherv1 "github.com/hesoyamTM/nbf-protos/gen/go/matcher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	FormServiceApi       matcherv1.FormServiceClient
	GroupQueryApi        matcherv1.GroupQueryServiceClient
	FindGroupsServiceApi matcherv1.FindGroupServiceClient
	GroupServiceApi      matcherv1.GroupServiceClient
}

func New(ctx context.Context, address string) (*Client, error) {
	cc, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		FormServiceApi:       matcherv1.NewFormServiceClient(cc),
		GroupQueryApi:        matcherv1.NewGroupQueryServiceClient(cc),
		FindGroupsServiceApi: matcherv1.NewFindGroupServiceClient(cc),
		GroupServiceApi:      matcherv1.NewGroupServiceClient(cc),
	}, nil
}

////////////////////////////////////

func (c *Client) CreateForm(ctx context.Context, uid string, protoParams *matcherv1.Parameters) error {
	_, err := c.FormServiceApi.CreateForm(ctx, &matcherv1.CreateFormRequest{
		UserId:     uid,
		Parameters: protoParams,
	},
	)

	return err
}

func (c *Client) GetFormByUser(ctx context.Context, uid string) (*dto.Form, error) {
	resp, err := c.FormServiceApi.GetFormByUser(ctx, &matcherv1.GetFormByUserRequest{
		UserId: uid,
	})
	if err != nil {
		return nil, err
	}

	return &dto.Form{
		Id:         resp.GetId(),
		UserID:     resp.GetUserId(),
		Parameters: ParametersFromProto(resp.GetParameters()),
		Active:     resp.GetActive(),
		Created_at: resp.GetCreatedAt().AsTime(),
		Updated_at: resp.GetUpdatedAt().AsTime(),
	}, nil
}
func (c *Client) UpdateForm(ctx context.Context, uid string, protoParams *matcherv1.Parameters) error {
	_, err := c.FormServiceApi.UpdateForm(ctx, &matcherv1.UpdateFormRequest{
		UserId:     uid,
		Parameters: protoParams,
	},
	)

	return err
}

func (c *Client) DeleteForm(ctx context.Context, uid string) error {
	_, err := c.FormServiceApi.DeleteForm(ctx, &matcherv1.DeleteFormRequest{
		UserId: uid,
	})

	return err
}

///////////////////////////////////////////

func (c *Client) GetGroup(ctx context.Context, gid string) (*dto.Group, error) {
	resp, err := c.GroupQueryApi.GetGroup(ctx, &matcherv1.GetGroupRequest{
		GroupId: gid,
	})
	if err != nil {
		return nil, err
	}

	return &dto.Group{
		Id:         resp.GetId(),
		OwnerID:    resp.GetOwnerId(),
		Parameters: ParametersFromProto(resp.GetParameters()),
		MaxUsers:   resp.GetMaxUsers(),
		Created_at: resp.GetCreatedAt().AsTime(),
		Updated_at: resp.GetUpdatedAt().AsTime(),
	}, nil
}

func (c *Client) DeleteGroup(ctx context.Context, oid string) error {
	_, err := c.GroupQueryApi.DeleteGroup(ctx, &matcherv1.DeleteGroupRequest{
		OwnerId: oid,
	})

	return err
}

func (c *Client) ListGroupMembers(ctx context.Context, gid string) ([]*dto.Form, error) {
	resp, err := c.GroupQueryApi.ListGroupMembers(ctx, &matcherv1.ListGroupMembersRequest{
		GroupId: gid,
	})
	if err != nil {
		return nil, err
	}

	forms := make([]*dto.Form, len(resp.GetMembers()))
	for i, form := range resp.GetMembers() {
		forms[i] = &dto.Form{
			Id:         form.GetId(),
			UserID:     form.GetUserId(),
			Parameters: ParametersFromProto(form.GetParameters()),
			Active:     form.GetActive(),
			Created_at: form.GetCreatedAt().AsTime(),
			Updated_at: form.GetUpdatedAt().AsTime(),
		}
	}

	return forms, nil
}

///////////////////////////////////////////

func (c *Client) FindGroups(ctx context.Context, uid string) ([]*dto.Group, error) {
	resp, err := c.FindGroupsServiceApi.FindGroups(ctx, &matcherv1.FindGroupsRequest{
		UserId: uid,
	})
	if err != nil {
		return nil, err
	}

	groups := make([]*dto.Group, len(resp.GetGroups()))
	for i, group := range resp.GetGroups() {
		groups[i] = &dto.Group{
			Id:         group.GetId(),
			OwnerID:    group.GetOwnerId(),
			Parameters: ParametersFromProto(group.GetParameters()),
			MaxUsers:   group.GetMaxUsers(),
			Created_at: group.GetCreatedAt().AsTime(),
			Updated_at: group.GetUpdatedAt().AsTime(),
		}
	}

	return groups, nil
}

///////////////////////////////////////////

func (c *Client) SendJoinRequest(ctx context.Context, uid string, gid string) error {
	_, err := c.GroupServiceApi.SendJoinRequest(ctx, &matcherv1.SendJoinRequestRequest{
		UserId:  uid,
		GroupId: gid,
	})

	return err
}

func (c *Client) AcceptJoinRequest(ctx context.Context, oid string, rid string) error {
	_, err := c.GroupServiceApi.AcceptJoinRequest(ctx, &matcherv1.AcceptJoinRequestRequest{
		OwnerId:   oid,
		RequestId: rid,
	})

	return err
}

func (c *Client) RejectJoinRequest(ctx context.Context, oid string, rid string) error {
	_, err := c.GroupServiceApi.RejectJoinRequest(ctx, &matcherv1.RejectJoinRequestRequest{
		OwnerId:   oid,
		RequestId: rid,
	})

	return err
}

// Внутрянка

func ParametersFromProto(protoParams *matcherv1.Parameters) dto.Parameters {
	if protoParams == nil {
		return dto.Parameters{}
	}

	return dto.Parameters{
		Name:           protoParams.GetName(),
		Surname:        protoParams.GetSurname(),
		Geo:            pointFromProto(protoParams.GetGeo()),
		Photos:         protoParams.GetPhotos(),
		Budget:         protoParams.GetBudget(),
		RoomCount:      protoParams.GetRoomCount(),
		RoommatesCount: protoParams.GetRoommatesCount(),
		Age:            protoParams.GetAge(),
		Smoking:        protoParams.GetSmoking(),
		Alko:           protoParams.GetAlko(),
		Pet:            protoParams.GetPet(),
		Sex:            sexToString(protoParams.GetSex()),
		UserType:       userTypeToString(protoParams.GetUserType()),
		Description:    protoParams.GetDescription(),
	}
}

func pointFromProto(protoPoint *matcherv1.Point) dto.Point {
	if protoPoint == nil {
		return dto.Point{}
	}
	return dto.Point{
		Lat: protoPoint.GetLat(),
		Lon: protoPoint.GetLon(),
	}
}

func sexToString(sex matcherv1.Sex) string {
	switch sex {
	case matcherv1.Sex_SEX_MALE:
		return "male"
	case matcherv1.Sex_SEX_FEMALE:
		return "female"
	default:
		return "unspecified"
	}
}

func userTypeToString(userType matcherv1.UserType) string {
	switch userType {
	case matcherv1.UserType_USER_TYPE_STUDENT:
		return "student"
	case matcherv1.UserType_USER_TYPE_WORKER:
		return "worker"
	case matcherv1.UserType_USER_TYPE_TOURIST:
		return "tourist"
	default:
		return "unspecified"
	}
}
