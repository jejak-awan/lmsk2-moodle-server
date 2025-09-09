package chaos_engineering

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestChaosEngineeringNetworkChaos tests network chaos scenarios
func TestChaosEngineeringNetworkChaos(t *testing.T) {
	// Test 1: Network latency injection
	t.Run("NetworkLatencyInjection", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test network latency injection
		chaosData := map[string]interface{}{
			"type":     "network_latency",
			"duration": 30,
			"latency":  1000, // 1 second
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Network latency injection test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "network_latency" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under chaos
		start := time.Now()
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		duration := time.Since(start)

		if w.Code != http.StatusOK {
			t.Errorf("System behavior test failed with status %d", w.Code)
		}

		// Check if latency was injected
		if duration < 500*time.Millisecond {
			t.Error("Network latency not properly injected")
		}
	})

	// Test 2: Network packet loss
	t.Run("NetworkPacketLoss", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test network packet loss
		chaosData := map[string]interface{}{
			"type":        "packet_loss",
			"duration":    30,
			"loss_rate":   0.1, // 10% packet loss
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Network packet loss test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "packet_loss" {
			t.Error("Chaos type not set correctly")
		}

		// Test system resilience
		successCount := 0
		totalRequests := 10

		for i := 0; i < totalRequests; i++ {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				successCount++
			}
		}

		// Check if system handled packet loss gracefully
		successRate := float64(successCount) / float64(totalRequests)
		if successRate < 0.8 { // At least 80% success rate
			t.Errorf("System not resilient to packet loss: %f success rate", successRate)
		}
	})

	// Test 3: Network partition
	t.Run("NetworkPartition", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test network partition
		chaosData := map[string]interface{}{
			"type":     "network_partition",
			"duration": 30,
			"target":   "database",
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Network partition test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "network_partition" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under partition
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// System should handle partition gracefully
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Errorf("System not handling network partition gracefully: status %d", w.Code)
		}
	})
}

// TestChaosEngineeringResourceChaos tests resource chaos scenarios
func TestChaosEngineeringResourceChaos(t *testing.T) {
	// Test 1: CPU stress
	t.Run("CPUStress", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test CPU stress
		chaosData := map[string]interface{}{
			"type":     "cpu_stress",
			"duration": 30,
			"load":     0.8, // 80% CPU load
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("CPU stress test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "cpu_stress" {
			t.Error("Chaos type not set correctly")
		}

		// Test system performance under stress
		start := time.Now()
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		duration := time.Since(start)

		if w.Code != http.StatusOK {
			t.Errorf("System performance test failed with status %d", w.Code)
		}

		// Check if system is still responsive
		if duration > 5*time.Second {
			t.Error("System not responsive under CPU stress")
		}
	})

	// Test 2: Memory stress
	t.Run("MemoryStress", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test memory stress
		chaosData := map[string]interface{}{
			"type":     "memory_stress",
			"duration": 30,
			"load":     0.9, // 90% memory usage
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Memory stress test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "memory_stress" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under memory stress
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System behavior test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check if memory usage is reported
		if stats["memory_usage"] == nil {
			t.Error("Memory usage not reported under stress")
		}
	})

	// Test 3: Disk I/O stress
	t.Run("DiskIOStress", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test disk I/O stress
		chaosData := map[string]interface{}{
			"type":     "disk_io_stress",
			"duration": 30,
			"load":     0.7, // 70% disk I/O load
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Disk I/O stress test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "disk_io_stress" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under disk I/O stress
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System behavior test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check if disk usage is reported
		if stats["disk_usage"] == nil {
			t.Error("Disk usage not reported under stress")
		}
	})
}

