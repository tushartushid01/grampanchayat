package helper

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"grampanchayat/database"
	"grampanchayat/models"
)

func DeathRegistration(deathDetails models.DeathRegistrationRequest, createdBy int, tx *sqlx.Tx) (int, error) {
	SQL := `INSERT INTO death_details(name, phone_no, age, gender, aadhar_number, status, created_by, gram_panchayat_id, date_of_death, gaon_id)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
            RETURNING id`

	var deathID int

	err := tx.Get(&deathID, SQL, deathDetails.Name, deathDetails.PhoneNo, deathDetails.Age, deathDetails.Gender, deathDetails.AadharNumber, "new", createdBy, deathDetails.PanchayatID, deathDetails.DateOfDeath, deathDetails.GaonID)
	if err != nil {
		logrus.Printf("DeathRegistration: cannot register death:%v", err)
		return deathID, err
	}

	SQL = `insert into task (death_id, task_type_id, status)
			select $1, id, 'new' from task_types`

	_, err = tx.Exec(SQL, deathID)
	if err != nil {
		logrus.Printf("DeathRegistration: cannot register tasks:%v", err)
		return deathID, err
	}

	return deathID, nil
}

func AddAddress(address string, tx *sqlx.Tx) (int, error) {
	SQL := `INSERT INTO address(address)
             VALUES ($1)
             RETURNING id`

	var addressID int

	err := tx.Get(&addressID, SQL, address)
	if err != nil {
		logrus.Printf("AddAddress: cannot add address:%v", err)
		return addressID, err
	}
	return addressID, nil
}

func AddDeathAddress(deathID, addressID int, tx *sqlx.Tx) error {
	SQL := `INSERT INTO death_details_address(death_detail_id, address_id)
            VALUES ($1, $2)`

	_, err := tx.Exec(SQL, deathID, addressID)
	if err != nil {
		logrus.Printf("AddDeathAddress: cannot add death address:%v", err)
		return err
	}
	return nil
}

func GetDeathsNew(taskTypes []string, status string, userId int, search string) ([]models.DeathDetails, error) {
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
             gaon.id as gaon_id,
             gaon.name as gaon_name,
             count(t.status) filter (where t.status = 'completed')  as completed_tasks,
             count(t.status) filter (where t.status = 'new')        as new_tasks,
             count(t.status) filter (where t.status = 'processing') as processing_tasks,
             count(*)                                               as all_tasks,
             json_agg(json_build_object('taskId', t.id::text, 'status', t.status::text, 'name',
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
               JOIN users on users.id = $2
               JOIN roles r on users.roles_id = r.id
               LEFT JOIN user_gram_panchayat
                         on death_details.gram_panchayat_id = user_gram_panchayat.gram_panchayat_id and
                            users.id = user_gram_panchayat.user_id
               JOIN gram_panchayat gp on death_details.gram_panchayat_id = gp.id
               JOIN gaon on death_details.gaon_id = gaon.id
               LEFT JOIN user_tehsil on gp.tehsil_id = user_tehsil.tehsil_id and
                                        users.id = user_tehsil.user_id
      		   LEFT JOIN user_gaon ug on gaon.id = ug.gaon_id and users.id = ug.user_id
      WHERE death_details.archived_at IS NULL
        and task_types.name = any ($1)
        and (r.is_district_level OR user_gram_panchayat.id is not null OR user_tehsil.id is not null OR ug.id is not null)
      group by (death_details.id,
                death_details.name,
                death_details.phone_no,
                age,
                gender,
                aadhar_number,
                created_by,
                address,
                death_details.created_at,
                death_details.date_of_death,
               gaon.id)
      ORDER BY death_details.name) as death_details
	  WHERE %s
`

	statusWhereClause := ""
	if status == "processing" {
		statusWhereClause = "(completed_tasks != all_tasks AND new_tasks = 0) "
	} else if status == "completed" {
		statusWhereClause = "completed_tasks = all_tasks "
	} else if status == "new" {
		statusWhereClause = "new_tasks > 0 "
	}

	SQL = fmt.Sprintf(SQL, statusWhereClause)
	deathDetails := make([]models.DeathDetails, 0)

	values := make([]interface{}, 0)
	values = append(values, pq.StringArray(taskTypes), userId)
	num := 2
	if search != "" {
		nameStr := fmt.Sprintf("AND (name ilike '%%' || $%d || '%%' OR phone_no ilike '%%' || $%d || '%%' OR address ilike '%%' || $%d || '%%')", num+1, num+1, num+1)
		SQL += nameStr
		num++
		values = append(values, search)
	}
	err := database.GramPanchayatDB.Select(&deathDetails, SQL, values...)
	if err != nil {
		logrus.Printf("GetDeaths: cannot get deaths:%v", err)
		return deathDetails, err
	}
	return deathDetails, nil
}

func MarkProcessing(taskID int) error {
	SQL := `UPDATE task 
            SET    status = $1,
                   start_date = now()
            WHERE  id = $2
            AND    archived_at IS NULL `

	_, err := database.GramPanchayatDB.Exec(SQL, "processing", taskID)
	if err != nil {
		logrus.Printf("MarkProcessing: cannot update status to processing:%v", err)
		return err
	}
	return nil
}

func MarkCompleted(taskID int) error {
	SQL := `UPDATE task 
            SET    status = $1,
                   completed_date = now()
            WHERE  id = $2
            AND    archived_at IS NULL `

	_, err := database.GramPanchayatDB.Exec(SQL, "completed", taskID)
	if err != nil {
		logrus.Printf("MarkCompleted: cannot update status to completed:%v", err)
		return err
	}
	return nil
}

func TaskRejected(taskID int, processing models.Processing) error {
	SQL := `UPDATE task 
            SET    status = $1,
                   is_rejected = true,
                   reason = $2,
                   completed_date = now()
            WHERE  id = $3
            AND    archived_at IS NULL `

	_, err := database.GramPanchayatDB.Exec(SQL, "completed", processing.Reason, taskID)
	if err != nil {
		logrus.Printf("TaskRejected: cannot update status to completed:%v", err)
		return err
	}
	return nil
}
