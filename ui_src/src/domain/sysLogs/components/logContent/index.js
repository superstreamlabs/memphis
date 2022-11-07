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

import LogBadge from '../../../../components/logBadge';
import { capitalizeFirst, cutInfoLog, parsingDate } from '../../../../services/valueConvertor';
import Copy from '../../../../components/copy';

const LogContent = ({ displayedLog }) => {
    return (
        <div className="log-content-wrapper">
            <log-header is="3xd">
                <p>Log details</p>
            </log-header>
            <log-payload is="3xd">
                <div className="log-details">
                    <div className="source">
                        <p className="title">Source</p>
                        <span className="des">{displayedLog?.source && capitalizeFirst(displayedLog?.source)}</span>
                    </div>
                    <div className="type">
                        <p className="title">Type</p>
                        <LogBadge type={displayedLog?.type} />
                    </div>
                    <div className="date">
                        <p className="title">Time</p>
                        <span className="des">{parsingDate(displayedLog?.creation_date)}</span>
                    </div>
                </div>
                <div></div>
            </log-payload>
            <log-content is="3xd">
                <p>{cutInfoLog(displayedLog?.data)}</p>
                <div className="copy-button">
                    <Copy data={cutInfoLog(displayedLog?.data)} />
                </div>
            </log-content>
        </div>
    );
};

export default LogContent;
