# Ingress Route Overview

## Overview 

The Ingress object was added to Kubernetes in version 1.2 to describe properties of a cluster-wide reverse HTTP proxy. 
The goal of the `IngressRoute` Custom Resource Definition (CRD) is to extend the functionality of the original Ingress object idea and allow for a richer user experience as well as solve many of the shortcomings that exist today. 

### Features of IngressRoute

- Multiple upstreams: An IngressRoute object and reference multiple services and load balance traffic across more than once service
- Delegation: Root IngressRoute objects can delegate portions of domains/paths to teams. This allows for safe sharing of resources without the risk of a single object breaking the entire routing system
- Security: Much like Delegation, Administrators can limit the access Teams have to paths of a Domain. Additionally, Administrators are empowered to manage TLS certificates in a more secure manner. 
- Validation: Better validation of resources at creation time
- Annotations: CRDs are a great way to extend the Kubernetes API and avoid the many Annotations that exist with typical Ingress resources

## Root / Delegation

A core feature of the `IngressRoute` CRD is delegation which follows the working model of DNS. 
As the owner of a DNS domain, for example .com, I delegate to another nameserver the responsibility for handing the subdomain heptio.com. 
Any nameserver can hold a record for heptio.com, but without the linkage from the parent .com TLD, its information is unreachable and non authoritative.   

The `Root` IngressRoute is the top level entry point for a domain and is used as the top level configuration of a cluster's ingress resources. 
Each Root IngressRoute starts at a virtual host, which describes properties such as the fully qualified name of the virtual host, any aliases of the vhost (for example, a www. prefix), TLS configuration, and possibly global access list details. 
The vertices of a graph do not contain virtual host information. Instead they are reachable from a root only by delegation. This permits the owner of an ingress root to both delegate the authority to publish a service on a portion of the route space inside a virtual host, and to further delegate authority to publish and delegate.

In practice the linkage, or delegation, from root to vertex, is performed with a specific type of route action. You can think of it as routing traffic to another ingress route for further processing, instead of routing traffic directly to a service.

#-- Roots are used by Administrators to abstract away functionality from Teams, such as TLS secrets, as well as and allow for route delegation.