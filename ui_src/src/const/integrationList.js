// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import datadogBannerPopup from '../assets/images/datadogBannerPopup.webp';
import elasticBannerPopup from '../assets/images/elasticBannerPopup.webp';
import grafanaBannerPopup from '../assets/images/grafanaBannerPopup.webp';
import debeziumBannerPopup from '../assets/images/debeziumBannerPopup.webp';
import slackBannerPopup from '../assets/images/slackBannerPopup.webp';
import pagerdutyBanner from '../assets/images/pagerdutyBanner.webp';
import influxDBBanner from '../assets/images/influxDBBanner.webp';
import newrelicBanner from '../assets/images/newrelicBanner.webp';
import elasticBanner from '../assets/images/elasticBanner.webp';
import s3BannerPopup from '../assets/images/s3BannerPopup.webp';
import datadogBanner from '../assets/images/datadogBanner.webp';
import grafanaBanner from '../assets/images/grafanaBanner.webp';
import debeziumBanner from '../assets/images/debeziumBanner.webp';
import pagerDutyIcon from '../assets/images/pagerDutyIcon.svg';
import githubIntegrationIcon from '../assets/images/githubIntegrationIcon.svg';
import githubBannerPopup from '../assets/images/githubBannerPopup.webp';
import githubBanner from '../assets/images/githubBanner.webp';
import newrelicIcon from '../assets/images/newrelicIcon.svg';
import influxDBIcon from '../assets/images/influxDBIcon.svg';
import slackBanner from '../assets/images/slackBanner.webp';
import datadogIcon from '../assets/images/datadogIcon.svg';
import grafanaIcon from '../assets/images/grafanaIcon.svg';
import debeziumIcon from '../assets/images/debeziumIcon.svg';
import elasticIcon from '../assets/images/elasticIcon.svg';
import slackLogo from '../assets/images/slackLogo.svg';
import s3Banner from '../assets/images/s3Banner.webp';
import s3Logo from '../assets/images/s3Logo.svg';

import { ColorPalette } from './globalConst';

export const CATEGORY_LIST = {
    All: {
        name: 'All',
        color: ColorPalette[13]
    },
    Monitoring: {
        name: 'Monitoring',
        color: ColorPalette[8],
        osOnly: true
    },
    Notifications: {
        name: 'Notifications',
        color: ColorPalette[0]
    },
    'Storage Tiering': {
        name: 'Storage Tiering',
        color: ColorPalette[4]
    },
    'Change-Data-Capture': {
        name: 'Change-Data-Capture',
        color: ColorPalette[11]
    }
    // SourceCode: {
    //     name: 'Source Code',
    //     color: ColorPalette[6]
    // }
};

