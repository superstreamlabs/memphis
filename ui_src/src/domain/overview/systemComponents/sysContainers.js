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
import { Divider } from 'antd';
import { Progress } from 'antd';
import SysContainer from '../../../assets/images/sysContainer.svg';
import TooltipComponent from '../../../components/tooltip/tooltip';
import { convertBytes } from '../../../services/valueConvertor';

const SysContainers = ({ component, k8sEnv, index }) => {
    const getColor = (percentage) => {
        if (percentage <= 33) return '#2ED47A';
        else if (percentage < 66) return '#4A3AFF';
        else return '#FF716E';
    };

    const getTooltipData = (item, name) => {
        return (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', textTransform: 'capitalize' }}>
                <label>current: {name === 'CPU' ? `${item?.current} CPU` : `${convertBytes(item?.current)}`}</label>
                <label>total: {name === 'CPU' ? `${item?.total} CPU` : `${convertBytes(item?.total)}`}</label>
            </div>
        );
    };
    const getContainerItem = (item, name) => {
        return (
            <TooltipComponent text={() => getTooltipData(item, name)}>
                <div className="system-container-item">
                    <p className="item-name">{name}</p>
                    <p className="item-usage">{item?.percentage}%</p>
                    <Progress percent={item?.percentage} showInfo={false} strokeColor={getColor(item?.percentage)} trailColor="#D9D9D9" size="small" />
                </div>
            </TooltipComponent>
        );
    };
    return (
        <div className="system-container" style={{ opacity: component.healthy ? 1 : 0.3 }}>
            <div className="system-container-header">
                <img src={SysContainer} alt="SysContainer" width="15" height="15" />
                <div className="cont-tls">
                    <p>{component?.name}</p>
                    <label>{k8sEnv ? `POD ${index + 1}` : `CONTAINER`}</label>
                </div>
            </div>

            <div className="system-container-body">
                {getContainerItem(component?.cpu, 'CPU')}
                <Divider type="vertical" />
                {getContainerItem(component?.memory, 'Memory')}
                <Divider type="vertical" />
                {getContainerItem(component?.storage, 'Storage')}
            </div>
        </div>
    );
};

export default SysContainers;
