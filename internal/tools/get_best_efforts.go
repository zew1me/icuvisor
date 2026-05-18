package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getBestEffortsName        = "get_best_efforts"
	getBestEffortsDescription = "Get upstream best efforts grouped by sport and default/requested power, heart-rate, and pace buckets. Defaults to Ride, Run, and Swim with terse bucket rows."
	fetchBestEffortsMessage   = "could not fetch best efforts; check intervals.icu credentials, athlete ID, sports, and date range"
)

var (
	defaultBestEffortSports    = []string{"Ride", "Run", "Swim"}
	defaultRunDistanceBuckets  = []int{400, 1000, 1609, 5000, 10000}
	defaultSwimDistanceBuckets = []int{50, 100, 200, 400, 1500}
)

// BestEffortsClient retrieves athlete curve sets for best efforts.
type BestEffortsClient interface {
	ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
	ListAthleteHRCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
	ListAthletePaceCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
}

type bestEffortsRequest struct {
	Oldest          string   `json:"oldest,omitempty"`
	Newest          string   `json:"newest,omitempty"`
	Sports          []string `json:"sports,omitempty"`
	DurationSeconds []int    `json:"duration_seconds,omitempty"`
	DistanceMeters  []int    `json:"distance_meters,omitempty"`
	IncludeFull     bool     `json:"include_full,omitempty"`
}

type bestEffortsResponse struct {
	Sports []bestEffortsSport `json:"sports"`
	Meta   bestEffortsMeta    `json:"_meta"`
}

type bestEffortsSport struct {
	Sport   string          `json:"sport"`
	Efforts []bestEffortRow `json:"efforts,omitempty"`
	Full    map[string]any  `json:"full,omitempty"`
}

type bestEffortRow struct {
	Family          string   `json:"family"`
	DurationSeconds int      `json:"duration_seconds,omitempty"`
	DistanceMeters  int      `json:"distance_meters,omitempty"`
	PowerWatts      *float64 `json:"power_watts,omitempty"`
	HeartRateBPM    *float64 `json:"heart_rate_bpm,omitempty"`
	PaceValue       *float64 `json:"pace_value,omitempty"`
	ActivityID      string   `json:"activity_id,omitempty"`
}

type bestEffortsMeta struct {
	ServerVersion   string           `json:"server_version"`
	SportsRequested []string         `json:"sports_requested"`
	Oldest          string           `json:"oldest,omitempty"`
	Newest          string           `json:"newest,omitempty"`
	CurveSpec       string           `json:"curve_spec"`
	DurationSeconds []int            `json:"duration_seconds"`
	DistanceMeters  []int            `json:"distance_meters"`
	MissingBuckets  map[string][]int `json:"missing_buckets,omitempty"`
	IncludeFull     bool             `json:"include_full"`
}

func newGetBestEffortsTool(client BestEffortsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getBestEffortsName, Description: getBestEffortsDescription, InputSchema: bestEffortsInputSchema(), OutputSchema: genericOutputSchema("Best efforts grouped by sport and bucket."), Handler: getBestEffortsHandler(client, version, debugMetadata, shapeCfg)})
}

func getBestEffortsHandler(client BestEffortsClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeBestEffortsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCurveArgumentsMessage, err)
		}
		curveSpec := bestEffortsCurveSpec(args)
		missing := map[string][]int{}
		payload := bestEffortsResponse{Sports: make([]bestEffortsSport, 0, len(args.Sports)), Meta: bestEffortsMeta{ServerVersion: normalizeVersion(version), SportsRequested: args.Sports, Oldest: args.Oldest, Newest: args.Newest, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds, DistanceMeters: args.DistanceMeters, MissingBuckets: missing, IncludeFull: args.IncludeFull}}
		for _, sport := range args.Sports {
			row, err := bestEffortsForSport(ctx, client, sport, curveSpec, args, missing)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(fetchBestEffortsMessage, err)
			}
			payload.Sports = append(payload.Sports, row)
		}
		if len(missing) == 0 {
			payload.Meta.MissingBuckets = nil
		}
		return encodeShaped(payload, args.IncludeFull, []string{"sports"}, version, debugMetadata, getBestEffortsName, response.UnitSystemMetric, shapeCfg)
	}
}

