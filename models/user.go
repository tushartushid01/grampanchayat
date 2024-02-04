package models

import (
	"database/sql"
	"time"
)

package models

import (
"database/sql"
"github.com/golang-jwt/jwt"
"github.com/lib/pq"
"time"
)

type FiltersCheck struct {
	IsSearched    bool
	SearchedName  string
	Limit         int
	Page          int
	SortBy        string
	GramPanchayat string
}

type ContextValues struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
}

type Claims struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
	jwt.StandardClaims
}

type UserCredentials struct {
	ID   int    `json:"id" db:"id"`
	Role string `json:"role" db:"role"`
}

type RoleDetails struct {
	Role string `json:"role" db:"role"`
}

type LoginWithOTP struct {
	Phone string `json:"phoneNo" db:"phone_no"`
	OTP   string `json:"otp" db:"otp"`
}

type SendOTP struct {
	Phone string `json:"phoneNo" db:"phone_no"`
}

type GramPanchayatCreateRequest struct {
	GramPanchayat string `json:"gramPanchayat" db:"gram_panchayat"`
	SachivName    string `json:"name" db:"name"`
	SachivPhoneNo string `json:"phoneNo" db:"phone_no"`
	TehsilID      int    `json:"tehsilID" db:"tehsil_id"`
	SahayakName   string `json:"sahayakName"`
	SahayakPhone  string `json:"sahayakPhoneNo"`
	BlockID       int    `json:"blockId"`
}

type TehsilCreateRequest struct {
	ID      int    `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	PhoneNo string `json:"phoneNo" db:"phone_no"`
	Tehsil  string `json:"tehsil" db:"tehsil"`
}

type GramPanchayatList struct {
	SachivName        string `json:"sachivName" db:"sachiv_name"`
	SachivID          int    `json:"sachivId" db:"sachiv_id"`
	SahayakID         int    `json:"sahayakId" db:"sahayak_id"`
	GramPanchayatName string `json:"gramPanchayatName" db:"gram_panchayat_name"`
	GramPanchayatID   int    `json:"gramPanchayatId" db:"gram_panchayat_id"`
	PhoneNo           string `json:"phoneNo" db:"sachiv_phone"`
	SahayakName       string `json:"sahayakName" db:"sahayak_name"`
	SahayakPhone      string `json:"sahayakPhone" db:"sahayak_phone"`
	TehsilName        string `json:"tehsilName" db:"tehsil_name"`
	TehsilID          int    `json:"tehsilID" db:"tehsil_id"`
	BlockName         string `json:"blockName" db:"block_name"`
	BlockID           int    `json:"blockId" db:"block_id"`
}

type GramPanchayatDetails struct {
	ID                int            `json:"id" db:"id"`
	SachivName        string         `json:"sachivName" db:"sachiv_name"`
	GramPanchayatName pq.StringArray `json:"gramPanchayatName" db:"gram_panchayat_name"`
	TehsilName        string         `json:"tehsilName" db:"tehsil_name"`
	PhoneNo           string         `json:"phoneNo" db:"phone_no"`
	TehsilID          int            `json:"tehsilId"`
}

type TehsilDetails struct {
	ID                int            `json:"id" db:"id"`
	SdmName           string         `json:"sdmName" db:"sdm_name"`
	TehsilName        string         `json:"tehsilName" db:"tehsil_name"`
	GramPanchayatName pq.StringArray `json:"gramPanchayatName" db:"gram_panchayat_name"`
	PhoneNo           string         `json:"phoneNo" db:"phone_no"`
}

type UserInfo struct {
	ID                   int               `json:"id" db:"id"`
	Name                 string            `json:"name" db:"name"`
	PhoneNumber          string            `db:"phone_no" json:"phoneNumber"`
	Role                 string            `db:"role" json:"role"`
	RegisterDeathEnabled bool              `db:"register_death_enabled" json:"registerDeathEnabled"`
	PanchayatList        []PanchayatOutput `json:"PanchayatList" db:"-"`
}

type PanchayatInfo struct {
	ID          int     `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	GaonDetails []uint8 `json:"gaonDetails" db:"gaon_details"`
}

