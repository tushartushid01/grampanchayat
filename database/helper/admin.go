package helper

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"grampanchayat/database"
	"grampanchayat/models"
	"grampanchayat/utilities"
	"strconv"
	"time"
)

func CreateSession(claims *models.Claims) error {
	// language=SQL
	SQL := `INSERT INTO sessions(user_id, expires_at)
            VALUES ($1, $2)`

	_, err := database.GramPanchayatDB.Exec(SQL, claims.ID, time.Now().Add(60*time.Hour))
	if err != nil {
		logrus.Printf("CreateSession: cannot create session:%v", err)
		return err
	}
	return nil
}

func CheckSession(userID int) (uuid.UUID, error) {
	// language=SQL
	SQL := `SELECT id
            FROM   sessions
            WHERE  expires_at > now()
            AND    user_id = $1
            ORDER BY created_at desc LIMIT 1`

	var sessionID uuid.UUID

	err := database.GramPanchayatDB.Get(&sessionID, SQL, userID)
	if err != nil {
		logrus.Printf("CheckSession: cannot get session id:%v", err)
		return sessionID, err
	}
	return sessionID, nil
}

func GetDisplayTypes(role string) ([]string, error) {

	var SQL string
	if role == utilities.Sahayak || role == utilities.Sachiv || role == utilities.LekhPal {
		// language=SQL
		SQL = `SELECT name
            FROM   task_types`
		types := make([]string, 0)

		err := database.GramPanchayatDB.Select(&types, SQL)
		return types, err
	} else {
		// language=SQL
		SQL = `SELECT name
            FROM   task_types
            join task_role tr on task_types.id = tr.task_type_id
            join roles r on tr.role_id = r.id
            where r.role = $1`
		types := make([]string, 0)

		err := database.GramPanchayatDB.Select(&types, SQL, role)
		return types, err
	}
}

func GetActionableTaskTypes(role string) ([]string, error) {
	// language=SQL
	SQL := `SELECT name
            FROM   task_types
            join task_role tr on task_types.id = tr.task_type_id
            join roles r on tr.role_id = r.id
            where r.role = $1`
	types := make([]string, 0)

	err := database.GramPanchayatDB.Select(&types, SQL, role)
	return types, err
}

func AddOtp(phone, otp string) error {
	// language=SQL
	SQL := `INSERT INTO otp(phone_no, otp, expiring_time)
            VALUES ($1, $2, $3)`

	_, err := database.GramPanchayatDB.Exec(SQL, phone, otp, time.Now().Add(5*time.Minute))
	if err != nil {
		logrus.Printf("AddOtp: cannot add otp for user:%v", err)
		return err
	}
	return nil
}

func FetchOTP(phone string) (string, error) {
	// language=SQL
	SQL := `SELECT otp
            FROM   otp
            WHERE phone_no= $1
            AND   expiring_time > now() 
            AND   archived_at IS NULL 
            ORDER BY expiring_time desc LIMIT 1`

	var otp string

	err := database.GramPanchayatDB.Get(&otp, SQL, phone)
	if err != nil {
		logrus.Printf("FetchOTP: caannot get otp:%v", err)
		return otp, err
	}
	return otp, nil
}

func AddRole(roleDetails models.RoleDetails) error {
	// language = SQL
	SQL := `INSERT INTO roles(role)
            VALUES ($1)`
	_, err := database.GramPanchayatDB.Exec(SQL, roleDetails.Role)
	if err != nil {
		logrus.Printf("AddRole:cannot add role:%v", err)
		return err
	}
	return nil
}
func BulkAddRole(roleDetails []models.RoleDetails) ([]int, error) {
	// language = SQL
	insertQuery := `INSERT INTO roles(role) VALUES($1)`
	for i := range roleDetails {
		numFields := 1
		n := i * numFields
		insertQuery += `(`
		for j := 0; j < numFields; j++ {
			insertQuery += `$` + strconv.Itoa(n+j+1) + `,`
		}
		insertQuery = insertQuery[:len(insertQuery)-1] + `),`
	}
	insertQuery = insertQuery[:len(insertQuery)-1] //remove the last trailing comma
	returnStmt := " RETURNING id"
	insertQuery += returnStmt
	values := make([]interface{}, 0)
	for _, value := range roleDetails {
		values = append(values, value.Role)
	}
	rows, err := database.GramPanchayatDB.Query(insertQuery, values...)
	if err != nil {
		logrus.Printf("AddRole:cannot add role:%v", err)
		return nil, err
	}
	roleIds := make([]int, 0)
	for rows.Next() {
		var id int
		scanErr := rows.Scan(&id)
		if scanErr != nil {
			return nil, scanErr
		}
		roleIds = append(roleIds, id)
	}
	return roleIds, nil
}

func FetchUserIDAndRole(phone string) (models.UserCredentials, error) {
	// language=SQL
	SQL := `SELECT users.id,
                   role
            FROM users JOIN roles r on r.id = users.roles_id
            WHERE phone_no = $1`

	var userCredentials models.UserCredentials

	err := database.GramPanchayatDB.Get(&userCredentials, SQL, phone)
	if err != nil {
		logrus.Printf("FetchUserIDAndRole: cannot fetch user id or role:%v", err)
		return userCredentials, err
	}
	return userCredentials, nil
}

func FetchRole(role string) (int, error) {
	// language=SQL
	SQL := `SELECT id
            FROM   roles
            WHERE  role = $1`

	var roleID int

	err := database.GramPanchayatDB.Get(&roleID, SQL, role)
	if err != nil {
		logrus.Printf("FetchRole:cannot get role id:%v", err)
		return roleID, err
	}
	return roleID, nil
}

