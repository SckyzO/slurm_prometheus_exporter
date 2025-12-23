# Slurm Exporter 🚀

A Prometheus exporter for Slurm metrics, because monitoring your HPC cluster should be as smooth as your jobs running on it!

## Features ✨

- ✅ Export Slurm OpenMetrics (version 25.11+)
- ✅ Support for multiple endpoints (jobs, nodes, partitions, scheduler)
- ✅ Basic Authentication and SSL/TLS support
- ✅ Customizable global labels for all metrics
- ✅ Easy configuration with YAML
- ✅ Built with Clean Architecture principles
- ✅ Comprehensive error handling and logging

## Prerequisites 📋

- Go 1.21 or higher
- Slurm 25.11 or higher with OpenMetrics enabled
- Access to Slurm REST API

## Installation 🔧

### From Source

```bash
git clone https://github.com/sckyzo/slurm_prometheus_exporter.git
cd slurm_prometheus_exporter
make build
```

The binary will be available at `bin/slurm_exporter`.

### Using Go Install

```bash
go install github.com/sckyzo/slurm_prometheus_exporter/cmd/slurm_exporter@latest
```

## Configuration ⚙️

Create a `config.yaml` file with your settings:

```yaml
# Configuration for the connection to Slurm API
slurm:
  url: "http://localhost:6817"
  timeout: "10s"
  tls_insecure_skip_verify: false  # Set to true for self-signed certificates (insecure)

# HTTP server configuration
server:
  port: 8080
  basic_auth:
    enabled: true
    username: "admin"
    password: "password"
  ssl:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"

# Endpoints to expose
endpoints:
  - name: "jobs"
    path: "/metrics/jobs"
    enabled: true
  - name: "nodes"
    path: "/metrics/nodes"
    enabled: true
  - name: "partitions"
    path: "/metrics/partitions"
    enabled: true
  - name: "jobs-users-accts"
    path: "/metrics/jobs-users-accts"
    enabled: true
  - name: "scheduler"
    path: "/metrics/scheduler"
    enabled: true

# Global custom labels
labels:
  cluster: "cluster01"
  env: "prod"
  region: "eu-west-1"

# Logging configuration
logging:
  level: "info"
  output: "stdout"
```

An example configuration file is available in [`configs/config.yaml`](configs/config.yaml).

## Usage 🚀

Run the exporter with your configuration file:

```bash
bin/slurm_exporter --config.file=config.yaml
```

### Command-line Options

```
Usage: slurm_exporter [<flags>]

Flags:
  --help                        Show help
  -v, --version                 Show version information
  --config.file="config.yaml"   Path to configuration file
  --web.listen-address=":8080"  Address to listen on for web interface and telemetry
  --log.level="info"            Log level (debug, info, warn, error)
  --log.format="text"           Log format (text, json)
```

### Examples

```bash
# Use a specific config file
bin/slurm_exporter --config.file=/etc/slurm_exporter/config.yaml

# Override listen address
bin/slurm_exporter --config.file=config.yaml --web.listen-address=":9100"

# Enable debug logging
bin/slurm_exporter --config.file=config.yaml --log.level=debug

# Use JSON logging format
bin/slurm_exporter --config.file=config.yaml --log.format=json
```

### Endpoints

The exporter exposes the following endpoints:

- `/` - Landing page with exporter information
- `/metrics` - Aggregated Prometheus metrics from all enabled Slurm endpoints

## Prometheus Configuration 📊

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'slurm'
    static_configs:
      - targets: ['localhost:8080']
    basic_auth:
      username: 'admin'
      password: 'password'
    scrape_interval: 30s
```

## Development 💻

### Project Structure

```
slurm_exporter/
├── cmd/
│   └── slurm_exporter/      # Main application entry point
├── internal/
│   ├── config/              # Configuration handling
│   ├── collector/           # Slurm metrics collection
│   ├── server/              # HTTP server
│   └── metrics/             # Prometheus metrics registry
├── pkg/                     # Public packages
├── configs/                 # Example configurations
├── test_data/               # Test data for development
└── .github/workflows/       # CI/CD workflows
```

### Building

```bash
make build        # Build the binary
make test         # Run tests
make lint         # Run linter
make clean        # Clean build artifacts
```

### Testing

Test data is available in the [`test_data/`](test_data/) directory with sample OpenMetrics output from Slurm.

## Contributing 🤝

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments 🙏

- The Prometheus team for their excellent client library
- The Slurm team for providing OpenMetrics support

## Support 💬

If you encounter any issues or have questions, please [open an issue](https://github.com/sckyzo/slurm_prometheus_exporter/issues) on GitHub.
