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

import React, { useState } from 'react';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';
import { RadialBarChart, RadialBar, Legend, ResponsiveContainer, PolarAngleAxis } from 'recharts';
import { PieChart, Pie, Sector, Cell, Label } from 'recharts';
import { Divider } from '@material-ui/core';

const getData = (resource) => {
    return [
        { name: `${resource.resource}total`, value: resource.total - resource.usage, fill: '#F5F5F5' },
        { name: `${resource.resource}used`, value: resource.usage, fill: getColor(resource) }
    ];
};

const getPercentage = (resource) => {
    return resource.total !== 0 ? (resource.usage / resource.total) * 100 : 0;
};

const getColor = (resource) => {
    const percentage = getPercentage(resource);
    if (percentage <= 35) return '#61DFC6';
    else if (percentage < 70) return '#6557FF';
    else return '#FF716E';
};

const Resources = () => {
    const [resourcesTotal, setResources] = useState([
        { resource: 'CPU', usage: 50, total: 100, units: 'Mb' },
        { resource: 'Memory', usage: 75, total: 100, units: 'Mb' },
        { resource: 'Storage', usage: 25, total: 100, units: 'Mb' }
    ]);

    return (
        <div className="overview-wrapper resources-container">
            <p className="overview-components-header">Resources</p>
            <div className="charts-wrapper">
                {resourcesTotal?.length > 0 &&
                    resourcesTotal.map((resource, index) => {
                        return (
                            <>
                                <div className="resource">
                                    <ResponsiveContainer height={'100%'} width={'40%'}>
                                        <PieChart>
                                            <Pie dataKey="value" data={getData(resource)} startAngle={-270} innerRadius={'55%'}>
                                                <Label value={`${getPercentage(resource)}%`} position="center" />
                                            </Pie>
                                        </PieChart>
                                    </ResponsiveContainer>
                                    <div className="resource-data">
                                        <p className="resource-name">{`${resource.resource} Usage`}</p>
                                        <p className="resource-value">
                                            <label className="value">{`${resource.usage}${resource.units} / `}</label>
                                            <label>{`${resource.total}${resource.units}`}</label>
                                        </p>
                                    </div>
                                </div>
                                {index < resourcesTotal.length - 1 && <Divider />}
                            </>
                        );
                    })}
            </div>
        </div>
    );
};

export default Resources;
