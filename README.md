![](https://memphis-public-files.s3.eu-central-1.amazonaws.com/Vector_page-0001.jpg)
<br><br>
![Github tag](https://img.shields.io/github/v/release/memphis-os/memphis-control-plane) [![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/Memphis-OS/memphis-broker/commit-activity) [![GoReportCard example](https://goreportcard.com/badge/github.com/nanomsg/mangos)](https://goreportcard.com/report/github.com/nanomsg/mangos)

Too many data sources and too many schemas? Looking for a messaging queue to scale your data-driven architecture? Require greater performance for your data streams? Your architecture is based on post-processing data, and you want to switch to real-time in minutes instead of months? Struggle to install, configure and update Kafka/RabbitMQ/and other MQs?

**Meet Memphis**

**[Memphis](https://memphis.dev)** is a dev-first, cloud-native, event processing platform made out of devs' struggles with tools like Kafka, RabbitMQ, NATS, and others, allowing you to achieve all other message brokers' benefits in a fraction of the time.<br><br>
**[Memphis](https://memphis.dev) delivers:**
- The most simple to use Message Broker (With the same behaivour as NATS and Kafka)
- State-of-the-art UI and CLI
- No need for Kafka Connect, Kafka Streams, ksql. All the tools you need are under the same roof
- An in-line data processing in any programming language
- Out-of-the-box deep observability of every component

RabbitMQ has Queues, Kafka as Topics, **Memphis has Stations.**
#### TL;DR
**On Day 1 (For the DevOps heros out there) -**<br>
Memphis platform provides the same old and loved behavior (Produce-Consume) of other data lakes and MQs, but removes completly the complexity barriers, messy documentation, ops, manual scale, orchestration and more.

**On Day 2 (For the Developers) -**
Developer lives with developing real-time, event-driven apps that are too complex.
Consumers and Producers are filled with logic, data orchestration is needed between the different services, no GUI to understand metrics and flows, lack of monitoring, hard to implement SDKs, etc.

No More.

In the coming versions, Memphis will answer the challenges above,<br>and recude 90% of dev work arround building a real-time / event-driven / data-driven apps.

---

**Purpose of this repo**<br>
For Memphis control-plane.
The control-plane is the operating system that controls Memphis platform.

**Table of Contents**
- [Memphis Components](#memphis-components)
- [Memphis repos](#memphis-repos)
- [Current SDKs](#current-sdks)
- [Installation](#installation)
  - [Kubernetes](#kubernetes)
    - [Install](#install)
    - [K8S Diagram](#k8s-diagram)
  - [Docker](#docker)
    - [Install](#install-1)
- [Next Steps](#next-steps)
  - [Kubernetes](#kubernetes-1)
    - [Localhost Environment](#localhost-environment)
    - [Production Environments](#production-environments)
  - [Docker](#docker-1)
- [Memphis Contributors](#memphis-contributors)
- [Contribution guidelines](#contribution-guidelines)
- [Documentation](#documentation)
- [Contact](#contact)
## Memphis Components
![](https://memphis-public-files.s3.eu-central-1.amazonaws.com/graphics+for+github/components+diagram+-+cp.png )

## Memphis repos
- [memphis-control-plane](https://github.com/Memphis-OS/memphis-control-plane "memphis-control-plane")
- [memphis-ui](https://github.com/Memphis-OS/memphis-ui "memphis-ui")
- [memphis-broker](https://github.com/Memphis-OS/memphis-broker "memphis-broker")
- [memphis-cli](https://github.com/Memphis-OS/memphis-cli "memphis-cli")
- [memphis-k8s](https://github.com/Memphis-OS/memphis-k8s "memphis-k8s")
- [memphis-docker](https://github.com/Memphis-OS/memphis-docker "memphis-docker")

## Current SDKs
- [memphis-js](https://github.com/Memphis-OS/memphis.js "Node.js")

## Installation

### Kubernetes
#### Install
```shell
helm repo add memphis https://k8s.memphis.dev/charts/
helm install my-memphis memphis/memphis --create-namespace --namespace memphis
```

**Helm chart options**<br>
Example:<br>
`helm install my-memphis --set cluster.replicas=1,rootPwd="rootpassword" memphis/memphis --create-namespace --namespace memphis`

|  Option |Description   |Default Value   |
| :------------ | :------------ | :------------ |
|rootPwd   |Root password for the dashboard   |`"memphis"`   |
|connectionToken   |Token for connecting an app to the Memphis Message Queue. Auto Generated   |`""`   |
|dashboard.port   |Dashboard's (GUI) port   |80   |
|cluster.replicas   |Amount of Message Queue workers   |3   |

#### K8S Diagram
![](https://memphis-public-files.s3.eu-central-1.amazonaws.com/Untitled+Diagram.png)

---

### Docker
#### Install
    curl -s https://memphis-os.github.io/memphis-docker/docker-compose.yml -o docker-compose.yml
    docker compose -f docker-compose.yml -p memphis up

The following will be deployed as docker containers
```shell
memphis-control-plane-1
memphis-ui-1
memphis-cluster-1
memphis-mongo-1
```



## Next Steps
### Kubernetes
#### Localhost Environment
```shell
Memphis UI can be accessed via port 80 on the following DNS name from within your cluster: 
memphis-ui.memphis.svc.cluster.local

To access Memphis from localhost, run the below commands:
  1. kubectl port-forward service/memphis-ui 9000:80 --namespace memphis &
  2. kubectl port-forward service/memphis-cluster 7766:7766 --namespace memphis &
  3. kubectl port-forward service/control-plane 6666:6666 6667:80 --namespace memphis &

Dashboard: http://localhost:9000
```
#### Production Environments
Please expose the UI, Cluster, and Control-plane via k8s ingress / load balancer / nodeport

------------

### Docker
**To access Memphis, run the below commands:**
Dashboard - `http://localhost:9000`<br>
Broker - `localhost:7766`<br>
Control-Plane for CLI - `localhost:5555`<br>
Control-Plane for SDK - `localhost:6666` + `localhost:5555`

## Memphis Contributors
<img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Alon+Avrahami.jpg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Ariel+Bar.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Arjun+Anjaria.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Carlos+Gasperi.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Daniel+Eliyahu.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Itay+Katz.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Jim+Doty.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Nikita+Aizenberg.jpg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Rado+Marina.jpg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"><img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Raghav+Ramesh.jpg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Tal+Goldberg.jpg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;"> <img src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/contributors-images/Yehuda+Mizrahi.jpeg" width="60" height="60" style="border-radius: 25px; border: 2px solid #61DFC6;">

## Contribution guidelines

soon

## Documentation

- [Official documentation](https://docs.memphis.dev)

## Contact 
- [Slack](https://bit.ly/37uwCPd): Q&A, Help, Feature requests, and more
- [Twitter](https://bit.ly/3xzkxTx): Follow us on Twitter!
- [Discord](https://bit.ly/3OfnuhX): Join our Discord Server!
- [Medium](https://bit.ly/3ryFDgS): Follow our Medium page!
- [Youtube](https://bit.ly/38Y8rcq): Subscribe our youtube channel!
