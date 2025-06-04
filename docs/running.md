## Running the applications locally

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