func decodeBestEffortsRequest(raw json.RawMessage) (bestEffortsRequest, error) {
	var args bestEffortsRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[bestEffortsRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	if (args.Oldest == "") != (args.Newest == "") {
		return args, errors.New("oldest and newest must be supplied together or both omitted")
	}
	if args.Oldest != "" && (!validDate(args.Oldest) || !validDate(args.Newest) || args.Newest < args.Oldest) {
		return args, errors.New("oldest/newest must be paired YYYY-MM-DD values in order")
	}
	args.Sports = normalizeSports(args.Sports)
	args.DurationSeconds = normalizePositiveInts(args.DurationSeconds, defaultDurationBuckets)
	args.DistanceMeters = normalizePositiveInts(args.DistanceMeters, nil)
	return args, nil
}

func normalizeSports(values []string) []string {
	if len(values) == 0 {
		return append([]string(nil), defaultBestEffortSports...)
	}
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		sport := strings.TrimSpace(value)
		if sport != "" && !seen[sport] {
			seen[sport] = true
			out = append(out, sport)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), defaultBestEffortSports...)
	}
	return out
}

func bestEffortsCurveSpec(args bestEffortsRequest) string {
	if args.Oldest == "" {
		return "all"
	}
	return rangeCurveSpec(args.Oldest, args.Newest)
}

func bestEffortsForSport(ctx context.Context, client BestEffortsClient, sport string, curveSpec string, args bestEffortsRequest, missing map[string][]int) (bestEffortsSport, error) {
	out := bestEffortsSport{Sport: sport}
	power, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds})
	if err != nil {
		return out, err
	}
	hr, err := client.ListAthleteHRCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds})
	if err != nil {
		return out, err
	}
	distances := args.DistanceMeters
	if len(distances) == 0 {
		distances = defaultDistanceBucketsForSport(sport)
	}
	pace, err := client.ListAthletePaceCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DistanceMeters: distances})
	if err != nil {
		return out, err
	}
	out.Efforts = append(out.Efforts, effortRowsFromDurationCurve("power", firstCurve(power), args.DurationSeconds, missing, sport)...)
	out.Efforts = append(out.Efforts, effortRowsFromDurationCurve("heart_rate", firstCurve(hr), args.DurationSeconds, missing, sport)...)
	out.Efforts = append(out.Efforts, effortRowsFromDistanceCurve(firstCurve(pace), distances, missing, sport)...)
	if args.IncludeFull {
		out.Full = map[string]any{"power": power.Raw, "heart_rate": hr.Raw, "pace": pace.Raw}
	}
	return out, nil
}

func effortRowsFromDurationCurve(family string, curve intervals.DataCurve, buckets []int, missing map[string][]int, sport string) []bestEffortRow {
	rows := []bestEffortRow{}
	for _, bucket := range buckets {
		value, idx, ok := valueAtBucket(curve.Secs, curve.Values, bucket)
		if !ok {
			missing[sport+":"+family] = append(missing[sport+":"+family], bucket)
			continue
		}
		row := bestEffortRow{Family: family, DurationSeconds: bucket, ActivityID: activityIDAt(curve, idx)}
		rounded := roundPtr(value)
		if family == "power" {
			row.PowerWatts = rounded
		} else {
			row.HeartRateBPM = rounded
		}
		rows = append(rows, row)
	}
	return rows
}

func effortRowsFromDistanceCurve(curve intervals.DataCurve, buckets []int, missing map[string][]int, sport string) []bestEffortRow {
	rows := []bestEffortRow{}
	for _, bucket := range buckets {
		value, idx, ok := valueAtBucket(curve.Distance, curve.Values, bucket)
		if !ok {
			missing[sport+":pace"] = append(missing[sport+":pace"], bucket)
			continue
		}
		rows = append(rows, bestEffortRow{Family: "pace", DistanceMeters: bucket, PaceValue: roundPtr(value), ActivityID: activityIDAt(curve, idx)})
	}
	return rows
}

func activityIDAt(curve intervals.DataCurve, idx int) string {
	if idx >= 0 && idx < len(curve.ActivityID) {
		return curve.ActivityID[idx]
	}
	return ""
}

func defaultDistanceBucketsForSport(sport string) []int {
	if strings.Contains(strings.ToLower(sport), "swim") {
		return append([]int(nil), defaultSwimDistanceBuckets...)
	}
	return append([]int(nil), defaultRunDistanceBuckets...)
}

func bestEffortsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Optional local start date YYYY-MM-DD. Supply with newest or omit both for all-time."}, "newest": map[string]any{"type": "string", "description": "Optional local end date YYYY-MM-DD. Supply with oldest or omit both for all-time."}, "sports": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Intervals.icu sports/types to fan out. Defaults to Ride, Run, Swim."}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Power/HR duration buckets. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "distance_meters": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Pace distance buckets. Defaults depend on sport: run-style 400,1000,1609,5000,10000m; swim 50,100,200,400,1500m."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
