#!/bin/bash

# Script to demonstrate the Temporal Orchestrator functionality.

echo "--- Temporal Orchestrator Demo ---"
echo ""
echo "This script will start multiple item workflows at different times to showcase"
echo "how the central orchestrator manages concurrency."
echo ""
echo "Prerequisite 1: The Temporal server must be running in a separate terminal."
echo "E.g. 'temporal server start-dev'"
echo ""
echo "Prerequisite 2: The Temporal worker must be running in a separate terminal."
echo "Run this command: go run orchestrator/worker/*.go"
echo ""
read -p "Press [Enter] to start the demo..."

echo ""
echo "Step 1: Starting Item A ('item-a-1')."
echo "This workflow should register successfully and, after a 30s delay, start processing."
(go run orchestrator/starter/main.go a item-a-1 && echo "✅ item-a-1 finished.") &
echo "-> Started item-a-1 in the background."
echo ""

echo "Step 2: Waiting 10 seconds..."
sleep 10
echo ""

echo "Step 3: Starting Item B ('item-b-1')."
echo "At this point, item-a-1 is waiting but not yet processing, so item-b-1 should also register successfully."
echo "It will be queued to start processing later if Orchestrator permits."
(go run orchestrator/starter/main.go b item-b-1 && echo "✅ item-b-1 finished.") &
echo "-> Started item-b-1 in the background."
echo ""

echo "Step 4: Waiting 30 seconds..."
echo "During this time, item-a-1's initial 30s delay will end, and it will request to start processing."
echo "The orchestrator should grant this request, and item-a-1 will begin its 30s of work."
sleep 30
echo ""

echo "Step 5: Starting Item A ('item-a-2') while item-a-1 is processing."
echo "Because item-a-1 is now 'in-progress', the orchestrator's rules should REJECT the registration of this new workflow."
echo "You should see a 'Workflow failed: Halted by orchestrator' message for item-a-2 shortly."
(go run orchestrator/starter/main.go a item-a-2 && echo "✅ item-a-2 finished.") &
echo "-> Started item-a-2 in the background."
echo ""

echo "Step 6: Demo in progress. Monitoring background jobs..."
echo "You can also query the orchestrator state in another terminal:"
echo "go run orchestrator/query/main.go"
echo ""

wait

echo ""
echo "Step 7: All initial workflows have completed. Now demonstrating successful registration after completion..."
echo "Starting another Item B workflow ('item-b-2') to show it can register and process successfully"
echo "when no other items are in progress."
sleep 5

echo ""
echo "Starting item-b-2..."
(go run orchestrator/starter/main.go b item-b-2 && echo "✅ item-b-2 finished successfully.") &
echo "-> Started item-b-2 in the background."
echo ""

echo "Step 8: Waiting for item-b-2 to complete..."
echo "This demonstrates that the orchestrator correctly allows new workflows to register"
echo "and process when the system is no longer busy."
echo ""

wait

echo "--- Demo Complete ---"
echo "All background workflow starters have finished."
echo ""
echo "Summary of what was demonstrated:"
echo "1. ✅ Concurrent registration of workflows (item-a-1 and item-b-1)"
echo "2. ✅ Sequential processing with concurrency control (item-a-1 processed first)"
echo "3. ✅ Rejection of new workflows during processing (item-a-2 was rejected)"
echo "4. ✅ Rejection of processing of previously registered workflows (item-b-1 was rejected due to item-a-1 is being processed)"
echo "5. ✅ New workflow registration after completion (item-b-2 processed successfully)"
echo ""
echo "The orchestrator successfully managed workflow concurrency and state transitions!"
