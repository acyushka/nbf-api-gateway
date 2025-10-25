package matcher_handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	matcherv1 "github.com/hesoyamTM/nbf-protos/gen/go/matcher"
	"go.uber.org/zap"
)

type MatcherClient interface {
	//Form Service
	CreateForm(ctx context.Context, uid string, protoParams *matcherv1.Parameters) error
	GetFormByUser(ctx context.Context, uid string) (*Form, error)
	UpdateForm(ctx context.Context, uid string, protoParams *matcherv1.Parameters) error
	DeleteForm(ctx context.Context, uid string) error
	//Group Query
	GetGroup(ctx context.Context, gid string) (*Group, error)
	DeleteGroup(ctx context.Context, oid string) error
	ListGroupMembers(ctx context.Context, gid string) ([]*Form, error)
	//FindGroup Service
	FindGroups(ctx context.Context, uid string) ([]*Group, error)
	//Group Service
	SendJoinRequest(ctx context.Context, uid string, gid string) error //айди вернуть
	AcceptJoinRequest(ctx context.Context, oid string, rid string) error
	RejectJoinRequest(ctx context.Context, oid string, rid string) error
}

type MatcherHandler struct {
	matcherClient MatcherClient
}

func NewMatcherHandler(m MatcherClient) *MatcherHandler {
	return &MatcherHandler{
		matcherClient: m,
	}
}

//////////////FORM SERVICE/////////////////

