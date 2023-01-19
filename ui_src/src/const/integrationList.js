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
import slackBannerPopup from '../assets/images/slackBannerPopup.svg';
import pagerdutyBanner from '../assets/images/pagerdutyBanner.svg';
import influxDBBanner from '../assets/images/influxDBBanner.svg';
import newrelicBanner from '../assets/images/newrelicBanner.svg';
import s3BannerPopup from '../assets/images/s3BannerPopup.svg';
import datadogBanner from '../assets/images/datadogBanner.svg';
import pagerDutyIcon from '../assets/images/pagerDutyIcon.svg';
import newrelicIcon from '../assets/images/newrelicIcon.svg';
import influxDBIcon from '../assets/images/influxDBIcon.svg';
import slackBanner from '../assets/images/slackBanner.svg';
import datadogIcon from '../assets/images/datadogIcon.svg';
import slackLogo from '../assets/images/slackLogo.svg';
import s3Banner from '../assets/images/s3Banner.svg';
import s3Logo from '../assets/images/s3Logo.svg';

import { FiberManualRecord } from '@material-ui/icons';
import { diffDate } from '../services/valueConvertor';
import { ColorPalette } from './globalConst';

export const CATEGORY_LIST = {
    All: {
        name: 'All',
        color: ColorPalette[13]
    },
    Notifications: {
        name: 'Notifications',
        color: ColorPalette[0]
    },
    Storage: {
        name: 'Storage',
        color: ColorPalette[4]
    },
    Monitoring: {
        name: 'Monitoring',
        color: ColorPalette[8]
    }
};

export const REGIONS_OPTIONS = [
    {
        name: 'US East (Ohio) [us-east-2]',
        value: 'us-east-2'
    },
    {
        name: 'US East (N. Virginia) [us-east-1]',
        value: 'us-east-1'
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
        name: 'Europe (Milan) [eu-south-1]',
        value: 'eu-south-1'
    },
    {
        name: 'Europe (Paris) [eu-west-3]',
        value: 'eu-west-3'
    },
    {
        name: 'Europe (Spain) [eu-south-2]',
        value: 'eu-south-2'
    },
    {
        name: 'Europe (Stockholm) [eu-north-1]',
        value: 'eu-north-1'
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

export const INTEGRATION_LIST = [
    {
        name: 'Slack',
        by: 'memphis',
        banner: <img className="banner" src={slackBanner} alt="slackBanner" />,
        insideBanner: <img className="insideBanner" src={slackBannerPopup} alt="slackBannerPopup" />,
        icon: <img src={slackLogo} alt="slackLogo" />,
        description: 'Receive alerts and notifications directly to your chosen slack channel for faster response and better real-time observability',
        date: 'Nov 19, 2022',
        category: CATEGORY_LIST['Notifications'],
        header: (
            <div className="header-left-side">
                <img src={slackLogo} alt="slackLogo" />
                <div className="details">
                    <p>Slack</p>
                    <>
                        <span>by memphis</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate('Nov 19, 2022')}</span>
                    </>
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
    {
        name: 'Amazon S3',
        by: 'memphis',
        banner: <img className="banner" src={s3Banner} alt="s3Banner" />,
        insideBanner: <img className="insideBanner" src={s3BannerPopup} alt="s3BannerPopup" />,
        icon: <img src={s3Logo} alt="s3Logo" />,
        description: 'Amazon S3 offers cost-efficient object storage and can act as a 2nd tier storage option for ingested messages',
        date: 'Jan 1, 2023',
        category: CATEGORY_LIST['Storage'],
        header: (
            <div className="header-left-side">
                <img src={s3Logo} alt="s3Logo" />
                <div className="details">
                    <p>Amazon S3</p>
                    <>
                        <span>by memphis</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate('Jan 1, 2023')}</span>
                    </>
                </div>
            </div>
        ),
        integrateDesc: (
            <div className="integrate-description">
                <p>Description</p>
                <span className="content">Amazon S3 offers cost-efficient object storage and can act as a 2nd tier storage option for ingested messages.</span>
            </div>
        )
    },
    {
        name: 'PagerDuty',
        by: 'memphis',
        banner: <img className="banner" src={pagerdutyBanner} alt="pagerdutyBanner" />,
        insideBanner: <img className="insideBanner" src={pagerdutyBanner} alt="pagerdutyBanner" />,
        icon: <img src={pagerDutyIcon} alt="pagerDutyIcon" />,
        description: 'In PagerDuty, you can configure operations schedules to allow for 24x7 monitoring by an operations team that can span the globe.',
        date: 'Nov 19, 2022',
        category: CATEGORY_LIST['Notifications'],
        comingSoon: true,
        header: (
            <div className="header-left-side">
                <img src={pagerDutyIcon} alt="pagerDutyIcon" />
                <div className="details">
                    <p>PagerDuty</p>
                    <>
                        <span>by memphis</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate('Nov 19, 2022')}</span>
                    </>
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
    {
        name: 'New Relic',
        by: 'memphis',
        banner: <img className="banner" src={newrelicBanner} alt="newrelicBanner" />,
        insideBanner: <img className="insideBanner" src={newrelicBanner} alt="newrelicBanner" />,
        icon: <img src={newrelicIcon} alt="newrelicIcon" />,
        description: 'New Relic is where dev, ops, security and business teams solve software. Integrate memphis logs and metrics with New Relic',
        date: 'Nov 19, 2022',
        category: CATEGORY_LIST['Monitoring'],
        comingSoon: true,
        header: (
            <div className="header-left-side">
                <img src={newrelicIcon} alt="newrelicIcon" />
                <div className="details">
                    <p>New Relic</p>
                    <>
                        <span>by memphis</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate('Nov 19, 2022')}</span>
                    </>
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
    {
        name: 'Datadog',
        by: 'memphis',
        banner: <img className="banner" src={datadogBanner} alt="datadogBanner" />,
        insideBanner: <img className="insideBanner" src={datadogBanner} alt="datadogBanner" />,
        icon: <img src={datadogIcon} alt="datadogIcon" />,
        description: 'Datadog is an end-to-end monitoring and observability platform. Memphis can integrate with your custom dashboard in datadog',
        date: 'Nov 19, 2022',
        category: CATEGORY_LIST['Monitoring'],
        comingSoon: true,
        header: (
            <div className="header-left-side">
                <img src={datadogIcon} alt="datadogIcon" />
                <div className="details">
                    <p>Datadog</p>
                    <>
                        <span>by memphis</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate('Nov 19, 2022')}</span>
                    </>
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
        )
    },
    {
        name: 'influxDB',
        by: 'memphis',
        banner: <img className="banner" src={influxDBBanner} alt="influxDBBanner" />,
        insideBanner: <img className="insideBanner" src={influxDBBanner} alt="influxDBBanner" />,
        icon: <img src={influxDBIcon} alt="influxDBIcon" />,
        description: 'Ship memphis logs to influxDB for near real-time monitoring with Grafana visualization',
        date: 'Nov 19, 2022',
        category: CATEGORY_LIST['Monitoring'],
        comingSoon: true,
        header: (
            <div className="header-left-side">
                <img src={influxDBIcon} alt="influxDBIcon" />
                <div className="details">
                    <p>influxDB</p>
                    <>
                        <span>by memphis</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate('Nov 19, 2022')}</span>
                    </>
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
];