func GetUserByPhoneNo(phone string, tx *sqlx.Tx) (models.UserAndRoleID, error) {
	//language=SQL
	SQL := `SELECT roles_id, 
       			   users.id as user_id
            FROM   users JOIN roles r on r.id = users.roles_id
            WHERE  users.phone_no = $1
            AND    users.archived_at IS NULL `

	var userAndRoleID models.UserAndRoleID

	err := tx.Get(&userAndRoleID, SQL, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			userAndRoleID.UserID = 0
			userAndRoleID.RolesId = 0
			return userAndRoleID, nil
		}
		logrus.Printf("GetUserByPhoneNo: cannot get role by phone no:%v", err)
		return userAndRoleID, err
	}
	return userAndRoleID, nil
}

func AddUser(name, phone, role string, userAndRoleID models.UserAndRoleID, tx *sqlx.Tx) (int, error) {
	var userID int
	roleID, err := FetchRole(role)
	if err != nil {
		logrus.Printf("AddUser: cannot add user:%v", err)
		return userID, err
	}
	if userAndRoleID.RolesId != 0 {
		if userAndRoleID.RolesId != roleID {
			logrus.Printf("AddUser: cannot add another user with same phone number:%v", err)
			return userID, errors.New("cannot add another user with same phone number")
		}
		return userAndRoleID.UserID, nil
	}
	// language=SQL
	SQL := `INSERT INTO users(name, phone_no, roles_id)
            VALUES ($1, $2, $3)
            RETURNING id`
	err = tx.Get(&userID, SQL, name, phone, roleID)
	if err != nil {
		logrus.Printf("AddUser: cannot add user:%v", err)
		return userID, err
	}

	return userID, nil
}

func AddTehsil(tehsil string, tx *sqlx.Tx) (int, error) {
	// language=SQL
	SQL := `INSERT INTO tehsil(name)
            VALUES ($1)
            RETURNING id`

	var tehsilID int

	err := tx.Get(&tehsilID, SQL, tehsil)
	if err != nil {
		logrus.Printf("tehsil: cannot enter tehsil:%v", err)
		return tehsilID, err
	}
	return tehsilID, nil
}

func AddGramPanchayat(gramPanchayat string, tehsilID, blockID int, tx *sqlx.Tx) (int, error) {
	// language=SQL
	SQL := `INSERT INTO gram_panchayat(name, tehsil_id, block_id)
            VALUES  ($1, $2, $3)
            RETURNING id`

	var gramPanchayatID int

	err := tx.Get(&gramPanchayatID, SQL, gramPanchayat, tehsilID, blockID)
	if err != nil {
		logrus.Printf("AddGramPanchayat: cannot add gram panchayat:%v", err)
		return gramPanchayatID, err
	}
	return gramPanchayatID, nil
}

func AddUserTehsil(userID, tehsilID int, tx *sqlx.Tx) error {
	// language=SQL
	SQL := `INSERT INTO user_tehsil(user_id, tehsil_id)
            VALUES ($1, $2)`

	_, err := tx.Exec(SQL, userID, tehsilID)
	if err != nil {
		logrus.Printf("AddUserTehsil: cannot add user_tehsil:%v", err)
		return err
	}
	return nil
}

func AddUserGramPanchayat(userID, gramPanchayatId int, tx *sqlx.Tx) error {
	// language=SQL
	SQL := `INSERT INTO user_gram_panchayat(user_id, gram_panchayat_id)
            VALUES ($1, $2)`

	_, err := tx.Exec(SQL, userID, gramPanchayatId)
	if err != nil {
		logrus.Printf("AddUserGramPanchayat: cannot add UserGramPanchayat:%v", err)
		return err
	}
	return nil
}