func (h *MatcherHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req struct {
		UserID     string     `json:"user_id"`
		Parameters Parameters `json:"parameters"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	sex, err := validateSex(req.Parameters.Sex)
	if err != nil {
		http.Error(w, "Invalid sex value"+err.Error(), http.StatusBadRequest)
		return
	}

	userType, err := validateUserType(req.Parameters.UserType)
	if err != nil {
		http.Error(w, "Invalid User Type value"+err.Error(), http.StatusBadRequest)
		return
	}

	protoParams := toProtoParams(req.Parameters, sex, userType)

	ctx := r.Context()
	if err := h.matcherClient.CreateForm(ctx, req.UserID, protoParams); err != nil {
		log.Error("Failed to create Form", zap.Error(err))
		http.Error(w, "Failed to create Form", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
}

func (h *MatcherHandler) GetFormByUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")

	ctx := r.Context()
	form, err := h.matcherClient.GetFormByUser(ctx, uid)
	if err != nil {
		log.Error("Failed to get Form", zap.Error(err))
		http.Error(w, "Failed to get Form", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, form)
}

func (h *MatcherHandler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req struct {
		UserID     string     `json:"user_id"`
		Parameters Parameters `json:"parameters"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	sex := 0
	userType := 0

	if req.Parameters.Sex != "" {
		sex, err = validateSex(req.Parameters.Sex)
		if err != nil {
			http.Error(w, "Invalid sex value"+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.Parameters.UserType != "" {
		userType, err = validateUserType(req.Parameters.UserType)
		if err != nil {
			http.Error(w, "Invalid User Type value"+err.Error(), http.StatusBadRequest)
			return
		}
	}

	protoParams := toProtoParams(req.Parameters, sex, userType)

	ctx := r.Context()
	if err := h.matcherClient.UpdateForm(ctx, req.UserID, protoParams); err != nil {
		log.Error("Failed to update Form", zap.Error(err))
		http.Error(w, "Failed to update Form", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *MatcherHandler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")
	if uid == "" {
		http.Error(w, "User id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.matcherClient.DeleteForm(ctx, uid); err != nil {
		log.Error("Failed to delete form", zap.Error(err))
		http.Error(w, "Failed to delete form", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

//////////////GROUP QUERY/////////////////

func (h *MatcherHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	gid := chi.URLParam(r, "gid")

	ctx := r.Context()
	group, err := h.matcherClient.GetGroup(ctx, gid)
	if err != nil {
		log.Error("Failed to get Group", zap.Error(err))
		http.Error(w, "Failed to get Group", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, group)
}

func (h *MatcherHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	oid := chi.URLParam(r, "oid")
	if oid == "" {
		http.Error(w, "Owner id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.matcherClient.DeleteGroup(ctx, oid); err != nil {
		log.Error("Failed to delete group", zap.Error(err))
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}
}

func (h *MatcherHandler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	gid := chi.URLParam(r, "gid")

	ctx := r.Context()
	forms, err := h.matcherClient.ListGroupMembers(ctx, gid)
	if err != nil {
		log.Error("Failed to get Members of group", zap.Error(err))
		http.Error(w, "Failed to get Members of group", http.StatusInternalServerError)
		return
	}

	response := &ListGroupMembersResponse{
		Forms: forms,
	}

	render.JSON(w, r, response)
	render.Status(r, http.StatusOK)
}

//////////////FIND GROUP SERVICE/////////////////

func (h *MatcherHandler) FindGroups(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")

	ctx := r.Context()
	groups, err := h.matcherClient.FindGroups(ctx, uid)
	if err != nil {
		log.Error("Failed to find Groups", zap.Error(err))
		http.Error(w, "Failed to find Groups", http.StatusInternalServerError)
		return
	}

	response := &FindGroupsResponse{
		Groups: groups,
	}

	render.JSON(w, r, response)
	render.Status(r, http.StatusOK)
}

//////////////GROUP SERVICE/////////////////

func (h *MatcherHandler) SendJoinRequest(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req struct {
		UserID  string `json:"user_id"`
		GroupID string `json:"group_id"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.GroupID == "" {
		http.Error(w, "User id or group id is empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.matcherClient.SendJoinRequest(ctx, req.UserID, req.GroupID); err != nil {
		log.Error("Failed to send Join Request", zap.Error(err))
		http.Error(w, "Failed to send Join Request", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *MatcherHandler) AcceptJoinRequest(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req struct {
		OwnerID   string `json:"owner_id"`
		RequestID string `json:"request_id"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OwnerID == "" || req.RequestID == "" {
		http.Error(w, "Owner id or Request id is empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.matcherClient.AcceptJoinRequest(ctx, req.OwnerID, req.RequestID); err != nil {
		log.Error("Failed to accept Join Request", zap.Error(err))
		http.Error(w, "Failed to accept Join Request", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *MatcherHandler) RejectJoinRequest(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req struct {
		OwnerID   string `json:"owner_id"`
		RequestID string `json:"request_id"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OwnerID == "" || req.RequestID == "" {
		http.Error(w, "Owner id or Request id is empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.matcherClient.RejectJoinRequest(ctx, req.OwnerID, req.RequestID); err != nil {
		log.Error("Failed to reject Join Request", zap.Error(err))
		http.Error(w, "Failed to reject Join Request", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

//Внутрянка

func validateSex(sex string) (int, error) {
	if sex == "unspecified" {
		return 0, nil
	}
	if sex == "male" {
		return 1, nil
	}
	if sex == "female" {
		return 2, nil
	}

	return 0, fmt.Errorf("expected 'male', 'female' or 'unspecified', got '%s'", sex)
}

func validateUserType(userType string) (int, error) {
	if userType == "unspecified" {
		return 0, nil
	}
	if userType == "student" {
		return 1, nil
	}
	if userType == "worker" {
		return 2, nil
	}
	if userType == "tourist" {
		return 3, nil
	}

	return 0, fmt.Errorf("expected 'student', 'worker', 'tourist' or 'unspecified', got '%s'", userType)
}

func toProtoParams(p Parameters, sex, userType int) *matcherv1.Parameters {
	return &matcherv1.Parameters{
		Name:    p.Name,
		Surname: p.Surname,
		Geo: &matcherv1.Point{
			Lat: p.Geo.Lat,
			Lon: p.Geo.Lon,
		},
		Photos:         p.Photos,
		Budget:         p.Budget,
		RoomCount:      p.RoomCount,
		RoommatesCount: p.RoommatesCount,
		Age:            p.Age,
		Smoking:        p.Smoking,
		Alko:           p.Alko,
		Pet:            p.Pet,
		Sex:            matcherv1.Sex(sex),
		UserType:       matcherv1.UserType(userType),
		Description:    p.Description,
	}
}
