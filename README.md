<div align="center">
  
  ![Memphis light logo](https://github.com/memphisdev/memphis/blob/master/logo-white.png?raw=true#gh-dark-mode-only)
  
</div>

<div align="center">
  
  ![Memphis light logo](https://github.com/memphisdev/memphis/blob/master/logo-black.png?raw=true#gh-light-mode-only)
  
</div>

<div align="center">
<h4>Simple as RabbitMQ, robust as Apache Kafka, and perfect for busy developers.</h4>
<img width="750" alt="Memphis UI" src="https://user-images.githubusercontent.com/70286779/204081372-186aae7b-a387-4253-83d1-b07dff69b3d0.png"><br>

  
  <a href="https://landscape.cncf.io/?selected=memphis"><img width="200" alt="CNCF Silver Member" src="https://github.com/cncf/artwork/raw/master/other/cncf-member/silver/white/cncf-member-silver-white.svg#gh-dark-mode-only"></a>
  
</div>

<div align="center">
  
  <img width="200" alt="CNCF Silver Member" src="https://github.com/cncf/artwork/raw/master/other/cncf-member/silver/color/cncf-member-silver-color.svg#gh-light-mode-only">
  
</div>
 
 <p align="center">
  <a href="https://sandbox.memphis.dev/" target="_blank">Sandbox</a> - <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">Twitter</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p>

<p align="center">
<a href="https://discord.gg/WZpysvAeTf"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a>
<a href="https://github.com/memphisdev/memphis/issues?q=is%3Aissue+is%3Aclosed"><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis?color=6557ff"></a> 
  <img src="https://img.shields.io/npm/dw/memphis-dev?color=ffc633&label=installations">
<a href="https://github.com/memphisdev/memphis/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> 
<a href="https://docs.memphis.dev/memphis/release-notes/releases/v0.4.2-beta"><img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis?color=61dfc6"></a>
<img src="https://img.shields.io/github/last-commit/memphisdev/memphis?color=61dfc6&label=last%20commit">
</p>

**[Memphis](https://memphis.dev)** is a next-generation alternative to traditional message brokers.<br><br>
A simple, robust, and durable cloud-native message broker wrapped with<br>
an entire ecosystem that enables cost-effective, fast, and reliable development of modern queue-based use cases.<br><br>
Memphis enables the building of modern queue-based applications that require<br>
large volumes of streamed and enriched data, modern protocols, zero ops, rapid development,<br>
extreme cost reduction, and a significantly lower amount of dev time for data-oriented developers and data engineers.

## 🫣 A world without Memphis
When your application requires a message broker or a queue,<br>
Implementing one will require you to -
- Build a dead-letter queue, create observability, and a retry mechanism
- Build a scalable environment
- Create client wrappers
- Tag events to achieve multi-tenancy
- Enforce schemas and handle transformations
- Handle back pressure. Client or queue side
- Configure monitoring and real-time alerts
- Create a cloud-agnostic implementation
- Create config alignment between production to a dev environment
- Spent weeks and months learning the internals through archival documentation, ebooks, and courses
- Onboard your developers<br>
And the list continues...<br>

**Or, you can just use [Memphis](https://memphis.dev)** and focus your resources on tasks that matter 😎
<br>

## ✨ Features [v0.4.5](https://docs.memphis.dev/memphis/release-notes/releases/v0.4.5-beta)

[**Roadmap**](https://github.com/orgs/memphisdev/projects/2/views/1)

![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Production-ready message broker in under 3 minutes<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Easy-to-use UI, CLI, and SDKs<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Data-level observability<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Dead-Letter Queue with automatic message retransmit<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Schemaverse - Embedded schema management for produced data (Protobuf/JSON/GraphQL/Avro)<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Storage tiering<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) SDKs: Node.JS, Go, Python, Typescript, NestJS, REST, .NET, Kotlin<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Kubernetes-native<br>
![20](https://user-images.githubusercontent.com/70286779/220196529-abb958d2-5c58-4c33-b5e0-40f5446515ad.png) Community driven<br>

## 🚀 Getting Started
Helm for Kubernetes☸
```shell
helm repo add memphis https://k8s.memphis.dev/charts/ --force-update && \
helm install my-memphis memphis/memphis --create-namespace --namespace memphis
```
Docker🐳 Compose
```shell
curl -s https://memphisdev.github.io/memphis-docker/docker-compose.yml -o docker-compose.yml && \
docker compose -f docker-compose.yml -p memphis up
```

<p align="center">
<a href="https://youtu.be/-5YmxYRQsdw"><img align="center" alt="connect your first app" src="https://img.youtube.com/vi/-5YmxYRQsdw/0.jpg"></a>
</p>

[Tutorials](https://docs.memphis.dev/memphis/getting-started/tutorials)<br>
[Sandbox](https://sandbox.memphis.dev)<br>
[Installation videos](https://www.youtube.com/playlist?list=PL_7iYjqhtXpWpZT2U0zDYo2eGOoGmg2mm)<br><br>

## High-Level Architecture

<a href="https://docs.memphis.dev/memphis/memphis/architecture">
<p align="center">
<img height="500" alt="memphis.dev Architecture" src="https://user-images.githubusercontent.com/70286779/212160911-aeae5b55-c0cf-46bc-a1aa-fcfdb6d10a2a.png">
</p>
</a>

## Local access
### Via Kubernetes
```shell
To access Memphis using UI/CLI/SDK from localhost, run the below commands:

  - kubectl port-forward service/memphis 6666:6666 9000:9000 7770:7770 --namespace memphis > /dev/null &

For interacting with the broker via HTTP:

  - kubectl port-forward service/memphis-rest-gateway 4444:4444 --namespace memphis > /dev/null &

Dashboard/CLI: http://localhost:9000
Broker: localhost:6666 (Client Connections)
REST gateway: localhost:4444 (Data + Mgmt)
```

**For Production Environments**
Please expose the UI, Cluster, and Control-plane via k8s ingress / load balancer / nodeport

### Via Docker
```shell
Dashboard/CLI: http://localhost:9000
Broker: localhost:6666
```

## 👉 Use-cases
- Async task management
- Real-time streaming pipelines
- Data ingestion
- Cloud Messaging
  - Services (microservices, service mesh)
  - Event/Data Streaming (observability, analytics, ML/AI)
- Queuing
- N:N communication patterns
- Ingest Grafana Loki logs at scale

## Support 🙋‍♂️🤝

### Ask a question ❓ about Memphis{dev} or something related to us:

We welcome you to our discord server with your questions, doubts and feedback.

<a href="https://discord.gg/WZpysvAeTf"><img src="https://amplication.com/images/discord_banner_purple.svg"/></a>

### Create a bug 🐞 report

If you see an error message or run into an issue, please [create bug report](https://github.com/memphisdev/memphis/issues/new?assignees=&labels=type%3A%20bug&template=bug_report.md&title=). This effort is valued and it will help all Memphis{dev} users.


### Submit a feature 💡 request 

If you have an idea, or you think that we're missing a capability that would make development easier and more robust, please [Submit feature request](https://github.com/memphisdev/memphis/issues/new?assignees=&labels=type%3A%20feature%20request).

If an issue❗with similar feature request already exists, don't forget to leave a "+1".
If you add some more information such as your thoughts and vision about the feature, your comments will be embraced warmly :)

## Contributing

Memphis{dev} is an open-source project.<br>
We are committed to a fully transparent development process and appreciate highly any contributions.<br>
Whether you are helping us fix bugs, proposing new features, improving our documentation or spreading the word - we would love to have you as part of the Memphis{dev} community.

Please refer to our [Contribution Guidelines](./CONTRIBUTING.md) and [Code of Conduct](./CODE_OF_CONDUCT.md).

## Contributors ✨

Thanks goes to these wonderful people ❤:<br><br>
 <a href = "https://github.com/memphisdev/memphis/graphs/contributors">
   <img src = "https://contrib.rocks/image?repo=memphisdev/memphis"/>
 </a>

## License 📃
Memphis is open-sourced and operates under the "Memphis Business Source License 1.0" license
Built out of Apache 2.0, the main difference between the licenses is:
"You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service. A “Service” is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services."
Please check out [License](./LICENSE) to read the full text.
