package matcher_handler

import (
	models "api-gateway/internal/ports/handlers/user_handler"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	authorization "github.com/hesoyamTM/nbf-auth/pkg/auth"
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
	LeaveGroup(ctx context.Context, uid string) error
	KickGroup(ctx context.Context, oid, uid string) error
	GetGroup(ctx context.Context, gid string) (*Group, error)
	GetGroupByUser(ctx context.Context, uid string) (*Group, error)
	DeleteGroup(ctx context.Context, oid string) error
	ListGroupMembers(ctx context.Context, gid string) ([]*Form, error)
	//FindGroup Service
	FindGroups(ctx context.Context, uid string) ([]*GroupWithScore, error)
	//Group Service
	GetRequests(ctx context.Context, gid string) ([]*GroupRequest, error)
	SendJoinRequest(ctx context.Context, uid string, gid string) (string, error)
	AcceptJoinRequest(ctx context.Context, oid string, rid string) error
	RejectJoinRequest(ctx context.Context, oid string, rid string) error
}

type FileStorageClient interface {
	UploadPhotos(ctx context.Context, userID string, files []*models.FilePhoto) ([]string, error)
	GetPhotoURL(ctx context.Context, userID string, photoID string) (string, error)
}

type MatcherHandler struct {
	matcherClient     MatcherClient
	fileStorageClient FileStorageClient
}

func NewMatcherHandler(m MatcherClient, storageClient FileStorageClient) *MatcherHandler {
	return &MatcherHandler{
		matcherClient:     m,
		fileStorageClient: storageClient,
	}
}

//////////////FORM SERVICE/////////////////

