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

import './../style.scss';

import { PriorityHighRounded } from '@material-ui/icons';
import { Popover } from 'antd';
import React from 'react';

const remainingPorstPopInnerStyle = { padding: '10px', borderRadius: '4px', boxShadow: '0px 1px 3px rgba(0, 0, 0, 0.12), 0px 23px 44px rgba(176, 183, 195, 0.14)' };

const SysCompInfo = ({ status, components }) => {
    const compList = () => {
        return (
            <div className="comp-list-wrapper">
                {components['unhealthy_components']?.length && (
                    <div className="comp-length unhealthy">
                        <p>Unhealthy</p>
                        <div className="number-wrapper">
                            <span>{components['unhealthy_components']?.length}</span>
                        </div>
                    </div>
                )}
                {components['risky_components']?.length && (
                    <div className="comp-length risky">
                        <p>Risky</p>
                        <div className="number-wrapper">
                            <span>{components['risky_components']?.length}</span>
                        </div>
                    </div>
                )}
                {components['dangerous_components']?.length && (
                    <div className="comp-length dangerous">
                        <p>Dangerous</p>
                        <div className="number-wrapper">
                            <span>{components['dangerous_components']?.length}</span>
                        </div>
                    </div>
                )}
                {components['healthy_components']?.length && (
                    <div className="comp-length healthy">
                        <p>Healthy</p>
                        <div className="number-wrapper">
                            <span>{components['healthy_components']?.length}</span>
                        </div>
                    </div>
                )}
            </div>
        );
    };

    const infoStatus = (background, iconColor) => {
        return (
            <Popover overlayInnerStyle={remainingPorstPopInnerStyle} placement="bottomLeft" content={compList}>
                <div className="sys-components-info" style={{ background: background }}>
                    <div className="error-icon" style={{ background: iconColor }}>
                        <PriorityHighRounded />
                    </div>
                </div>
            </Popover>
        );
    };

    switch (status) {
        case 'risky':
            return infoStatus('#FFF6ED', '#FFA043');
        case 'dangerous':
            return infoStatus('#FFF6E0', '#FFC633');
        case 'unhealthy':
            return infoStatus('#FFEBE6', '#FC3400');
        default:
            return null;
    }
};

export default SysCompInfo;