func GetUserInfo(userID int) (models.UserInfo, error) {
	SQL := `
			select users.name,
				   users.phone_no,
				   users.id,
				   roles.role,
				   roles.is_district_level as register_death_enabled
			from users
					 join roles on users.roles_id = roles.id
			where users.id = $1
`

	var info models.UserInfo
	err := database.GramPanchayatDB.Get(&info, SQL, userID)
	if err != nil {
		logrus.Printf("GetSdm: cannot get tehsil list: %v", err)
		return info, err
	}
	if info.Role == utilities.SDM {
		SQL := `
			select gp.id,
				   gp.name,
				   json_agg(json_build_object('gaonId',g.id,'gaonName', g.name::text))    as gaon_details
			from users
					 join user_tehsil on users.id = user_tehsil.user_id
					 join gram_panchayat gp on user_tehsil.tehsil_id = gp.tehsil_id
			         join gaon g on gp.id = g.gram_panchayat_id
			where users.id = $1
            GROUP BY gp.id, gp.name
`

		panchayatList := make([]models.PanchayatInfo, 0)
		err := database.GramPanchayatDB.Select(&panchayatList, SQL, userID)
		if err != nil {
			logrus.Printf("GetSdm: cannot get tehsil list: %v", err)
			return info, err
		}
		panchayatOutput := make([]models.PanchayatOutput, 0)
		for i := range panchayatList {
			var out []models.GaonInfo
			err = json.Unmarshal(panchayatList[i].GaonDetails, &out)
			if err != nil {
				logrus.Printf("GetUserInfo: unmarshal error:%v", err)
				return info, err
			}

			panchayatOut := models.PanchayatOutput{
				ID:          panchayatList[i].ID,
				Name:        panchayatList[i].Name,
				GaonDetails: out,
			}
			panchayatOutput = append(panchayatOutput, panchayatOut)

		}
		info.PanchayatList = panchayatOutput
	} else if info.Role == utilities.Sachiv {
		SQL := `
			select gp.id,
				   gp.name,
				   json_agg(json_build_object('gaonId',g.id,'gaonName', g.name::text))    as gaon_details
			from users
					 join user_gram_panchayat ugp on users.id = ugp.user_id
					 join gram_panchayat gp on ugp.gram_panchayat_id = gp.id
					 join gaon g on gp.id = g.gram_panchayat_id

			where users.id = $1
            group by gp.id, gp.name
`
		panchayatList := make([]models.PanchayatInfo, 0)
		err := database.GramPanchayatDB.Select(&panchayatList, SQL, userID)
		if err != nil {
			logrus.Printf("GetSdm: cannot get tehsil list: %v", err)
			return info, err
		}
		panchayatOutput := make([]models.PanchayatOutput, 0)
		for i := range panchayatList {
			var out []models.GaonInfo
			err = json.Unmarshal(panchayatList[i].GaonDetails, &out)
			if err != nil {
				logrus.Printf("GetUserInfo: unmarshal error:%v", err)
				return info, err
			}

			panchayatOut := models.PanchayatOutput{
				ID:          panchayatList[i].ID,
				Name:        panchayatList[i].Name,
				GaonDetails: out,
			}
			panchayatOutput = append(panchayatOutput, panchayatOut)

		}
		info.PanchayatList = panchayatOutput
	} else if info.Role == utilities.LekhPal {
		SQL := `
			select gp.id as id,
			       gp.name as name,
			       json_agg(json_build_object('gaonId',g.id,'gaonName', g.name::text))    as gaon_details
			from  users JOIN user_gaon ug on users.id = ug.user_id
			            JOIN gaon g on g.id = ug.gaon_id
			            JOIN gram_panchayat gp on gp.id = g.gram_panchayat_id
			where users.id = $1
            GROUP BY gp.id, gp.name
`
		panchayatList := make([]models.PanchayatInfo, 0)
		err := database.GramPanchayatDB.Select(&panchayatList, SQL, userID)
		if err != nil {
			logrus.Printf("GetSdm: cannot get tehsil list: %v", err)
			return info, err
		}
		panchayatOutput := make([]models.PanchayatOutput, 0)
		for i := range panchayatList {
			var out []models.GaonInfo
			err = json.Unmarshal(panchayatList[i].GaonDetails, &out)
			if err != nil {
				logrus.Printf("GetUserInfo: unmarshal error:%v", err)
				return info, err
			}

			panchayatOut := models.PanchayatOutput{
				ID:          panchayatList[i].ID,
				Name:        panchayatList[i].Name,
				GaonDetails: out,
			}
			panchayatOutput = append(panchayatOutput, panchayatOut)

		}
		info.PanchayatList = panchayatOutput
	}
	// district level person cannot add new deaths
	info.RegisterDeathEnabled = info.Role == utilities.Sachiv || info.Role == utilities.Sahayak
	return info, nil
}

func GetAdminInfo(userID int) (models.UserInfo, error) {
	SQL := `
			select users.name,
				   users.phone_no,
				   users.id,
				   roles.role
			from users
					 join roles on users.roles_id = roles.id
			where users.id = $1
`

	var info models.UserInfo
	err := database.GramPanchayatDB.Get(&info, SQL, userID)
	if err != nil {
		logrus.Printf("GetSdm: cannot get tehsil list: %v", err)
		return info, err
	}
	return info, nil
}

func GetGramPanchayatID(name string) (int, error) {
	// language = SQL
	SQL := `SELECT id FROM gram_panchayat
		  WHERE name=$1
			AND archived_at IS NULL
`
	var gpID int
	err := database.GramPanchayatDB.Get(&gpID, SQL, name)
	if err != nil {
		logrus.Printf("GetGramPanchayatID: cannot get gram panchayat id:%v", err)
		return 0, err
	}
	return gpID, nil
}

func GetGramPanchayatList(filterCheck *models.FiltersCheck) ([]models.GramPanchayatList, error) {
	SQL := `
		WITH cte_sahayak AS (SELECT ugp.user_id as sahayak_id,
									users.name  as sahayak_name,
									phone_no    as sahayak_phone,
									g.id        as sahayak_gid,
									g.name      as sahayak_gram_panchayat,
									t.name 		as tehsil_name,
									t.id        as tehsil_id
							 FROM users
									  JOIN roles r on r.id = users.roles_id
									  JOIN user_gram_panchayat ugp on users.id = ugp.user_id
									  JOIN gram_panchayat g on g.id = ugp.gram_panchayat_id
							 		  JOIN tehsil t on g.tehsil_id = t.id
							 WHERE r.role = 'Sahayak'
							   AND users.archived_at IS NULL
							   AND g.archived_at IS NULL),
			 cte_sachiv AS (SELECT ugp.user_id as sachiv_id,
								   users.name  as sachiv_name,
								   phone_no    as sachiv_phone,
								   g.id        as sachiv_gid,
		                           b2.name	   as block_name,
		                           b2.id	   as block_id
							FROM users
									 JOIN roles r on r.id = users.roles_id
									 JOIN user_gram_panchayat ugp on users.id = ugp.user_id
									 JOIN gram_panchayat g on g.id = ugp.gram_panchayat_id
							         LEFT JOIN block b2 on b2.id = g.block_id
		
							WHERE r.role = 'Sachiv'
							  AND users.archived_at IS NULL
							  AND g.archived_at IS NULL)
		SELECT cte_sahayak.sahayak_name   as sahayak_name,
		       cte_sahayak.sahayak_id	  as sahayak_id,
			   cte_sachiv.sachiv_name     as sachiv_name,
			   cte_sachiv.sachiv_id	      as sachiv_id,
			   cte_sahayak.sahayak_phone  as sahayak_phone,
			   cte_sachiv.sachiv_phone    as sachiv_phone,
			   sahayak_gram_panchayat     as gram_panchayat_name,
			   sahayak_gid				  as gram_panchayat_id,
			   cte_sahayak.tehsil_name    as tehsil_name,
			   cte_sahayak.tehsil_id      as tehsil_id,
			   COALESCE (cte_sachiv.block_name, '') as block_name,
			   cte_sachiv.block_id					as block_id
		FROM cte_sachiv
				 JOIN cte_sahayak ON cte_sachiv.sachiv_gid = cte_sahayak.sahayak_gid
		WHERE ($1 or sahayak_gram_panchayat ilike '%' || $2 || '%')
		LIMIT $3 OFFSET $4
`

	gramPanchayatList := make([]models.GramPanchayatList, 0)

	err := database.GramPanchayatDB.Select(&gramPanchayatList, SQL, !filterCheck.IsSearched, filterCheck.SearchedName, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)
	if err != nil {
		logrus.Printf("GetGramPanchayatInformation: cannot get gram panchayat list:%v", err)
		return gramPanchayatList, err
	}
	return gramPanchayatList, nil
}

