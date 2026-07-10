// Package planning provides a pure, deterministic constraint model and validator
// for the plan-filler scheduling domain. It validates candidate sessions against
// structured week, day, and slot constraints and computes reconciliation totals.
//
// This package contains no intervals.icu client calls, no calendar writes, no model
// inference, and no physiology classification. All inputs are caller-supplied structs
// with numeric fields; free-text instructions are never treated as hard constraints.
package planning

import "slices"

// SlotConstraint defines limits for one available training window within a day.
// Two slots are independent — a session must fit within exactly one slot and
// cannot span multiple slots. Each slot can hold at most one session; once a session
// claims a slot in a batch validation pass, that slot is unavailable to later candidates.
type SlotConstraint struct {
	// MaxDurationMinutes caps session duration for this slot. Zero means uncapped.
	MaxDurationMinutes float64
	// MaxIndoorMinutes caps duration specifically for indoor sessions.
	// Zero means no indoor cap. Outdoor sessions are not affected by this field.
	MaxIndoorMinutes float64
	// AllowedSports restricts which sports may fill this slot.
	// Empty means any sport is allowed.
	AllowedSports []string
	// AllowedModes restricts which training modes may fill this slot.
	// Empty means any mode is allowed.
	AllowedModes []string
}

// DayConstraints defines training availability for one calendar day.
// A day that is absent from WeekConstraints.AvailableDays is considered unavailable.
type DayConstraints struct {
	// Date is the athlete-local calendar date in YYYY-MM-DD format.
	Date string
	// MaxSessionsPerDay is the upper bound on sessions placed on this day.
	// Zero means the day is effectively unavailable regardless of slot count.
	MaxSessionsPerDay int
	// MaxTotalDailyMinutes caps the combined duration of all sessions on this day.
	// Zero means uncapped. Two sessions of 45 min each produce a combined total of 90 min.
	MaxTotalDailyMinutes float64
	// Slots lists the independent training windows available on this day.
	// A candidate session must fit within exactly one slot; slots do not combine.
	// Each slot can hold at most one session (consumed in batch validation).
	Slots []SlotConstraint
}

// WeekConstraints encodes the planning parameters for one calendar week.
//
// Availability (AvailableDays) captures where sessions may be placed.
// RequestedSessionCount captures how many sessions the caller wants placed.
// These are separate concepts: having 5 available days does not imply 5 sessions
// are requested, and requesting 3 sessions does not create availability on days
// that are absent from AvailableDays.
type WeekConstraints struct {
	// WeekStartDate is the athlete-local Monday in YYYY-MM-DD format.
	WeekStartDate string
	// WeeklyTargetMinutes is the full-week training-time target (e.g. for a complete week).
	// For an in-progress week, use the original full target here and report
	// actual completed time in CompletedMinutes; the validator derives RemainingMinutes
	// from WeeklyTargetMinutes - CompletedMinutes - FixedMinutes.
	WeeklyTargetMinutes float64
	// WeeklyTargetLoad is the full-week training-load target (e.g. TSS, ATL points).
	WeeklyTargetLoad float64
	// CompletedMinutes is already-logged training time this week (read-only past data).
	// Callers must not redistribute or zero this to create headroom.
	CompletedMinutes float64
	// CompletedLoad is already-logged training load this week (read-only past data).
	CompletedLoad float64
	// FixedMinutes is committed future training time from locked events
	// (e.g. races, A-priority events, unavailable blocks). These reduce the
	// remaining scheduling budget without being candidates themselves.
	FixedMinutes float64
	// FixedLoad is committed future training load from locked events.
	FixedLoad float64
	// RequestedSessionCount is the number of sessions the caller wants placed
	// into available slots. It is a scheduling intent and a hard cap —
	// ValidateCandidates marks valid candidates beyond this count as excess.
	// Zero means no session-count cap is applied.
	RequestedSessionCount int
	// AvailableDays lists the days within this week where sessions may be placed.
	// Days absent from this list are unavailable for scheduling.
	AvailableDays []DayConstraints
}