export const REGIONS_OPTIONS = [
    {
        name: 'US East (N. Virginia) [us-east-1]',
        value: 'us-east-1'
    },
    {
        name: 'US East (Ohio) [us-east-2]',
        value: 'us-east-2'
    },
    {
        name: 'US West (N. California) [us-west-1]',
        value: 'us-west-1'
    },
    {
        name: 'US West (Oregon) [us-west-2]',
        value: 'us-west-2'
    },
    {
        name: 'Asia Pacific (Mumbai) [ap-south-1]',
        value: 'ap-south-1'
    },
    {
        name: 'Asia Pacific (Osaka) [ap-northeast-3]',
        value: 'ap-northeast-3'
    },
    {
        name: 'Asia Pacific (Seoul) [ap-northeast-2]',
        value: 'ap-northeast-2'
    },
    {
        name: 'Asia Pacific (Singapore) [ap-southeast-1]',
        value: 'ap-southeast-1'
    },
    {
        name: 'Asia Pacific (Sydney) [ap-southeast-2]',
        value: 'ap-southeast-2'
    },
    {
        name: 'Asia Pacific (Tokyo) [ap-northeast-1]',
        value: 'ap-northeast-1'
    },
    {
        name: 'Canada (Central) [ca-central-1]',
        value: 'ca-central-1'
    },
    {
        name: 'Africa (Cape Town) [af-south-1]',
        value: 'af-south-1'
    },
    {
        name: 'Asia Pacific (Hong Kong) [ap-east-1]',
        value: 'ap-east-1'
    },
    {
        name: 'Asia Pacific (Hyderabad) [ap-south-2]',
        value: 'ap-south-2'
    },
    {
        name: 'Asia Pacific (Jakarta) [ap-southeast-3]',
        value: 'ap-southeast-3'
    },
    {
        name: 'Europe (Frankfurt) [eu-central-1]',
        value: 'eu-central-1'
    },
    {
        name: 'Europe (Ireland) [eu-west-1]',
        value: 'eu-west-1'
    },
    {
        name: 'Europe (London) [eu-west-2]',
        value: 'eu-west-2'
    },
    {
        name: 'Europe (Paris) [eu-west-3]',
        value: 'eu-west-3'
    },
    {
        name: 'Europe (Stockholm) [eu-north-1]',
        value: 'eu-north-1'
    },
    {
        name: 'Europe (Milan) [eu-south-1]',
        value: 'eu-south-1'
    },
    {
        name: 'Europe (Spain) [eu-south-2]',
        value: 'eu-south-2'
    },
    {
        name: 'Europe (Zurich) [eu-central-2]',
        value: 'eu-central-2'
    },
    {
        name: 'Middle East (Bahrain) [me-south-1]',
        value: 'me-south-1'
    },
    {
        name: 'Middle East (UAE) [me-central-1]',
        value: 'me-central-1'
    },
    {
        name: 'South America (São Paulo) [sa-east-1]',
        value: 'sa-east-1'
    }
];

