
# Go Samples Repository

This repository contains a collection of Go code samples, demonstrating various patterns, libraries, and technologies. It is intended as a playground and reference for Go developers. The first category of samples focuses on [Temporal](https://temporal.io/) workflows using the Temporal Go SDK, but additional categories and samples may be added over time.

## Purpose

- **Temporal workflows**: Offer practical examples of building, running, and testing workflows with Temporal.
- **Experiment and learn**: Provide a space to try out new ideas, patterns, and technologies in Go.
- **Showcase Go patterns**: Demonstrate idiomatic Go code, libraries, and best practices.

## Repository Structure

- `temporal/` — Samples using the Temporal Go SDK
  - Each sample is in its own folder:
    - `README.md` — Brief description of the sample
    - Workflow definition for the sample
    - `worker/main.go` — Worker implementation
    - `starter/main.go` — Workflow starter/client
- Other categories (e.g., `grpc/`, `http/`, `concurrency/`, etc.) may be added in the future
- `README.md` — This file

## Getting Started

1. **Clone the repository:**
   ```sh
   git clone https://github.com/zambzhya/my-samples-go.git
   cd my-samples-go
   ```
2. **Install dependencies:**
   Ensure you have Go installed (1.22+ recommended). Then run:
   ```sh
   go mod tidy
   ```
3. **Run samples:**
   Each sample will have its own instructions in its directory. Typically, you can run a sample with:
   ```sh
   go run ./<category>/<sample>/worker/main.go
   go run ./<category>/<sample>/starter/main.go
   ```

### Running a Temporal Server Locally

To run Temporal samples, you need access to a Temporal server. There are several ways to run a local Temporal server. For example, you can run it locally as described in the [Temporal Getting Started Guide](https://learn.temporal.io/getting_started/go/dev_environment/):

Refer to the [official documentation](https://docs.temporal.io/) for more options and details.

## Resources

- [Temporal Go SDK Documentation](https://docs.temporal.io/go/introduction/)
- [Temporal Samples (Official)](https://github.com/temporalio/samples-go)


## License

MIT License
