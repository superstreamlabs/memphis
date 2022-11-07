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

const LogBadge = ({ type }) => {
    return (
        <div className="log-badge-container">
            {type && type === 'info' && (
                <div className="info">
                    <p>Info</p>
                </div>
            )}
            {type && type === 'warn' && (
                <div className="warn">
                    <p>Warn</p>
                </div>
            )}
            {type && type === 'err' && (
                <div className="error">
                    <p>Error</p>
                </div>
            )}
        </div>
    );
};

export default LogBadge;