// CandidateSession describes a proposed training session to be validated.
type CandidateSession struct {
	// Date is the proposed athlete-local date in YYYY-MM-DD format.
	Date string
	// Sport identifies the training discipline (e.g. "Ride", "Run", "Swim").
	Sport string
	// Mode identifies the training mode (e.g. "EnduranceRide", "Intervals").
	Mode string
	// Indoor indicates an indoor trainer, treadmill, pool, or similar facility.
	Indoor bool
	// DurationMinutes is the proposed session length.
	DurationMinutes float64
	// Load is the proposed training load contribution (e.g. TSS, ATL points).
	Load float64
}

// ViolationCode identifies a hard constraint breach that blocks session placement.
type ViolationCode string

const (
	// ViolationDayUnavailable fires when the candidate date has no DayConstraints
	// in WeekConstraints.AvailableDays, or when MaxSessionsPerDay is zero.
	ViolationDayUnavailable ViolationCode = "day_unavailable"

	// ViolationDailySessionCount fires when adding the candidate would exceed MaxSessionsPerDay.
	ViolationDailySessionCount ViolationCode = "daily_session_count_exceeded"

	// ViolationDailyTimeExceeded fires when the candidate would push the combined daily
	// duration over DayConstraints.MaxTotalDailyMinutes.
	ViolationDailyTimeExceeded ViolationCode = "daily_time_exceeded"

	// ViolationSlotDuration fires when the candidate duration exceeds every available
	// slot's MaxDurationMinutes. Two 45-minute slots cannot accommodate a 95-minute session.
	// This code is only emitted when ALL available (unconsumed) slots reject for this reason.
	ViolationSlotDuration ViolationCode = "slot_duration_exceeded"

	// ViolationIndoorDuration fires when an indoor candidate's duration exceeds
	// every available slot's MaxIndoorMinutes. Outdoor sessions with the same duration
	// are not affected. Only emitted when ALL available slots reject for this reason.
	ViolationIndoorDuration ViolationCode = "indoor_duration_exceeded"

	// ViolationSportNotAllowed fires when the candidate sport is excluded by
	// every available slot's AllowedSports list. Only emitted when ALL slots reject.
	ViolationSportNotAllowed ViolationCode = "sport_not_allowed"

	// ViolationModeNotAllowed fires when the candidate mode is excluded by
	// every available slot's AllowedModes list. Only emitted when ALL slots reject.
	ViolationModeNotAllowed ViolationCode = "mode_not_allowed"

	// ViolationNoCompatibleSlot fires when no available slot can accommodate the
	// candidate and no single constraint reason is universal across all slots.
	// This is the deterministic fallback for mixed-reason slot rejections
	// (e.g. slot A rejects for duration, slot B rejects for sport).
	ViolationNoCompatibleSlot ViolationCode = "no_compatible_slot"

	// ViolationNoAvailableSlot fires when all slots for the day have been consumed
	// by prior candidates in a batch validation pass, even if MaxSessionsPerDay has
	// not been reached.
	ViolationNoAvailableSlot ViolationCode = "no_available_slot"

	// ViolationWeeklyLoadOvershoot fires when the candidate load would push the
	// projected weekly load over the remaining load budget
	// (WeeklyTargetLoad - CompletedLoad - FixedLoad - prior candidate load).
	ViolationWeeklyLoadOvershoot ViolationCode = "weekly_load_overshoot"

	// ViolationWeeklyTimeOvershoot fires when the candidate duration would push the
	// projected weekly time over the remaining time budget.
	ViolationWeeklyTimeOvershoot ViolationCode = "weekly_time_overshoot"

	// ViolationRequestedSessionCountExceeded fires when the candidate would be the
	// (N+1)th valid session and RequestedSessionCount is N. Position within the
	// candidate batch determines which sessions are accepted.
	ViolationRequestedSessionCountExceeded ViolationCode = "requested_session_count_exceeded"
)

// WarningCode identifies a soft constraint concern that does not block placement
// but deserves caller attention.
type WarningCode string

