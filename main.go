package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/microsoft/durabletask-go/api"
	"github.com/microsoft/durabletask-go/backend"
	"github.com/microsoft/durabletask-go/backend/sqlite"
	"github.com/microsoft/durabletask-go/task"
)

var client backend.TaskHubClient

func main() {
	// Configure SQLite backend
	logger := backend.DefaultLogger()
	sqliteBackend := sqlite.NewSqliteBackend(sqlite.NewSqliteOptions("poc-durabletask.db"), logger)
	client = backend.NewTaskHubClient(sqliteBackend)

	// Register orchestrators and activities
	r := task.NewTaskRegistry()
	r.AddOrchestrator(ExternalEventOrchestrator)
	r.AddOrchestrator(RetryOrchestrator)
	r.AddActivity(UnreliableActivity)
	r.AddActivity(ReliableActivity)
	r.AddActivity(ProcessingActivity)

	// Create workers
	ctx := context.Background()
	executor := task.NewTaskExecutor(r)
	orchestrationWorker := backend.NewOrchestrationWorker(sqliteBackend, executor, logger)
	activityWorker := backend.NewActivityTaskWorker(sqliteBackend, executor, logger)
	taskHubWorker := backend.NewTaskHubWorker(sqliteBackend, orchestrationWorker, activityWorker, logger)

	go func() {
		if err := taskHubWorker.Start(ctx); err != nil {
			log.Fatalf("Error starting worker: %v", err)
		}
	}()

	// Configure HTTP routes
	router := mux.NewRouter()
	router.HandleFunc("/start-external-event", startExternalEventHandler).Methods("POST")
	router.HandleFunc("/start-retry", startRetryHandler).Methods("POST")
	router.HandleFunc("/send-event/{instanceId}", sendEventHandler).Methods("POST")
	router.HandleFunc("/status/{instanceId}", statusHandler).Methods("GET")

	fmt.Println("🚀 Server started at http://localhost:8080")
	fmt.Println("📝 Available routes:")
	fmt.Println("  POST /start-external-event - Start orchestrator waiting for external event")
	fmt.Println("  POST /start-retry - Start orchestrator with retry policy")
	fmt.Println("  POST /send-event/{instanceId} - Send external event")
	fmt.Println("  GET /status/{instanceId} - Query status")

	log.Fatal(http.ListenAndServe(":8080", router))
}

// Handler to start orchestrator waiting for external event
func startExternalEventHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n📥 [HTTP] POST /start-external-event - Starting external event orchestrator\n")

	instanceId := api.InstanceID(fmt.Sprintf("external-event-%d", time.Now().Unix()))

	if _, err := client.ScheduleNewOrchestration(context.Background(), ExternalEventOrchestrator, api.WithInstanceID(instanceId)); err != nil {
		fmt.Printf("❌ [HTTP] Failed to start external event orchestrator: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ [HTTP] External event orchestrator started with ID: %s\n", instanceId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"instanceId": string(instanceId),
		"message":    "Orchestrator started. Waiting for external event...",
	})
}

// Handler to start orchestrator with retry policy
func startRetryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n📥 [HTTP] POST /start-retry - Starting retry orchestrator\n")

	instanceId := api.InstanceID(fmt.Sprintf("retry-%d", time.Now().Unix()))

	if _, err := client.ScheduleNewOrchestration(context.Background(), RetryOrchestrator, api.WithInstanceID(instanceId)); err != nil {
		fmt.Printf("❌ [HTTP] Failed to start retry orchestrator: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ [HTTP] Retry orchestrator started with ID: %s\n", instanceId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"instanceId": string(instanceId),
		"message":    "Retry orchestrator started",
	})
}

// Handler to send external event
func sendEventHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := api.InstanceID(vars["instanceId"])

	fmt.Printf("\n📥 [HTTP] POST /send-event/%s - Sending external event\n", instanceId)

	var request struct {
		EventName string      `json:"eventName"`
		Data      interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Printf("❌ [HTTP] Invalid JSON in request body: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("📤 [HTTP] Sending event '%s' with data: %v\n", request.EventName, request.Data)

	if err := client.RaiseEvent(context.Background(), instanceId, request.EventName, api.WithEventPayload(request.Data)); err != nil {
		fmt.Printf("❌ [HTTP] Failed to send event: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ [HTTP] Event '%s' sent successfully to %s\n", request.EventName, instanceId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Event '%s' sent to %s", request.EventName, instanceId),
	})
}

// Handler to query status
func statusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := api.InstanceID(vars["instanceId"])

	fmt.Printf("\n📥 [HTTP] GET /status/%s - Querying orchestration status\n", instanceId)

	metadata, err := client.FetchOrchestrationMetadata(context.Background(), instanceId)
	if err != nil {
		fmt.Printf("❌ [HTTP] Failed to fetch metadata: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("📊 [HTTP] Status for %s: %s\n", instanceId, metadata.RuntimeStatus.String())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"instanceId":    string(instanceId),
		"name":          metadata.Name,
		"runtimeStatus": metadata.RuntimeStatus.String(),
	})
}
