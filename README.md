<div align="center">
  
  ![Memphis light logo](https://github.com/memphisdev/memphis-broker/blob/master/logo-white.png?raw=true#gh-dark-mode-only)
  
</div>

<div align="center">
  
  ![Memphis light logo](https://github.com/memphisdev/memphis-broker/blob/master/logo-black.png?raw=true#gh-light-mode-only)
  
</div>

<div align="center">
<h1>Real-Time Data Processing Platform</h1>

<img width="750" alt="Memphis UI" src="https://user-images.githubusercontent.com/70286779/182241744-2016dc1a-c758-48ba-8666-40b883242ea9.png">


<a target="_blank" href="https://twitter.com/intent/tweet?text=Probably+The+Easiest+Message+Broker+In+The+World%21+%0D%0Ahttps%3A%2F%2Fgithub.com%2Fmemphisdev%2Fmemphis-broker+%0D%0A%0D%0A%23MemphisDev"><img src="https://user-images.githubusercontent.com/70286779/174467733-e7656c1e-cfeb-4877-a5f3-1bd4fccc8cf1.png" width="60"></a> 
</div>
 
 <p align="center">
  <a href="https://demo.memphis.dev/">Playground</a> - <a href="https://sandbox.memphis.dev/" target="_blank">Sandbox</a> - <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">Twitter</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p>

<p align="center">
<a href="https://discord.gg/WZpysvAeTf"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a> <a href=""><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis-broker?color=6557ff"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"></a> <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis-broker?color=61dfc6"> <img src="https://img.shields.io/github/last-commit/memphisdev/memphis-broker?color=61dfc6&label=last%20commit">
</p>

**[Memphis{dev}](https://memphis.dev)** is an open-source real-time data processing platform<br>
that provides end-to-end support for in-app streaming use cases using Memphis distributed message broker.<br>
Memphis' platform requires zero ops, enables rapid development, extreme cost reduction, <br>
eliminates coding barriers, and saves a great amount of dev time for data-oriented developers and data engineers.

## 📸 Screenshots
Dashboard             |  Station (Topic) overview|  CLI
:-------------------------:|:-------------------------:|:-------------------------:
<img width="300" alt="Dashboard" src="https://user-images.githubusercontent.com/70286779/182221769-3aa953cc-df71-4c0e-b0d2-9dd4ab83fea9.png">|<img width="300" alt="Station Overview" src="https://user-images.githubusercontent.com/70286779/182221788-0a159007-ab93-46aa-9c81-222671144a05.png">|<img src="https://user-images.githubusercontent.com/70286779/175806007-9a37e130-3e5a-4606-bdda-a71a89efae7f.png" alt="drawing" width="300"/>

## ⭐️ Why
Working with data streaming is HARD.<br>

As a developer, you need to build a dedicated pipeline for each data source,<br>
work with schemas, formats, serializations, analyze each source individually,<br>
enrich the data with other sources, constantly change APIs, and scale for better performance 🥵.<br>
Besides that, it constantly crashes and requires adaptation to different rate limits.<br>
**It takes time and resources that you probably don't have.**<br>

Message broker acts as the middleman and supports streaming architecture,<br>
but then you encounter Apache Kafka and its documentation and run back to the monolith and batch jobs.<br>
**Give memphis{dev} a spin before.**

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

[**Roadmap**](https://github.com/orgs/memphisdev/projects/2/views/1)

**v0.3.6**

- 🚀 Fully optimized message broker in under 3 minutes
- 💻 Easy-to-use UI, CLI, and SDKs
- 📺 Data-level observability
- ☠️ Dead-Letter Queue with automatic message retransmit
- ⛓  SDKs: Node.JS, Go, Python, Typescript, NestJS
- 🐳☸ Runs on your Docker or Kubernetes
- 👨‍💻 Community driven

## 🚀 Getting Started
[Sandbox](https://sandbox.memphis.dev)<br>
[Installation videos](https://www.youtube.com/playlist?list=PL_7iYjqhtXpWpZT2U0zDYo2eGOoGmg2mm)<br><br>
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

<p align="center">
<a href="https://medium.com/memphis-dev/how-to-build-your-own-wolt-app-b220d738bb71"> Build an event-driven food delivery app </a>

</p>

## High-Level Architecture

<p align="center">
<img alt="memphis.dev-logo" height="500" alt="memphis.dev Architecture" src="https://user-images.githubusercontent.com/70286779/193022748-42ed5f37-3083-4fee-a2d0-dabf2ccafb23.png">

</p>

## Local access
### Via Kubernetes
To access Memphis UI from localhost, run the below commands:
```shell
kubectl port-forward service/memphis-ui 9000:80 --namespace memphis > /dev/null &
```

To access Memphis using CLI or SDK from localhost, run the below commands:</br>
```shell
kubectl port-forward service/memphis-cluster 6666:6666 5555:5555 --namespace memphis > /dev/null &
```
Dashboard: http://localhost:9000</br>
Memphis broker: http://localhost:6666

**For Production Environments**
Please expose the UI, Cluster, and Control-plane via k8s ingress / load balancer / nodeport

### Via Docker
UI - http://localhost:9000<br>
Broker - http://localhost:6666<br>

## Beta
Memphis{dev} is currently in Beta version. This means that we are still working on essential features like real-time messages tracing, schema registry and inline processing as well as making more SDKs and supporting materials.

How does it affect you? Well... mostly it doesn't.<br>
(a) The core of memphis broker is highly stable<br>
(b) We learn and fix fast<br><br>
But we need your love, and any help we can get by stars, PR, feedback, issues, and enhancments.<br>
Read more on [Memphis{dev} Documentation 📃](https://memphis.dev/docs).

## Support 🙋‍♂️🤝

### Ask a question ❓ about Memphis{dev} or something related to us:

We welcome you to our discord server with your questions, doubts and feedback.

<a href="https://discord.gg/WZpysvAeTf"><img src="https://amplication.com/images/discord_banner_purple.svg"/></a>

### Create a bug 🐞 report

If you see an error message or run into an issue, please [create bug report](https://github.com/memphisdev/memphis-broker/issues/new?assignees=&labels=type%3A%20bug&template=bug_report.md&title=). This effort is valued and it will help all Memphis{dev} users.


### Submit a feature 💡 request 

If you have an idea, or you think that we're missing a capability that would make development easier and more robust, please [Submit feature request](https://github.com/memphisdev/memphis-broker/issues/new?assignees=&labels=type%3A%20feature%20request).

If an issue❗with similar feature request already exists, don't forget to leave a "+1".
If you add some more information such as your thoughts and vision about the feature, your comments will be embraced warmly :)

## Contributing

Memphis{dev} is an open-source project.<br>
We are committed to a fully transparent development process and appreciate highly any contributions.<br>
Whether you are helping us fix bugs, proposing new features, improving our documentation or spreading the word - we would love to have you as part of the Memphis{dev} community.

Please refer to our [Contribution Guidelines](./CONTRIBUTING.md) and [Code of Conduct](./code_of_conduct.md).

## Contributors ✨

Thanks goes to these wonderful people ❤:<br><br>
 <a href = "https://github.com/memphisdev/memphis-broker/graphs/contributors">
   <img src = "https://contrib.rocks/image?repo=memphisdev/memphis-broker"/>
 </a>

## License 📃
Please check out [License](./LICENSE) to read the full text.
