package main

import (
	"bufio"
	"context"
	"flag"
	"net/http"
	"os"
	"strconv"

	"yeetikuserver/config"
	"yeetikuserver/utils"

	h "yeetikuserver/handler"
	"yeetikuserver/model"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

//DefaultFieldsHook :
type DefaultFieldsHook struct {
}

//Fire :
func (df *DefaultFieldsHook) Fire(entry *log.Entry) error {
	entry.Data["UserName"] = ""
	return nil
}

//Levels :
func (df *DefaultFieldsHook) Levels() []log.Level {
	return log.AllLevels
}

//BasicAuth :
func BasicAuth(h httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Get the Basic Authentication credentials
		jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Config.SecretKey), nil
			},
			SigningMethod: jwt.SigningMethodHS256,
		})
		token, _ := jwtMiddleware.Options.Extractor(r)
		var ctx context.Context
		if len(token) > 0 {
			id := utils.ParseUserProperty(token)
			ctx = utils.SaveUserInfoToContext(r.Context(), id)
		}

		if jwtMiddleware.CheckJWT(rw, r) != nil {
			rw.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			h(rw, r.WithContext(ctx), ps)
		}
	}
}

var runServer = true
var defaultPort = flag.Int("port", 8080, "designated ports")
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
	// n.Use(m.CheckAuthMiddleware())
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
	router.POST("/api/v1/users", BasicAuth(h.SaveUser))
	router.GET("/api/v1/users/:id", h.GetUser)
	router.PUT("/api/v1/users/:id", BasicAuth(h.SaveUser))
	router.POST("/api/v1/users/:id/reset", BasicAuth(h.ResetPasswordUser))
	router.DELETE("/api/v1/users/:id", BasicAuth(h.DeleteUsers))
	router.DELETE("/api/v1/users", BasicAuth(h.DeleteUsers))
	router.PUT("/api/v1/users/:id/avatar", BasicAuth(h.ChangeAvatar))
	router.PUT("/api/v1/users/:id/password", BasicAuth(h.ChangePassword))
	router.GET("/api/v1/users/:id/banks", h.GetUserBanks)
	router.GET("/api/v1/users/:id/exams", h.GetUserExams)
	router.GET("/api/v1/users/:id/records", h.GetUserRecords)
	router.GET("/api/v1/users/:id/favorites", h.GetUserFavorites)
	router.GET("/api/v1/users/:id/bank/:bankid/wrong", h.GetUserWrong)
	router.POST("/api/v1/users/:id/records", BasicAuth(h.InsertRecords))
	router.GET("/api/v1/me", BasicAuth(h.GetCurrentUser))

	router.GET("/api/v1/groups", h.GetGroups)
	router.GET("/api/v1/groups/:id", h.GetGroup)
	router.GET("/api/v1/groups/:id/users", h.GetRelatedUsers)
	router.POST("/api/v1/groups/:id/users", BasicAuth(h.AddRelatedUsers))
	router.POST("/api/v1/groups", BasicAuth(h.SaveGroup))
	router.PUT("/api/v1/groups", BasicAuth(h.SaveGroup))

	router.GET("/api/v1/tags", h.GetTags)
	router.GET("/api/v1/tags/:id", h.GetTag)
	router.PUT("/api/v1/tags", BasicAuth(h.SaveTag))
	router.POST("/api/v1/tags/delete", BasicAuth(h.DeleteTags))

	router.GET("/api/v1/banks", h.GetBanks)
	router.POST("/api/v1/banks", BasicAuth(h.CreateBank))
	router.GET("/api/v1/banks/:id", h.GetBank)
	router.PUT("/api/v1/banks/:id", BasicAuth(h.UpdateBank))
	router.DELETE("/api/v1/banks/:id", BasicAuth(h.RemoveBank))
	router.GET("/api/v1/banks/:id/questions", h.GetRelatedQuestions)
	router.POST("/api/v1/banks/:id/questions", BasicAuth(h.AddRelateQuestions))
	router.POST("/api/v1/banks/:id/tags", BasicAuth(h.SaveRelatedBankTags))
	router.GET("/api/v1/banks/:id/tags", h.GetBankTags)
	router.DELETE("/api/v1/banks/:id/tags", BasicAuth(h.RemoveRelatedTags))
	router.DELETE("/api/v1/banks/:id/questions", BasicAuth(h.RemoveRelatedQuestions))
	router.POST("/api/v1/banks/:id/upload", BasicAuth(h.UploadBankImage))
	router.POST("/api/v1/banks/:id/status", BasicAuth(h.ChangeStatus))
	router.GET("/api/v1/banks/:id/records", h.GetRecords)
	router.GET("/api/v1/banks/:id/user/:userid/records", h.QueryUserRecords)
	router.GET("/api/v1/banktags", h.GetAllBankTags)
	router.GET("/api/v1/banktags/:id/banks", h.GetRelatedBanks)
	router.POST("/api/v1/banktags", BasicAuth(h.SaveBankTags))
	router.DELETE("/api/v1/banktags", BasicAuth(h.DeleteBankTags))

	router.POST("/api/v1/categories", BasicAuth(h.CreateCategory))
	router.DELETE("/api/v1/categories", BasicAuth(h.DeleteCategory))
	router.PUT("/api/v1/categories", BasicAuth(h.UpdateCategory))
	router.GET("/api/v1/categories", h.GetCategories)

	router.GET("/api/v1/questions", h.GetQuestions)
	router.GET("/api/v1/questions/:id", h.GetQuestion)
	router.GET("/api/v1/questions/:id/favorites/:userid", h.IsUserFavorites)
	router.POST("/api/v1/questions/:id/favorites", BasicAuth(h.AddFavorites))
	router.DELETE("/api/v1/questions/:id/favorites", BasicAuth(h.RemoveFavorites))
	router.PUT("/api/v1/questions/:id/category", BasicAuth(h.ChangeCategory))
	router.POST("/api/v1/questions", BasicAuth(h.SaveQuestion))
	router.PUT("/api/v1/questions", BasicAuth(h.SaveQuestion))
	router.DELETE("/api/v1/questions", BasicAuth(h.DeleteQuestion))
	router.GET("/api/v1/questions/:id/comments", h.GetQuestionComments)

	router.POST("/api/v1/import/questions", BasicAuth(h.ImportFromExcel))
	// router.GET("/api/v1/import/questions/users/:id/result", h.ImportQuestionResult)
	router.GET("/api/v1/notification/users/:id/questions/import/", h.GetQuestionImportResult)
	router.DELETE("/api/v1/notification/users/:id/questions/import/", BasicAuth(h.RemoveQuestionImportResult))

	router.GET("/api/v1/feedback", h.GetFeedBacks)
	router.POST("/api/v1/feedback", BasicAuth(h.CreateFeedBack))

	router.GET("/api/v1/client/config", h.GetAppConfig)
	router.POST("/api/v1/client/config", BasicAuth(h.SaveAppConfig))
	router.POST("/api/v1/client/splash/upload", BasicAuth(h.UploadClientSplashImage))
	router.POST("/api/v1/client/icon/upload", BasicAuth(h.UploadClientIconImage))

	router.GET("/api/v1/comments/all", h.GetALlComments)
	router.GET("/api/v1/comments/parent/:id", h.GetChildComments)
	router.PUT("/api/v1/comments/:commentid/users/:userid/like", BasicAuth(h.LikeComments))
	router.PUT("/api/v1/comments/:commentid/users/:userid/dislike", BasicAuth(h.DislikeComments))
	router.DELETE("/api/v1/comments", BasicAuth(h.DeleteComments))
	router.POST("/api/v1/comments", BasicAuth(h.CreateComments))

	router.GET("/api/v1/exams", h.GetUserExams)
	router.GET("/api/v1/exams/:id", h.GetExam)
	router.POST("/api/v1/exams", BasicAuth(h.CreateExam))
	router.PUT("/api/v1/exams/:id/score", BasicAuth(h.UpdateExamScore))

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
