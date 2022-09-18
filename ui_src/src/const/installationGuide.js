// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
import gcpIcon from '../assets/images/gcpIcon.svg';
import awsIcon from '../assets/images/awsIcon.svg';

export const INSTALLATION_GUIDE = {
    Main: {
        header: 'Installation',
        description: (
            <span>
                Please choose your preferred environment to deploy memphis on{' '}
                <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                    Learn More
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
                command: `kubectl port-forward service/memphis-cluster 6666:6666 9000:9000 --namespace memphis > /dev/null &`,
                icon: 'copy'
            },
            {
                title: (
                    <span>
                        Step 3 - Open memphis{' '}
                        <a href="http://localhost:5555" target="_blank">
                            UI
                        </a>
                    </span>
                ),
                command: (
                    <a href="http://localhost:5555" target="_blank">
                        http://localhost:5555
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
                        <a href="http://localhost:5555" target="_blank">
                            UI
                        </a>
                    </span>
                ),
                command: (
                    <a href="http://localhost:5555" target="_blank">
                        http://localhost:5555
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
                src: <img src={awsIcon} />,
                docsLink: 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/deployment/cloud-deployment/deploy-on-aws'
            },
            {
                name: 'gcp',
                src: <img src={gcpIcon} />,
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