func GetAllTehsil(limit, page int) (models.Tehsils, error) {
	var tehsils models.Tehsils
	TehsilDetail := make([]models.TehsilDetail, 0)
	count := 0
	offset := limit * page

	egp := errgroup.Group{}
	egp.Go(func() error {
		// language = SQL
		SQL := `SELECT id,
                   name as tehsil_name
              FROM tehsil 
              WHERE archived_at IS NULL
              order by name 
              LIMIT $1 OFFSET $2`
		err := database.GramPanchayatDB.Select(&TehsilDetail, SQL, limit, offset)
		if err != nil {
			logrus.Printf("GetAllTehsil: cannot get tehsil names:%v", err)
			return err
		}
		return nil
	})

	egp.Go(func() error {
		// language=sql
		SQL := `SELECT count(*) FROM tehsil where tehsil.archived_at is null`
		err := database.GramPanchayatDB.Get(&count, SQL)
		if err != nil && err != sql.ErrNoRows {
			logrus.Printf("GetTehsils: cannot get tehsils:%v", err)
			return err
		}

		return nil
	})

	err := egp.Wait()
	if err != nil {
		return tehsils, err
	}

	tehsils.Tehsils = TehsilDetail
	tehsils.TotalCount = count
	return tehsils, nil
}

func GetTehsils(filter models.FiltersCheck) (models.Tehsils, error) {
	var count int
	TehsilDetails := make([]models.TehsilDetail, 0)
	var tehsils models.Tehsils
	egp := errgroup.Group{}
	egp.Go(func() error {
		// language = SQL
		SQL := `SELECT t.name as tehsil_name,
                       t.id,
                   u.name,
                   u.id as user_id,
                   u.phone_no as phone
            FROM tehsil t
                join user_tehsil ut 
                      on t.id = ut.tehsil_id
                join users u on ut.user_id = u.id
            WHERE ut.archived_at IS NULL
               AND t.archived_at is NULL
               AND u.archived_at is null
               `
		values := make([]interface{}, 0)
		num := 0
		if filter.SearchedName != "" {
			nameStr := fmt.Sprintf("AND t.name ilike '%%' || $%d || '%%' ", num+1)
			SQL += nameStr
			num++
			values = append(values, filter.SearchedName)
		}

		pagination := fmt.Sprintf("limit $%d offset $%d", num+1, num+2)
		SQL += pagination
		values = append(values, filter.Limit)
		offset := filter.Limit * filter.Page
		values = append(values, offset)

		err := database.GramPanchayatDB.Select(&TehsilDetails, SQL, values...)
		if err != nil {
			logrus.Printf("GetTehsils: cannot get tehsils:%v", err)
			return err
		}

		return nil
	})

	egp.Go(func() error {
		// language=sql
		SQL := `SELECT count(*) FROM tehsil where tehsil.archived_at is null`
		err := database.GramPanchayatDB.Get(&count, SQL)
		if err != nil && err != sql.ErrNoRows {
			logrus.Printf("GetTehsils: cannot get tehsils:%v", err)
			return err
		}

		return nil
	})

	err := egp.Wait()
	if err != nil {
		return tehsils, err
	}

	tehsils.Tehsils = TehsilDetails
	tehsils.TotalCount = count
	return tehsils, nil
}

func GetTasks(searchText string, limit, page int) (models.Tasks, error) {
	TaskDetails := make([]models.TaskDetails, 0)
	offset := limit * page
	var Tasks models.Tasks
	var count int

	egp := errgroup.Group{}
	egp.Go(func() error {
		// language=sql
		SQL := `SELECT tt.name as task_name,
                       tt.id as task_id,
                       r.id as role_id,
                      r.role as role_name
                FROM task_role t
                 join roles r on t.role_id = r.id
                 join task_types tt on t.task_type_id = tt.id
               WHERE t.archived_at is null
                 AND ($3 OR tt.name ilike '%' || $4 || '%')
                 order by t.created_at desc 
               LIMIT $1 OFFSET $2`

		err := database.GramPanchayatDB.Select(&TaskDetails, SQL, limit, offset, searchText == "", searchText)
		if err != nil && err != sql.ErrNoRows {
			logrus.Printf("GetTehsils: cannot get tehsils:%v", err)
			return err
		}

		return nil
	})

	egp.Go(func() error {
		// language=sql
		SQL := `SELECT count(*)
                FROM task_role t
                JOIN task_types tt on t.task_type_id = tt.id
                WHERE t.archived_at is null
                AND ($1 OR tt.name ilike '%' || $2 || '%')`

		err := database.GramPanchayatDB.Get(&count, SQL, searchText == "", searchText)
		if err != nil && err != sql.ErrNoRows {
			logrus.Printf("GetTehsils: cannot get tehsils:%v", err)
			return err
		}

		return nil
	})
	err := egp.Wait()
	if err != nil {
		return Tasks, err
	}

	Tasks.Tasks = TaskDetails
	Tasks.TotalCount = count

	return Tasks, nil
}

