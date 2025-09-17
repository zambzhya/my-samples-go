# Temporal Orchestrator Sample

This sample demonstrates a sophisticated "Central Orchestrator" pattern using Temporal. A long-running, singleton `OrchestratorWorkflow` orchestrates multiple, shorter-lived `ItemWorkflow` instances.

## Core Concepts & Features

- **Singleton Orchestrator**: The `OrchestratorWorkflow` acts as a central point of control. It is robustly started using the `SignalWithStartWorkflow` pattern, ensuring there's only ever one instance running.
- **Dynamic Child Workflows**: `ItemWorkflow` instances are created dynamically to perform work on specific items. The sample demonstrates this with different item types (`ItemA`, `ItemB`), showcasing how to handle varied payloads.
- **Stateful Orchestration**: The orchestrator maintains a complete state of all `ItemWorkflow` instances it's aware of, including their registration status, processing status, and their specific data payloads.
- **Two-Step Concurrency Control**: A strict rule is enforced: **only one item can be processing at a time**. This is achieved with a two-step locking mechanism:
  1.  An `ItemWorkflow` first signals the orchestrator to `Register`.
  2.  After a delay, it must explicitly signal again to `StartProcessing`. The orchestrator will deny this request if another item is already processing.
- **Two-Way Signaling**: The pattern uses bidirectional communication. The `ItemWorkflow` sends signals to request state changes, and the `OrchestratorWorkflow` sends `ItemInstructionSignal`s back to grant or deny those requests ("go/no-go").
- **Graceful Timeout**: The singleton `OrchestratorWorkflow` will only time out and complete after a period of inactivity *and* when no items are currently registered.

## Workflow Lifecycle

1.  A **Starter** application initiates an `ItemWorkflow`.
2.  The `ItemWorkflow` immediately sends a `RegisterSignal` to the `OrchestratorWorkflow`.
3.  The `OrchestratorWorkflow` receives the signal, updates its internal state, and sends an instruction signal back to the `ItemWorkflow` to confirm registration.
4.  The `ItemWorkflow` waits for a period (30s) and then sends a `StartProcessingSignal`.
5.  The `OrchestratorWorkflow` checks if any other item is currently processing.
    - If **yes**, it sends a "no-go" signal back, and the `ItemWorkflow` terminates.
    - If **no**, it marks the item as "in-progress" and sends a "go" signal.
6.  The `ItemWorkflow` receives the "go" signal, performs its work (simulated by a 30s sleep), and sends status `UpdateSignal`s to the orchestrator.
7.  Upon completion, the `ItemWorkflow` sends a `StopProcessingSignal` and a `DeregisterSignal`.
8.  The `OrchestratorWorkflow` updates its state, freeing up the processing slot for another item.

## Query Support

The `OrchestratorWorkflow` supports a query (`orchestrator-query-list-orchestrated-items`) that returns a detailed snapshot of its current state, including:
- Total number of items being tracked.
- A map of all `OrchestratedItem`s with their full state (ID, workflow IDs, payload, in-progress status).
- The total number of signals handled.

## How to Run

First, ensure you have a [local Temporal server running](../../README.md#running-a-temporal-server-locally).

1. **Start the worker:**
   ```sh
   go run orchestrator/worker/*.go
   ```

2.  **Start Item Workflows:**
    In separate terminals, start one or more item workflows. The starter takes two arguments: `item_type` and `item_id`.
    ```sh
    # Start a workflow for item-1 of type 'a'
    go run orchestrator/starter/main.go a item-1

    # Start a workflow for item-2 of type 'b'
    go run orchestrator/starter/main.go b item-2
    ```
    You will observe that `item-1` registers and, after 30 seconds, gets approval to process. When `item-2` attempts to register, it will be accepted, but its subsequent request to *process* will be denied because `item-1` is already processing.

3.  **Query the Orchestrator's State:**
    While the workflows are running, you can query the `OrchestratorWorkflow` to see the state of all items.
    ```sh
    go run orchestrator/query/main.go
    ```
    The output will be a JSON representation of the items managed by the orchestrator.

4.  **Run the Automated Demo:**
    A shell script is provided to demonstrate the full lifecycle, including the concurrency control.
    ```sh
    ./orchestrator/run_demo.sh
    ```

## Code Structure

- `worker/main.go`: Contains the main `OrchestratorWorkflow` logic and the worker registration.
- `worker/item_workflow.go`: Defines the `ItemWorkflow` that performs the actual work.
- `starter/main.go`: The client application to start new `ItemWorkflow` instances.
- `query/main.go`: The client application to query the `OrchestratorWorkflow`.
- `orchestrator.go`: Defines the core orchestration logic and state management, decoupled from the workflow itself.
- `*.go` (at root of `orchestrator/`): These files (`signals.go`, `payload.go`, `item.go`, etc.) define the shared data structures, constants, and interfaces used across the sample.
