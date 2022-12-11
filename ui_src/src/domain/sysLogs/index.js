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

import React, { useEffect, useContext } from 'react';

import Button from '../../components/button';
import { Context } from '../../hooks/store';
import LogsWrapper from './components/logsWrapper';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';

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
                <h1 className="main-header-h1">System Logs</h1>
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
                    onClick={downloadLogs}
                />
            </div>

            <LogsWrapper />
        </div>
    );
};

export default SysLogs;