const (
	// WarnInfeasibleSessionCount fires when RequestedSessionCount exceeds the
	// total structural slot capacity across all available days. Structural capacity
	// is min(MaxSessionsPerDay, len(Slots)) per day; sport/mode filtering is not applied.
	WarnInfeasibleSessionCount WarningCode = "infeasible_session_count"

	// WarnInfeasibleLoad fires when the total candidate load (including invalid candidates)
	// is less than the remaining weekly load target, meaning the target cannot be met
	// with the provided candidates.
	WarnInfeasibleLoad WarningCode = "infeasible_load"

	// WarnZeroRemainingLoad fires when the remaining load budget is zero or negative.
	// Fires unconditionally when remaining ≤ 0, regardless of the candidate's Load value.
	// Completed and fixed events already meet the weekly load target.
	WarnZeroRemainingLoad WarningCode = "zero_remaining_load"

	// WarnZeroRemainingTime fires when the remaining time budget is zero or negative.
	// Parallel to WarnZeroRemainingLoad for the time dimension.
	WarnZeroRemainingTime WarningCode = "zero_remaining_time"
)

// Violation reports a hard constraint breach.
type Violation struct {
	Code    ViolationCode `json:"code"`
	Message string        `json:"message"`
	Field   string        `json:"field,omitempty"`
	Value   any           `json:"value,omitempty"`
}

// Warning reports a soft constraint concern.
type Warning struct {
	Code    WarningCode `json:"code"`
	Message string      `json:"message"`
	Field   string      `json:"field,omitempty"`
	Value   any         `json:"value,omitempty"`
}

// Reconciliation holds computed weekly time and load totals for a set of candidates.
// All fields are derived from WeekConstraints and caller-supplied candidates;
// no values are redistributed, inferred, or smoothed.
type Reconciliation struct {
	WeeklyTargetMinutes float64 `json:"weekly_target_minutes"`
	WeeklyTargetLoad    float64 `json:"weekly_target_load"`
	CompletedMinutes    float64 `json:"completed_minutes"`
	CompletedLoad       float64 `json:"completed_load"`
	FixedMinutes        float64 `json:"fixed_minutes"`
	FixedLoad           float64 `json:"fixed_load"`
	CandidateMinutes    float64 `json:"candidate_minutes"`
	CandidateLoad       float64 `json:"candidate_load"`
	// RemainingMinutes is WeeklyTargetMinutes - CompletedMinutes - FixedMinutes.
	// This is the scheduling budget for new sessions; negative when already over target.
	RemainingMinutes float64 `json:"remaining_minutes"`
	// RemainingLoad is WeeklyTargetLoad - CompletedLoad - FixedLoad.
	RemainingLoad float64 `json:"remaining_load"`
	// ProjectedMinutes is CompletedMinutes + FixedMinutes + CandidateMinutes.
	ProjectedMinutes float64 `json:"projected_minutes"`
	// ProjectedLoad is CompletedLoad + FixedLoad + CandidateLoad.
	ProjectedLoad float64 `json:"projected_load"`
}

// CandidateResult is the validation outcome for a single CandidateSession.
type CandidateResult struct {
	Candidate  CandidateSession `json:"candidate"`
	Valid      bool             `json:"valid"`
	Violations []Violation      `json:"violations"`
	Warnings   []Warning        `json:"warnings,omitempty"`
}

// BatchResult is the validation outcome for all candidate sessions in a week.
type BatchResult struct {
	// Results contains one CandidateResult per input candidate, in order.
	Results []CandidateResult `json:"results"`
	// Warnings contains week-level warnings that apply to the batch as a whole
	// rather than to any individual candidate.
	Warnings []Warning `json:"warnings,omitempty"`
	// Reconciliation holds the computed weekly totals for all candidates combined.
	Reconciliation Reconciliation `json:"reconciliation"`
}