type PanchayatOutput struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	GaonDetails []GaonInfo `json:"gaonDetails" db:"gaon_details"`
}

type TehsilInfo struct {
	ID                int            `json:"id" db:"id"`
	Name              string         `json:"name" db:"name"`
	GramPanchayatId   pq.Int64Array  `json:"gramPanchayatId" db:"gram_panchayat_id"`
	GramPanchayatName pq.StringArray `json:"gramPanchayatName" db:"gram_panchayat_name"`
	GaonId            pq.Int64Array  `json:"gaonId" db:"gaon_id"`
	GaonName          pq.StringArray `json:"gaonName" db:"gaon_name"`
}

type GramPanchayatInfo struct {
	GramPanchayatId   pq.Int64Array  `json:"gramPanchayatId" db:"gram_panchayat_id"`
	GramPanchayatName pq.StringArray `json:"gramPanchayatName" db:"gram_panchayat_name"`
	GaonId            pq.Int64Array  `json:"gaonId" db:"gaon_id"`
	GaonName          pq.StringArray `json:"gaonName" db:"gaon_name"`
}

type GaonInfo struct {
	GaonId   int    `json:"gaonId" db:"gaon_id"`
	GaonName string `json:"gaonName" db:"gaon_name"`
}

type DeathDetails struct {
	ID                int       `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	PhoneNo           string    `json:"phoneNo" db:"phone_no"`
	Age               int       `json:"age" db:"age"`
	Gender            string    `json:"gender" db:"gender"`
	AadharNumber      string    `json:"aadharNumber" db:"aadhar_number"`
	Status            string    `json:"status" db:"status"`
	Address           string    `json:"address" db:"address"`
	CreatedBy         int       `json:"createdBy" db:"created_by"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	DateOfDeath       time.Time `json:"dateOfDeath" db:"date_of_death"`
	TaskDetails       []uint8   `json:"task_Details" db:"task_details"`
	GramPanchayatId   int       `json:"gramPanchayatId" db:"gram_panchayat_id"`
	GramPanchayatName string    `json:"gramPanchayatName" db:"gram_panchayat_name"`
	TehsilId          int       `json:"tehsilId" db:"tehsil_id"`
	TehsilName        string    `json:"tehsilName" db:"tehsil_name"`
	BlockId           int       `json:"blockId" db:"block_id"`
	BlockName         string    `json:"blockName" db:"block_name"`
	GaonId            int       `json:"gaonId" db:"gaon_id"`
	GaonName          string    `json:"gaonName" db:"gaon_name"`
}

type DeathRegistrationRequest struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	PhoneNo      string    `json:"phoneNo"`
	Age          int       `json:"age"`
	Gender       string    `json:"gender"`
	AadharNumber string    `json:"aadharNumber"`
	Address      string    `json:"address"`
	DateOfDeath  time.Time `json:"dateOfDeath"`
	PanchayatID  int       `json:"gramPanchayatID" db:"gram_panchayat_id"`
	GaonID       int       `json:"gaonId" db:"gaon_id"`
}

type Processing struct {
	Started bool   `json:"started"`
	Reason  string `json:"reason"`
}

