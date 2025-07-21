# POC: Microsoft DurableTask Framework for Go

This POC demonstrates the use of Microsoft's DurableTask framework in Go, focusing on essential features:

## ✨ Features Demonstrated

1. **External Events**: Multi-step orchestrator with multiple pause/resume points
2. **Retry Policies**: Multi-activity orchestrator showing exactly where execution restarts
3. **Durable Execution**: Activities that completed successfully are never re-executed

## 🚀 How to Run

```bash
# Install dependencies
go mod tidy

# Run the server
go run .
```

The server will start at `http://localhost:8080`

## 📋 API Endpoints

### 1. Start External Event Orchestrator
```bash
curl -X POST http://localhost:8080/start-external-event
```

### 2. Send External Events (Multiple events needed)
```bash
# First event: step2
curl -X POST http://localhost:8080/send-event/external-event-1735124567 \
  -H "Content-Type: application/json" \
  -d '{"eventName": "step2", "data": "ok"}'

# Second event: step4  
curl -X POST http://localhost:8080/send-event/external-event-1735124567 \
  -H "Content-Type: application/json" \
  -d '{"eventName": "step4", "data": "done"}'
```

### 3. **Start Retry Orchestrator (MAIN DEMONSTRATION)**
```bash
curl -X POST http://localhost:8080/start-retry
```

**What you'll see in the console:**
```
🎯 Retry orchestrator started...
🚀 Calling Activity 1 (Always succeeds)...
🔄 ReliableActivity execution #1: Step 1: Initialize
✅ ReliableActivity succeeded: Step 1: Initialize completed at 15:30:42
✅ Activity 1 completed: Step 1: Initialize completed at 15:30:42

🚀 Calling Activity 2 (Always succeeds)...
🔄 ReliableActivity execution #2: Step 2: Prepare data
✅ ReliableActivity succeeded: Step 2: Prepare data completed at 15:30:43
✅ Activity 2 completed: Step 2: Prepare data completed at 15:30:43

🚀 Calling Activity 3 (Will fail and retry)...
🔄 UnreliableActivity - Attempt #1
❌ Attempt #1: Network connection error (simulated)
⏳ Framework will retry in 1s...
💡 NOTE: Activities 1 & 2 will NOT re-execute on retry!

🎯 Retry orchestrator started...
🔄 UnreliableActivity - Attempt #2
❌ Attempt #2: Database timeout (simulated)
⏳ Framework will retry in 2s...
💡 NOTE: Activities 1 & 2 will NOT re-execute on retry!

🎯 Retry orchestrator started...
🔄 UnreliableActivity - Attempt #3
❌ Attempt #3: Service temporarily unavailable (simulated)
⏳ Framework will retry in 4s...
💡 NOTE: Activities 1 & 2 will NOT re-execute on retry!

🎯 Retry orchestrator started...
🔄 UnreliableActivity - Attempt #4
✅ Success on attempt #4: Step 3: Critical operation completed at 15:30:55
🎉 RETRY POLICY WORKED! Activity succeeded after 4 attempts!
✅ Activity 3 completed: Step 3: Critical operation completed at 15:30:55

🚀 Calling Activity 4 (Final step)...
🔄 ReliableActivity execution #3: Step 4: Finalize
✅ ReliableActivity succeeded: Step 4: Finalize completed at 15:30:55
✅ Activity 4 completed: Step 4: Finalize completed at 15:30:55

🎉 ORCHESTRATION COMPLETED SUCCESSFULLY!
```

### 4. Query Status
```bash
curl http://localhost:8080/status/external-event-1735124567
```

## 🔧 How Durable Execution Works

### **Key Insight: No Re-execution of Completed Activities**
When the retry orchestrator restarts due to Activity 3 failing:
- ✅ **Activity 1** (completed) - **NOT re-executed**
- ✅ **Activity 2** (completed) - **NOT re-executed**  
- ❌ **Activity 3** (failed) - **Retried from this point**
- ⏳ **Activity 4** (not reached yet) - **Executed after Activity 3 succeeds**

### Retry Configuration:
- **MaxAttempts**: 5 maximum attempts
- **InitialRetryInterval**: 1 second initial interval
- **MaxRetryInterval**: 10 seconds maximum interval  
- **BackoffCoefficient**: 2.0 (doubles interval each attempt)

### Retry Intervals:
- Attempt 1 → Failure → Wait 1s
- Attempt 2 → Failure → Wait 2s  
- Attempt 3 → Failure → Wait 4s
- Attempt 4 → **Success!** ✅

### External Events with Multiple Pause Points
The external event orchestrator demonstrates:
1. **Step 1**: Initial processing activity
2. **⏸️  PAUSE**: Wait for `step2` event
3. **▶️  RESUME**: Process step2 data  
4. **⏸️  PAUSE**: Wait for `step4` event
5. **▶️  RESUME**: Final processing

## 💡 Step-by-Step Demonstration

1. **Test Retry Policy (MAIN - Shows Restart Behavior):**
   ```bash
   # Execute and observe console - notice Activities 1&2 are NOT re-executed
   curl -X POST http://localhost:8080/start-retry
   ```

2. **Test External Events (Shows Multiple Pause/Resume):**
   ```bash
   # Start orchestrator
   curl -X POST http://localhost:8080/start-external-event
   
   # Send first event (step2)
   curl -X POST http://localhost:8080/send-event/YOUR_INSTANCE_ID \
     -H "Content-Type: application/json" \
     -d '{"eventName": "step2", "data": "ok"}'
   
   # Send second event (step4)
   curl -X POST http://localhost:8080/send-event/YOUR_INSTANCE_ID \
     -H "Content-Type: application/json" \
     -d '{"eventName": "step4", "data": "done"}'
   ```

## 🎯 What This Demonstrates

- **Durable Execution**: Completed activities are never re-executed, even on retries
- **Precise Restart Points**: Orchestrator resumes exactly where it left off
- **Multiple Pause/Resume**: External events can pause workflow at any point
- **State Persistence**: All progress is preserved across restarts
- **Retry Policies**: Automatic retry with exponential backoff

## 📁 File Structure

- `main.go`: HTTP server and framework configuration
- `external_events_example.go`: Multi-step orchestrator with pause/resume points
- `retry_example.go`: Multi-activity orchestrator demonstrating restart behavior 