func (h *MatcherHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(52428800); err != nil {
		log.Error("Failed to parse formdata", zap.Error(err))
		http.Error(w, "Invalid formdata", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID     string     `json:"user_id"`
		Parameters Parameters `json:"parameters"`
	}

	if err := json.Unmarshal([]byte(r.FormValue("data")), &req); err != nil {
		log.Error("Failed to parse JSON", zap.Error(err))
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

	ctx := r.Context()

	photoIDs, err := h.uploadPhotos(ctx, r, req.UserID)
	if err != nil {
		log.Error("Failed to upload photos", zap.Error(err))
		http.Error(w, "Failed to upload photos", http.StatusBadRequest)
		return
	}

	if len(photoIDs) > 0 {
		req.Parameters.Photos = photoIDs
	}

	protoParams := toProtoParams(req.Parameters, sex, userType)

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

	for i, photoID := range form.Parameters.Photos {
		url, err := h.fileStorageClient.GetPhotoURL(ctx, form.UserID, photoID)
		if err != nil {
			log.Error("Failed to get presigned URL", zap.Error(err))
			continue
		}
		form.Parameters.Photos[i] = url
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

	if err := r.ParseMultipartForm(52428800); err != nil {
		log.Error("Failed to parse formdata", zap.Error(err))
		http.Error(w, "Invalid formdata", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID     string     `json:"user_id"`
		Parameters Parameters `json:"parameters"`
	}

	if err := json.Unmarshal([]byte(r.FormValue("data")), &req); err != nil {
		log.Error("Failed to parse JSON", zap.Error(err))
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

	ctx := r.Context()

	photoIDs, err := h.uploadPhotos(ctx, r, req.UserID)
	if err != nil {
		log.Error("Failed to upload photos", zap.Error(err))
		http.Error(w, "Failed to upload photos", http.StatusBadRequest)
		return
	}

	if len(photoIDs) > 0 {
		req.Parameters.Photos = photoIDs
	}

	protoParams := toProtoParams(req.Parameters, sex, userType)

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

func (h *MatcherHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	uid, ok := ctx.Value(authorization.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.matcherClient.LeaveGroup(ctx, uid); err != nil {
		log.Error("Failed to leave from group", zap.Error(err))
		http.Error(w, "Failed to leave from group", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *MatcherHandler) KickGroup(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	oid := chi.URLParam(r, "oid")

	ctx := r.Context()
	uid, ok := ctx.Value(authorization.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.matcherClient.KickGroup(ctx, oid, uid); err != nil {
		log.Error("Failed to kick group", zap.Error(err))
		http.Error(w, "Failed to kick group", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

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

	for i, photoID := range group.Parameters.Photos {
		url, err := h.fileStorageClient.GetPhotoURL(ctx, group.Id, photoID)
		if err != nil {
			log.Error("Failed to get presigned URL", zap.Error(err))
			continue
		}
		group.Parameters.Photos[i] = url
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, group)
}

func (h *MatcherHandler) GetGroupByUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")

	ctx := r.Context()
	group, err := h.matcherClient.GetGroupByUser(ctx, uid)
	if err != nil {
		log.Error("Failed to get Group by User", zap.Error(err))
		http.Error(w, "Failed to get Group by User", http.StatusInternalServerError)
		return
	}

	for i, photoID := range group.Parameters.Photos {
		url, err := h.fileStorageClient.GetPhotoURL(ctx, group.Id, photoID)
		if err != nil {
			log.Error("Failed to get presigned URL", zap.Error(err))
			continue
		}
		group.Parameters.Photos[i] = url
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

	render.Status(r, http.StatusOK)
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

	for _, form := range forms {
		for i, photoID := range form.Parameters.Photos {
			photoURL, err := h.fileStorageClient.GetPhotoURL(ctx, form.UserID, photoID)
			if err != nil {
				log.Error("Failed to get presigned URL", zap.Error(err))
				continue
			}
			form.Parameters.Photos[i] = photoURL
		}
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
	GroupsWithScore, err := h.matcherClient.FindGroups(ctx, uid)
	if err != nil {
		log.Error("Failed to find Groups", zap.Error(err))
		http.Error(w, "Failed to find Groups", http.StatusInternalServerError)
		return
	}

	for _, groupWithScore := range GroupsWithScore {
		for i, photoID := range groupWithScore.Group.Parameters.Photos {
			url, err := h.fileStorageClient.GetPhotoURL(ctx, groupWithScore.Group.Id, photoID)
			if err != nil {
				log.Error("Failed to get group photo presigned URL",
					zap.String("photo_id", photoID),
					zap.String("group_id", groupWithScore.Group.Id),
					zap.Error(err))
				continue
			}
			groupWithScore.Group.Parameters.Photos[i] = url
		}
	}

	response := &FindGroupsResponse{
		GroupsWithScore: GroupsWithScore,
	}

	render.JSON(w, r, response)
	render.Status(r, http.StatusOK)
}

//////////////GROUP SERVICE/////////////////

func (h *MatcherHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	gid := chi.URLParam(r, "gid")

	ctx := r.Context()
	resp, err := h.matcherClient.GetRequests(ctx, gid)
	if err != nil {
		log.Error("Failed to get GroupRequest", zap.Error(err))
		http.Error(w, "Failed to get GroupRequest", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, resp)
	render.Status(r, http.StatusOK)
}

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
	rid, err := h.matcherClient.SendJoinRequest(ctx, req.UserID, req.GroupID)
	if err != nil {
		log.Error("Failed to send Join Request", zap.Error(err))
		http.Error(w, "Failed to send Join Request", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{
		"request_id": rid,
	})
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
		Month:          p.Months,
		Age:            p.Age,
		Smoking:        p.Smoking,
		Alko:           p.Alko,
		Pet:            p.Pet,
		Sex:            matcherv1.Sex(sex),
		UserType:       matcherv1.UserType(userType),
		Description:    p.Description,
		Address:        p.Address,
	}
}

func (h *MatcherHandler) uploadPhotos(ctx context.Context, r *http.Request, userID string) ([]string, error) {
	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		return nil, nil
	}

	photoFiles := make([]*models.FilePhoto, 0, len(files))

	for _, fileHeader := range files {
		if fileHeader.Size > 52428800 {
			return nil, fmt.Errorf("Invalid photo size")
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("Failed to open file: %w", err)
		}
		defer file.Close()

		photoFiles = append(photoFiles, &models.FilePhoto{
			Data:        file,
			FileName:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-type"),
		})

	}

	photoIDs, err := h.fileStorageClient.UploadPhotos(ctx, userID, photoFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to upload photos: %w", err)
	}

	return photoIDs, nil
}
