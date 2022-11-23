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

export const INTEGRATION_LIST = [
    {
        name: 'Slack',
        by: 'memphis',
        banner: <img className="banner" src={r} alt="banner" />,
        insideBanner: <img className="insideBanner" src={insideBanner} alt="insideBanner" />,
        icon: <img src={figmaIcon} alt="figmaIcon" />,
        description: 'Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC',
        date: 'Nov 19, 2022',
        header: (
            <div className="integrate-header">
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
                    Receive alerts and notifications directly to your chosen slack channel for faster response and better real-time observability. Read More
                </span>
            </div>
        )
    }
];
