# LLM SSE Service

## Description

LLM SSE Service is a lightweight Go-based orchestration layer for managing multi-agent LLM interactions with real-time streaming. It concurrently executes multiple LLM tasks (e.g., LLM1, LLM2), aggregates their outputs through a summarizing LLM (LLM3), and streams the final response back to the client using Server-Sent Events (SSE).

### Configuration

There is an ```.env``` file which contains main configuration options that are presented here:

* ```LLM_KEY=``` should be filled with an actual OpenAI key for interactions with the LLM.

* ```USE_LLM_MOCK=``` tells the service to use mocked LLM for manual or functionality testing. Default is *"true"*.

If ```USE_LLM_MOCK``` is false and ```LLM_KEY``` is not presented there will be an API error related to an empty API key.

* ```LOG_LEVEL=``` tells the logger what log level to use. Default is *"info"*.

* ```PRODUCTION=``` tells what environment the service is operating in. Currently it affects only log configuration. Default value is *"false"* in the config and left as *"true"* in the config file for convenience.

* ```HTTP_ADDR=``` what address and port system should be operating with. Default value is *":8080"*.

### Commands for service operations

```make service-build``` - for building the service

```make service-up``` - for starting the service. It should be build beforehand

```make service-logs``` - for viewing logs

```make service-down``` - for stopping the service

```make test``` - for test execution

```make test-coverage``` - for test execution with coverage report generation

```make test-coverage-detailed``` - for detailed coverage report presentation in the cli

```make test-coverage-visualize``` - for visualization of the coverage as html via browser

### Examples of requests

```
curl -N -X POST http://localhost:8080/api/process \
  -H "Content-Type: application/json" \
  -d '{"message":"Explain gravity like I am 5", "message_id":"message_123"}'
```

```
curl -N -X POST http://localhost:8080/api/process \
  -H "Content-Type: application/json" \
  -d '{"message":"Build me a robot", "message_id":"message_123"}'
```

### Future improvements list

* **Configurable Model Usage**
Support configurable model selection to be agnostic about the underlying LLM API provider. This would allow switching between cloud-based models and local models (e.g., Ollama) seamlessly.

* **Partial Response Handling from LLMs**
In cases where some LLM calls fail but others succeed, allow the system to proceed with the successful responses. These partial results can be sent to the combining LLM, while the failures are logged and monitored. This improves resilience and degrades gracefully instead of halting the entire pipeline.

* **External and Persistent Prompt Configuration**
Move prompts to an external configuration or persistent storage. This provides better flexibility and allows dynamic updates to prompts without code changes or redeployments.

* **Standardized and Layered Error Handling**
Introduce a standard error structure (with wrapping and unwrapping across layers) to improve debugging, error propagation, and control flow — especially as the system grows more complex.

* **Distributed Tracing and Monitoring**
Add tracing (e.g., OpenTelemetry) and improved metrics to observe the system's performance and internal behavior. This can help detect bottlenecks, failed requests, and latency spikes.


* **Conversation Storage**
Store user conversations to improve UX (resumable sessions, history, etc.) and unlock future capabilities such as analytics, fine-tuning, or recommendations. This can provide product insights and drive feature development.

* **Distributed Rate Limiting**
Evolve the in-memory rate limiter to a distributed version using tools like Redis with sorted sets or token buckets. This ensures scalability across multiple server instances.

* **Authentication & Access Control**
Introduce support for authentication (e.g., JWT, API keys, or OAuth) to protect endpoints from unauthorized access. This enables user-level restrictions, usage-based billing, and role-based access control in multi-tenant deployments. It would also allow tracking message ownership (message_id / conversation_id) per authenticated user for improved analytics and security.
