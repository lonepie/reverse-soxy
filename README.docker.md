# Docker Setup for Reverse-SOXY

This document provides instructions for running Reverse-SOXY in Docker containers.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/) (optional, for running multi-container setups)

## Building the Docker Image

To build the Docker image:

```bash
docker build -t reverse-soxy .
```

## Running with Docker

### Proxy Mode

```bash
docker run -p 1080:1080 -p 9000:9000 reverse-soxy --proxy-listen-addr 0.0.0.0:1080 --tunnel-listen-port 9000 --secret yourSecretHere
```

### Agent Mode

```bash
docker run reverse-soxy --tunnel-addr proxy.host:9000 --secret yourSecretHere
```

### Relay Mode

```bash
docker run -p 9000:9000 reverse-soxy --mode relay --relay-listen-port 9000 --secret yourSecretHere
```

### Proxy via Relay Mode

```bash
docker run -p 1080:1080 reverse-soxy --mode proxy --register --relay-addr relay.host:9000 --proxy-listen-addr 0.0.0.0:1080 --secret yourSecretHere
```

### Agent via Relay Mode

```bash
docker run reverse-soxy --mode agent --relay-addr relay.host:9000 --secret yourSecretHere
```

## Running with Docker Compose

The included `docker-compose.yml` file provides configurations for all modes of operation.

### Setting a Secure Secret

Before running, set a secure secret:

```bash
export SECRET=yourSecretHere
```

### Running Different Setups

#### Direct Proxy and Agent

```bash
# Start the proxy
docker-compose up proxy

# In another terminal, start the agent
docker-compose up agent
```

#### Relay Server with Proxy and Agent

```bash
# Start the relay server
docker-compose up relay

# In another terminal, start the proxy via relay
docker-compose up proxy-via-relay

# In another terminal, start the agent via relay
docker-compose up agent-via-relay
```

## Configuration

### Environment Variables

- `SECRET`: The shared secret for encryption/authentication
- `PROXY_HOST`: The hostname of the proxy (for agent mode)
- `RELAY_HOST`: The hostname of the relay server (for proxy-via-relay and agent-via-relay modes)

### Custom Configuration File

You can mount a custom YAML configuration file:

```bash
docker run -v /path/to/your/config.yml:/app/config/config.yml reverse-soxy --config /app/config/config.yml --secret yourSecretHere
```

## Security Considerations

- Always use a strong, unique secret for each deployment
- Consider using Docker secrets or environment variables for the secret in production
- The default configuration exposes ports to all interfaces (0.0.0.0) within the container, so be careful with port mappings
- In production, consider using a non-root user in the container

## Troubleshooting

### Debugging

Add the `--debug` flag to enable debug logging:

```bash
docker run reverse-soxy --secret yourSecretHere --debug
```

### Checking Container Logs

```bash
docker logs <container_id>
```

### Inspecting a Running Container

```bash
docker exec -it <container_id> /bin/sh
```
