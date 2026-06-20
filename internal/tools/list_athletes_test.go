package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/coach"
)

func TestListAthletesRemainsConfigSourcedAndSessionActive(t *testing.T) {
	t.Parallel()

	cfg := coach.Config{
		DefaultAthleteID: "i111",
		Athletes: []coach.Athlete{
			{ID: "i222", Label: "Read Only"},
			{ID: "i111", Label: "Default"},
		},
	}
	store := coach.NewSelectionStore(cfg.DefaultAthleteID)
	key, scope := store.Key("session-a")
	store.Select(key, "i222")
	ctx := coach.WithSelectionContext(context.Background(), coach.SelectionContext{Store: store, Key: key, Scope: scope})

	tool := newListAthletesTool(cfg)
	result, err := tool.Handler(ctx, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("list_athletes Handler() error = %v", err)
	}

	var got listAthletesResponse
	if err := json.Unmarshal([]byte(resultText(t, result)), &got); err != nil {
		t.Fatalf("unmarshal list_athletes response: %v", err)
	}
	if got.Meta.Source != "config" || got.Meta.Count != 2 || got.Meta.DefaultAthleteID != "i111" || got.Meta.ActiveAthleteID != "i222" {
		t.Fatalf("list_athletes meta = %#v, want config/default/active", got.Meta)
	}
	if got.Athletes[0].AthleteID != "i111" || !got.Athletes[0].Default || got.Athletes[0].Active {
		t.Fatalf("first sorted athlete = %#v, want default inactive i111", got.Athletes[0])
	}
	if got.Athletes[1].AthleteID != "i222" || got.Athletes[1].Default || !got.Athletes[1].Active {
		t.Fatalf("second sorted athlete = %#v, want active non-default i222", got.Athletes[1])
	}
}
