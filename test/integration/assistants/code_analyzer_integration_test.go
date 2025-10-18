package assistants

import (
	"context"
	"testing"
	"time"

	"github.com/andranikuz/aiwf/test/integration/assistants/generated/code_analyzer"
)

func TestCodeAnalyzer_Integration(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := code_analyzer.NewService(openaiClient)

	// Test Go code with potential issues
	testCode := `package main

import "fmt"

func main() {
	var x int
	x = 10
	y := x + 5
	fmt.Println(x)
	// y is declared but not used

	// Potential SQL injection
	query := "SELECT * FROM users WHERE id = " + userInput
	db.Query(query)
}
`

	input := code_analyzer.CodeAnalysisRequest{
		Code:           testCode,
		Language:       "go",
		AnalysisFocus:  "all",
	}

	result, trace, err := service.Agents().CodeAnalyzer.Run(ctx, input)
	if err != nil {
		t.Fatalf("CodeAnalyzer agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if trace == nil {
		t.Fatal("Expected trace, got nil")
	}

	t.Logf("✓ Code analysis completed")
	t.Logf("  Total issues: %d", result.Summary["total_issues"])
	if result.Summary["critical_count"] != nil {
		t.Logf("  Critical issues: %v", result.Summary["critical_count"])
	}
	if result.Summary["quality_score"] != nil {
		t.Logf("  Quality score: %v/100", result.Summary["quality_score"])
	}

	if len(result.Issues) > 0 {
		t.Logf("  Issues found:")
		for i, issue := range result.Issues {
			t.Logf("    [%d] %s (severity: %s)", i, issue.Message, issue.Severity)
		}
	}

	t.Logf("  Tokens (in/out): %d/%d", trace.InputTokens, trace.OutputTokens)
}

func TestCodeAnalyzer_MultipleLanguages(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	service := code_analyzer.NewService(openaiClient)

	tests := []struct {
		name     string
		language string
		code     string
		focus    string
	}{
		{
			name:     "Python - Security",
			language: "python",
			code: `import pickle
data = pickle.loads(user_input)  # Unsafe deserialization
exec(code_string)  # Code injection vulnerability
password = "admin123"  # Hardcoded secret
`,
			focus: "security",
		},
		{
			name:     "JavaScript - Performance",
			language: "javascript",
			code: `function processArray(arr) {
  for (let i = 0; i < arr.length; i++) {
    for (let j = 0; j < arr.length; j++) {
      console.log(arr[i] * arr[j]);
    }
  }
}
`,
			focus: "performance",
		},
		{
			name:     "Rust - Maintainability",
			language: "rust",
			code: `fn main() {
	let x = 5;
	let y = {
		let x = 3;
		x + 1
	};
	let z = y + 1;
}
`,
			focus: "maintainability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := code_analyzer.CodeAnalysisRequest{
				Code:          tt.code,
				Language:      tt.language,
				AnalysisFocus: tt.focus,
			}

			result, trace, err := service.Agents().CodeAnalyzer.Run(ctx, input)
			if err != nil {
				t.Fatalf("CodeAnalyzer failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Language: %s | Focus: %s", tt.language, tt.focus)
			t.Logf("  Issues: %d | Metrics: %d", len(result.Issues), len(result.Metrics))
			if len(result.Metrics) > 0 {
				for _, metric := range result.Metrics {
					t.Logf("    - %s: %v %s", metric.Name, metric.Value, metric.Unit)
				}
			}
		})
	}
}

func TestCodeAnalyzer_ComplexCode(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := code_analyzer.NewService(openaiClient)

	// More realistic code sample
	testCode := `package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func getUser(db *sql.DB, userID string) (*User, error) {
	// Potential SQL injection
	query := fmt.Sprintf("SELECT id, name, email FROM users WHERE id = %s", userID)
	row := db.QueryRow(query)

	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	return user, err
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	db := getDB()
	user, err := getUser(db, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User: %+v\n", user)
}

func getDB() *sql.DB {
	// Missing error handling
	db, _ := sql.Open("mysql", "root:password@/mydb")
	return db
}

func main() {
	http.HandleFunc("/user", handleRequest)
	http.ListenAndServe(":8080", nil)
}
`

	input := code_analyzer.CodeAnalysisRequest{
		Code:          testCode,
		Language:      "go",
		AnalysisFocus: "all",
	}

	result, trace, err := service.Agents().CodeAnalyzer.Run(ctx, input)
	if err != nil {
		t.Fatalf("CodeAnalyzer failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Complex code analysis completed")
	t.Logf("  Issues found: %d", len(result.Issues))

	// Categorize issues by severity
	severityCount := make(map[string]int)
	typeCount := make(map[string]int)

	for _, issue := range result.Issues {
		severityCount[issue.Severity]++
		typeCount[issue.Type]++
	}

	if len(severityCount) > 0 {
		t.Logf("  By severity:")
		for severity, count := range severityCount {
			t.Logf("    - %s: %d", severity, count)
		}
	}

	if len(typeCount) > 0 {
		t.Logf("  By type:")
		for issueType, count := range typeCount {
			t.Logf("    - %s: %d", issueType, count)
		}
	}

	t.Logf("  Tokens (in/out): %d/%d", trace.InputTokens, trace.OutputTokens)
}
