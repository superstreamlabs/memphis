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
import { PieChart, Pie } from 'recharts';
import OverflowTip from '../../../components/tooltip/overflowtip';
import { Add } from '@material-ui/icons';
import { Popover, Divider } from 'antd';
import ComponentIcon from '../../../assets/images/componentIcon.svg';
import HealthyBadge from '../../../components/healthyBadge';

const remainingPorstPopInnerStyle = { padding: '10px', borderRadius: '12px', border: '1px solid #f0f0f0' };

const Component = ({ comp, i }) => {
    const getData = (comp) => {
        let data = [];
        if (comp?.actual_pods > 0) {
            for (let i = 0; i < comp?.actual_pods; i++) data.push({ name: `actual${i}`, value: 1, fill: '#6557FF' });
        }
        if (comp?.desired_pods > comp?.actual_pods) {
            for (let i = 0; i < comp?.desired_pods - comp?.actual_pods; i++) data.push({ name: `desired${i}`, value: 1, fill: '#EBEAED' });
        }
        if (comp?.desired_pods === 0 && comp?.actual_pods === 0) data.push({ name: `desired${i}`, value: 1, fill: '#EBEAED' });
        return data;
    };

    return (
        <div className="sys-components-container" key={`${comp?.podName}${i}`}>
            <img src={ComponentIcon} className="component-img" alt="ComponentIcon" width="18" height="18" />
            <div className="component">
                <div className="sys-components">
                    <OverflowTip text={comp?.name}>
                        <p className="component-name">{comp?.name}</p>
                    </OverflowTip>
                    <div className="pie-status-component">
                        <HealthyBadge status={comp?.status} />
                        <div className="pie-status">
                            <PieChart height={30} width={30}>
                                <Pie dataKey="value" data={getData(comp)} startAngle={-270}></Pie>
                            </PieChart>
                            <p>
                                {comp?.actual_pods}/{comp?.desired_pods}
                            </p>
                        </div>
                    </div>
                </div>
                <div className="pods-container">
                    {comp?.hosts?.length > 0 && (
                        <>
                            <div className="hosts">
                                <label className="comp-label">Hosts</label>
                                <OverflowTip maxWidth="9vw" text={comp?.hosts[0]}>
                                    <label className="value">{comp?.hosts[0]}</label>
                                </OverflowTip>
                                {comp?.hosts?.length > 1 && (
                                    <Popover
                                        overlayInnerStyle={remainingPorstPopInnerStyle}
                                        placement="bottomLeft"
                                        content={comp?.hosts?.slice(1)?.map((host) => {
                                            return <p className="comp-plus-popover">{host}</p>;
                                        })}
                                    >
                                        <div className="plus-comp">
                                            <Add className="add" />
                                            <p>{comp?.hosts?.length - 1}</p>
                                        </div>
                                    </Popover>
                                )}
                            </div>
                            <Divider type="vertical" />
                        </>
                    )}
                    <div className="ports">
                        <label className="comp-label">Ports</label>
                        <label className="value">{comp?.ports.length > 0 ? comp?.ports[0] : 'None'}</label>
                        {comp?.ports?.length > 1 && (
                            <Popover
                                overlayInnerStyle={remainingPorstPopInnerStyle}
                                placement="bottomLeft"
                                content={comp?.ports?.slice(1)?.map((port) => {
                                    return <p className="comp-plus-popover">{port}</p>;
                                })}
                            >
                                <div className="plus-comp">
                                    <Add className="add" />
                                    <p>{comp?.ports?.length - 1}</p>
                                </div>
                            </Popover>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Component;