func GetDeathCount(month time.Month, year, week int) (models.TotalDeaths, error) {
	//language=SQL
	SQL := `
			SELECT (SELECT count(*)
        			FROM death_details
        			WHERE EXTRACT(MONTH FROM date_of_death) = $1
          				AND EXTRACT(YEAR FROM date_of_death) = $2) AS month,
       				(SELECT count(*)
        			FROM death_details
        			WHERE EXTRACT(WEEK FROM date_of_death) = $3 
        			  AND EXTRACT(YEAR FROM date_of_death) = $2)   AS week,
       				(SELECT count(*)
        			FROM death_details
        			WHERE date_of_death >= now()::DATE)            AS day;
			`

	var DeathCount models.TotalDeaths
	err := database.GramPanchayatDB.Get(&DeathCount, SQL, month, year, week)

	return DeathCount, err
}

func GetDistrictPost() ([]models.DistrictPost, error) {
	//language=SQL
	SQL := `
			select u.name, phone_no, r.role, json_agg(json_build_object('taskName',tt.name)) as task_name
			from users u
         			join roles r on u.roles_id = r.id
         			join task_role tr on r.id = tr.role_id
         			join task_types tt on tr.task_type_id = tt.id
			where r.is_district_level = true
			group by r.role, u.name, phone_no
			`

	DistrictLevelPost := make([]models.DistrictPost, 0)
	err := database.GramPanchayatDB.Select(&DistrictLevelPost, SQL)

	return DistrictLevelPost, err
}

func GetGraphTesting(filter models.DeathFilter) ([]models.GraphDeatils, error) {
	SQL := `SELECT date, coalesce(registered,0) as registered,coalesce(completed,0) as completed
			FROM  (SELECT death_details.created_at::TIMESTAMP::DATE as created_at,
                  count(death_details.id) as registered,
                  count(death_details.id) filter ( where status = 'completed' )as completed

			FROM death_details JOIN gram_panchayat gp on death_details.gram_panchayat_id = gp.id
			WHERE death_details.created_at BETWEEN (now() - '9 days'::interval) AND now()
`
	values := make([]interface{}, 0)
	num := 0
	str := ""

	if len(filter.GramPanchayatID) > 0 {
		gramStr := fmt.Sprintf("AND gram_panchayat_id =ANY($%d) ", num+1)
		SQL += gramStr
		num++
		values = append(values, pq.Array(filter.GramPanchayatID))
		str = `GROUP BY death_details.created_at::TIMESTAMP::DATE) as counter
					  right join
						( select date from
				generate_series(
				now()::date - '9 days'::interval,
				now(),
				INTERVAL '1 day'
				) as date ) as timer on timer.date=counter.created_at
			group by timer.date,registered,completed
			order by date`
	}
	if len(filter.TehsilID) > 0 {
		tehsilStr := fmt.Sprintf("AND tehsil_id =ANY($%d) ", num+1)
		SQL += tehsilStr
		num++
		values = append(values, pq.Array(filter.TehsilID))
		str = `GROUP BY death_details.created_at::TIMESTAMP::DATE) as counter
					  right join
						( select date from
				generate_series(
				now()::date - '9 days'::interval,
				now(),
				INTERVAL '1 day'
				) as date ) as timer on timer.date=counter.created_at
			group by timer.date,registered,completed
			order by date`
	}
	if len(filter.BlockID) > 0 {
		blockStr := fmt.Sprintf("AND block_id =ANY($%d) ", num+1)
		SQL += blockStr
		num++
		values = append(values, pq.Array(filter.BlockID))
		str = `GROUP BY death_details.created_at::TIMESTAMP::DATE) as counter
					  right join
						( select date from
				generate_series(
				now()::date - '9 days'::interval,
				now(),
				INTERVAL '1 day'
				) as date ) as timer on timer.date=counter.created_at
			group by timer.date,registered,completed
			order by date`
	}

	if len(filter.GramPanchayatID) == 0 && len(filter.TehsilID) == 0 && len(filter.BlockID) == 0 {
		str = `GROUP BY death_details.created_at::TIMESTAMP::DATE) as counter
					  right join
						( select date from
				generate_series(
				now()::date - '9 days'::interval,
				now(),
				INTERVAL '1 day'
				) as date ) as timer on timer.date=counter.created_at
			group by timer.date,registered,completed
			order by date`
	}

	SQL += str

	graphDetails := make([]models.GraphDeatils, 0)
	err := database.GramPanchayatDB.Select(&graphDetails, SQL, values...)
	if err != nil {
		logrus.Printf("GetGraph: cannot get registered count:%v", err)
		return graphDetails, err
	}
	return graphDetails, nil
}

func GetGraph() ([]models.GraphDeatils, error) {
	SQL := `SELECT date, coalesce(registered,0) as registered,coalesce(completed,0) as completed
			FROM  (SELECT death_details.created_at::TIMESTAMP::DATE as created_at,
                  count(death_details.id) as registered,
                  count(death_details.id) filter ( where status = 'completed' )as completed

			FROM death_details JOIN gram_panchayat gp on death_details.gram_panchayat_id = gp.id
			WHERE death_details.created_at BETWEEN (now() - '10 days'::interval) AND now()
			GROUP BY death_details.created_at::TIMESTAMP::DATE) as counter
          right join
      		( select date from
          			generate_series(
                     		 now()::date - '10 days'::interval,
                      			now(),
                     			 INTERVAL '1 day'
            ) as date ) as timer on timer.date=counter.created_at
	group by timer.date,registered,completed
	order by date
`
	graphDetails := make([]models.GraphDeatils, 0)
	err := database.GramPanchayatDB.Select(&graphDetails, SQL)
	if err != nil {
		logrus.Printf("GetGraph: cannot get registered count:%v", err)
		return graphDetails, err
	}
	return graphDetails, nil
}

