package routes

import (
	"net/http"
	"time"

	"github.com/sashank33v/recallengine/internal/handler"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
	"github.com/sashank33v/recallengine/web"
)

func NewRouter(Authhandler *handler.Authhandler, notehandler *handler.NoteHandler, questionhandler *handler.QuestionHandler, coachhandler *handler.CoachHandler, searchhandler *handler.SearchHandler, tutorhandler *handler.TutorHandler, conversationhandler *handler.ConversationHandler, studyhandler *handler.StudyHandler, masteryhandler *handler.MasteryHandler, generationhandler *handler.GenerationHandler, authservice *service.AuthService) http.Handler {
	mux := http.NewServeMux()
	auth := middleware.Auth(authservice)
	ratelimit := middleware.NewRateLimiter(10, time.Minute)
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
	mux.Handle("GET /mastery", auth(http.HandlerFunc(masteryhandler.List)))
	mux.Handle("POST /notes/{id}/generate", auth(ratelimit.Limit(http.HandlerFunc(generationhandler.Generate))))
	mux.Handle("POST /questions/{id}/approve", auth(http.HandlerFunc(questionhandler.Approve)))
	mux.Handle("POST /questions/{id}/reject", auth(http.HandlerFunc(questionhandler.Reject)))
	mux.Handle("PUT /questions/{id}", auth(http.HandlerFunc(questionhandler.Update)))
	mux.Handle("GET /search", auth(http.HandlerFunc(searchhandler.Search)))
	mux.Handle("POST /ask", auth(ratelimit.Limit(http.HandlerFunc(tutorhandler.Ask))))
	mux.Handle("POST /chat", auth(ratelimit.Limit(http.HandlerFunc(tutorhandler.Chat))))
	mux.Handle("GET /conversations", auth(http.HandlerFunc(conversationhandler.List)))
	mux.Handle("GET /conversations/{id}/messages", auth(http.HandlerFunc(conversationhandler.Messages)))
	mux.Handle("POST /coach", auth(ratelimit.Limit(http.HandlerFunc(coachhandler.Ask))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, web.FS, "index.html")
	})
	return mux
}
