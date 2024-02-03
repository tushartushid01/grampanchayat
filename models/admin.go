package models

import (
	"database/sql"
	"time"
)

type AdminLoginCredentials struct {
	Phone    string `json:"phone" db:"phone_no"`
	Password string `json:"password" db:"password"`
}

type TehsilAndUserID struct {
	TehsilID int
	UserID   int
}

type TehsilDetail struct {
	TehsilID   int    `json:"id" db:"id"`
	UserID     int    `json:"userId" db:"user_id"`
	TehsilName string `json:"tehsilName" db:"tehsil_name"`
	SDMName    string `json:"sdmName" db:"name"`
	PhoneNo    string `json:"phoneNo" db:"phone"`
}

type Tehsils struct {
	TotalCount int            `json:"totalCount"`
	Tehsils    []TehsilDetail `json:"tehsils"`
}

type TaskDetails struct {
	TaskID   int    `json:"taskId" db:"task_id"`
	TaskName string `json:"taskName" db:"task_name"`
	RoleID   int    `json:"roleId" db:"role_id"`
	RoleName string `json:"roleName" db:"role_name"`
}

type Tasks struct {
	TotalCount int           `json:"totalCount"`
	Tasks      []TaskDetails `json:"tasks"`
}

type TotalDeaths struct {
	Month int `db:"month"`
	Week  int `db:"week"`
	Day   int `db:"day"`
}

type DistrictPost struct {
	Role     string  `json:"role" db:"role"`
	Name     string  `json:"name" db:"name"`
	PhoneNo  string  `json:"phoneNo" db:"phone_no"`
	TaskName []uint8 `json:"taskName" db:"task_name"`
}

type DistrictPostOutput struct {
	Role     string     `json:"role" db:"role"`
	Name     string     `json:"name" db:"name"`
	PhoneNo  string     `json:"phoneNo" db:"phone_no"`
	TaskName []TaskName `json:"taskName" db:"task_name"`
}

type GraphDeatils struct {
	Date       time.Time `json:"date" db:"date"`
	Registered int       `json:"registered" db:"registered"`
	Completed  int       `json:"completed" db:"completed"`
}

type DeathFilter struct {
	GramPanchayatID []int
	TehsilID        []int
	BlockID         []int
	GaonID          []int
	TaskID          []int
	TaskName        []string
	FromDate        time.Time
	ToDate          time.Time
	Search          string
	Status          string
	OrderBy         string
	IsAscending     bool
	Limit           int
	Page            int
}

type NewUserDetails struct {
	userID          int    `json:"userID" db:"user_id"`
	RoleID          int    `json:"roleID" db:"role_id"`
	Name            string `json:"name" db:"name"`
	PhoneNo         string `json:"phoneNo" db:"phone_no"`
	GramPanchayatID int    `json:"gramPanchayatID" db:"gram_panchayat_id"`
	TehsilID        int    `json:"tehsilID" db:"tehsil_id"`
}

type BlockDetails struct {
	BlockID   int    `json:"blockId" db:"id"`
	BlockName string `json:"blockName" db:"name"`
}

type RandomDeathDetails struct {
	ID                int            `json:"id" db:"id"`
	DeathId           int            `json:"deathId" db:"death_id"`
	Name              string         `json:"name" db:"name"`
	PhoneNo           string         `json:"phoneNo" db:"phone_no"`
	Age               int            `json:"age" db:"age"`
	Gender            string         `json:"gender" db:"gender"`
	AadharNumber      string         `json:"aadharNumber" db:"aadhar_number"`
	Status            string         `json:"status" db:"status"`
	Address           string         `json:"address" db:"address"`
	CreatedBy         int            `json:"createdBy" db:"created_by"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
	DateOfDeath       time.Time      `json:"dateOfDeath" db:"date_of_death"`
	TaskDetails       []uint8        `json:"task_Details" db:"task_details"`
	GramPanchayatId   int            `json:"gramPanchayatId" db:"gram_panchayat_id"`
	GramPanchayatName string         `json:"gramPanchayatName" db:"gram_panchayat_name"`
	TehsilId          int            `json:"tehsilId" db:"tehsil_id"`
	TehsilName        string         `json:"tehsilName" db:"tehsil_name"`
	BlockId           int            `json:"blockId" db:"block_id"`
	BlockName         string         `json:"blockName" db:"block_name"`
	GaonId            int            `json:"gaonId" db:"gaon_id"`
	GaonName          string         `json:"gaonName" db:"gaon_name"`
	IsReviewed        bool           `json:"isReviewed" db:"is_reviewed"`
	Comment           sql.NullString `json:"comment" db:"comment"`
	ReviewedBy        sql.NullInt64  `json:"reviewedBy" db:"reviewed_by"`
	ReviewedAt        sql.NullTime   `json:"reviewedAt" db:"reviewed_at"`
}

type RandomDeath struct {
	ID            int       `json:"id" db:"id"`
	DeathID       int       `json:"deathId" db:"death_detail_id"`
	IsReviewed    bool      `json:"IsReviewed" db:"is_reviewed"`
	ReviewComment string    `json:"reviewComment" db:"comment"`
	ReviewedBy    int       `json:"reviewedBy" db:"review_by"`
	ReviewedAt    time.Time `json:"reviewedAt" db:"reviewed_at"`
}

type GaonDetails struct {
	ID                int    `json:"id" db:"id"`
	GaonName          string `json:"name" db:"name"`
	LekhPalID         int    `json:"lekhPalId" db:"lekhpal_id"`
	LekhPalName       string `json:"lekhPalName" db:"lekhpal_name"`
	LekhPalPhone      string `json:"lekhPalPhone" db:"phone_no"`
	GramPanchayatID   int    `json:"gramPanchayatId" db:"gram_panchayat_id"`
	GramPanchayatName string `json:"gramPanchayatName" db:"gram_panchayat_name"`
}

type TaskName struct {
	TaskName string `json:"taskName"`
}
