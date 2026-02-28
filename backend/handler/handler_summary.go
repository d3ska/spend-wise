package handler

import (
	"net/http"
	"time"

	"backend/auth"
	"backend/model"
	"backend/service"
)

// SummaryHandler handles the spending summary HTTP endpoint.
type SummaryHandler struct {
	svc *service.SummaryService
}

// NewSummaryHandler creates a SummaryHandler.
func NewSummaryHandler(svc *service.SummaryService) *SummaryHandler {
	return &SummaryHandler{svc: svc}
}

// GetSummary handles GET /api/v1/workspaces/{id}/summary?from=&to=.
func (h *SummaryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid 'from' date, expected YYYY-MM-DD")
		return
	}

	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid 'to' date, expected YYYY-MM-DD")
		return
	}
	// The SQL queries use half-open intervals (date < $to), so add 1 day
	// to include the end date the caller specified.
	to = to.AddDate(0, 0, 1)

	summary, err := h.svc.GetSummary(r.Context(), userID, model.WorkspaceID(wsID), from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, summaryResponse(summary))
}

func summaryResponse(s model.SpendingSummary) map[string]any {
	byCategory := make([]map[string]any, 0, len(s.ByCategory))
	for _, c := range s.ByCategory {
		entry := map[string]any{
			"category_id": c.CategoryID,
			"name":        c.Name,
			"icon":        c.Icon,
			"slug":        c.Slug,
			"spent": map[string]any{
				"amount":   c.Spent.Amount().StringFixed(2),
				"currency": c.Spent.Currency(),
			},
			"prev_spent": map[string]any{
				"amount":   c.PrevSpent.Amount().StringFixed(2),
				"currency": c.PrevSpent.Currency(),
			},
			"budget": nil,
		}
		if c.Budget != nil {
			entry["budget"] = map[string]any{
				"amount":   c.Budget.Amount().StringFixed(2),
				"currency": c.Budget.Currency(),
			}
		}
		byCategory = append(byCategory, entry)
	}

	byParticipant := make([]map[string]any, 0, len(s.ByParticipant))
	for _, p := range s.ByParticipant {
		byParticipant = append(byParticipant, map[string]any{
			"user_id":      p.UserID,
			"display_name": p.DisplayName,
			"spent": map[string]any{
				"amount":   p.Spent.Amount().StringFixed(2),
				"currency": p.Spent.Currency(),
			},
			"funded": map[string]any{
				"amount":   p.Funded.Amount().StringFixed(2),
				"currency": p.Funded.Currency(),
			},
			"balance": map[string]any{
				"amount":   p.Balance.Amount().StringFixed(2),
				"currency": p.Balance.Currency(),
			},
		})
	}

	var overallBudget any
	if s.OverallBudget != nil {
		overallBudget = map[string]any{
			"amount":   s.OverallBudget.Amount().StringFixed(2),
			"currency": s.OverallBudget.Currency(),
		}
	}

	dailyTrend := make([]map[string]any, 0, len(s.DailyTrend))
	for _, d := range s.DailyTrend {
		dailyTrend = append(dailyTrend, map[string]any{
			"date": d.Date.Format("2006-01-02"),
			"spent": map[string]any{
				"amount":   d.Spent.Amount().StringFixed(2),
				"currency": d.Spent.Currency(),
			},
		})
	}

	return map[string]any{
		"workspace_id":      s.WorkspaceID,
		"from":              s.From.Format("2006-01-02"),
		"to":                s.To.AddDate(0, 0, -1).Format("2006-01-02"),
		"transaction_count": s.TransactionCount,
		"total_spent": map[string]any{
			"amount":   s.TotalSpent.Amount().StringFixed(2),
			"currency": s.TotalSpent.Currency(),
		},
		"prev_transaction_count": s.PrevTransactionCount,
		"prev_total_spent": map[string]any{
			"amount":   s.PrevTotalSpent.Amount().StringFixed(2),
			"currency": s.PrevTotalSpent.Currency(),
		},
		"overall_budget": overallBudget,
		"by_category":    byCategory,
		"by_participant": byParticipant,
		"daily_trend":    dailyTrend,
	}
}
