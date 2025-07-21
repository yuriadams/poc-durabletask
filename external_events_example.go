package main

import (
	"fmt"
	"time"

	"github.com/microsoft/durabletask-go/task"
)

// ExternalEventOrchestrator - Demonstrates multiple pause/resume points with external events
func ExternalEventOrchestrator(ctx *task.OrchestrationContext) (any, error) {
	fmt.Printf("\n🎯 [ORCHESTRATOR] External Event Orchestrator started\n")

	// Step 1: Initial processing
	var step1Result string
	fmt.Printf("🚀 [ORCHESTRATOR] Step 1: Starting initial setup activity...\n")
	if err := ctx.CallActivity(ProcessingActivity, task.WithActivityInput("Step 1: Initialize")).Await(&step1Result); err != nil {
		fmt.Printf("❌ [ORCHESTRATOR] Step 1 FAILED: %v\n", err)
		return fmt.Sprintf("❌ Step 1 failed: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Step 1 COMPLETED: %s\n", step1Result)

	// Step 2: Wait for step2 event
	fmt.Printf("\n⏸️  [ORCHESTRATOR] PAUSING EXECUTION - Waiting for 'step2' event...\n")
	fmt.Printf("📞 [ORCHESTRATOR] Timeout: 60 seconds\n")
	fmt.Printf("💡 [HINT] Send event: curl -X POST http://localhost:8080/send-event/{instanceId} -H 'Content-Type: application/json' -d '{\"eventName\":\"step2\",\"data\":\"ok\"}'\n")

	var step2Data string
	if err := ctx.WaitForSingleEvent("step2", 60*time.Second).Await(&step2Data); err != nil {
		fmt.Printf("⏰ [ORCHESTRATOR] TIMEOUT - No step2 event received in 60 seconds\n")
		return "❌ Timeout: No step2 event received in 60 seconds", nil
	}

	fmt.Printf("▶️  [ORCHESTRATOR] RESUMING EXECUTION - Step2 event received: %v\n", step2Data)

	// Step 3: Process step2 data
	var step3Result string
	fmt.Printf("🚀 [ORCHESTRATOR] Step 3: Processing step2 data...\n")
	if err := ctx.CallActivity(ProcessingActivity, task.WithActivityInput(fmt.Sprintf("Step 3: Process %s", step2Data))).Await(&step3Result); err != nil {
		fmt.Printf("❌ [ORCHESTRATOR] Step 3 FAILED: %v\n", err)
		return fmt.Sprintf("❌ Step 3 failed: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Step 3 COMPLETED: %s\n", step3Result)

	// Step 4: Wait for step4 event
	fmt.Printf("\n⏸️  [ORCHESTRATOR] PAUSING EXECUTION - Waiting for 'step4' event...\n")
	fmt.Printf("📞 [ORCHESTRATOR] Timeout: 60 seconds\n")
	fmt.Printf("💡 [HINT] Send event: curl -X POST http://localhost:8080/send-event/{instanceId} -H 'Content-Type: application/json' -d '{\"eventName\":\"step4\",\"data\":\"done\"}'\n")

	var step4Data string
	if err := ctx.WaitForSingleEvent("step4", 60*time.Second).Await(&step4Data); err != nil {
		fmt.Printf("⏰ [ORCHESTRATOR] TIMEOUT - No step4 event received in 60 seconds\n")
		return "❌ Timeout: No step4 event received in 60 seconds", nil
	}

	fmt.Printf("▶️  [ORCHESTRATOR] RESUMING EXECUTION - Step4 event received: %v\n", step4Data)

	// Step 5: Final processing
	var step5Result string
	fmt.Printf("🚀 [ORCHESTRATOR] Step 5: Final processing...\n")
	if err := ctx.CallActivity(ProcessingActivity, task.WithActivityInput(fmt.Sprintf("Step 5: Finalize %s", step4Data))).Await(&step5Result); err != nil {
		fmt.Printf("❌ [ORCHESTRATOR] Step 5 FAILED: %v\n", err)
		return fmt.Sprintf("❌ Step 5 failed: %v", err), nil
	}
	fmt.Printf("✅ [ORCHESTRATOR] Step 5 COMPLETED: %s\n", step5Result)

	// Final result
	finalResult := fmt.Sprintf("✅ Workflow completed! Events: [%s, %s], Results: [%s] -> [%s] -> [%s]",
		step2Data, step4Data, step1Result, step3Result, step5Result)

	fmt.Printf("\n🎉 [ORCHESTRATOR] EXTERNAL EVENT WORKFLOW COMPLETED SUCCESSFULLY!\n")
	fmt.Printf("📊 [ORCHESTRATOR] Final Result: %s\n\n", finalResult)
	return finalResult, nil
}

// ProcessingActivity - Simulates processing steps in the workflow
func ProcessingActivity(ctx task.ActivityContext) (any, error) {
	var input string
	if err := ctx.GetInput(&input); err != nil {
		input = "No input"
	}

	fmt.Printf("🔄 [ACTIVITY] ProcessingActivity STARTED: %s\n", input)

	// Simulate processing time
	fmt.Printf("⚙️  [ACTIVITY] Processing for 300ms...\n")
	time.Sleep(300 * time.Millisecond)

	timestamp := time.Now().Format("15:04:05")
	result := fmt.Sprintf("%s (processed at %s)", input, timestamp)

	fmt.Printf("✅ [ACTIVITY] ProcessingActivity COMPLETED: %s\n", result)
	return result, nil
}
