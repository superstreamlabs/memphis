<p align="center">
  <a href="https://memphis.dev" target="_blank">
    <img alt="memphis.dev-logo" height="70" alt="memphis.dev Logo" src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/graphics+for+github/color+logo.svg">
  </a>
 </p>
 <p align="center">
  <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">Twitter</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p>

<p align="center">
  <a href="https://discord.gg/WZpysvAeTf"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a> <a href=""><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis-broker?color=61dfc6"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor-v1.0-ff69b4.svg?color=ffc633" alt="Contributor"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/LICENSE"><img src="https://img.shields.io/github/license/memphisdev/memphis-broker?color=6557ff" alt="License"></a> <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis-broker?color=61dfc6"> <img src="https://img.shields.io/github/last-commit/memphisdev/memphis-broker?color=ffc633&label=last%20commit">
</p>

### Probably the easiest message broker in the world.

**[Memphis{dev}](https://memphis.dev)** is a modern replacement for Apache Kafka.<br>A message broker for developers made out of devs' struggles with using message brokers,<br>building complex data/event-driven apps, and troubleshooting them.<br><br>Allowing developers to achieve all other message brokers' benefits in a fraction of the time.<br>

# Features
**Current**
- Fully optimized message broker in under 3 minutes
- Easy-to-use UI, CLI, and SDKs
- Data-level observability
- Runs on your Docker or Kubernetes

**Coming**
- Embedded schema registry using dbt
- Message Journey
- More SDKs
- Inline processing
- Ready-to-use connectors and analysis functions

# Getting Started
Kubernetes
```shell
helm repo add memphis https://k8s.memphis.dev/charts/ && \
helm install my-memphis memphis/memphis --create-namespace --namespace memphis
```
Docker Compose
```shell
curl -s https://memphisdev.github.io/memphis-docker/docker-compose.yml -o docker-compose.yml && \
docker compose -f docker-compose.yml -p memphis up
```
[Installation Videos](https://www.youtube.com/playlist?list=PL_7iYjqhtXpWpZT2U0zDYo2eGOoGmg2mm)