func GetDeathsAdmin(filter models.DeathFilter) ([]models.DeathDetails, error) {
	SQL := `
select id,
       name,
       phone_no,
       age,
       gender,
       aadhar_number,
       created_by,
       address,
       created_at,
       date_of_death,
       gram_panchayat_id,
       gram_panchayat_name,
       tehsil_id,
       tehsil_name,
       block_id,
       block_name,
       gaon_id,
       gaon_name,
       task_details
from (SELECT death_details.id,
             death_details.name,
             death_details.phone_no,
             age,
             gender,
             aadhar_number,
             created_by,
             address,
             death_details.created_at,
             death_details.date_of_death,
             death_details.gram_panchayat_id as gram_panchayat_id,
             gp.name as gram_panchayat_name,
             t2.name as tehsil_name,
             t2.id as tehsil_id,
             b.name as block_name,
             b.id as block_id,
             g.id as gaon_id,
             g.name as gaon_name,
             count(t.status) filter (where t.status = 'completed')  as completed_tasks,
             count(t.status) filter (where t.status = 'new')        as new_tasks,
             count(t.status) filter (where t.status = 'processing') as processing_tasks,
             count(*)                                               as all_tasks,
             json_agg(json_build_object('taskId',t.id::text,'status', t.status::text, 'name',
                                        task_types.name, 'startDate',
                                        t.start_date::DATE, 'completeDate',
                                        t.completed_date::DATE,
                 						'isRejected',t.is_rejected,
                 						'reason',t.reason))    as task_details
      FROM death_details
               JOIN death_details_address dda on death_details.id = dda.death_detail_id
               JOIN address a on a.id = dda.address_id
               JOIN task t on death_details.id = t.death_id
               JOIN task_types on task_types.id = t.task_type_id
      		   JOIN gram_panchayat gp on death_details.gram_panchayat_id = gp.id
               JOIN gaon g on death_details.gaon_id = g.id
      		   JOIN tehsil t2 on gp.tehsil_id = t2.id
      		   JOIN block b on b.id = gp.block_id
      WHERE death_details.archived_at IS NULL
         `

	values := make([]interface{}, 0)
	num := 0

	if len(filter.TaskName) > 0 {
		taskStr := fmt.Sprintf("AND task_types.name =ANY($%d) ", num+1)
		SQL += taskStr
		num++
		values = append(values, pq.StringArray(filter.TaskName))
	}

	if len(filter.TaskID) > 0 {
		taskStr := fmt.Sprintf("AND task_type_id =ANY($%d) AND completed_date IS NULL ", num+1)
		SQL += taskStr
		num++
		values = append(values, pq.Array(filter.TaskID))
	}

	groupByClause := "group by (death_details.id, death_details.name, death_details.phone_no, age, gender, aadhar_number, created_by, address, death_details.created_at, death_details.date_of_death, gp.name, b.id, t2.id, g.id) ORDER BY death_details.name) as death_details WHERE "
	SQL += groupByClause
	statusWhereClause := ""
	if filter.Status == "processing" {
		statusWhereClause = "(completed_tasks != all_tasks AND new_tasks = 0) "
	} else if filter.Status == "completed" {
		statusWhereClause = "completed_tasks = all_tasks "
	} else if filter.Status == "new" {
		statusWhereClause = "new_tasks > 0 "
	} else {
		statusWhereClause = "true "
	}
	SQL += statusWhereClause

	if len(filter.GramPanchayatID) > 0 {
		gramStr := fmt.Sprintf("AND gram_panchayat_id =ANY($%d) ", num+1)
		SQL += gramStr
		num++
		values = append(values, pq.Array(filter.GramPanchayatID))
	}
	if len(filter.GaonID) > 0 {
		gaonStr := fmt.Sprintf("AND gaon_id =ANY($%d) ", num+1)
		SQL += gaonStr
		num++
		values = append(values, pq.Array(filter.GaonID))
	}
	if len(filter.TehsilID) > 0 {
		tehsilStr := fmt.Sprintf("AND tehsil_id =ANY($%d) ", num+1)
		SQL += tehsilStr
		num++
		values = append(values, pq.Array(filter.TehsilID))
	}
	if len(filter.BlockID) > 0 {
		blockStr := fmt.Sprintf("AND block_id =ANY($%d) ", num+1)
		SQL += blockStr
		num++
		values = append(values, pq.Array(filter.BlockID))
	}
	if filter.FromDate.Unix() > 0 {
		nameStr := fmt.Sprintf("AND created_at::date >= $%d::date ", num+1)
		SQL += nameStr
		num++
		values = append(values, filter.FromDate)
	}
	if filter.ToDate.Unix() > 0 {
		nameStr := fmt.Sprintf("AND created_at::date <= $%d::date ", num+1)
		SQL += nameStr
		num++
		values = append(values, filter.ToDate)
	}
	if filter.Search != "" {
		nameStr := fmt.Sprintf("AND (name ilike '%%' || $%d || '%%' OR phone_no ilike '%%' || $%d || '%%' OR address ilike '%%' || $%d || '%%')", num+1, num+1, num+1)
		SQL += nameStr
		num++
		values = append(values, filter.Search)
	}

	orderBy := ""
	if filter.OrderBy == "date" {
		if filter.IsAscending {
			orderBy = "order by death_details.created_at asc "
		} else {
			orderBy = "order by death_details.created_at desc "
		}
	} else if filter.OrderBy == "tasks" {
		if filter.IsAscending {
			orderBy = "order by completed_tasks desc "
		} else {
			orderBy = "order by completed_tasks asc "
		}
	} else {
		orderBy = "order by death_details.created_at desc, completed_tasks asc "
	}

	SQL += orderBy

	pagination := fmt.Sprintf("limit $%d offset $%d", num+1, num+2)
	SQL += pagination
	values = append(values, filter.Limit)
	offset := filter.Limit * filter.Page
	values = append(values, offset)

	deathDetails := make([]models.DeathDetails, 0)

	err := database.GramPanchayatDB.Select(&deathDetails, SQL, values...)
	if err != nil {
		logrus.Printf("GetDeaths: cannot get deaths:%v", err)
		return deathDetails, err
	}
	return deathDetails, nil

}