// Reconcile computes weekly time and load totals from WeekConstraints and a set
// of candidates. It does not validate any constraints; call ValidateCandidates for
// full constraint checking.
func Reconcile(wc WeekConstraints, candidates []CandidateSession) Reconciliation {
	var candMin, candLoad float64
	for _, c := range candidates {
		candMin += c.DurationMinutes
		candLoad += c.Load
	}
	return buildReconciliation(wc, candMin, candLoad)
}

// ValidateCandidate validates a single candidate session against the week constraints,
// assuming it is the first (and only) session being considered for its date and no
// prior candidates have been processed in the current batch. All slots in the day are
// treated as available (no slot consumption).
//
// For batch validation with slot-consumption and session-count cap tracking, use ValidateCandidates.
func ValidateCandidate(wc WeekConstraints, candidate CandidateSession) CandidateResult {
	day, ok := findDay(wc.AvailableDays, candidate.Date)
	if !ok || day.MaxSessionsPerDay == 0 {
		return CandidateResult{
			Candidate: candidate,
			Valid:     false,
			Violations: []Violation{{
				Code:    ViolationDayUnavailable,
				Message: "session date is not available for scheduling",
				Field:   "date",
				Value:   candidate.Date,
			}},
		}
	}
	return validateAgainstDay(wc, day, day.Slots, 0, 0, 0, 0, candidate)
}

// ValidateCandidates validates all candidates in order with slot-consumption tracking,
// daily session-count tracking, and RequestedSessionCount cap enforcement.
//
// Slot consumption: each slot holds at most one session. When a candidate passes all
// slot-level constraints and claims a slot, that slot is removed from the available set
// for subsequent candidates on the same day. This is independent of other violations —
// a candidate that finds a compatible slot claims it even if it also has a weekly overshoot
// violation (pessimistic accumulation).
//
// Session-count cap: once RequestedSessionCount valid (violation-free) candidates have been
// accepted, subsequent valid candidates receive ViolationRequestedSessionCountExceeded.
// Position within the slice determines which candidates are accepted.
//
// Accumulation: all candidates (valid or not) increment the per-day session counter and
// per-day minute total, and contribute to the priorLoad/priorMinutes weekly budget for
// subsequent candidates. Invalid candidates are included because they are proposed
// positions in the schedule — a caller who provides an invalid candidate still intends
// that time window to be occupied, which pessimistically affects budget checks for later candidates.
func ValidateCandidates(wc WeekConstraints, candidates []CandidateSession) BatchResult {
	type dayState struct {
		day            DayConstraints
		sessions       int
		minutes        float64
		availableSlots []SlotConstraint
	}

	dayStates := map[string]*dayState{}
	for _, day := range wc.AvailableDays {
		d := day
		dayStates[day.Date] = &dayState{
			day:            d,
			availableSlots: slices.Clone(d.Slots),
		}
	}

	var priorLoad, priorMinutes float64
	var validCount int
	results := make([]CandidateResult, len(candidates))

	for i, candidate := range candidates {
		ds := dayStates[candidate.Date]

		var result CandidateResult
		if ds == nil {
			// Day not in AvailableDays.
			result = CandidateResult{
				Candidate: candidate,
				Valid:     false,
				Violations: []Violation{{
					Code:    ViolationDayUnavailable,
					Message: "session date is not available for scheduling",
					Field:   "date",
					Value:   candidate.Date,
				}},
			}
		} else {
			result = validateAgainstDay(wc, ds.day, ds.availableSlots, ds.sessions, ds.minutes, priorLoad, priorMinutes, candidate)

			// Consume the first compatible slot regardless of other violations
			// (pessimistic: the candidate occupies a time window in the proposed schedule).
			slotIdx := findCompatibleSlotIndex(ds.availableSlots, candidate)
			if slotIdx >= 0 {
				ds.availableSlots = slices.Delete(ds.availableSlots, slotIdx, slotIdx+1)
			}

			// Enforce RequestedSessionCount cap: if this candidate is otherwise valid
			// but the cap has already been reached, mark it as excess.
			if result.Valid && wc.RequestedSessionCount > 0 && validCount >= wc.RequestedSessionCount {
				result.Valid = false
				result.Violations = append(result.Violations, Violation{
					Code:    ViolationRequestedSessionCountExceeded,
					Message: "requested session count already reached; this candidate is excess",
					Field:   "requested_session_count",
					Value:   wc.RequestedSessionCount,
				})
			}

			// Update day state (all candidates, including invalid, consume the daily counter).
			ds.sessions++
			ds.minutes += candidate.DurationMinutes
		}

		if result.Valid {
			validCount++
		}

		// All candidates consume the weekly accumulation budget (pessimistic).
		priorLoad += candidate.Load
		priorMinutes += candidate.DurationMinutes

		results[i] = result
	}

	var weekWarnings []Warning

	// Warn when the requested count cannot be met structurally.
	// Structural capacity is min(MaxSessionsPerDay, len(Slots)) per day;
	// sport/mode filtering is not applied here.
	totalSlots := availableSlotCount(wc)
	if wc.RequestedSessionCount > 0 && wc.RequestedSessionCount > totalSlots {
		weekWarnings = append(weekWarnings, Warning{
			Code:    WarnInfeasibleSessionCount,
			Message: "requested session count exceeds total structural slot capacity across available days",
			Field:   "requested_session_count",
			Value:   wc.RequestedSessionCount,
		})
	}

	var candMin, candLoad float64
	for _, c := range candidates {
		candMin += c.DurationMinutes
		candLoad += c.Load
	}
	recon := buildReconciliation(wc, candMin, candLoad)

	// Warn when candidates cannot satisfy the remaining load target.
	// Includes invalid candidates' load (pessimistic).
	if recon.RemainingLoad > 0 && candLoad < recon.RemainingLoad {
		weekWarnings = append(weekWarnings, Warning{
			Code:    WarnInfeasibleLoad,
			Message: "candidate load total is less than remaining weekly load target",
			Field:   "remaining_load",
			Value:   recon.RemainingLoad,
		})
	}

	return BatchResult{
		Results:        results,
		Warnings:       weekWarnings,
		Reconciliation: recon,
	}
}