type DeathDetailsOutput struct {
	ID                int            `json:"id" db:"id"`
	DeathId           int            `json:"deathId"`
	Name              string         `json:"name" db:"name"`
	PhoneNo           string         `json:"phoneNo" db:"phone_no"`
	Age               int            `json:"age" db:"age"`
	Gender            string         `json:"gender" db:"gender"`
	AadharNumber      string         `json:"aadharNumber" db:"aadhar_number"`
	Status            string         `json:"status" db:"status"`
	Address           string         `json:"address" db:"address"`
	CreatedBy         int            `json:"createdBy" db:"created_by"`
	RegisterBy        UserInfo       `json:"registerBy" db:"-"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
	DateOfDeath       time.Time      `json:"dateOfDeath" db:"date_of_death"`
	GramPanchayatId   int            `json:"gramPanchayatId" db:"gram_panchayat_id"`
	GramPanchayatName string         `json:"gramPanchayatName" db:"gram_panchayat_name"`
	GaonId            int            `json:"gaonId" db:"gaon_id"`
	GaonName          string         `json:"gaonName" db:"gaon_name"`
	TehsilId          int            `json:"tehsilId" db:"tehsil_id"`
	TehsilName        string         `json:"tehsilName" db:"tehsil_name"`
	BlockId           int            `json:"blockId" db:"block_id"`
	BlockName         string         `json:"blockName" db:"block_name"`
	TaskDetails       []TaskDetail   `json:"taskDetails" db:"task_details"`
	IsReviewed        bool           `json:"isReviewed" db:"is_reviewed"`
	Comment           sql.NullString `json:"comment" db:"comment"`
	ReviewedBy        sql.NullInt64  `json:"reviewedBy" db:"reviewed_by"`
	Reviewer          UserInfo       `json:"reviewer" db:"-"`
	ReviewedAt        sql.NullTime   `json:"reviewedAt" db:"reviewed_at"`
}
type TaskDetail struct {
	TaskID       string `json:"taskId" db:"task_id"`
	Status       string `json:"status" db:"status"`
	TaskType     string `json:"name"`
	StartDate    string `json:"startDate" db:"start_date"`
	CompleteDate string `json:"completeDate" db:"completed_date"`
	IsEditable   bool   `json:"isEditable" db:"is_editable"`
	IsRejected   bool   `json:"isRejected" db:"is_rejected"`
	Reason       string `json:"reason"`
}

type DashBoardDetails struct {
	MonthCount  int `json:"monthCount" db:"month_count"`
	WeekCount   int `json:"weekCount" db:"week_count"`
	DayReported int `json:"dayReported" db:"day_reported"`
	DayResolved int `json:"dayResolved" db:"day_resolved"`
}

type TaskRole struct {
	TaskId int `json:"taskId" db:"id"`
	RoleId int `json:"roleId" db:"role_id"`
}

type GramPanchayatAndTehsilID struct {
	GramPanchayatID int `json:"gramPanchayatID" db:"gram_panchayat_id"`
	TehsilID        int `json:"tehsilID" db:"tehsil_id"`
}

type UserAndRoleID struct {
	UserID  int    `json:"userID" db:"user_id"`
	PhoneNo string `json:"phoneNo" db:"phone_no"`
	RolesId int    `json:"rolesId" db:"roles_id"`
}

type GramPanchayatAndUserID struct {
	GramPanchayatID int
	UserID          int
}

type GramUserDetails struct {
	GramPanchayatID   int    `json:"gramPanchayatID" db:"gram_panchayat_id"`
	GramPanchayatName string `json:"gramPanchayatName" db:"gram_panchayat_name"`
	SachivID          int    `json:"sachivId" db:"-"`
	SahayakID         int    `json:"sahayakId" db:"-"`
	SachivName        string `json:"sachivName" db:"-"`
	SachivPhoneNo     string `json:"sachivPhoneNo" db:"-"`
	SahayakName       string `json:"sahayakName" db:"-"`
	SahayakPhoneNo    string `json:"sahayakPhoneNo" db:"-"`
	BlockID           int    `json:"blockID" db:"block_id"`
	TehsilID          int    `json:"tehsilId" db:"tehsil_id"`
}

type TehsilUserDetails struct {
	TehsilID   int    `json:"tehsilID" db:"tehsil_id"`
	UserID     int    `json:"userID" db:"user_id"`
	TehsilName string `json:"tehsilName" db:"tehsil_name"`
	Name       string `json:"name" db:"name"`
	PhoneNo    string `json:"phoneNo" db:"phone_no"`
}