// TestChaosEngineeringServiceChaos tests service chaos scenarios
func TestChaosEngineeringServiceChaos(t *testing.T) {
	// Test 1: Service failure
	t.Run("ServiceFailure", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test service failure
		chaosData := map[string]interface{}{
			"type":     "service_failure",
			"duration": 30,
			"service":  "database",
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Service failure test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "service_failure" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under service failure
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// System should handle service failure gracefully
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Errorf("System not handling service failure gracefully: status %d", w.Code)
		}
	})

	// Test 2: Service restart
	t.Run("ServiceRestart", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test service restart
		chaosData := map[string]interface{}{
			"type":     "service_restart",
			"duration": 30,
			"service":  "web_server",
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Service restart test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "service_restart" {
			t.Error("Chaos type not set correctly")
		}

		// Test system recovery after restart
		time.Sleep(2 * time.Second) // Wait for restart

		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System recovery test failed with status %d", w.Code)
		}
	})

	// Test 3: Service degradation
	t.Run("ServiceDegradation", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test service degradation
		chaosData := map[string]interface{}{
			"type":     "service_degradation",
			"duration": 30,
			"service":  "cache",
			"level":    0.5, // 50% performance degradation
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Service degradation test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "service_degradation" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under degradation
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System behavior test failed with status %d", w.Code)
		}

		// Check if system is still functional
		var stats map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		if stats["cpu_usage"] == nil {
			t.Error("System not functional under degradation")
		}
	})
}

// TestChaosEngineeringDataChaos tests data chaos scenarios
func TestChaosEngineeringDataChaos(t *testing.T) {
	// Test 1: Data corruption
	t.Run("DataCorruption", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test data corruption
		chaosData := map[string]interface{}{
			"type":     "data_corruption",
			"duration": 30,
			"target":   "database",
			"rate":     0.01, // 1% corruption rate
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data corruption test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "data_corruption" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under data corruption
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// System should handle data corruption gracefully
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("System not handling data corruption gracefully: status %d", w.Code)
		}
	})

	// Test 2: Data loss
	t.Run("DataLoss", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test data loss
		chaosData := map[string]interface{}{
			"type":     "data_loss",
			"duration": 30,
			"target":   "cache",
			"rate":     0.05, // 5% data loss rate
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data loss test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "data_loss" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under data loss
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// System should handle data loss gracefully
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Errorf("System not handling data loss gracefully: status %d", w.Code)
		}
	})

	// Test 3: Data inconsistency
	t.Run("DataInconsistency", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test data inconsistency
		chaosData := map[string]interface{}{
			"type":     "data_inconsistency",
			"duration": 30,
			"target":   "replication",
			"rate":     0.02, // 2% inconsistency rate
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data inconsistency test failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		// Check for chaos injection
		if chaos["chaos_id"] == nil {
			t.Error("Chaos ID not generated")
		}

		if chaos["status"] != "injected" {
			t.Error("Chaos status not injected")
		}

		if chaos["type"] != "data_inconsistency" {
			t.Error("Chaos type not set correctly")
		}

		// Test system behavior under data inconsistency
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// System should handle data inconsistency gracefully
		if w.Code != http.StatusOK && w.Code != http.StatusConflict {
			t.Errorf("System not handling data inconsistency gracefully: status %d", w.Code)
		}
	})
}

