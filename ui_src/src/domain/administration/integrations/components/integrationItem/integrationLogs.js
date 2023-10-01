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

import React, { useState, useEffect } from 'react';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Copy from '../../../../../components/copy';

const IntegrationLogs = ({ integrationName }) => {
    const [logs, setLogs] = useState([]);
    const getIntegrationLogs = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_LOGS}?name=${integrationName}`);
            const logList = data?.map((log) => {
                return (
                    <div>
                        <p style={{ display: 'flex', alignItems: 'center' }}>
                            <lavel style={{ fontSize: 6 }}>{'\u2B24'} </lavel>
                            {`${log?.created_at} ${log?.message}`}
                        </p>
                    </div>
                );
            });
            setLogs(logList);
        } catch (err) {
            return;
        }
    };
    useEffect(() => {
        getIntegrationLogs();
    }, []);

    return (
        <div className="integration-log-content-wrapper">
            <log-content is="3xd">
                <>{logs}</>
                <div className="copy-button">
                    <Copy data={logs} />
                </div>
            </log-content>
        </div>
    );
};

export default IntegrationLogs;
