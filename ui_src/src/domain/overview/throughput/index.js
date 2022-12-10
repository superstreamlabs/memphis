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

import { Segmented } from 'antd';
import React, { useState } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import DatePickerComponent from '../../../components/datePicker';
import ThroughputInterval from './throughputInterval';
import { getBackgroundColor, getFontColor } from '../../../utils/styleTemplates';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';
import { keys } from '@material-ui/core/styles/createBreakpoints';

const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
        return (
            <div className="custom-tooltip">
                <p>{`Time: ${label}`}</p>
                <p style={{ textTransform: 'capitalize' }}>
                    {payload[0].dataKey}: {payload[0].value}
                </p>
            </div>
        );
    }

    return null;
};

const axisStyle = {
    fontSize: '12px',
    fontFamily: 'InterSemiBold',
    margin: '0px'
};

const data = [
    {
        name: '00:00',
        throughput: 4000
    },
    {
        name: '01:00',
        throughput: 3000
    },
    {
        name: '02:00',
        throughput: 2000
    },
    {
        name: '03:00',
        throughput: 2780
    },
    {
        name: '04:00',
        throughput: 1890
    },
    {
        name: '05:00',
        throughput: 2390
    },
    {
        name: '06:00',
        throughput: 3490
    },
    {
        name: '07:00',
        throughput: 4000
    },
    {
        name: '08:00',
        throughput: 3000
    },
    {
        name: '09:00',
        throughput: 2000
    },
    {
        name: '10:00',
        throughput: 2780
    },
    {
        name: '11:00',
        throughput: 1890
    },
    {
        name: '12:00',
        throughput: 2390
    },
    {
        name: '13:00',
        throughput: 4000
    },
    {
        name: '14:00',
        throughput: 3000
    },
    {
        name: '15:00',
        throughput: 2000
    },
    {
        name: '16:00',
        throughput: 2780
    },
    {
        name: '17:00',
        throughput: 1890
    },
    {
        name: '18:00',
        throughput: 2390
    },
    {
        name: '19:00',
        throughput: 3490
    },
    {
        name: '20:00',
        throughput: 4000
    },
    {
        name: '21:00',
        throughput: 3000
    },
    {
        name: '22:00',
        throughput: 2000
    },
    {
        name: '23:00',
        throughput: 2780
    }
];

const Throughput = () => {
    const [throughputType, setThroughputType] = useState('consumers');
    return (
        <div className="overview-wrapper throughput-overview-container">
            <div className="throughput-header">
                <div className="throughput-header-side">
                    <p className="overview-components-header">Throughput</p>
                    <Segmented options={['Producers', 'Consumers']} onChange={(e) => setThroughputType(e)} />
                </div>
                <ThroughputInterval />

                {/* <DatePickerComponent /> */}
            </div>

            <div className="throughput-chart">
                <ResponsiveContainer>
                    <AreaChart
                        data={data}
                        margin={{
                            top: 30,
                            right: 0,
                            left: 0,
                            bottom: 20
                        }}
                    >
                        <defs>
                            <linearGradient id="colorThroughput" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="2%" stopColor="#6557FF" stopOpacity={0.4} />
                                <stop offset="95%" stopColor="#6557FF" stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="6 3" stroke="#f5f5f5" />
                        <XAxis style={axisStyle} dataKey="name" />
                        <YAxis style={axisStyle} />
                        <Tooltip content={<CustomTooltip />} />
                        <Area type="monotone" dataKey="throughput" stroke="#6557FF" fill="url(#colorThroughput)" />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default Throughput;
