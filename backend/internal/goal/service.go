package goal

import (
	"context"
	"strings"
	"time"
)

// Service coordinates business logic for savings goals and month-end sweeps.
type Service struct {
	repo Repository
}

// NewService constructs a goal Service instance.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput defines parameters required to create a savings goal.
type CreateInput struct {
	Title              string
	TargetPaise        int64
	Deadline           *string
	MonthlyAmountPaise *int64
	Priority           *int
}

// UpdateInput defines parameters required to edit an existing savings goal.
type UpdateInput struct {
	Title              string
	TargetPaise        int64
	Deadline           *string
	MonthlyAmountPaise *int64
	Priority           int
	Status             string
}

// CreateGoal validates input rules and persists a new savings goal.
func (s *Service) CreateGoal(ctx context.Context, userID string, input CreateInput) (GoalResponse, error) {
	title := strings.TrimSpace(input.Title)
	if len(title) < 1 || len(title) > 60 {
		return GoalResponse{}, ErrInvalidTitle
	}

	if input.TargetPaise <= 0 {
		return GoalResponse{}, ErrInvalidTarget
	}

	// Validate deadline and monthly_amount_paise constraints
	hasDeadline := input.Deadline != nil && strings.TrimSpace(*input.Deadline) != ""
	hasMonthlyAmount := input.MonthlyAmountPaise != nil && *input.MonthlyAmountPaise > 0

	if hasDeadline && hasMonthlyAmount {
		return GoalResponse{}, ErrConflictingTimeline
	}

	var deadlineStr *string
	if hasDeadline {
		trimmedDeadline := strings.TrimSpace(*input.Deadline)
		if _, err := time.Parse("2006-01-02", trimmedDeadline); err != nil {
			return GoalResponse{}, ErrInvalidDeadline
		}
		deadlineStr = &trimmedDeadline
	}

	if input.MonthlyAmountPaise != nil && *input.MonthlyAmountPaise <= 0 {
		return GoalResponse{}, ErrInvalidMonthlyAmount
	}

	// Priority resolution: if omitted or <= 0, use (highest existing priority + 1)
	priority := 1
	if input.Priority != nil && *input.Priority > 0 {
		priority = *input.Priority
	} else {
		maxP, err := s.repo.GetMaxPriority(ctx, userID)
		if err != nil {
			return GoalResponse{}, err
		}
		priority = maxP + 1
	}

	goal := Goal{
		UserID:             userID,
		Title:              title,
		TargetPaise:        input.TargetPaise,
		SavedPaise:         0,
		Deadline:           deadlineStr,
		MonthlyAmountPaise: input.MonthlyAmountPaise,
		Priority:           priority,
		Status:             "active",
	}

	created, err := s.repo.CreateGoal(ctx, goal)
	if err != nil {
		return GoalResponse{}, err
	}

	return s.toResponse(created, time.Now().UTC()), nil
}

// ListGoals retrieves all goals for a user and computes dynamic progress metrics.
func (s *Service) ListGoals(ctx context.Context, userID string) ([]GoalResponse, error) {
	goals, err := s.repo.ListGoals(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	responses := make([]GoalResponse, 0, len(goals))
	for _, g := range goals {
		responses = append(responses, s.toResponse(g, now))
	}

	return responses, nil
}

// UpdateGoal edits title, target_paise, deadline, monthly_amount_paise, priority, and status.
func (s *Service) UpdateGoal(ctx context.Context, userID, goalID string, input UpdateInput) (GoalResponse, error) {
	title := strings.TrimSpace(input.Title)
	if len(title) < 1 || len(title) > 60 {
		return GoalResponse{}, ErrInvalidTitle
	}

	if input.TargetPaise <= 0 {
		return GoalResponse{}, ErrInvalidTarget
	}

	if input.Priority < 1 {
		return GoalResponse{}, ErrInvalidPriority
	}

	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status != "active" && status != "paused" && status != "completed" {
		return GoalResponse{}, ErrInvalidStatus
	}

	hasDeadline := input.Deadline != nil && strings.TrimSpace(*input.Deadline) != ""
	hasMonthlyAmount := input.MonthlyAmountPaise != nil && *input.MonthlyAmountPaise > 0

	if hasDeadline && hasMonthlyAmount {
		return GoalResponse{}, ErrConflictingTimeline
	}

	var deadlineStr *string
	if hasDeadline {
		trimmedDeadline := strings.TrimSpace(*input.Deadline)
		if _, err := time.Parse("2006-01-02", trimmedDeadline); err != nil {
			return GoalResponse{}, ErrInvalidDeadline
		}
		deadlineStr = &trimmedDeadline
	}

	if input.MonthlyAmountPaise != nil && *input.MonthlyAmountPaise <= 0 {
		return GoalResponse{}, ErrInvalidMonthlyAmount
	}

	// Verify goal exists and belongs to user
	existing, err := s.repo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return GoalResponse{}, err
	}

	existing.Title = title
	existing.TargetPaise = input.TargetPaise
	existing.Deadline = deadlineStr
	existing.MonthlyAmountPaise = input.MonthlyAmountPaise
	existing.Priority = input.Priority
	existing.Status = status

	updated, err := s.repo.UpdateGoal(ctx, existing)
	if err != nil {
		return GoalResponse{}, err
	}

	return s.toResponse(updated, time.Now().UTC()), nil
}

// DeleteGoal removes a goal belonging to the authenticated user.
func (s *Service) DeleteGoal(ctx context.Context, userID, goalID string) error {
	return s.repo.DeleteGoal(ctx, userID, goalID)
}

// CloseMonth initiates the month-end sweep, passing the pure Allocate function into repository.
func (s *Service) CloseMonth(ctx context.Context, userID string, month time.Time) (CloseResult, error) {
	return s.repo.CloseMonth(ctx, userID, month, Allocate)
}

// GetUnclosedPreviousMonths retrieves pending rollover jobs for the background worker.
func (s *Service) GetUnclosedPreviousMonths(ctx context.Context, previousMonth time.Time) ([]RolloverJob, error) {
	return s.repo.GetUnclosedPreviousMonths(ctx, previousMonth)
}

// toResponse calculates pure computed fields and formats GoalResponse.
func (s *Service) toResponse(g Goal, now time.Time) GoalResponse {
	remaining := RemainingPaise(g.TargetPaise, g.SavedPaise)
	pct := ProgressPercent(g.TargetPaise, g.SavedPaise)

	var deadlineTime *time.Time
	if g.Deadline != nil && *g.Deadline != "" {
		if t, err := time.Parse("2006-01-02", *g.Deadline); err == nil {
			deadlineTime = &t
		}
	}

	monthlyNeeded := MonthlyNeededPaise(remaining, deadlineTime, now)
	etaMonth := ETAMonth(remaining, g.MonthlyAmountPaise, now)

	return GoalResponse{
		ID:                 g.ID,
		UserID:             g.UserID,
		Title:              g.Title,
		TargetPaise:        g.TargetPaise,
		SavedPaise:         g.SavedPaise,
		Deadline:           g.Deadline,
		MonthlyAmountPaise: g.MonthlyAmountPaise,
		Priority:           g.Priority,
		Status:             g.Status,
		RemainingPaise:     remaining,
		ProgressPercent:    pct,
		MonthlyNeededPaise: monthlyNeeded,
		EtaMonth:           etaMonth,
		CreatedAt:          g.CreatedAt,
		UpdatedAt:          g.UpdatedAt,
	}
}

