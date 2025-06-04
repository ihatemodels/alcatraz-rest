## Contributing

### Requirements

- Go 1.24
- GNU Make
- Docker (optional)
- Terraform (optional)

### Environment

This guide is for osx and Linux. If you are on Windows, you can use the Windows Subsystem for Linux (WSL). 

### Run the applications locally

```bash
make run-server
make run-sender
```

- test it

```bash
curl http://localhost:9000/api/ping
```

### Before submitting a PR

- Make sure to sync the vendor if the dependencies have changed.

```bash
make tidy
```

- Make sure to run the tests and lints.

```bash
make check
```