package handler

import (
	"errors"
	"net/http"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/appraisal-crm/notification-service/internal/httputil"
	"github.com/appraisal-crm/notification-service/internal/middleware"
	"github.com/appraisal-crm/notification-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// List returns all notifications for the authenticated recipient.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifications, err := h.svc.ListUserNotifications(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, notifications)
}

// GetByID returns a single notification.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	n, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "notification not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "failed to get notification")
		return
	}

	if n.RecipientID != nil && *n.RecipientID != userID {
		httputil.RespondError(w, http.StatusForbidden, "forbidden")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, n)
}

// MarkRead marks a notification as read by the recipient.
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "notification not found")
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			httputil.RespondError(w, http.StatusForbidden, "forbidden")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
