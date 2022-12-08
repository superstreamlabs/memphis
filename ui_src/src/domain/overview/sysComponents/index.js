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

import React, { useContext, useState } from 'react';
import { Divider } from '@material-ui/core';

import HealthyBadge from '../../../components/healthyBadge';
import { Context } from '../../../hooks/store';
import { Link } from 'react-router-dom';
import pathDomains from '../../../router';
import { ResponsiveContainer, PieChart, Pie, Legend, Cell } from 'recharts';
// import { PieChart, Pie, Sector, Cell, ResponsiveContainer } from 'recharts';

const SysComponents = () => {
    const [state, dispatch] = useContext(Context);

    const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];
    const getData = (comp) => {
        console.log(comp);
        return [
            { name: 'actual', value: comp.actual_pods, color: 'red' },
            { name: 'desired', value: comp.desired_pods, color: 'yellow' }
        ];
    };
    return (
        <div className="overview-wrapper sys-components-container">
            <span className="overview-components-header">System components</span>
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
                                    <div style={{ display: 'flex', alignItems: 'center' }}>
                                        <PieChart height={35} width={35}>
                                            <Pie
                                                dataKey="value"
                                                data={getData(comp)}
                                                //     [
                                                //     { name: 'Group A', value: 400 },
                                                //     { name: 'Group B', value: 300 },
                                                //     { name: 'Group C', value: 300 },
                                                //     { name: 'Group D', value: 200 }
                                                // ]}
                                            >
                                                {/* {' '}
                                                {data.map((entry, index) => (
                                                    // <Cell fill={COLORS[index % COLORS.length]} />
                                                ))} */}
                                            </Pie>
                                        </PieChart>
                                        <p>
                                            {comp.actual_pods}/{comp.desired_pods}
                                        </p>
                                    </div>
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
