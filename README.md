# Alcatraz rest

https://alcatraz.rest/api/ping

The why you should hire me in a single repo :). I may consider selling the domain if hired.

![GitHub License](https://img.shields.io/github/license/ihatemodels/alcatraz-rest)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/ihatemodels/alcatraz-rest/ci.yml)
![GitHub go.mod Go version (branch)](https://img.shields.io/github/go-mod/go-version/ihatemodels/alcatraz-rest/main)

## Table of Contents

- **[About](#about)**
    - **[Structure](#structure)**
    - **[Requirements](#requirements)**
    - **[Running Locally](#running-locally)**
- **[Build](#build)**
- **[CI](#ci)**
- **[IaC](#iac)**

### About 

Implementation of an API that returns the hostname of the underlying node and a sender application that:

- Lists the node hostnames
- Counts the number of requests handled by each node
- Counts the number of available nodes

#### **Structure**

```shell
├── cmd # Entrypoints 
│   ├── sender # Sender application
│   └── server # Server application
├── iac      # Infrastructure as Code implementation
├── internal # Internal sharable code
│   ├── api
│   │   └── v1 # API v1
|   |   └── v2 # API v2
│   ├── config
│   └── observability
```

#### **Requirements**

Go, GNU Make, Docker, Terraform, Git

#### **Running Locally**

First start the server:

```shell
make run-server
##################
{"time":"2025-06-04T07:46:20.198708793Z","level":"INFO","msg":"starting...","application":"alcatraz-rest","version":"local"}
```

Then run the sender:

```shell
make run-sender
##################
=== Load Balancer Test Results ===
Total Requests: 100
Successful Requests: 100
Failed Requests: 0
Available Nodes: 1
Average Response Time: 0 ms

=== Node Hostnames ===
1. dev

=== Requests Per Node ===
dev                 :  100 requests (100.0%)

=== Response Time Statistics (ms) ===
dev                 : avg=  0ms, min=  0ms, max=  1ms, count=100
```


### Build

### CI

### IaC