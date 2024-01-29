package cmd

import (
	"fmt"
	"os"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	databaseName := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	serverPort := os.Getenv("SERVER_PORT")
	err := database.ConnectAndMigrate(host, port, databaseName, user, password, database.SSLModeDisable)
	if err != nil {
		logrus.Printf("ConnectAndMigrate: error is:%v", err)
		return
	}
	fmt.Println("connected")
	srv := server.SetupRoutes()
	go cron.RunCronJob()
	err = srv.Run(fmt.Sprintf(":%s", serverPort))
	if err != nil {
		logrus.Printf("could not run the server:%v", err)
		return
	}
	// log.Printf("check 101 for main")
}
