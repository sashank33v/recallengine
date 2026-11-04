package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sashank33v/recallengine/internal/config"
	"github.com/sashank33v/recallengine/internal/handler"
	"github.com/sashank33v/recallengine/internal/platform/ai"
	"github.com/sashank33v/recallengine/internal/platform/postgres"
	"github.com/sashank33v/recallengine/internal/repository"
	"github.com/sashank33v/recallengine/internal/routes"
	"github.com/sashank33v/recallengine/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration: %w", err)
	}
	fmt.Printf("Starting %s....\n", cfg.AppName)

	db, err := postgres.ConnectDB(cfg)
	if err != nil {
		log.Fatal("Failed to Connect database:", err)
	}
	defer db.Close()

	fmt.Println("database connected succesfully!")
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	masteryRepo := repository.NewMasteryRepository(db)
	generationRepo := repository.NewGenerationRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	authService := service.NewAuthService(userRepo, sessionRepo)
	noteService := service.NewNoteService(noteRepo)
	questionService := service.NewQuestionService(questionRepo)
	studyService := service.NewStudyService(questionRepo)
	masteryService := service.NewMasterService(masteryRepo)
	generationService := service.NewGenerationService(generationRepo)
	searchService := service.NewSearchService(chunkRepo)
	aiClient := ai.NewClient(cfg.GroqAPIKey, cfg.GroqModel)
	tutorService := service.NewTutorService(searchService, aiClient)
	authHandler := handler.NewAuthHandler(authService)
	noteHandler := handler.NewNoteHandler(noteService)
	questionHandler := handler.NewQuestionHandler(questionService)
	studyHandler := handler.NewStudyHandler(studyService)
	masteryHandler := handler.NewMasterHandler(masteryService)
	generationHandler := handler.NewGenerationHandler(generationService)
	searchHandler := handler.NewSearchHandler(searchService)
	tutorHandler := handler.NewTutorHandler(tutorService)
	router := routes.NewRouter(authHandler, noteHandler, questionHandler, searchHandler, tutorHandler, studyHandler, masteryHandler, generationHandler, authService)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	go func() {
		fmt.Printf("Server running on http://localhost:%s\n", cfg.Port)
		err = http.ListenAndServe(":"+cfg.Port, router)
		if err != nil {
			log.Fatal("server error:", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down the server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Println("shutdown error:", err)
	}
	fmt.Println("Server stopped")

}
