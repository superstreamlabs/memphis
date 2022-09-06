// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useContext, useState } from 'react';
import { Divider } from '@material-ui/core';

import HealthyBadge from '../../../components/healthyBadge';
import { Context } from '../../../hooks/store';
import { Link } from 'react-router-dom';
import pathDomains from '../../../router';

const SysComponents = () => {
    const [state, dispatch] = useContext(Context);

    return (
        <div className="overview-wrapper sys-components-container">
            <span className="overview-components-header">
                System Components
                <span className="actions-side">
                    <div className="hover-section">
                        <div className="action overview">
                            {/* <Link style={{ cursor: 'pointer' }} to={`${pathDomains.sysLogs}`}>
                                <span className="link-row" style={{ width: '100px' }}>
                                    Logs
                                </span>
                            </Link> */}
                        </div>
                    </div>
                </span>
            </span>
            <div className="sys-components sys-components-header">
                <p>Component</p>
                <p>Pods</p>
                <p>Status</p>
            </div>
            {!state?.monitor_data?.system_components && <Divider />}
            <div className="component-list">
                {state?.monitor_data?.system_components &&
                    state?.monitor_data?.system_components?.map((comp, i) => {
                        return (
                            <div style={{ lineHeight: '30px' }} key={`${comp.podName}${i}`}>
                                <Divider />
                                <div className="sys-components">
                                    <p>{comp.component}</p>
                                    <p>
                                        {comp.actual_pods}/{comp.desired_pods}
                                    </p>
                                    <HealthyBadge status={comp.actual_pods / comp.desired_pods} />
                                </div>
                            </div>
                        );
                    })}
            </div>
        </div>
    );
};

export default SysComponents;
