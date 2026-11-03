package routes

import (
	"net/http"

	"github.com/sashank33v/recallengine/internal/handler"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

func NewRouter(Authhandler *handler.Authhandler, notehandler *handler.NoteHandler, questionhandler *handler.QuestionHandler, studyhandler *handler.StudyHandler, authservice *service.AuthService) http.Handler {
	mux := http.NewServeMux()
	auth := middleware.Auth(authservice)
	mux.HandleFunc("GET /healthz", handler.HealthHandler)
	mux.HandleFunc("POST /signup", Authhandler.Signup)
	mux.HandleFunc("POST /login", Authhandler.Login)

	mux.Handle("GET /me", auth(http.HandlerFunc(Authhandler.Me)))
	mux.Handle("POST /notes", auth(http.HandlerFunc(notehandler.Create)))
	mux.Handle("GET /notes", auth(http.HandlerFunc(notehandler.List)))
	mux.Handle("POST /questions", auth(http.HandlerFunc(questionhandler.Create)))
	mux.Handle("GET /study/due", auth(http.HandlerFunc(studyhandler.Due)))
	mux.Handle("POST /study/answer", auth(http.HandlerFunc(studyhandler.Answer)))
	mux.Handle("GET /questions", auth(http.HandlerFunc(questionhandler.List)))
	return mux
}