// validateAgainstDay is the core per-candidate validator.
// availableSlots is the set of slots still available on this day. In single-candidate
// validation this equals day.Slots; in batch validation it is the unconsumed subset.
func validateAgainstDay(wc WeekConstraints, day DayConstraints, availableSlots []SlotConstraint, sessionsAlreadyOnDay int, dailyMinutesAlready float64, priorLoad float64, priorMinutes float64, candidate CandidateSession) CandidateResult {
	var violations []Violation
	var warnings []Warning

	// 1. Daily session count.
	if sessionsAlreadyOnDay >= day.MaxSessionsPerDay {
		violations = append(violations, Violation{
			Code:    ViolationDailySessionCount,
			Message: "maximum sessions per day already reached",
			Field:   "max_sessions_per_day",
			Value:   day.MaxSessionsPerDay,
		})
	}

	// 2. Combined daily duration.
	if day.MaxTotalDailyMinutes > 0 {
		newDailyTotal := dailyMinutesAlready + candidate.DurationMinutes
		if newDailyTotal > day.MaxTotalDailyMinutes {
			violations = append(violations, Violation{
				Code:    ViolationDailyTimeExceeded,
				Message: "combined daily training duration would exceed the daily cap",
				Field:   "max_total_daily_minutes",
				Value:   day.MaxTotalDailyMinutes,
			})
		}
	}

	// 3. Slot constraints (against available/unconsumed slots).
	if len(day.Slots) > 0 {
		if len(availableSlots) == 0 {
			// All slots have been consumed by prior candidates.
			violations = append(violations, Violation{
				Code:    ViolationNoAvailableSlot,
				Message: "all training slots for this day are already claimed",
				Field:   "date",
				Value:   candidate.Date,
			})
		} else {
			slotViolations := universalSlotViolations(availableSlots, candidate)
			violations = append(violations, slotViolations...)
		}
	}

	// 4. Weekly remaining load.
	// Remaining is target minus already-committed (completed + fixed) and prior candidates.
	remainingLoad := wc.WeeklyTargetLoad - wc.CompletedLoad - wc.FixedLoad - priorLoad
	if wc.WeeklyTargetLoad > 0 {
		if remainingLoad <= 0 {
			warnings = append(warnings, Warning{
				Code:    WarnZeroRemainingLoad,
				Message: "remaining weekly load budget is zero or negative; no additional load is needed",
				Field:   "remaining_load",
				Value:   remainingLoad,
			})
		} else if candidate.Load > remainingLoad {
			violations = append(violations, Violation{
				Code:    ViolationWeeklyLoadOvershoot,
				Message: "candidate load exceeds remaining weekly load budget",
				Field:   "weekly_target_load",
				Value:   remainingLoad,
			})
		}
	}

	// 5. Weekly remaining time.
	remainingMin := wc.WeeklyTargetMinutes - wc.CompletedMinutes - wc.FixedMinutes - priorMinutes
	if wc.WeeklyTargetMinutes > 0 {
		if remainingMin <= 0 {
			warnings = append(warnings, Warning{
				Code:    WarnZeroRemainingTime,
				Message: "remaining weekly time budget is zero or negative; no additional time is needed",
				Field:   "remaining_minutes",
				Value:   remainingMin,
			})
		} else if candidate.DurationMinutes > remainingMin {
			violations = append(violations, Violation{
				Code:    ViolationWeeklyTimeOvershoot,
				Message: "candidate duration exceeds remaining weekly time budget",
				Field:   "weekly_target_minutes",
				Value:   remainingMin,
			})
		}
	}

	return CandidateResult{
		Candidate:  candidate,
		Valid:      len(violations) == 0,
		Violations: violations,
		Warnings:   warnings,
	}
}

