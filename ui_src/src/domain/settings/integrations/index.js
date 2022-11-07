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

import '../style.scss';
import './style.scss';

import React, { useState } from 'react';

import Switcher from '../../../components/switcher';

const Integrations = () => {
    const [hubIntegration, setHubIntegration] = useState(false);
    const [slackIntegration, setSlackIntegration] = useState(false);
    return (
        <div className="alerts-integrations-container">
            <h3 className="title">Some sentence</h3>
            <div>
                <div className="hub-connect-integration">
                    <div className="alert-integration-type">
                        <label className="integration-label-bold">Memphis hub</label>
                        <Switcher onChange={() => setHubIntegration(!hubIntegration)} checked={hubIntegration} checkedChildren="on" unCheckedChildren="off" />
                    </div>
                    {!hubIntegration && <p>Signin placeholder</p>}
                </div>
                <div className="alert-integration-type">
                    <label className="alert-integration-label">Slack</label>
                    <Switcher onChange={() => setSlackIntegration(!slackIntegration)} checked={slackIntegration} checkedChildren="on" unCheckedChildren="off" />
                </div>
            </div>
        </div>
    );
};

export default Integrations;
