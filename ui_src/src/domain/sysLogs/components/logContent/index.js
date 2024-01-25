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

import './style.scss';

import React from 'react';

import LogBadge from 'components/logBadge';
import { parsingDate } from 'services/valueConvertor';
import Copy from 'components/copy';

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
                        <span className="des">{displayedLog?.source}</span>
                    </div>
                    <div className="type">
                        <p className="title">Type</p>
                        <LogBadge type={displayedLog?.type} />
                    </div>
                    <div className="date">
                        <p className="title">Time</p>
                        <span className="des">{parsingDate(displayedLog?.created_at)}</span>
                    </div>
                </div>
                <div></div>
            </log-payload>
            <log-content is="3xd">
                <p>{displayedLog?.data}</p>
                <div className="copy-button">
                    <Copy data={displayedLog?.data} />
                </div>
            </log-content>
        </div>
    );
};

export default LogContent;
