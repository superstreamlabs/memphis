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

import React, { useEffect, useContext } from 'react';

import Button from 'components/button';
import { Context } from 'hooks/store';
import LogsWrapper from './components/logsWrapper';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';

const SysLogs = () => {
    const [state, dispatch] = useContext(Context);

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'logs' });
    }, []);

    const downloadLogs = async () => {
        try {
            const response = await httpRequest('GET', `${ApiEndpoints.DOWNLOAD_SYS_LOGS}`, {}, {}, {}, true, 0);
            const url = window.URL.createObjectURL(new Blob([response]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', `logs.log`);
            document.body.appendChild(link);
            link.click();
        } catch (error) {}
    };

    return (
        <div className="logs-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <h1 className="main-header-h1">System Logs</h1>
                    <span className="memphis-label">Memphis platform system logs.</span>
                </div>
                <Button
                    className="modal-btn"
                    width="160px"
                    height="36px"
                    placeholder="Download logs"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="14px"
                    fontWeight="600"
                    aria-haspopup="true"
                    boxShadowStyle="float"
                    onClick={downloadLogs}
                />
            </div>
            <LogsWrapper />
        </div>
    );
};

export default SysLogs;
