### Steps to run this sample:
1) Run a [Temporal service](../../README.md#running-a-temporal-server-locally).
2) Run the following command to start the worker (from `temporal` dir)
    ```
    go run helloworld/worker/main.go
    ```
3) Run the following command to start the example (from `temporal` dir)
    ```
    go run helloworld/starter/main.go
    ```