export const INTEGRATION_LIST = {
    Datadog: {
        name: 'Datadog',
        by: 'memphis',
        banner: <img className="banner" src={datadogBanner} alt="datadogBanner" />,
        insideBanner: <img className="insideBanner" src={datadogBannerPopup} alt="datadogBannerPopup" />,
        icon: <img src={datadogIcon} alt="datadogIcon" />,
        description: 'Datadog is an end-to-end monitoring and observability platform. Memphis can integrate with your custom dashboard in datadog',
        category: CATEGORY_LIST['Monitoring'],
        osOnly: true,
        header: (
            <div className="header-left-side">
                <img src={datadogIcon} alt="datadogIcon" />
                <div className="details">
                    <p>Datadog</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    Datadog is an end-to-end monitoring and observability platform. Memphis can integrate with your custom dashboard in datadog
                </span>
            </div>
        ),
        steps: [
            {
                title: 'Step 1: Make sure your Memphis Prometheus exporter is on',
                key: 0
            },
            {
                title: 'Step 2: Add Datadog annotation to Memphis statefulset',
                key: 1
            },
            {
                title: 'Step 3: Check Datadog for Memphis metrics',
                key: 2
            },
            {
                title: 'Step 4: Import the Memphis dashboard',
                key: 3
            }
        ]
    },
    Slack: {
        name: 'Slack',
        by: 'memphis',
        banner: <img className="banner" src={slackBanner} alt="slackBanner" />,
        insideBanner: <img className="insideBanner" src={slackBannerPopup} alt="slackBannerPopup" />,
        icon: <img src={slackLogo} alt="slackLogo" />,
        description: 'Receive alerts and notifications directly to your chosen slack channel for faster response and better real-time observability',
        category: CATEGORY_LIST['Notifications'],
        header: (
            <div className="header-left-side">
                <img src={slackLogo} alt="slackLogo" />
                <div className="details">
                    <p>Slack</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    Receive alerts and notifications directly to your chosen slack channel for faster response and better real-time observability
                </span>
            </div>
        )
    },
    S3: {
        name: 'S3',
        by: 'memphis',
        banner: <img className="banner" src={s3Banner} alt="s3Banner" />,
        insideBanner: <img className="insideBanner" src={s3BannerPopup} alt="s3BannerPopup" />,
        icon: <img src={s3Logo} alt="s3Logo" />,
        description:
            'S3-compatible storage providers offer cost-efficient object storage and can act as a 2nd tier storage option for ingested messages—vendor examples: AWS S3, Backblaze B2, DigitalOcean Spaces, or Minio.',
        date: 'Jan 1, 2023',
        category: CATEGORY_LIST['Storage Tiering'],
        header: (
            <div className="header-left-side">
                <img src={s3Logo} alt="s3Logo" />
                <div className="details">
                    <p>S3 Compatible Object Storage</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    S3-compatible storage providers offer cost-efficient object storage and can act as a 2nd tier storage option for ingested messages—vendor examples:
                    AWS S3, Backblaze B2, DigitalOcean Spaces, or Minio.
                </span>
            </div>
        )
    },
    // GitHub: {
    //     name: 'Github',
    //     by: 'memphis',
    //     banner: <img className="banner" src={githubBanner} alt="gitHubBanner" />,
    //     insideBanner: <img className="insideBanner" src={githubBannerPopup} alt="slackBannerPopup" />,
    //     icon: <img src={githubIntegrationIcon} alt="gitHubIcon" />,
    //     description:
    //         'GitHub is an open source code repository and collaborative software development platform. Use GitHub repositories to manage your Schemaverse schemas and Functions source code.',
    //     category: CATEGORY_LIST['SourceCode'],
    //     header: (
    //         <div className="header-left-side">
    //             <img src={githubIntegrationIcon} alt="gitHubLogo" />
    //             <div className="details">
    //                 <p>GitHub</p>
    //                 <span>by memphis</span>
    //             </div>
    //         </div>
    //     ),
    //     integrateDesc: (
    //         <div className="integrate-description">
    //             <p>Description</p>
    //             <span className="content">
    //                 GitHub is an open source code repository and collaborative software development platform. Use GitHub repositories to manage your Schemaverse schemas
    //                 and Functions source code.
    //             </span>
    //         </div>
    //     )
    // },
    Elasticsearch: {
        name: 'Elasticsearch observability',
        by: 'memphis',
        banner: <img className="banner" src={elasticBanner} alt="elasticBanner" />,
        insideBanner: <img className="insideBanner" src={elasticBannerPopup} alt="elasticBannerPopup" />,
        icon: <img src={elasticIcon} alt="elasticIcon" />,
        description: 'Monitor and observe Memphis infrastructure using Elasticsearch Observability and Kibana',
        category: CATEGORY_LIST['Monitoring'],
        experimental: true,
        osOnly: true,
        header: (
            <div className="header-left-side">
                <img src={elasticIcon} alt="elasticIcon" />
                <div className="details">
                    <p>Elasticsearch observability</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">Monitor and observe Memphis infrastructure using Elasticsearch Observability and Kibana</span>
            </div>
        ),
        steps: [
            {
                title: 'Step 1: Download the Elastic Agent manifest',
                key: 0
            },
            {
                title: 'Step 2: Configure Elastic Agent policy',
                key: 1
            },
            {
                title: 'Step 3: Enroll Elastic Agent to the policy',
                key: 2
            },
            {
                title: 'Step 4: Deploy the Elastic Agent',
                key: 3
            }
        ]
    },
    Grafana: {
        name: 'Grafana',
        by: 'memphis',
        banner: <img className="banner" src={grafanaBanner} alt="grafanaBanner" />,
        insideBanner: <img className="insideBanner" src={grafanaBannerPopup} alt="grafanaBannerPopup" />,
        icon: <img src={grafanaIcon} alt="grafanaIcon" />,
        description: 'Visualize Memphis metrics using Grafana and prometheus',
        category: CATEGORY_LIST['Monitoring'],
        osOnly: true,
        header: (
            <div className="header-left-side">
                <img src={grafanaIcon} alt="grafanaIcon" />
                <div className="details">
                    <p>Grafana</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">Visualize Memphis metrics using Grafana and prometheus</span>
            </div>
        ),
        steps: [
            {
                title: `Step 0: Configuring Prometheus to collect pods' logs`,
                key: 0
            },
            {
                title: 'Step 1: Enabling Memphis Prometheus exporter',
                key: 1
            },
            {
                title: 'Step 2: Import Memphis dashboard',
                key: 2
            }
        ]
    },
    'Debezium and Postgres': {
        name: 'Debezium and Postgres',
        by: 'memphis',
        banner: <img className="banner" src={debeziumBanner} alt="debeziumBanner" />,
        insideBanner: <img className="insideBanner" src={debeziumBannerPopup} alt="debeziumBannerPopup" />,
        icon: <img src={debeziumIcon} alt="debeziumIcon" />,
        description:
            'Debezium is one of the most popular frameworks for collecting "Change Data Capture (CDC)" events from various databases and can now be easily integrated with Memphis.dev for collecting CDC events from various databases.',
        category: CATEGORY_LIST['Change-Data-Capture'],
        header: (
            <div className="header-left-side">
                <img src={debeziumIcon} alt="debeziumIcon" />
                <div className="details">
                    <p>Debezium and Postgres</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    Debezium is one of the most popular frameworks for collecting "Change Data Capture (CDC)" events from various databases and can now be easily
                    integrated with Memphis.dev for collecting CDC events from various databases.
                </span>
            </div>
        ),
        steps: [
            {
                title: 'Step 0: Create an client-type Memphis user for Debezium',
                key: 0
            },
            {
                title: 'Step 1: Setup Debezium',
                key: 1
            }
        ]
    },
    PagerDuty: {
        name: 'PagerDuty',
        by: 'memphis',
        banner: <img className="banner" src={pagerdutyBanner} alt="pagerdutyBanner" />,
        insideBanner: <img className="insideBanner" src={pagerdutyBanner} alt="pagerdutyBanner" />,
        icon: <img src={pagerDutyIcon} alt="pagerDutyIcon" />,
        description: 'In PagerDuty, you can configure operations schedules to allow for 24x7 monitoring by an operations team that can span the globe.',
        category: CATEGORY_LIST['Notifications'],
        comingSoon: true,
        osOnly: true,
        header: (
            <div className="header-left-side">
                <img src={pagerDutyIcon} alt="pagerDutyIcon" />
                <div className="details">
                    <p>PagerDuty</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    In PagerDuty, you can configure operations schedules to allow for 24x7 monitoring by an operations team that can span the globe.
                </span>
            </div>
        )
    },
    'New Relic': {
        name: 'New Relic',
        by: 'memphis',
        banner: <img className="banner" src={newrelicBanner} alt="newrelicBanner" />,
        insideBanner: <img className="insideBanner" src={newrelicBanner} alt="newrelicBanner" />,
        icon: <img src={newrelicIcon} alt="newrelicIcon" />,
        description: 'New Relic is where dev, ops, security and business teams solve software. Integrate memphis logs and metrics with New Relic',
        comingSoon: true,
        category: CATEGORY_LIST['Monitoring'],
        osOnly: true,
        header: (
            <div className="header-left-side">
                <img src={newrelicIcon} alt="newrelicIcon" />
                <div className="details">
                    <p>New Relic</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    New Relic is where dev, ops, security and business teams solve software. Integrate memphis logs and metrics with New Relic
                </span>
            </div>
        )
    },
    influxDB: {
        name: 'influxDB',
        by: 'memphis',
        banner: <img className="banner" src={influxDBBanner} alt="influxDBBanner" />,
        insideBanner: <img className="insideBanner" src={influxDBBanner} alt="influxDBBanner" />,
        icon: <img src={influxDBIcon} alt="influxDBIcon" />,
        description: 'Ship memphis logs to influxDB for near real-time monitoring with Grafana visualization',
        category: CATEGORY_LIST['Monitoring'],
        comingSoon: true,
        osOnly: true,
        header: (
            <div className="header-left-side">
                <img src={influxDBIcon} alt="influxDBIcon" />
                <div className="details">
                    <p>influxDB</p>
                    <span>by memphis</span>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">Ship memphis logs to influxDB for near real-time monitoring with Grafana visualization</span>
            </div>
        )
    }
};