// slotFits returns true if the candidate satisfies all constraints of the given slot.
func slotFits(slot SlotConstraint, candidate CandidateSession) bool {
	if slot.MaxDurationMinutes > 0 && candidate.DurationMinutes > slot.MaxDurationMinutes {
		return false
	}
	if candidate.Indoor && slot.MaxIndoorMinutes > 0 && candidate.DurationMinutes > slot.MaxIndoorMinutes {
		return false
	}
	if len(slot.AllowedSports) > 0 && !slices.Contains(slot.AllowedSports, candidate.Sport) {
		return false
	}
	if len(slot.AllowedModes) > 0 && !slices.Contains(slot.AllowedModes, candidate.Mode) {
		return false
	}
	return true
}

// findCompatibleSlotIndex returns the index of the first slot in slots that the
// candidate fits, or -1 if none fits.
func findCompatibleSlotIndex(slots []SlotConstraint, candidate CandidateSession) int {
	for i, slot := range slots {
		if slotFits(slot, candidate) {
			return i
		}
	}
	return -1
}

// universalSlotViolations returns violations when no slot can accommodate the candidate.
// Violation codes are only emitted when they apply universally — every available slot
// rejects the candidate for that reason. When reasons are mixed across slots (slot A
// rejects for duration, slot B rejects for sport), the generic ViolationNoCompatibleSlot
// code is emitted instead, since neither reason is universally true.
func universalSlotViolations(slots []SlotConstraint, candidate CandidateSession) []Violation {
	if len(slots) == 0 {
		return nil
	}

	// First check if any slot fits at all.
	if findCompatibleSlotIndex(slots, candidate) >= 0 {
		return nil
	}

	// No slot fits. Determine which constraints are universally violated.
	// A constraint is "universal" if every slot rejects the candidate for that reason.
	allDuration := true
	allIndoor := true
	allSport := true
	allMode := true

	for _, slot := range slots {
		// Duration: does this slot reject for duration?
		rejectsDuration := slot.MaxDurationMinutes > 0 && candidate.DurationMinutes > slot.MaxDurationMinutes
		if !rejectsDuration {
			allDuration = false
		}

		// Indoor: does this slot reject for indoor cap?
		rejectsIndoor := candidate.Indoor && slot.MaxIndoorMinutes > 0 && candidate.DurationMinutes > slot.MaxIndoorMinutes
		if !rejectsIndoor {
			allIndoor = false
		}

		// Sport: does this slot reject for sport?
		rejectsSport := len(slot.AllowedSports) > 0 && !slices.Contains(slot.AllowedSports, candidate.Sport)
		if !rejectsSport {
			allSport = false
		}

		// Mode: does this slot reject for mode?
		rejectsMode := len(slot.AllowedModes) > 0 && !slices.Contains(slot.AllowedModes, candidate.Mode)
		if !rejectsMode {
			allMode = false
		}
	}

	var violations []Violation
	if allDuration {
		violations = append(violations, Violation{
			Code:    ViolationSlotDuration,
			Message: "session duration exceeds every available slot duration cap",
			Field:   "duration_minutes",
			Value:   candidate.DurationMinutes,
		})
	}
	if allIndoor {
		violations = append(violations, Violation{
			Code:    ViolationIndoorDuration,
			Message: "indoor session duration exceeds every available slot indoor cap",
			Field:   "duration_minutes",
			Value:   candidate.DurationMinutes,
		})
	}
	if allSport {
		violations = append(violations, Violation{
			Code:    ViolationSportNotAllowed,
			Message: "session sport is excluded by every available slot",
			Field:   "sport",
			Value:   candidate.Sport,
		})
	}
	if allMode {
		violations = append(violations, Violation{
			Code:    ViolationModeNotAllowed,
			Message: "session mode is excluded by every available slot",
			Field:   "mode",
			Value:   candidate.Mode,
		})
	}

	// Fallback: no single reason is universally true across all slots.
	if len(violations) == 0 {
		violations = append(violations, Violation{
			Code:    ViolationNoCompatibleSlot,
			Message: "no available slot can accommodate this candidate; constraints differ across slots",
			Field:   "date",
			Value:   candidate.Date,
		})
	}

	return violations
}

