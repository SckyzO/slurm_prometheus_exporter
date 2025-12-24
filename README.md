# Slurm Exporter 🚀

A Prometheus exporter for Slurm metrics, because monitoring your HPC cluster should be as smooth as your jobs running on it! This exporter unifies all Slurm metrics endpoints (see [Slurm Metrics Documentation](https://slurm.schedmd.com/metrics.html)) into a single Prometheus-compatible endpoint.

## Features ✨

- ✅ Export Native OpenMetrics from Slurm (version 25.11+)
- ✅ Support for multiple endpoints (jobs, jobs-users-accts, nodes, partitions, scheduler)
- ✅ Basic Authentication and SSL/TLS support
- ✅ Customizable global labels for all metrics
- ✅ Easy configuration with YAML
- ✅ Built with Clean Architecture principles
- ✅ Comprehensive error handling and logging

## Prerequisites 📋

- Go 1.23 or higher
- Slurm 25.11 or higher with OpenMetrics enabled
- Access to Slurm Metrics (https://slurm.schedmd.com/metrics.html)

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

## Metrics

The exporter exposes the following metrics:

| Metric | Description |
|---|---|
| `slurm_node_cpus` | Total number of cpus in the node |
| `slurm_node_cpus_alloc` | Allocated cpus in the node |
| `slurm_node_cpus_effective` | CPUs allocatable to jobs not reserved for system usage |
| `slurm_node_cpus_idle` | Idle cpus in the node |
| `slurm_node_memory_alloc_bytes` | Bytes allocated to jobs in the node |
| `slurm_node_memory_effective_bytes` | Memory allocatable to jobs not reserved for system usage |
| `slurm_node_memory_free_bytes` | Free memory in bytes of the node |
| `slurm_node_memory_bytes` | Total memory in bytes of the node |
| `slurm_nodes` | Total number of nodes |
| `slurm_nodes_alloc` | Number of nodes in Allocated state |
| `slurm_nodes_blocked` | Number of nodes in Blocked state |
| `slurm_nodes_completing` | Number of nodes with Completing flag |
| `slurm_nodes_cloud` | Number of Cloud nodes |
| `slurm_nodes_down` | Number of nodes in Down state |
| `slurm_nodes_drain` | Number of nodes with Drain flag |
| `slurm_nodes_drained` | Number of drained nodes |
| `slurm_nodes_draining` | Number of nodes in draining condition (Drain state with active jobs) |
| `slurm_nodes_dyn_future` | Number of future dynamic nodes |
| `slurm_nodes_dyn_normal` | Number of dynamic nodes |
| `slurm_nodes_external` | Number of external nodes |
| `slurm_nodes_fail` | Number of nodes with Fail flag |
| `slurm_nodes_future` | Number of nodes in Future state |
| `slurm_nodes_idle` | Number of nodes in Idle state |
| `slurm_nodes_invalid_reg` | Number of nodes with Invalid Registration flag |
| `slurm_nodes_maint` | Number of nodes with Maintenance flag |
| `slurm_nodes_mixed` | Number of nodes in Mixed state |
| `slurm_nodes_noresp` | Number of nodes with Not Responding flag |
| `slurm_nodes_planned` | Number of nodes with Planned flag |
| `slurm_nodes_power_down` | Number of nodes marked to be powered down |
| `slurm_nodes_power_up` | Number of nodes marked to be powered up |
| `slurm_nodes_powered_down` | Number of nodes powered down |
| `slurm_nodes_powering_up` | Number of nodes powering up |
| `slurm_nodes_reboot_issued` | Number of nodes with Reboot Issued flag |
| `slurm_nodes_reboot_req` | Number of nodes with Reboot Requested flag |
| `slurm_nodes_resv` | Number of nodes with Reserved flag |
| `slurm_nodes_unknown` | Number of nodes in Unknown state |
| `slurm_partitions` | Total number of partitions |
| `slurm_partition_jobs` | Number of jobs in this partition |
| `slurm_partition_jobs_bootfail` | Number of jobs in BootFail state |
| `slurm_partition_jobs_cancelled` | Number of jobs in Cancelled state |
| `slurm_partition_jobs_completed` | Number of jobs in Completed state |
| `slurm_partition_jobs_completing` | Number of jobs in Completing state |
| `slurm_partition_jobs_configuring` | Number of jobs in Configuring state |
| `slurm_partition_jobs_cpus_alloc` | Total number of Cpus allocated by jobs |
| `slurm_partition_jobs_deadline` | Number of jobs in Deadline state |
| `slurm_partition_jobs_expediting` | Number of jobs in Expediting state |
| `slurm_partition_jobs_failed` | Number of jobs in Failed state |
| `slurm_partition_jobs_fed_requeued` | Number of jobs requeued in a federation |
| `slurm_partition_jobs_finished` | Number of jobs in Finished |
| `slurm_partition_jobs_hold` | Number of jobs in Hold state |
| `slurm_partition_jobs_max_job_nodes` | Max of the max_nodes required of all pending jobs in that partition |
| `slurm_partition_jobs_max_job_nodes_nohold` | Max of the max_nodes required of all pending jobs in that partition excluding Held jobs |
| `slurm_partition_jobs_memory_alloc` | Total memory bytes allocated by jobs |
| `slurm_partition_jobs_min_job_nodes` | Max of the min_nodes required of all pending jobs in that partition |
| `slurm_partition_jobs_min_job_nodes_nohold` | Max of the min_nodes required of all pending jobs in that partition excluding Held jobs |
| `slurm_partition_jobs_node_failed` | Number of jobs in Node Failed state |
| `slurm_partition_jobs_outofmemory` | Number of jobs in Out of Memory state |
| `slurm_partition_jobs_pending` | Number of jobs in Pending state |
| `slurm_partition_jobs_powerup_node` | Number of jobs in PowerUp Node state |
| `slurm_partition_jobs_preempted` | Number of jobs in Preempted state |
| `slurm_partition_jobs_requeued` | Number of jobs in Requeued state |
| `slurm_partition_jobs_resizing` | Number of jobs in Resizing state |
| `slurm_partition_jobs_revoked` | Number of revoked jobs |
| `slurm_partition_jobs_running` | Number of jobs in Running state |
| `slurm_partition_jobs_signaling` | Number of jobs in Signaling state |
| `slurm_partition_jobs_stageout` | Number of jobs in StageOut state |
| `slurm_partition_jobs_started` | Number of jobs started |
| `slurm_partition_jobs_suspended` | Number of jobs in Suspended state |
| `slurm_partition_jobs_timeout` | Number of jobs in Timeout state |
| `slurm_partition_jobs_wait_part_node_limit` | Jobs wait partition node limit |
| `slurm_partition_nodes_alloc` | Nodes allocated |
| `slurm_partition_nodes_blocked` | Nodes blocked |
| `slurm_partition_nodes_cg` | Nodes in completing state |
| `slurm_partition_nodes_cloud` | Cloud nodes |
| `slurm_partition_nodes_cpus_efctv` | Number of effective CPUs on all nodes, excludes CoreSpec |
| `slurm_partition_nodes_cpus_idle` | Number of idle CPUs on all nodes |
| `slurm_partition_nodes_cpus_alloc` | Number of allocated cpus |
| `slurm_partition_nodes_down` | Nodes in Down state |
| `slurm_partition_nodes_drain` | Nodes in Drain state |
| `slurm_partition_nodes_drained` | Nodes in Drained state |
| `slurm_partition_nodes_draining` | Number of nodes in draining condition (Drain state with active jobs) |
| `slurm_partition_nodes_dyn_future` | Dynamic nodes in Future state |
| `slurm_partition_nodes_dyn_normal` | Dynamic nodes |
| `slurm_partition_nodes_external` | External nodes |
| `slurm_partition_nodes_fail` | Nodes in Fail state |
| `slurm_partition_nodes_future` | Nodes in Future state |
| `slurm_partition_nodes_idle` | Nodes in Idle state |
| `slurm_partition_nodes_invalid_reg` | Number of nodes with Invalid Registration flag |
| `slurm_partition_nodes_maint` | Nodes in maintenance state |
| `slurm_partition_nodes_mem_alloc` | Amount of allocated memory of all nodes |
| `slurm_partition_nodes_mem_avail` | Amount of available memory of all nodes |
| `slurm_partition_nodes_mem_free` | Amount of free memory in all nodes |
| `slurm_partition_nodes_mem_tot` | Total amount of memory of all nodes |
| `slurm_partition_nodes_mixed` | Nodes in Mixed state |
| `slurm_partition_nodes_no_resp` | Nodes in Not Responding state |
| `slurm_partition_nodes_planned` | Nodes in Planned state |
| `slurm_partition_nodes_power_down` | Nodes marked to Power Down |
| `slurm_partition_nodes_power_up` | Nodes marked to Power Up |
| `slurm_partition_nodes_powered_down` | Powered down nodes |
| `slurm_partition_nodes_powering_down` | Powering down nodes |
| `slurm_partition_nodes_powering_up` | Powering up nodes |
| `slurm_partition_nodes_reboot_issued` | Nodes which initiated reboot |
| `slurm_partition_nodes_reboot_requested` | Nodes with Reboot Requested flag |
| `slurm_partition_nodes_resv` | Nodes with Reserved flag |
| `slurm_partition_nodes_unknown` | Nodes in Unknown state |
| `slurm_partition_cpus` | Partition total cpus |
| `slurm_partition_nodes` | Partition total nodes |
| `slurm_agent_cnt` | Number of agent threads |
| `slurm_agent_queue_size` | Outgoing RPC retry queue length |
| `slurm_agent_thread_cnt` | Total active agent-created threads |
| `slurm_bf_depth_mean` | Mean backfill cycle depth |
| `slurm_bf_mean_cycle` | Mean backfill cycle time |
| `slurm_bf_mean_table_sz` | Mean backfill table size |
| `slurm_bf_queue_len_mean` | Mean backfill queue length |
| `slurm_bf_try_depth_mean` | Mean depth attempts in backfill |
| `slurm_backfilled_het_jobs` | Heterogeneous components backfilled |
| `slurm_backfilled_jobs` | Total backfilled jobs since reset |
| `slurm_bf_active` | Backfill scheduler active jobs |
| `slurm_bf_cycle_cnt` | Backfill cycle count |
| `slurm_bf_cycle_last` | Last backfill cycle time |
| `slurm_bf_cycle_max` | Max backfill cycle time |
| `slurm_bf_cycle_tot` | Sum of backfill cycle times |
| `slurm_bf_depth_tot` | Sum of backfill job depths |
| `slurm_bf_depth_try_tot` | Sum of backfill depth attempts |
| `slurm_bf_last_depth` | Last backfill depth |
| `slurm_bf_last_depth_try` | Last backfill depth attempts |
| `slurm_bf_queue_len` | Backfill queue length |
| `slurm_bf_queue_len_tot` | Sum of backfill queue lengths |
| `slurm_bf_table_size` | Backfill table size |
| `slurm_bf_table_size_tot` | Sum of backfill table sizes |
| `slurm_bf_when_last_cycle` | Timestamp of last backfill cycle |
| `slurm_sdiag_jobs_canceled` | Jobs canceled since reset |
| `slurm_sdiag_jobs_completed` | Jobs completed since reset |
| `slurm_sdiag_jobs_failed` | Jobs failed since reset |
| `slurm_sdiag_jobs_pending` | Jobs pending at timestamp |
| `slurm_sdiag_jobs_running` | Jobs running at timestamp |
| `slurm_sdiag_jobs_started` | Jobs started since reset |
| `slurm_sdiag_jobs_submitted` | Jobs submitted since reset |
| `slurm_sdiag_job_states_ts` | Job states timestamp |
| `slurm_last_backfilled_jobs` | Backfilled jobs since last cycle |
| `slurm_sdiag_latency` | Measurement latency |
| `slurm_schedule_cycle_cnt` | Scheduling cycle count |
| `slurm_schedule_cycle_depth` | Processed jobs depth total |
| `slurm_schedule_cycle_last` | Last scheduling cycle time |
| `slurm_schedule_cycle_max` | Max scheduling cycle time |
| `slurm_schedule_cycle_tot` | Sum of scheduling cycle times |
| `slurm_schedule_queue_len` | Jobs pending queue length |
| `slurm_sched_exit_end` | End of job queue |
| `slurm_sched_exit_max_depth` | Hit default_queue_depth |
| `slurm_sched_exit_max_job_start` | Hit sched_max_job_start |
| `slurm_sched_exit_lic` | Blocked on licenses |
| `slurm_sched_exit_rpc_cnt` | Hit max_rpc_cnt |
| `slurm_sched_exit_timeout` | Timeout (max_sched_time) |
| `slurm_bf_exit_end` | End of job queue |
| `slurm_bf_exit_max_job_start` | Hit bf_max_job_start |
| `slurm_bf_exit_max_job_test` | Hit bf_max_job_test |
| `slurm_bf_exit_state_changed` | System state changed |
| `slurm_bf_exit_table_limit` | Hit table size limit (bf_node_space_size) |
| `slurm_bf_exit_timeout` | Timeout (bf_max_time) |
| `slurm_sched_mean_cycle` | Mean scheduling cycle time |
| `slurm_sched_mean_depth_cycle` | Mean depth of scheduling cycles |
| `slurm_server_thread_cnt` | Active slurmctld threads count |
| `slurm_slurmdbd_queue_size` | Queued messages to SlurmDBD |
| `slurm_last_proc_req_start` | Timestamp of last process request start |
| `slurm_sched_stats_timestamp` | Statistics snapshot timestamp |
| `slurm_jobs_bootfail` | Number of jobs in BootFail state |
| `slurm_jobs_cancelled` | Number of jobs in Cancelled state |
| `slurm_jobs_completed` | Number of jobs in Completed state |
| `slurm_jobs_completing` | Number of jobs in Completing state |
| `slurm_jobs_configuring` | Number of jobs in Configuring state |
| `slurm_jobs_cpus_alloc` | Total number of Cpus allocated by jobs |
| `slurm_jobs_deadline` | Number of jobs in Deadline state |
| `slurm_jobs_expediting` | Number of jobs in Expediting state |
| `slurm_jobs_failed` | Number of jobs in Failed state |
| `slurm_jobs_fed_requeued` | Number of jobs requeued in a federation |
| `slurm_jobs_finished` | Number of finished jobs |
| `slurm_jobs_hold` | Number of jobs in Hold state |
| `slurm_jobs` | Total number of jobs |
| `slurm_jobs_memory_alloc` | Total memory bytes allocated by jobs |
| `slurm_jobs_node_failed` | Number of jobs in Node Failed state |
| `slurm_jobs_nodes_alloc` | Total number of nodes allocated by jobs |
| `slurm_jobs_outofmemory` | Number of jobs in Out of Memory state |
| `slurm_jobs_pending` | Number of jobs in Pending state |
| `slurm_jobs_powerup_node` | Number of jobs in PowerUp Node state |
| `slurm_jobs_preempted` | Number of jobs in Preempted state |
| `slurm_jobs_requeued` | Number of jobs in Requeued state |
| `slurm_jobs_resizing` | Number of jobs in Resizing state |
| `slurm_jobs_revoked` | Number of jobs in Rvoked state |
| `slurm_jobs_running` | Number of jobs in Running state |
| `slurm_jobs_signaling` | Number of jobs being signaled |
| `slurm_jobs_stageout` | Number of jobs in StageOut state |
| `slurm_jobs_started` | Number of started jobs |
| `slurm_jobs_suspended` | Number of jobs in Suspended state |
| `slurm_jobs_timeout` | Number of jobs in Timeout state |
| `slurm_exporter_build_info` | A metric with a constant '1' value labeled by version, git_commit, and build_time |
| `slurm_exporter_http_request_duration_seconds` | Duration of HTTP requests |
| `slurm_exporter_http_requests_total` | Total number of HTTP requests received by the exporter |
| `slurm_exporter_scrape_success` | Whether the last scrape was successful (1 = success, 0 = failure) |