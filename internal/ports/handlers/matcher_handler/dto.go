package matcher_handler

import (
	"time"
)

type Form struct {
	Id         string
	UserID     string
	Parameters Parameters
	Active     bool
	Created_at time.Time
	Updated_at time.Time
}

type Group struct {
	Id         string
	OwnerID    string
	Parameters Parameters
	MaxUsers   int32
	Created_at time.Time
	Updated_at time.Time
}

type GroupWithScore struct {
	Group Group
	Score float32
}

type Point struct {
	Lat float64
	Lon float64
}

type Parameters struct {
	Name           string
	Surname        string
	Geo            Point
	Photos         []string
	Budget         int32
	RoomCount      int32
	RoommatesCount int32
	Age            int32
	Smoking        bool
	Alko           bool
	Pet            bool
	Sex            string
	UserType       string
	Description    string
}

type ListGroupMembersResponse struct {
	Forms []*Form `json:"Forms"`
}

type FindGroupsResponse struct {
	GroupsWithScore []*GroupWithScore `json:"GroupsWithScore"`
}

// Constants

// type Sex string

// const (
// 	SexUnspecified Sex = "unspecified"
// 	SexMale        Sex = "male"
// 	SexFemale      Sex = "female"
// )

// type UserType string

// const (
// 	UserTypeUnspecified UserType = "unspecified"
// 	UserTypeStudent     UserType = "student"
// 	UserTypeWorker      UserType = "worker"
// 	UserTypeTourist     UserType = "tourist"
// )