// buildReconciliation computes a Reconciliation from WeekConstraints and candidate totals.
func buildReconciliation(wc WeekConstraints, candMin, candLoad float64) Reconciliation {
	remainingMin := wc.WeeklyTargetMinutes - wc.CompletedMinutes - wc.FixedMinutes
	remainingLoad := wc.WeeklyTargetLoad - wc.CompletedLoad - wc.FixedLoad
	return Reconciliation{
		WeeklyTargetMinutes: wc.WeeklyTargetMinutes,
		WeeklyTargetLoad:    wc.WeeklyTargetLoad,
		CompletedMinutes:    wc.CompletedMinutes,
		CompletedLoad:       wc.CompletedLoad,
		FixedMinutes:        wc.FixedMinutes,
		FixedLoad:           wc.FixedLoad,
		CandidateMinutes:    candMin,
		CandidateLoad:       candLoad,
		RemainingMinutes:    remainingMin,
		RemainingLoad:       remainingLoad,
		ProjectedMinutes:    wc.CompletedMinutes + wc.FixedMinutes + candMin,
		ProjectedLoad:       wc.CompletedLoad + wc.FixedLoad + candLoad,
	}
}

// availableSlotCount returns the total structural slot capacity across all available days.
// Capacity per day is min(MaxSessionsPerDay, len(Slots)); if no slots are defined,
// the capacity is MaxSessionsPerDay (slots are optional constraint templates).
// Sport/mode filtering is not applied.
func availableSlotCount(wc WeekConstraints) int {
	total := 0
	for _, day := range wc.AvailableDays {
		if day.MaxSessionsPerDay <= 0 {
			continue
		}
		cap := day.MaxSessionsPerDay
		if len(day.Slots) > 0 && len(day.Slots) < cap {
			cap = len(day.Slots)
		}
		total += cap
	}
	return total
}

// findDay returns the DayConstraints for the given date, if present.
func findDay(days []DayConstraints, date string) (DayConstraints, bool) {
	for _, d := range days {
		if d.Date == date {
			return d, true
		}
	}
	return DayConstraints{}, false
}
