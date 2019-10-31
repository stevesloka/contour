# Contour [![Build Status](https://travis-ci.org/projectcontour/contour.svg?branch=master)](https://travis-ci.org/projectcontour/contour) [![Go Report Card](https://goreportcard.com/badge/github.com/projectcontour/contour)](https://goreportcard.com/report/github.com/projectcontour/contour) ![GitHub release](https://img.shields.io/github/release/projectcontour/contour.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

![Contour is fun at parties!](/img/Contour.svg)

## Overview
Contour is an Ingress controller for Kubernetes that works by deploying the [Envoy proxy](https://www.envoyproxy.io/) as a reverse proxy and load balancer.
Contour supports dynamic configuration updates out of the box while maintaining a lightweight profile.

Contour also introduces a new ingress API ([HTTPProxy]({{ httpproxy.md | prepend: page.version | prepend: "docs" | relative_url }})) which is implemented via a Custom Resource Definition (CRD).
Its goal is to expand upon the functionality of the Ingress API to allow for a richer user experience as well as solve shortcomings in the original design.

## Prerequisites
Contour is tested with Kubernetes clusters running version 1.10 and later, but should work with earlier versions where Custom Resource Definitions are supported (Kubernetes 1.7+).

RBAC must be enabled on your cluster.

## Get started
Getting started with Contour is as simple as one command.
See the [Getting Started](https://projectcontour.io/getting-started) document.
