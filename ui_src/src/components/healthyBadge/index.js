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

import CheckCircleSharpIcon from '@material-ui/icons/CheckCircleSharp';
import ErrorSharpIcon from '@material-ui/icons/ErrorSharp';
import Cancel from '@material-ui/icons/Cancel';
import React from 'react';

const HealthyBadge = ({ status }) => {
    return (
        <div className="healthy-badge-container">
            {status > 0.6 && (
                <div className="healthy">
                    <CheckCircleSharpIcon className="badge-icon" theme="outlined" />
                    <p>Healthy</p>
                </div>
            )}
            {status > 0.3 && status <= 0.6 && (
                <div className="risky">
                    <ErrorSharpIcon className="badge-icon" theme="outlined" />
                    <p>Risky</p>
                </div>
            )}
            {status <= 0.3 && (
                <div className="unhealthy">
                    <Cancel className="badge-icon" theme="outlined" />
                    <p>Unhealthy</p>
                </div>
            )}
        </div>
    );
};

export default HealthyBadge;
