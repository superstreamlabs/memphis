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
import { ReactComponent as SysContainerIcon } from 'assets/images/sysContainer.svg';
import { ReactComponent as ErrorIndicationIcon } from 'assets/images/errorindication.svg';
import TooltipComponent from 'components/tooltip/tooltip';
import { convertBytes } from 'services/valueConvertor';

const SysContainers = ({ component, k8sEnv, metricsEnabled, index }) => {
    const getColor = (percentage) => {
        if (percentage <= 33) return '#2ED47A';
        else if (percentage < 66) return '#FFC633';
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
        const details = (
            <div className="system-container-item">
                <p className="item-name">{name}</p>
                <p className="item-usage">{item?.percentage}%</p>
                <Progress percent={item?.percentage} showInfo={false} strokeColor={getColor(item?.percentage)} trailColor="#D9D9D9" size="small" />
            </div>
        );
        return !component.healthy ? (
            <>{details}</>
        ) : !metricsEnabled ? (
            <>{details}</>
        ) : (
            <TooltipComponent text={() => getTooltipData(item, name)}>{details}</TooltipComponent>
        );
    };
    return (
        <div className="system-container">
            {(!component.healthy || !metricsEnabled) && (
                <div className="warn-msg">
                    <div className="msg-wrapper">
                        <ErrorIndicationIcon />
                        {!component.healthy ? (
                            k8sEnv ? (
                                <p>Pod {index + 1} is down</p>
                            ) : (
                                <p>Container is down</p>
                            )
                        ) : (
                            <p>
                                No metrics server found.&nbsp;
                                <a className="learn-more" href="https://docs.memphis.dev/memphis/dashboard-gui/overview#fix-no-metrics-server-found" target="_blank">
                                    Learn more
                                </a>
                            </p>
                        )}
                    </div>
                </div>
            )}
            <div className={!component.healthy ? 'blury' : !metricsEnabled ? 'blury' : null}>
                <div className="system-container-header">
                    <SysContainerIcon alt="syscontainer" width={15} height={15} />
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
        </div>
    );
};

export default SysContainers;
