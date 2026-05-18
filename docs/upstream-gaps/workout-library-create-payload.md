# Workout-library create payload

During TP-077 live probing against the dedicated test athlete, intervals.icu accepted workout-library creates only when the payload included an existing folder ID owned by the athlete.

Observed contract for `POST /api/v1/athlete/{id}/workouts`:

- Send the sport/activity type as JSON key `type`, for example `"type": "Ride"`.
- Include `folder_id` with a real workout-library folder ID. Omitted `folder_id` and explicit `"folder_id": null` both returned `422 Folder is required`.
- A description string containing the workout DSL is accepted alongside `name`, `type`, and `folder_id`.
- A payload using `sport` instead of `type` returned `422 Missing type`.

Implications for icuvisor:

- `create_workout` requires `folder_id` locally and documents that it must identify an existing folder owned by the athlete.
- The client still serializes the tool's `sport` argument to upstream JSON key `type`.
- `update_workout` behavior is unchanged; this note only describes create semantics.

Sanitized request/response fixtures live under `internal/intervals/testdata/workout_library/create_request.json` and `create_response.json`.
