package matcher_handler

import (
	"time"
)

type Form struct {
	Id         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Parameters Parameters `json:"parameters"`
	Active     bool       `json:"active"`
	Created_at time.Time  `json:"created_at"`
	Updated_at time.Time  `json:"updated_at"`
}

type Group struct {
	Id         string     `json:"id"`
	OwnerID    string     `json:"owner_id"`
	Parameters Parameters `json:"parameters"`
	MaxUsers   int32      `json:"max_users"`
	Created_at time.Time  `json:"created_at"`
	Updated_at time.Time  `json:"updated_at"`
}

type GroupWithScore struct {
	Group Group   `json:"group"`
	Score float32 `json:"score"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Parameters struct {
	Name           string   `json:"name,omitempty"`
	Surname        string   `json:"surname,omitempty"`
	Geo            Point    `json:"geo,omitempty"`
	Photos         []string `json:"photos,omitempty"`
	Budget         int32    `json:"budget,omitempty"`
	RoomCount      int32    `json:"room_count,omitempty"`
	RoommatesCount int32    `json:"roommates_count,omitempty"`
	Months         int32    `json:"months,omitempty"`
	Age            int32    `json:"age,omitempty"`
	Smoking        bool     `json:"smoking,omitempty"`
	Alko           bool     `json:"alko,omitempty"`
	Pet            bool     `json:"pet,omitempty"`
	Sex            string   `json:"sex,omitempty"`
	UserType       string   `json:"user_type,omitempty"`
	Description    string   `json:"description,omitempty"`
	Address        string   `json:"address,omitempty"`
}
type ListGroupMembersResponse struct {
	Forms []*Form `json:"forms"`
}

type FindGroupsResponse struct {
	GroupsWithScore []*GroupWithScore `json:"groups_with_score"`
}

type GroupRequest struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
