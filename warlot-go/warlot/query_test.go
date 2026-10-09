package warlot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

type complexUser struct {
	ID       int              `json:"id"`
	Username string           `json:"username"`
	Profile  userProfile      `json:"profile"`
	Tags     []string         `json:"tags"`
	Metadata *userMetaDetails `json:"metadata"`
}

type userProfile struct {
	Bio string `json:"bio"`
	Age int    `json:"age"`
}

type userMetaDetails struct {
	LoginCount int    `json:"login_count"`
	IP         string `json:"ip"`
}

func TestQuery_ComplexNestedStructs(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rows := []map[string]any{
			{
				"id":       101,
				"username": "alice_dev",
				"profile": map[string]any{
					"bio": "Distributed systems engineer",
					"age": 30,
				},
				"tags": []string{"admin", "core"},
				"metadata": map[string]any{
					"login_count": 42,
					"ip":          "192.168.1.5",
				},
			},
			{
				"id":       102,
				"username": "bob_ops",
				"profile": map[string]any{
					"bio": "Site reliability engineer",
					"age": 28,
				},
				"tags":     []string{"ops"},
				"metadata": nil,
			},
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rows":      rows,
			"row_count": len(rows),
		})
	})
	defer srv.Close()

	proj := cl.Project("test-proj")
	users, err := Query[complexUser](context.Background(), proj, "SELECT * FROM users", nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Username != "alice_dev" || users[0].Profile.Age != 30 {
		t.Errorf("user 0 mismatch: %+v", users[0])
	}
	if users[0].Metadata == nil || users[0].Metadata.LoginCount != 42 {
		t.Errorf("user 0 metadata mismatch: %+v", users[0].Metadata)
	}
	if users[1].Metadata != nil {
		t.Errorf("user 1 expected nil metadata, got %+v", users[1].Metadata)
	}
}

// queryOldLoop reproduces the pre-S2 row-by-row re-serialization loop for benchmark comparison.
func queryOldLoop[T any](rows []map[string]interface{}) ([]T, error) {
	out := make([]T, 0, len(rows))
	for _, row := range rows {
		b, _ := json.Marshal(row)
		var t T
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, fmt.Errorf("row decoding failed: %w", err)
		}
		out = append(out, t)
	}
	return out, nil
}

// queryBatch implements the optimized batch unmarshaling.
func queryBatch[T any](rows []map[string]interface{}) ([]T, error) {
	b, err := json.Marshal(rows)
	if err != nil {
		return nil, fmt.Errorf("row marshaling failed: %w", err)
	}
	var out []T
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("row decoding failed: %w", err)
	}
	return out, nil
}

func generateBenchmarkRows(n int) []map[string]interface{} {
	rows := make([]map[string]interface{}, n)
	for i := 0; i < n; i++ {
		rows[i] = map[string]interface{}{
			"id":       i,
			"username": fmt.Sprintf("user_%d", i),
			"profile": map[string]interface{}{
				"bio": fmt.Sprintf("Bio for user %d", i),
				"age": 20 + (i % 50),
			},
			"tags": []string{"tag1", "tag2"},
			"metadata": map[string]interface{}{
				"login_count": i * 3,
				"ip":          "127.0.0.1",
			},
		}
	}
	return rows
}

func BenchmarkQuery_OldRowByRow(b *testing.B) {
	rows := generateBenchmarkRows(500)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := queryOldLoop[complexUser](rows)
		if err != nil || len(res) != 500 {
			b.Fatal(err)
		}
	}
}

func BenchmarkQuery_NewBatch(b *testing.B) {
	rows := generateBenchmarkRows(500)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := queryBatch[complexUser](rows)
		if err != nil || len(res) != 500 {
			b.Fatal(err)
		}
	}
}

func queryDirectSinglePass[T any](body []byte) ([]T, error) {
	var env struct {
		Rows     []T    `json:"rows"`
		RowCount *int   `json:"row_count"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	return env.Rows, nil
}

func BenchmarkQuery_DirectSinglePass(b *testing.B) {
	rows := generateBenchmarkRows(500)
	rawJSON, _ := json.Marshal(map[string]any{"rows": rows, "row_count": len(rows)})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := queryDirectSinglePass[complexUser](rawJSON)
		if err != nil || len(res) != 500 {
			b.Fatal(err)
		}
	}
}
