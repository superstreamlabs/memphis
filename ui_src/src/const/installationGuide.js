// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
import gcpIcon from '../assets/images/gcpIcon.svg';
import awsIcon from '../assets/images/awsIcon.svg';

export const INSTALLATION_GUIDE = {
    Main: {
        header: 'Installation',
        description: (
            <span>
                Please choose your preferred environment to deploy memphis on{' '}
                <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                    Learn more
                </a>
            </span>
        )
    },
    Kubernetes: {
        header: 'Installation/Kubernetes',
        description: <span>Memphis can be deployed over any kubernetes cluster above version 1.20, including minikube</span>,
        steps: [
            {
                title: 'Step 1 - Copy & Paste to your terminal',
                command: `helm repo add memphis https://k8s.memphis.dev/charts/ --force-update &&  \nhelm install my-memphis memphis/memphis --create-namespace --namespace memphis`,
                icon: 'copy'
            },
            {
                title: 'Step 2 - Expose memphis to your localhost',
                command: `kubectl port-forward service/memphis-cluster 6666:6666 9000:9000 7770:7770 --namespace memphis > /dev/null &`,
                icon: 'copy'
            },
            {
                title: (
                    <span>
                        Step 3 - Open memphis{' '}
                        <a href="http://localhost:9000" target="_blank">
                            UI
                        </a>
                    </span>
                ),
                command: (
                    <a href="http://localhost:9000" target="_blank">
                        http://localhost:9000
                    </a>
                ),
                icon: 'link'
            }
        ],
        showLinks: true,
        videoLink: 'https://youtu.be/OmUJXqvFK4M',
        docsLink: 'https://docs.memphis.dev/memphis-new/deployment/kubernetes'
    },
    'Docker Compose': {
        header: 'Installation/Docker',
        description: <span>Memphis can be deployed over docker engine, swarm, and compose</span>,
        steps: [
            {
                title: 'Step 1 - Copy & Paste to your terminal',
                command: `curl -s https://memphisdev.github.io/memphis-docker/docker-compose.yml -o docker-compose.yml && \ndocker compose -f docker-compose.yml -p memphis up`,
                icon: 'copy'
            },
            {
                title: (
                    <span>
                        Step 2 - Open memphis{' '}
                        <a href="http://localhost:9000" target="_blank">
                            UI
                        </a>
                    </span>
                ),
                command: (
                    <a href="http://localhost:9000" target="_blank">
                        http://localhost:9000
                    </a>
                ),
                icon: 'link'
            }
        ],
        showLinks: true,
        videoLink: 'https://youtu.be/cXAk60hMtHs',
        docsLink: 'https://docs.memphis.dev/memphis-new/deployment/docker-compose#step-1-download-compose.yaml-file'
    },
    'Cloud Providers': {
        header: 'Installation/Cloud Providers',
        description: <span>Deploy Memphis to your preferred cloud provider directly. Dedicated kubernetes cluster with memphis installed will be deployed.</span>,
        clouds: [
            {
                name: 'aws',
                src: <img src={awsIcon} alt="awsIcon" />,
                docsLink: 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/deployment/cloud-deployment/deploy-on-aws'
            },
            {
                name: 'gcp',
                src: <img src={gcpIcon} alt="gcpIcon" />,
                docsLink: 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/deployment/cloud-deployment/deploy-on-gcp'
            }
        ],
        aws: [
            {
                title: 'Step 0 - Clone Memphis-Terraform repo',
                command: `git clone git@github.com:memphisdev/memphis-terraform.git && cd memphis-terraform`,
                icon: 'copy'
            },
            {
                title: 'Step 1 - Deploy',
                command: `make -C ./AWS/EKS/ allinone`,
                icon: 'copy'
            }
        ],
        gcp: [
            {
                title: 'Step 0 - Clone Memphis-Terraform repo',
                command: `git clone git@github.com:memphisdev/memphis-terraform.git && cd memphis-terraform`,
                icon: 'copy'
            },
            {
                title: 'Step 1 - Deploy',
                command: `make -C ./GCP/GKE/ allinone`,
                icon: 'copy'
            }
        ]
    }
};
