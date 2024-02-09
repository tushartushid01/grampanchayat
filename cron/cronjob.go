package cron

import (
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
	"grampanchayat/database/helper"
	"math"
	"math/rand"
	"time"
)

func RunCronJob() {
	s := gocron.NewScheduler(time.Local)
	_, err := s.Every(1).Day().At("02:00").Do(func() {
		date := time.Now()
		date = date.AddDate(0, 0, -1)
		deathIDs, err := helper.GetCompletedDeaths(date)
		if err != nil {
			logrus.Printf("RunCronJob: unable to get death id's. %v", err)
		}
		random := math.Ceil((float64(len(deathIDs)) * 10) / 100)
		deathId := make([]int, 0)
		for i := 0; i < int(random); i++ {
			randomIndex := rand.Intn(len(deathIDs))
			id := deathIDs[randomIndex]
			deathId = append(deathId, id)
		}
		err = helper.BulkInsertDeathReviewDetails(deathId)
		if err != nil {
			logrus.Printf("RunCronJob: unable to insert death id's. %v", err)
		}
	})
	if err != nil {
		return
	}

	s.StartBlocking()
}
