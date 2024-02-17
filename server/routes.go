package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"grampanchayat/handler"
	"net/http"
)

type Server struct {
	chi.Router
}

func SetupRoutes() *Server {
	router := chi.NewRouter()
	router.Use(middleware.CommonMiddlewares()...)

	router.Route("/health", func(r chi.Router) {
		r.Get("/api", func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintf(w, "Remote State Team")
			if err != nil {
				return
			}
		})
	})

	router.Route("/gram-panchayat", func(gramPanchayat chi.Router) {
		gramPanchayat.Post("/send-otp", handler.SendOTP)
		gramPanchayat.Post("/verify-otp", handler.LoginWithOTP)
		gramPanchayat.Route("/user", func(user chi.Router) {
			user.Use(middleware.AuthMiddleware)
			//TODO send grampanchayats(id and name) under this person
			user.Get("/info", handler.GetUserInfo)
			user.Route("/death", func(death chi.Router) {
				death.Post("/register", handler.DeathRegistration)
				death.Get("/new", handler.GetDeathsNew)
				death.Get("/processing", handler.GetDeathsProcessing)
				death.Get("/completed", handler.GetDeathsCompleted)
				death.Route("/{taskID}", func(task chi.Router) {
					//TODO user can deny that this task does not need to be done.
					//Task need to be completed in case of no, but the reason also need to be stored
					//and also we need to know that it was a no, not a general complete
					//Also for every task store processing start date
					task.Put("/start-processing", handler.ProcessingTask)
					task.Put("/completed", handler.MarkCompleted)
				})
			})

			user.Route("/admin", func(admin chi.Router) {
				admin.Use(middleware.AdminMiddleware)
				//admin.Get("/", handler.GetDeathDetails)
				admin.Get("/graph", handler.GetGraph)
				admin.Get("/info", handler.GetAdminInfo)
				admin.Post("/role", handler.AddRole)
				admin.Get("/role", handler.AddRole)
				admin.Post("/bulk-role", handler.BulkAddRole)

				admin.Get("/tehsils", handler.GetTehsils)
				admin.Get("/all-tehsil", handler.GetTehsilList)
				admin.Post("/tehsil", handler.AddSdm)
				admin.Get("/tasks", handler.GetTasks)

				admin.Post("/gram-panchayat-information", handler.AddGramPanchayatInformation)
				admin.Get("/gram-panchayat-information", handler.GetGramPanchayatInformation)

				admin.Get("/district-post", handler.GetDistrictPostInfo)
				admin.Get("/total-deaths", handler.GetTotalDeaths)

				admin.Get("/deaths", handler.GetDeathDetailsAdmin)

				admin.Put("/edit-gram-panchayat", handler.EditGramPanchayat)
				admin.Put("/edit-tehsil", handler.EditTehsil)
				admin.Post("/block", handler.AddBlock)
				admin.Put("/block", handler.EditBlock)
				admin.Get("/block", handler.GetBlock)
				admin.Get("/death-review", handler.FetchDeathReview)
				admin.Put("/death-review", handler.ReviewDeathDetails)

				admin.Post("/gaon", handler.AddGaon)
				admin.Get("/gaon", handler.GetGaon)
				admin.Put("/gaon", handler.EditGaon)
			})
		})
	})

	return &Server{router}
}

func (svc *Server) Run(port string) error {
	return http.ListenAndServe(port, svc)
}
