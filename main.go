package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"
	"strconv"

	"yeetikuserver/config"

	h "yeetikuserver/handler"
	m "yeetikuserver/middleware"
	"yeetikuserver/model"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type DefaultFieldsHook struct {
}

func (df *DefaultFieldsHook) Fire(entry *log.Entry) error {
	entry.Data["UserName"] = ""
	return nil
}

func (df *DefaultFieldsHook) Levels() []log.Level {
	return log.AllLevels
}

var runServer = true
var defaultPort = flag.Int("port", 0, "designated ports")
var resetAdminPassword = flag.String("reset-admin-password", "", "reset admin password")
var configFile = flag.String("config", "", "specified config file")

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.AddHook(&DefaultFieldsHook{})

	file, err := os.OpenFile("./logs/log.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	handleArguments()

	n := negroni.Classic()
	n.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		Debug:            true,
	}))
	n.Use(m.CheckAuthMiddleware())
	//n.Use(m.HeaderMiddleware())

	router := httprouter.New()
	router.GET("/home", h.Home)
	router.ServeFiles("/static/*filepath", http.Dir("./upload/img/"))
	router.ServeFiles("/download/*filepath", http.Dir("./download/"))
	router.ServeFiles("/assets/*filepath", http.Dir("./assets/"))
	router.POST("/api/v1/auth/admin/login", h.AdminLogin)
	router.POST("/api/v1/auth/login", h.Login)
	router.GET("/api/v1/auth/logout", h.Logout)
	router.POST("/api/v1/auth/register", h.Register)

	router.GET("/api/v1/users", h.GetUsers)
	router.POST("/api/v1/users", h.SaveUser)
	router.GET("/api/v1/users/:id", h.GetUser)
	router.PUT("/api/v1/users", h.SaveUser)
	router.PUT("/api/v1/users/:id", h.UpdateUser)
	router.POST("/api/v1/users/:id/reset", h.ResetPasswordUser)
	router.DELETE("/api/v1/users/:id", h.DeleteUsers)
	router.DELETE("/api/v1/users", h.DeleteUsers)
	router.PUT("/api/v1/users/:id/avatar", h.ChangeAvatar)
	router.PUT("/api/v1/users/:id/password", h.ChangePassword)
	router.GET("/api/v1/users/:id/banks", h.GetUserBanks)
	router.GET("/api/v1/users/:id/records", h.GetUserRecords)
	router.GET("/api/v1/users/:id/favorites", h.GetUserFavorites)
	router.GET("/api/v1/users/:id/wrong", h.GetUserWrong)
	router.POST("/api/v1/users/:id/records", h.InsertRecords)
	router.GET("/api/v1/me", h.GetCurrentUser)

	router.GET("/api/v1/groups", h.GetGroups)
	router.GET("/api/v1/groups/:id", h.GetGroup)
	router.GET("/api/v1/groups/:id/users", h.GetRelatedUsers)
	router.POST("/api/v1/groups/:id/users", h.AddRelatedUsers)
	router.POST("/api/v1/groups", h.SaveGroup)
	router.PUT("/api/v1/groups", h.SaveGroup)

	router.GET("/api/v1/tags", h.GetTags)
	router.GET("/api/v1/tags/:id", h.GetTag)
	router.PUT("/api/v1/tags", h.SaveTag)
	router.POST("/api/v1/tags/delete", h.DeleteTags)

	router.GET("/api/v1/banks", h.GetBanks)
	router.POST("/api/v1/banks", h.CreateBank)
	router.GET("/api/v1/banks/:id", h.GetBank)
	router.PUT("/api/v1/banks/:id", h.UpdateBank)
	router.DELETE("/api/v1/banks/:id", h.RemoveBank)
	router.GET("/api/v1/banks/:id/questions", h.GetRelatedQuestions)
	router.POST("/api/v1/banks/:id/questions", h.AddRelateQuestions)
	router.POST("/api/v1/banks/:id/tags", h.SaveRelatedBankTags)
	router.GET("/api/v1/banks/:id/tags", h.GetBankTags)
	router.DELETE("/api/v1/banks/:id/tags", h.RemoveRelatedTags)
	router.DELETE("/api/v1/banks/:id/questions", h.RemoveRelatedQuestions)
	router.POST("/api/v1/banks/:id/upload", h.UploadBankImage)
	router.POST("/api/v1/banks/:id/status", h.ChangeStatus)
	// router.POST("/api/v1/banks/:id/remove", h.RemoveBank)
	router.GET("/api/v1/banks/:id/records", h.GetRecords)
	router.GET("/api/v1/banks/:id/user/:userid/records", h.QueryUserRecords)
	router.GET("/api/v1/banktags", h.GetAllBankTags)
	router.GET("/api/v1/banktags/:id/banks", h.GetRelatedBanks)
	router.POST("/api/v1/banktags", h.SaveBankTags)
	router.DELETE("/api/v1/banktags", h.DeleteBankTags)

	router.POST("/api/v1/categories", h.CreateCategory)
	router.DELETE("/api/v1/categories", h.DeleteCategory)
	router.PUT("/api/v1/categories", h.UpdateCategory)
	router.GET("/api/v1/categories", h.GetCategories)

	router.GET("/api/v1/questions", h.GetQuestions)
	router.GET("/api/v1/questions/:id", h.GetQuestion)
	router.GET("/api/v1/questions/:id/favorites/:userid", h.IsUserFavorites)
	router.POST("/api/v1/questions/:id/favorites", h.AddFavorites)
	router.DELETE("/api/v1/questions/:id/favorites", h.RemoveFavorites)
	router.PUT("/api/v1/questions/:id/category", h.ChangeCategory)
	router.POST("/api/v1/questions", h.SaveQuestion)
	router.PUT("/api/v1/questions", h.SaveQuestion)
	router.DELETE("/api/v1/questions", h.DeleteQuestion)

	router.POST("/api/v1/import/questions", h.ImportFromExcel)
	// router.GET("/api/v1/import/questions/users/:id/result", h.ImportQuestionResult)
	router.GET("/api/v1/notification/users/:id/questions/import/", h.GetQuestionImportResult)
	router.DELETE("/api/v1/notification/users/:id/questions/import/", h.RemoveQuestionImportResult)
	router.POST("/api/v1/feedback", h.CreateFeedBack)

	router.GET("/api/v1/client/config", h.GetAppConfig)
	router.POST("/api/v1/client/config", h.SaveAppConfig)
	router.POST("/api/v1/client/splash/upload", h.UploadClientSplashImage)
	router.POST("/api/v1/client/icon/upload", h.UploadClientIconImage)
	n.UseHandler(router)
	//defer  db.GetInstance().Close()
	if runServer {
		n.Run(":" + strconv.Itoa(*defaultPort))
	}
}

func handleArguments() {
	flag.Parse()

	if len(*configFile) > 0 {
		config.ParseConfig(*configFile)
	} else {
		config.ParseConfig("./config/config.json")
	}
	//读取完配置文件一定要先配置model
	model.InitDatabaseTable()
	var cfg = config.GetConfig()

	if len(*resetAdminPassword) > 0 {
		u := model.User{}
		err := u.ResetPassword(cfg.AdminAccount, *resetAdminPassword)
		if err != nil {
			panic("cannot reset password " + err.Error())
		}
		runServer = false
	}

	if *defaultPort < 1 {
		defaultPort = &cfg.Port
	}
}

func makeReader() func(string) string {
	reader := bufio.NewReader(os.Stdin)

	return func(s string) string {
		txt, _ := reader.ReadString('\n')
		return txt
	}
}
