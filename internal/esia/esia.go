package esia

import (
	"net/http"

	"github.com/vibe-gaming/backend/internal/esia/handler"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func RunMockServer() {
	h := handler.New()

	// ESIA OAuth2 endpoints
	http.HandleFunc("/aas/oauth2/ac", h.Authorize)
	http.HandleFunc("/aas/oauth2/authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.AuthorizeSubmit(w, r)
		} else {
			h.Authorize(w, r)
		}
	})
	http.HandleFunc("/aas/oauth2/te", h.Token)
	http.HandleFunc("/rs/prns/", h.GetPerson)
	http.HandleFunc("/userinfo", h.UserInfo)

	port := "8085"

	addr := ":" + port
	logger.Info("ESIA Mock Server started", zap.String("addr", addr))

	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
