// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
import r from '../assets/images/R.svg';
import figmaIcon from '../assets/images/figmaIcon.svg';
import insideBanner from '../assets/images/insideBanner.svg';
import { FiberManualRecord } from '@material-ui/icons';
import { diffDate } from '../services/valueConvertor';
import Button from '../components/button';

export const INTEGRATION_LIST = [
    {
        name: 'Slack',
        by: 'memphis',
        banner: <img className="banner" src={r} alt="banner" />,
        insideBanner: <img className="insideBanner" src={insideBanner} alt="insideBanner" />,
        icon: <img src={figmaIcon} alt="figmaIcon" />,
        description: 'Receive alerts and notifications directly to your chosen slack channel for faster response and better real-time observability',
        date: 'Nov 19, 2022',
        header: (
            <div className="header-left-side">
                <img src={figmaIcon} alt="figmaIcon" />
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
        name: 'PagerDuty',
        by: 'memphis',
        banner: <img className="banner" src={r} alt="banner" />,
        insideBanner: <img className="insideBanner" src={insideBanner} alt="insideBanner" />,
        icon: <img src={figmaIcon} alt="figmaIcon" />,
        description: 'In PagerDuty, you can configure operations schedules to allow for 24x7 monitoring by an operations team that can span the globe.',
        date: 'Nov 19, 2022',
        header: (
            <div className="header-left-side">
                <img src={figmaIcon} alt="figmaIcon" />
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
        banner: <img className="banner" src={r} alt="banner" />,
        insideBanner: <img className="insideBanner" src={insideBanner} alt="insideBanner" />,
        icon: <img src={figmaIcon} alt="figmaIcon" />,
        description: 'New Relic is where dev, ops, security and business teams solve software. Integrate memphis logs and metrics with New Relic',
        date: 'Nov 19, 2022',
        header: (
            <div className="header-left-side">
                <img src={figmaIcon} alt="figmaIcon" />
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
        banner: <img className="banner" src={r} alt="banner" />,
        insideBanner: <img className="insideBanner" src={insideBanner} alt="insideBanner" />,
        icon: <img src={figmaIcon} alt="figmaIcon" />,
        description: 'Datadog is an end-to-end monitoring and observability platform. Memphis can integrate with your custom dashboard in datadog',
        date: 'Nov 19, 2022',
        header: (
            <div className="header-left-side">
                <img src={figmaIcon} alt="figmaIcon" />
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
        banner: <img className="banner" src={r} alt="banner" />,
        insideBanner: <img className="insideBanner" src={insideBanner} alt="insideBanner" />,
        icon: <img src={figmaIcon} alt="figmaIcon" />,
        description: 'Ship memphis logs to influxDB for near real-time monitoring with Grafana visualization',
        date: 'Nov 19, 2022',
        header: (
            <div className="header-left-side">
                <img src={figmaIcon} alt="figmaIcon" />
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
