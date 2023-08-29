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

import { parsingDate } from '../../../../services/valueConvertor';
import { ReactComponent as SourceIcon } from '../../../../assets/images/sourceIcon.svg';
import LogBadge from '../../../../components/logBadge';

const LogPayload = ({ value, onSelected, selectedRow }) => {
    return (
        <div className={selectedRow === value?.message_seq ? 'log-payload log-selected' : 'log-payload'} onClick={() => onSelected(value?.message_seq)}>
            {selectedRow === value?.message_seq && <div className="selected"></div>}
            <p className="title">{value?.data}</p>
            <p className="created-date">{parsingDate(value?.created_at)}</p>
            <div className="log-info">
                <div className="source">
                    <SourceIcon alt="sourceIcon" />
                    <p>{value?.source}</p>
                </div>
                <LogBadge type={value?.type} />
            </div>
        </div>
    );
};

export default LogPayload;
