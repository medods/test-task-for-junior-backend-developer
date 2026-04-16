package common

import (
    "encoding/json"
    "testing"
    "time"
)

func TestDate_MarshalUnmarshal(t *testing.T) {
    dt := NewDate(time.Date(2026, 4, 16, 12, 34, 0, 0, time.UTC))
    b, err := dt.MarshalJSON()
    if err != nil {
        t.Fatalf("marshal err: %v", err)
    }
    if string(b) != "\"2026-04-16\"" {
        t.Fatalf("unexpected marshal: %s", string(b))
    }

    var parsed DateType
    if err := parsed.UnmarshalJSON(b); err != nil {
        t.Fatalf("unmarshal err: %v", err)
    }
    if !parsed.Time.Equal(dt.Time) {
        t.Fatalf("expected equal dates")
    }
}

func TestNewDate_TruncatesTime(t *testing.T) {
    src := time.Date(2026, 4, 16, 15, 10, 5, 0, time.FixedZone("X", 0))
    d := NewDate(src)
    if d.Hour() != 0 || d.Minute() != 0 || d.Second() != 0 {
        t.Fatalf("expected midnight, got %v", d.Time)
    }
    var js struct{ D DateType `json:"d"` }
    js.D = d
    _, err := json.Marshal(js)
    if err != nil {
        t.Fatalf("json marshal failed: %v", err)
    }
}
