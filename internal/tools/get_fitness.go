package tools

import (
	"context"
	"errors"
	"sort"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getFitnessName        = "get_fitness"
	getFitnessDescription = "Get CTL, ATL, and TSB fitness trends for a local date range. Dates are athlete-local YYYY-MM-DD values."
	fetchFitnessMessage   = "could not fetch fitness data; check intervals.icu credentials, athlete ID, and date range"
)

type fitnessResponse struct {
	Rows []fitnessRow `json:"fitness"`
	Meta fitnessMeta  `json:"_meta"`
}

type fitnessRow struct {
	Date string         `json:"date"`
	CTL  *float64       `json:"ctl,omitempty"`
	ATL  *float64       `json:"atl,omitempty"`
	TSB  *float64       `json:"tsb,omitempty"`
	Full map[string]any `json:"full,omitempty"`
}

type fitnessMeta struct {
	ServerVersion string `json:"server_version"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Timezone      string `json:"timezone"`
	Count         int    `json:"count"`
	IncludeFull   bool   `json:"include_full"`
}

func newGetFitnessTool(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getFitnessName, Description: getFitnessDescription, InputSchema: dateRangeInputSchema("local start date for fitness rows"), OutputSchema: genericOutputSchema("Fitness rows with CTL, ATL, and TSB."), Handler: getFitnessHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getFitnessHandler(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDateRangeRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidFitnessArgumentsMessage, err)
		}
		unitSystem, timezone, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchFitnessMessage, err)
		}
		rows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.StartDate, End: args.EndDate})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchFitnessMessage, err)
		}
		payload := fitnessResponse{Rows: shapeFitnessRows(rows, args.IncludeFull), Meta: fitnessMeta{ServerVersion: normalizeVersion(version), StartDate: args.StartDate, EndDate: args.EndDate, Timezone: timezone, Count: len(rows), IncludeFull: args.IncludeFull}}
		return encodeShaped(payload, args.IncludeFull, []string{"fitness"}, version, debugMetadata, getFitnessName, unitSystem, shapeCfg)
	}
}

func shapeFitnessRows(rows []intervals.SummaryWithCats, includeFull bool) []fitnessRow {
	out := make([]fitnessRow, 0, len(rows))
	for _, row := range rows {
		ctl, atl, tsb := roundPtr(row.Fitness), roundPtr(row.Fatigue), roundPtr(row.Form)
		shaped := fitnessRow{Date: row.Date, CTL: ctl, ATL: atl, TSB: tsb}
		if includeFull {
			shaped.Full = row.Raw
		}
		out = append(out, shaped)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}
