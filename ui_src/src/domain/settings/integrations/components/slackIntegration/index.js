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

import './style.scss';

import React from 'react';
import r from '../../../../../assets/images/R.svg';
import figmaIcon from '../../../../../assets/images/figmaIcon.svg';
import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { FiberManualRecord } from '@material-ui/icons';
import { diffDate } from '../../../../../services/valueConvertor';
import Input from '../../../../../components/Input';

const SlackIntegration = () => {
    const slackConfiguration = INTEGRATION_LIST[0];

    return (
        <slack-integration is="3xd" className="integration-modal-container">
            {slackConfiguration?.insideBanner}
            <div className="header">
                {slackConfiguration?.icon}
                <div className="details">
                    <p>{slackConfiguration?.name}</p>
                    <>
                        <span>by {slackConfiguration.by}</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate(slackConfiguration.date)} </span>
                    </>
                </div>
            </div>
            <div className="description">
                <p>Description</p>
                <span className="content">
                    There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or
                    randomised words which don't look even slightly believable. Read More
                </span>
            </div>
            <div className="api-details">
                <p className="title">API details</p>
                <div className="api-key">
                    <p>API KEY</p>
                    <span className="desc">
                        There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour
                    </span>
                    <Input
                        placeholder="Insert auth token"
                        type="text"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="purple"
                        borderColorType="none"
                        height="40px"
                        // onBlur={(e) => getStarted && updateFormState('name', e.target.value)}
                        // onChange={(e) => getStarted && updateFormState('name', e.target.value)}
                        // value={getStartedStateRef?.formFieldsCreateStation?.name}
                        // disabled={!allowEdit}
                    />
                </div>
                <div className="channel-id">
                    <p>Channel ID</p>
                    <span className="desc">
                        There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour
                    </span>
                    <Input
                        placeholder="Insert channel id"
                        type="text"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        height="40px"
                        // onBlur={(e) => getStarted && updateFormState('name', e.target.value)}
                        // onChange={(e) => getStarted && updateFormState('name', e.target.value)}
                        // value={getStartedStateRef?.formFieldsCreateStation?.name}
                        // disabled={!allowEdit}
                    />
                </div>
                <div className="notification-option">
                    <p>Notify me when:</p>
                </div>
            </div>
        </slack-integration>
    );
};

export default SlackIntegration;
