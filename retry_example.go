package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/microsoft/durabletask-go/task"
)

// Global counter to simulate deterministic attempts
var (
	attemptCounter int
	counterMutex   sync.Mutex
	activityCount  int
	activityMutex  sync.Mutex
)

// RetryOrchestrator - Tests retry policy with multiple activities to show restart behavior
func RetryOrchestrator(ctx *task.OrchestrationContext) (any, error) {
	fmt.Printf("\n🎯 [ORCHESTRATOR] Retry orchestrator started/restarted\n")
	fmt.Printf("📋 [ORCHESTRATOR] This orchestrator demonstrates that completed activities are NOT re-executed\n")

	// Configure retry policy
	retryPolicy := &task.RetryPolicy{
		MaxAttempts:          5,
		InitialRetryInterval: 1 * time.Second,
		MaxRetryInterval:     10 * time.Second,
		BackoffCoefficient:   2.0,
	}

	fmt.Printf("⚙️  [ORCHESTRATOR] Retry Policy: MaxAttempts=%d, InitialInterval=1s, BackoffCoeff=2.0\n", retryPolicy.MaxAttempts)

	// Activity 1: Always succeeds (to show it doesn't re-execute)
	var result1 string
	fmt.Printf("\n🚀 [ORCHESTRATOR] Calling Activity 1 (ReliableActivity - Always succeeds)...\n")
	if err := ctx.CallActivity(ReliableActivity, task.WithActivityInput("Step 1: Initialize")).Await(&result1); err != nil {
		fmt.Printf("❌ [ORCHESTRATOR] Activity 1 FAILED: %v\n", err)
		return fmt.Sprintf("❌ Activity 1 failed: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Activity 1 COMPLETED: %s\n", result1)

	// Activity 2: Always succeeds (to show it doesn't re-execute)
	var result2 string
	fmt.Printf("\n🚀 [ORCHESTRATOR] Calling Activity 2 (ReliableActivity - Always succeeds)...\n")
	if err := ctx.CallActivity(ReliableActivity, task.WithActivityInput("Step 2: Prepare data")).Await(&result2); err != nil {
		fmt.Printf("❌ [ORCHESTRATOR] Activity 2 FAILED: %v\n", err)
		return fmt.Sprintf("❌ Activity 2 failed: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Activity 2 COMPLETED: %s\n", result2)

	// Activity 3: Fails a few times then succeeds (with retry policy)
	fmt.Printf("\n🚀 [ORCHESTRATOR] Calling Activity 3 (UnreliableActivity - Will fail and trigger retries)...\n")
	fmt.Printf("🔄 [ORCHESTRATOR] Activity 3 has retry policy enabled - will auto-retry on failures\n")
	var result3 string
	if err := ctx.CallActivity(UnreliableActivity, task.WithActivityRetryPolicy(retryPolicy)).Await(&result3); err != nil {
		// Reset counter after complete failure
		counterMutex.Lock()
		attemptCounter = 0
		counterMutex.Unlock()
		fmt.Printf("❌ [ORCHESTRATOR] Activity 3 FAILED after all retry attempts: %v\n", err)
		return fmt.Sprintf("❌ Activity 3 failed after all attempts: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Activity 3 COMPLETED: %s\n", result3)

	// Activity 4: Always succeeds (to show orchestrator continues normally)
	var result4 string
	fmt.Printf("\n🚀 [ORCHESTRATOR] Calling Activity 4 (ReliableActivity - Final step)...\n")
	if err := ctx.CallActivity(ReliableActivity, task.WithActivityInput("Step 4: Finalize")).Await(&result4); err != nil {
		fmt.Printf("❌ [ORCHESTRATOR] Activity 4 FAILED: %v\n", err)
		return fmt.Sprintf("❌ Activity 4 failed: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Activity 4 COMPLETED: %s\n", result4)

	// Reset counters after success
	counterMutex.Lock()
	attemptCounter = 0
	counterMutex.Unlock()

	activityMutex.Lock()
	activityCount = 0
	activityMutex.Unlock()

	finalResult := fmt.Sprintf("✅ All activities completed! Results: [%s] -> [%s] -> [%s] -> [%s]",
		result1, result2, result3, result4)
	fmt.Printf("\n🎉 [ORCHESTRATOR] RETRY ORCHESTRATION COMPLETED SUCCESSFULLY!\n")
	fmt.Printf("📊 [ORCHESTRATOR] Final Result: %s\n\n", finalResult)
	return finalResult, nil
}

// ReliableActivity - Always succeeds to demonstrate orchestrator replay behavior
func ReliableActivity(ctx task.ActivityContext) (any, error) {
	var input string
	if err := ctx.GetInput(&input); err != nil {
		input = "No input"
	}

	activityMutex.Lock()
	activityCount++
	currentCount := activityCount
	activityMutex.Unlock()

	fmt.Printf("🔄 [ACTIVITY] ReliableActivity execution #%d STARTED: %s\n", currentCount, input)

	// Simulate some processing time
	fmt.Printf("⚙️  [ACTIVITY] ReliableActivity processing for 200ms...\n")
	time.Sleep(200 * time.Millisecond)

	timestamp := time.Now().Format("15:04:05")
	result := fmt.Sprintf("%s completed at %s", input, timestamp)

	fmt.Printf("✅ [ACTIVITY] ReliableActivity execution #%d SUCCEEDED: %s\n", currentCount, result)
	return result, nil
}

// UnreliableActivity - Simulates an activity that fails deterministically on first attempts
func UnreliableActivity(ctx task.ActivityContext) (any, error) {
	// Increment attempt counter
	counterMutex.Lock()
	attemptCounter++
	currentAttempt := attemptCounter
	counterMutex.Unlock()

	fmt.Printf("🔄 [ACTIVITY] UnreliableActivity - Attempt #%d STARTED\n", currentAttempt)

	// Simulate processing
	fmt.Printf("⚙️  [ACTIVITY] UnreliableActivity processing for 500ms...\n")
	time.Sleep(500 * time.Millisecond)

	// Fail deterministically on first 3 attempts
	if currentAttempt <= 3 {
		errorTypes := map[int]string{
			1: "Network connection error (simulated)",
			2: "Database timeout (simulated)",
			3: "Service temporarily unavailable (simulated)",
		}

		errorMsg := errorTypes[currentAttempt]
		fmt.Printf("❌ [ACTIVITY] UnreliableActivity - Attempt #%d FAILED: %s\n", currentAttempt, errorMsg)

		// Calculate wait time for next attempt (exponential backoff)
		waitTime := 1 << (currentAttempt - 1) // 1s, 2s, 4s
		fmt.Printf("⏳ [FRAMEWORK] Framework will retry in %ds...\n", waitTime)
		fmt.Printf("💡 [NOTE] Activities 1 & 2 will NOT re-execute on retry - only failed activity retries!\n")

		return nil, fmt.Errorf(errorMsg)
	}

	// Success on 4th attempt or later
	timestamp := time.Now().Format("15:04:05")
	successMsg := fmt.Sprintf("Step 3: Critical operation completed at %s", timestamp)
	fmt.Printf("✅ [ACTIVITY] UnreliableActivity - Attempt #%d SUCCEEDED: %s\n", currentAttempt, successMsg)
	fmt.Printf("🎉 [ACTIVITY] RETRY POLICY WORKED! Activity succeeded after %d attempts!\n", currentAttempt)

	return successMsg, nil
}
