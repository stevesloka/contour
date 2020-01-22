# Envoy Status Probes

Status: _Draft_

This proposal describes how to add better support for identifying the status of Envoy connections during a restart or reploy of Envoy.

## Goals

- Provide a mechanism to provide feedback on open connections & load inside an Envoy process
- Allow for a Envoy fleet roll-out minimizing connection loss

## Non Goals

- 

## Background