func EditGramPanchayat(gramPanchayatDetails models.GramUserDetails, tx *sqlx.Tx) error {
	SQL := `UPDATE gram_Panchayat
            SET    name = $1,
                   block_id = $2,
                   tehsil_id = $4
            WHERE  gram_panchayat.id = $3
            AND    archived_at IS NULL `

	_, err := tx.Exec(SQL, gramPanchayatDetails.GramPanchayatName, gramPanchayatDetails.BlockID, gramPanchayatDetails.GramPanchayatID, gramPanchayatDetails.TehsilID)
	if err != nil {
		logrus.Printf("EditGramPanchayat: cannot edit gramPanchayat name:%v", err)
		return err
	}
	return nil
}

func EditTehsil(tehsilDetails models.TehsilUserDetails, tx *sqlx.Tx) error {
	SQL := `UPDATE tehsil
            SET    name = $1
            WHERE  tehsil.id = $2
            AND    archived_at IS NULL `

	_, err := tx.Exec(SQL, tehsilDetails.TehsilName, tehsilDetails.TehsilID)
	if err != nil {
		logrus.Printf("EditGramPanchayat: cannot edit gramPanchayat name:%v", err)
		return err
	}
	return nil
}

func EditUser(name, phoneNo string, userID int, tx *sqlx.Tx) error {
	SQL := `UPDATE  users
           SET     name = $1, 
                   phone_no = $2
           WHERE id = $3
           AND  archived_at IS NULL 
`
	_, err := tx.Exec(SQL, name, phoneNo, userID)
	if err != nil {
		logrus.Printf("EditUser: cannot edit user:%v", err)
		return err
	}
	return nil
}

func AddBlock(block models.BlockDetails) (int, error) {
	// language=SQL
	SQL := `
			INSERT INTO block (name)
			VALUES ($1)
			RETURNING id
			`
	var id int
	err := database.GramPanchayatDB.Get(&id, SQL, block.BlockName)

	return id, err
}
func UpdateBlock(block models.BlockDetails) error {
	// language=SQL
	SQL := `
			UPDATE block 
            SET name=$1
			WHERE id=$2
			AND archived_at IS NULL 
			`
	_, err := database.GramPanchayatDB.Exec(SQL, block.BlockName, block.BlockID)

	return err
}

func GetBlockDetails() ([]models.BlockDetails, error) {
	// language=SQL
	SQL := `
			SELECT id, name
			FROM block
			WHERE archived_at is null
			`

	blockDetails := make([]models.BlockDetails, 0)
	err := database.GramPanchayatDB.Select(&blockDetails, SQL)

	return blockDetails, err
}

func GetDeathReview(filter models.DeathFilter) ([]models.RandomDeathDetails, error) {
	// language=SQL
	SQL := `
select id,
       death_id,
       name,
       phone_no,
       age,
       gender,
       aadhar_number,
       created_by,
       address,
       created_at,
       date_of_death,
       gram_panchayat_id,
       gram_panchayat_name,
       task_details,
       tehsil_id,
       tehsil_name,
       block_id,
       block_name,
       gaon_id,
       gaon_name, 
       is_reviewed,
       comment,
       reviewed_by,
       reviewed_at
from (SELECT death_review.id,
          	 death_details.id as death_id,
             death_details.name,
             death_details.phone_no,
             age,
             gender,
             aadhar_number,
             created_by,
             address,
             death_details.created_at,
             death_details.date_of_death,
             death_details.gram_panchayat_id as gram_panchayat_id,
             gp.name as gram_panchayat_name,
             t2.name as tehsil_name,
             t2.id as tehsil_id,
             b.name as block_name,
             b.id as block_id,
             g.id gaon_id,
             g.name as gaon_name, 
             json_agg(json_build_object('taskId',t.id::text,'status', t.status::text, 'name',
                                        task_types.name, 'startDate',
                                        t.start_date::DATE, 'completeDate',
                                        t.completed_date::DATE,
                 						'isRejected',t.is_rejected,
                 						'reason',t.reason))    as task_details,
          death_review.is_reviewed,
          case when is_reviewed = 'true' then (death_review.comment) end as comment,
          case when is_reviewed = 'true' then (death_review.review_by) end as reviewed_by,
          case when is_reviewed = 'true' then (death_review.reviewed_at) end as reviewed_at
      FROM death_details
          	   JOIN death_review on death_details.id = death_review.death_detail_id
               JOIN death_details_address dda on death_details.id = dda.death_detail_id
               JOIN address a on a.id = dda.address_id
               JOIN task t on death_details.id = t.death_id
               JOIN task_types on task_types.id = t.task_type_id
      		   JOIN gram_panchayat gp on death_details.gram_panchayat_id = gp.id
               JOIN gaon g on g.id = death_details.gaon_id
      		   JOIN tehsil t2 on gp.tehsil_id = t2.id
      		   JOIN block b on b.id = gp.block_id
      WHERE death_details.archived_at IS NULL
			AND death_review.archived_at IS NULL
      `

	values := make([]interface{}, 0)
	num := 0

	if len(filter.GramPanchayatID) > 0 {
		gramStr := fmt.Sprintf("AND gram_panchayat_id =ANY($%d) ", num+1)
		SQL += gramStr
		num++
		values = append(values, pq.Array(filter.GramPanchayatID))
	}
	if len(filter.TehsilID) > 0 {
		tehsilStr := fmt.Sprintf("AND tehsil_id =ANY($%d) ", num+1)
		SQL += tehsilStr
		num++
		values = append(values, pq.Array(filter.TehsilID))
	}
	if len(filter.BlockID) > 0 {
		blockStr := fmt.Sprintf("AND block_id =ANY($%d) ", num+1)
		SQL += blockStr
		num++
		values = append(values, pq.Array(filter.BlockID))
	}
	SQL += `GROUP BY death_details.id, death_details.name, death_details.phone_no, age, gender, aadhar_number, created_by, address, death_details.created_at, death_details.date_of_death, death_details.gram_panchayat_id, gp.name, death_review.created_at, death_review.is_reviewed, death_review.comment, death_review.review_by, death_review.reviewed_at, death_review.id, t2.id, b.id, g.id
			ORDER BY death_review.created_at DESC
			  ) as detail
				 `
	death := make([]models.RandomDeathDetails, 0)
	err := database.GramPanchayatDB.Select(&death, SQL, values...)

	return death, err
}

