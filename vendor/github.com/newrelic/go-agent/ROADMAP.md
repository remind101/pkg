# Go Agent Roadmap

## Product Vision
The goal of the Go agent is to provide complete visibility into the health of your service. The agent provides metrics about the runtime health of your service and the process it runs in, and traces that show how specific requests are performing. It also provides information about the environment in which it is running, so you can identify issues with specific hosts, regions, deployments, and other facets. 

New Relic is moving toward OpenTelemetry. OpenTelemetry is a unified standard for service instrumentation. You can use our [OpenTelemetry Exporter](https://github.com/newrelic/opentelemetry-exporter-go) today, and will soon see a shim will convert New Relic Go Agent data to OpenTelemetry. OpenTelemetry will include a broad set of high-quality community-contributed instrumentation and a powerful vendor-neutral API for adding your own instrumentation.


## Roadmap
**The Go instrumentation roadmap project is found [here](https://github.com/orgs/newrelic/projects/24)**.  

This roadmap project is broken down into the following sections:

- **Done**:
    - This section contains features that were recently completed.
- **Now**:
    - This section contains features that are currently in progress.
- **Next**:
    - This section contains work planned within the next three months. These features may still be de-prioritized and moved to Future.
- **Future**:
    - This section is for ideas for future work that is aligned with the product vision and possible opportunities for community contribution. It contains a list of features that anyone can implement. No guarantees can be provided on if or when these features will be completed.
     


### Disclaimers
This roadmap is subject to change at any time. Future items should not be considered commitments.
