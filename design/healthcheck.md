# Envoy Status Probes

Status: _Draft_

This proposal describes how to add better support for identifying the status of Envoy connections during a restart or reploy of Envoy.

## Goals

- Provide a mechanism to provide feedback on open connections & load inside an Envoy process
- Allow for a Envoy fleet roll-out minimizing connection loss

## Non Goals

- 

## Background

The Envoy process, the data path component of Contour, at times needs to be re-deployed.
This could be due to an upgrade or a node-failure forcing a redeployment.

The proper way to implement this process is to first send a `/healthcheck/fail` request to Envoy which triggers the current readiness probe
to fail, which in turn causes that instance of Envoy to stop accepting new connections and begin draining existing.

The main problem is that the `preStop` hook (which sends the `/healthcheck/fail` request) does not wait until Envoy has drained all connections.

## High-Level Design

This design looks to add a new component to Contour which will allow for a way to understand if open connections exist in Envoy before sending a `SIGTERM`.

1. A new sub-command of Contour named `healthcheck` which will serve `/envoy/healthcheck`
2. Upon a request to this endpoint, it will respond with 200 if Envoy can be killed and 503 if not

## Detailed Design

1. A new sub-command will be added to Contour which serves up 
2. 
2. This will serve up an endpoint named `/envoy/healthcheck/fail` which will 
3. 