// TestChaosEngineeringRecovery tests chaos recovery scenarios
func TestChaosEngineeringRecovery(t *testing.T) {
	// Test 1: Chaos recovery
	t.Run("ChaosRecovery", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Inject chaos first
		chaosData := map[string]interface{}{
			"type":     "cpu_stress",
			"duration": 30,
			"load":     0.8,
		}

		jsonData, _ := json.Marshal(chaosData)
		req, _ := http.NewRequest("POST", "/api/chaos/inject", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Chaos injection failed with status %d", w.Code)
		}

		var chaos map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chaos)
		if err != nil {
			t.Fatalf("Failed to unmarshal chaos response: %v", err)
		}

		chaosID := chaos["chaos_id"].(string)

		// Test chaos recovery
		req, _ = http.NewRequest("POST", "/api/chaos/recover", bytes.NewBufferString(`{"chaos_id": "`+chaosID+`"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Chaos recovery test failed with status %d", w.Code)
		}

		var recovery map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &recovery)
		if err != nil {
			t.Fatalf("Failed to unmarshal recovery response: %v", err)
		}

		// Check for recovery
		if recovery["chaos_id"] != chaosID {
			t.Error("Recovery chaos ID mismatch")
		}

		if recovery["status"] != "recovered" {
			t.Error("Recovery status not recovered")
		}

		if recovery["recovery_time"] == nil {
			t.Error("Recovery time not reported")
		}
	})

	// Test 2: System recovery
	t.Run("SystemRecovery", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test system recovery
		req, _ := http.NewRequest("POST", "/api/chaos/system-recovery", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System recovery test failed with status %d", w.Code)
		}

		var recovery map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &recovery)
		if err != nil {
			t.Fatalf("Failed to unmarshal recovery response: %v", err)
		}

		// Check for system recovery
		if recovery["recovery_id"] == nil {
			t.Error("Recovery ID not generated")
		}

		if recovery["status"] != "completed" {
			t.Error("Recovery status not completed")
		}

		if recovery["components_recovered"] == nil {
			t.Error("Components recovered not listed")
		}

		if recovery["recovery_time"] == nil {
			t.Error("Recovery time not reported")
		}
	})

	// Test 3: Data recovery
	t.Run("DataRecovery", func(t *testing.T) {
		router := setupChaosEngineeringTestRouter(t)
		token := getChaosEngineeringAuthToken(t, router)

		// Test data recovery
		req, _ := http.NewRequest("POST", "/api/chaos/data-recovery", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data recovery test failed with status %d", w.Code)
		}

		var recovery map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &recovery)
		if err != nil {
			t.Fatalf("Failed to unmarshal recovery response: %v", err)
		}

		// Check for data recovery
		if recovery["recovery_id"] == nil {
			t.Error("Recovery ID not generated")
		}

		if recovery["status"] != "completed" {
			t.Error("Recovery status not completed")
		}

		if recovery["data_integrity"] == nil {
			t.Error("Data integrity not checked")
		}

		if recovery["recovery_time"] == nil {
			t.Error("Recovery time not reported")
		}
	})
}

// Helper functions

func setupChaosEngineeringTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:  8080,
			Host:  "0.0.0.0",
			Debug: false,
		},
		Moodle: config.MoodleConfig{
			Path:       "/tmp/moodle",
			ConfigPath: "/tmp/moodle/config.php",
			DataPath:   "/tmp/moodledata",
		},
		Security: config.SecurityConfig{
			JWTSecret:     "test-secret-key",
			SessionTimeout: 3600,
			RateLimit:     100,
			AllowedIPs:   []string{"127.0.0.1"},
		},
		Monitoring: config.MonitoringConfig{
			UpdateInterval: 30,
			LogRetention:   7,
			AlertThresholds: config.AlertThresholdsConfig{
				CPU:    80.0,
				Memory: 85.0,
				Disk:   90.0,
			},
		},
	}

	db := setupChaosEngineeringTestDB(t)

	authService := services.NewAuthService(cfg.Security.JWTSecret, db)
	monitorService := services.NewMonitorService(cfg.Monitoring)
	moodleService := services.NewMoodleService(cfg.Moodle)
	securityService := services.NewSecurityService(cfg.Security)

	authHandler := handlers.NewAuthHandler(authService)
	apiHandler := handlers.NewAPIHandler(monitorService, moodleService, securityService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(securityService.RateLimitMiddleware())

	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)

	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/stats", apiHandler.GetStats)
		protected.GET("/moodle/status", apiHandler.GetMoodleStatus)
		protected.POST("/moodle/start", apiHandler.StartMoodle)
		protected.POST("/moodle/stop", apiHandler.StopMoodle)
		protected.POST("/moodle/restart", apiHandler.RestartMoodle)
		protected.GET("/users", apiHandler.GetUsers)
		protected.POST("/users", apiHandler.CreateUser)
		protected.GET("/users/stats", apiHandler.GetUserStats)
		protected.GET("/alerts", apiHandler.GetAlerts)
		protected.GET("/logs", apiHandler.GetSystemLogs)
	}

	return router
}

func getChaosEngineeringAuthToken(t *testing.T, router *gin.Engine) string {
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	token, ok := response["token"].(string)
	if !ok {
		t.Fatal("Token not found in login response")
	}

	return token
}

func setupChaosEngineeringTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_activities (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			action TEXT NOT NULL,
			resource TEXT,
			ip_address TEXT,
			user_agent TEXT,
			success BOOLEAN DEFAULT 1,
			message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE TABLE IF NOT EXISTS security_events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			severity TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			severity TEXT NOT NULL,
			resolved BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_logs (
			id TEXT PRIMARY KEY,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			source TEXT,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	passwordHash, _ := utils.HashPassword("admin123")
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "admin-id", "admin", "admin@k2net.id", passwordHash, "admin", true)

	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	return db
}