func GetCompletedDeaths(date time.Time) ([]int, error) {
	// language=SQL
	SQL := `
			SELECT id
			FROM (SELECT death_details.id,
             			 count(t.status) filter (where t.status = 'completed')           as completed_tasks,
             			 count(*)                                                        as all_task,
             			 count(*) filter ( where t.completed_date::DATE = $1 ) as completed_yesterday
      			  from death_details
               			   join task t on death_details.id = t.death_id
      			  group by death_details.id) as death_re
			where completed_tasks = all_task
  			  and completed_yesterday > 0;
`
	deathID := make([]int, 0)
	err := database.GramPanchayatDB.Select(&deathID, SQL, date)

	return deathID, err
}

func BulkInsertDeathReviewDetails(deathIDs []int) error {
	psql := sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)
	sqlQuery := psql.Insert("death_review").Columns("death_detail_id")
	for i, _ := range deathIDs {
		sqlQuery.Values(deathIDs[i])
	}

	SQL, args, err := sqlQuery.ToSql()
	if err != nil {
		logrus.Printf("BulkInsertdeathReviewDetails: not able to create sql string:%v", err)
		return err
	}
	_, err = database.GramPanchayatDB.Exec(SQL, args...)
	if err != nil {
		logrus.Printf("BulkInsertdeathReviewDetails:not able to insert death id's:%v", err)
		return err
	}

	return nil
}

func ReviewDeathDetails(deathDetailsReview models.RandomDeath, userID int) error {
	SQL := `UPDATE death_review
            SET    is_reviewed = true,
                   comment = $1,
                   review_by = $2,
                   reviewed_at = now()
            WHERE  death_detail_id = $3
            AND    id = $4
            `
	_, err := database.GramPanchayatDB.Exec(SQL, deathDetailsReview.ReviewComment, userID, deathDetailsReview.DeathID, deathDetailsReview.ID)
	if err != nil {
		logrus.Printf("ReviewDeathDetails:  cannot rview death:%v", err)
		return err
	}
	return nil
}

func AddGaon(gaon models.GaonDetails, tx *sqlx.Tx) (int, error) {
	// language=SQL
	SQL := `INSERT INTO gaon(name, gram_panchayat_id)
            VALUES ($1, $2)
            RETURNING id`

	var gaonID int

	err := tx.Get(&gaonID, SQL, gaon.GaonName, gaon.GramPanchayatID)
	if err != nil {
		logrus.Printf("AddGaon: cannot enter gaon:%v", err)
		return gaonID, err
	}
	return gaonID, nil
}

func AddUserGaon(userID, gaonID int, tx *sqlx.Tx) error {
	// language=SQL
	SQL := `INSERT INTO user_gaon(user_id, gaon_id)
            VALUES ($1, $2)`

	_, err := tx.Exec(SQL, userID, gaonID)
	if err != nil {
		logrus.Printf("AddUserGaon: cannot add user_gaon:%v", err)
		return err
	}
	return nil
}

func GetGaon() ([]models.GaonDetails, error) {
	SQL := `SELECT gaon.id,
                   u.name as lekhpal_name,
                   u.phone_no,
                   g.name as gram_panchayat_name,
                   g.id as gram_panchayat_id,
                   gaon.name,
                   u.id as lekhpal_id
            FROM   gaon  JOIN gram_panchayat g ON gaon.gram_panchayat_id =g.id 
                         JOIN user_gaon ug ON gaon.id = ug.gaon_id
                         JOIN users u ON ug.user_id = u.id`

	gaonDetails := make([]models.GaonDetails, 0)

	err := database.GramPanchayatDB.Select(&gaonDetails, SQL)
	if err != nil {
		logrus.Printf("GetGaon: not able to get Gaons:%v", err)
		return gaonDetails, err
	}
	return gaonDetails, nil
}

func UpdateGaon(gaon models.GaonDetails) error {
	// language=SQL
	SQL := `
			UPDATE gaon 
            SET name=$1,
                gram_panchayat_id = $2	
			WHERE id=$3
			AND archived_at IS NULL 
			`
	_, err := database.GramPanchayatDB.Exec(SQL, gaon.GaonName, gaon.GramPanchayatID, gaon.ID)

	return err
}
