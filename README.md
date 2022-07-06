<a href="https://discord.gg/WZpysvAeTf"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a> <a href=""><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis-broker?color=6557ff"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/LICENSE"><img src="https://img.shields.io/github/license/memphisdev/memphis-broker?color=ffc633" alt="License"></a> <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis-broker?color=61dfc6"> <img src="https://img.shields.io/github/last-commit/memphisdev/memphis-broker?color=61dfc6&label=last%20commit">
</p>

**[Memphis{dev}](https://memphis.dev)** is a message broker for developers made out of devs' struggles develop around message brokers.<br>Enables devs to achieve all other message brokers' benefits in a fraction of the time.<br>
Focusing on automatic optimization, schema management, inline processing,  and troubleshooting abilities. All under the same hood.
Utilizing NATS core.

## ⭐️ Why
Working with data streaming is HARD.<br>
As a developer, you need to build a dedicated pipeline per data source,<br>change the schema, individual analysis, enrich the data with other sources, it constantly crashes, it requires adaptation to different rate limits, constantly change APIs, and scale for better performance 🥵 .<br>
**It takes time and resources that you don't have.**<br><br>
Message broker is the answer. In short - It's an event-store.<br>
Message broker acts as the middleman and supports streaming architecture,<br>but then you encounter Apache Kafka and its documentation and run back to the monolith and batch jobs.<br>
Give memphis{dev} a spin before.

## 👉 Use-cases
- Async task management
- Real-time streaming pipelines
- Data ingestion
- Cloud Messaging
  - Services (microservices, service mesh)
  - Event/Data Streaming (observability, analytics, ML/AI)
- Queuing
- N:N communication patterns

## ✨ Features

**v0.2.2**

- 🚀 Fully optimized message broker in under 3 minutes
- 💻 Easy-to-use UI, CLI, and SDKs
- 📺 Data-level observability
- 🐳☸Runs on your Docker or Kubernetes
- 👨‍💻 Community driven

**Coming soon v0.2.5-1.0.0**
- Embedded schema registry using dbt
- Message Journey - Real-time messages tracing
- More SDKs (GoLang, Python, Kafka compatible)
- Inline processing
- Ready-to-use connectors and analysis functions

## 📸 Screenshots
Dashboard             |  Station overview|  CLI
:-------------------------:|:-------------------------:|:-------------------------:
<img src="https://user-images.githubusercontent.com/70286779/175805888-f08e2078-79e1-43f1-a841-1d7115bf15a8.png" alt="drawing" width="300"/>|<img src="https://user-images.githubusercontent.com/70286779/175805897-349dde51-427f-4c9b-95cd-12876a846f1a.png" alt="drawing" width="300"/>|<img src="https://user-images.githubusercontent.com/70286779/175806007-9a37e130-3e5a-4606-bdda-a71a89efae7f.png" alt="drawing" width="300"/>



## 🚀 Getting Started
[Installation videos](https://www.youtube.com/playlist?list=PL_7iYjqhtXpWpZT2U0zDYo2eGOoGmg2mm)<br><br>
Helm for Kubernetes
```shell
helm repo add memphis https://k8s.memphis.dev/charts/ && \
helm install my-memphis memphis/memphis --create-namespace --namespace memphis
