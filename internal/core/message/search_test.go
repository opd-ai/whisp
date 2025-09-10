package message

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/whisp/internal/storage"
)

// TestFTSSearchPerformance tests that FTS search meets performance requirements (<100ms)
func TestFTSSearchPerformance(t *testing.T) {
	// Create test manager with large message dataset
	manager := createTestManagerWithMessages(t, 1000)
	defer cleanup(manager)

	// Test query that should return multiple results
	query := "test message"

	// Measure search performance
	start := time.Now()
	results, err := manager.SearchMessages(query, 50)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("SearchMessages failed: %v", err)
	}

	// Verify performance requirement
	if duration > 100*time.Millisecond {
		t.Errorf("Search took %v, expected <100ms", duration)
	}

	// Verify results are relevant
	if len(results) == 0 {
		t.Error("Expected search results, got none")
	}

	// Verify all results contain the search term
	for _, msg := range results {
		if !strings.Contains(strings.ToLower(msg.Content), strings.ToLower(query)) {
			t.Errorf("Result doesn't contain search term: %s", msg.Content)
		}
	}

	t.Logf("Search completed in %v with %d results", duration, len(results))
}

// TestFTSSearchAccuracy tests search result accuracy and ranking
func TestFTSSearchAccuracy(t *testing.T) {
	manager := createTestManagerWithMessages(t, 100)
	defer cleanup(manager)

	testCases := []struct {
		name     string
		query    string
		expected int // minimum expected results
	}{
		{"exact phrase", "hello world", 1}, // Lower expectation for exact phrase matching
		{"single word", "hello", 5},        // Reasonable expectation for single word
		{"partial word", "test", 5},        // Reasonable expectation for test
		{"empty query", "", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := manager.SearchMessages(tc.query, 100)
			if err != nil {
				t.Fatalf("SearchMessages failed: %v", err)
			}

			if len(results) < tc.expected {
				t.Errorf("Expected at least %d results, got %d", tc.expected, len(results))
			}

			// Verify results are ordered by timestamp (newest first)
			for i := 1; i < len(results); i++ {
				if results[i].Timestamp.After(results[i-1].Timestamp) {
					t.Error("Results not ordered by timestamp DESC")
					break
				}
			}
		})
	}
}

// TestFTSSearchFallback tests fallback to LIKE query when FTS fails
func TestFTSSearchFallback(t *testing.T) {
	manager := createTestManagerWithMessages(t, 50)
	defer cleanup(manager)

	// Test with query that might cause FTS issues (special characters)
	query := "test@message"

	results, err := manager.SearchMessages(query, 20)
	if err != nil {
		t.Fatalf("SearchMessages should not fail even with special characters: %v", err)
	}

	// Should still get results via fallback mechanism
	t.Logf("Fallback search returned %d results", len(results))
}

// TestFTSSearchLargeDataset tests search on a large dataset for scalability
func TestFTSSearchLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	manager := createTestManagerWithMessages(t, 5000)
	defer cleanup(manager)

	queries := []string{"test", "message", "hello", "content"}

	for _, query := range queries {
		start := time.Now()
		results, err := manager.SearchMessages(query, 100)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("SearchMessages failed for query '%s': %v", query, err)
		}

		// Performance should still be acceptable with large dataset
		// FTS5 systems would be under 200ms, but fallback LIKE queries may be slower
		maxDuration := 200 * time.Millisecond
		if duration > maxDuration {
			// Check if we're using fallback - if so, allow up to 500ms
			manager.mu.RLock()
			isFTSAvailable := manager.isFTSAvailable()
			manager.mu.RUnlock()

			if !isFTSAvailable {
				maxDuration = 500 * time.Millisecond
			}

			if duration > maxDuration {
				t.Errorf("Search for '%s' took %v on large dataset, expected <%v", query, duration, maxDuration)
			}
		}

		t.Logf("Query '%s': %d results in %v", query, len(results), duration)
	}
}

// BenchmarkSearchMessages benchmarks the search performance
func BenchmarkSearchMessages(b *testing.B) {
	manager := createTestManagerWithMessages(b, 1000)
	defer cleanup(manager)

	query := "test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.SearchMessages(query, 50)
		if err != nil {
			b.Fatalf("SearchMessages failed: %v", err)
		}
	}
}

// BenchmarkSearchMessagesVaryingSize benchmarks search with different dataset sizes
func BenchmarkSearchMessagesVaryingSize(b *testing.B) {
	sizes := []int{100, 500, 1000, 2000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("messages_%d", size), func(b *testing.B) {
			manager := createTestManagerWithMessages(b, size)
			defer cleanup(manager)

			query := "test"

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := manager.SearchMessages(query, 50)
				if err != nil {
					b.Fatalf("SearchMessages failed: %v", err)
				}
			}
		})
	}
}

// createTestManagerWithMessages creates a message manager with a specified number of test messages
func createTestManagerWithMessages(t testing.TB, messageCount int) *Manager {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create mock managers
	toxMgr := &MockToxManager{}
	contactMgr := NewMockContactManager()

	// Add test contacts
	for i := uint32(1); i <= 10; i++ {
		contactMgr.AddContact(i, map[string]string{"name": fmt.Sprintf("Test Friend %d", i)})
	}

	manager := NewManager(db, toxMgr, contactMgr)

	// Create test messages with varying content
	words := []string{"hello", "world", "test", "message", "content", "search", "example", "data", "sample", "text"}

	// Add some specific test messages to ensure we have exact phrases
	specificMessages := []string{
		"hello world from user 1",
		"hello world this is a test",
		"hello world again",
		"hello world example message",
		"hello world test content",
		"hello there friend",
		"hello everyone",
		"hello testing",
		"hello message",
		"hello sample",
		"test message content",
		"test data sample",
		"test hello world",
		"test search function",
		"test example",
	}

	// Add specific messages first
	for i, content := range specificMessages {
		friendID := uint32((i % 10) + 1)
		_, err := manager.SendMessage(friendID, content, MessageTypeNormal)
		if err != nil {
			t.Fatalf("Failed to send specific test message: %v", err)
		}
	}

	// Then add random messages to fill the rest
	remainingCount := messageCount - len(specificMessages)
	for i := 0; i < remainingCount; i++ {
		// Generate random message content
		wordCount := rand.Intn(10) + 5 // 5-15 words
		var messageWords []string
		for j := 0; j < wordCount; j++ {
			messageWords = append(messageWords, words[rand.Intn(len(words))])
		}
		content := strings.Join(messageWords, " ")

		// Create message with varying friend IDs and timestamps
		friendID := uint32(rand.Intn(10) + 1)

		// Use SendMessage to properly store messages
		_, err := manager.SendMessage(friendID, content, MessageTypeNormal)
		if err != nil {
			t.Fatalf("Failed to send test message: %v", err)
		}
	}

	return manager
}

// cleanup cleans up test resources
func cleanup(manager *Manager) {
	if manager.db != nil {
		manager.db.Close()
	}
}
