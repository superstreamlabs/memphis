// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { Segmented } from 'antd';
import React, { useState } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Legend, Line } from 'recharts';
import ThroughputInterval from './throughputInterval';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';
import SysContainer from '../../../assets/images/sysContainer.svg';
import SelectComponent from '../../../components/select';
import SelectThroughput from '../../../components/selectThroughput';

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
        throughput1: 4000,
        throughput2: 5000,
        throughput3: 4000
    },
    {
        name: '01:00',
        throughput1: 3000,
        throughput2: 5000,
        throughput3: 4000
    },
    {
        name: '02:00',
        throughput1: 2000,
        throughput2: 5000,
        throughput3: 4000
    },
    {
        name: '03:00',
        throughput1: 2780,
        throughput2: 1000,
        throughput3: 4000
    },
    {
        name: '04:00',
        throughput1: 1890,
        throughput2: 2000,
        throughput3: 4000
    },
    {
        name: '05:00',
        throughput1: 2390,
        throughput2: 3000,
        throughput3: 4000
    },
    {
        name: '06:00',
        throughput1: 3490,
        throughput2: 1000,
        throughput3: 4000
    },
    {
        name: '07:00',
        throughput1: 4000,
        throughput2: 2000,
        throughput3: 4000
    },
    {
        name: '08:00',
        throughput1: 3000,
        throughput2: 5000,
        throughput3: 4000
    },
    {
        name: '09:00',
        throughput1: 2000,
        throughput2: 2000,
        throughput3: 4000
    },
    {
        name: '10:00',
        throughput1: 2780,
        throughput2: 2000,
        throughput3: 4000
    },
    {
        name: '11:00',
        throughput1: 1890,
        throughput2: 4000,
        throughput3: 4000
    },
    {
        name: '12:00',
        throughput1: 2390,
        throughput2: 7000,
        throughput3: 4000
    },
    {
        name: '13:00',
        throughput1: 4000,
        throughput2: 6000,
        throughput3: 4000
    }
];

// let data = [
//     {
//         name: '00:00',
//         throughput: 4000
//     },
//     {
//         name: '01:00',
//         throughput: 3000
//     },
//     {
//         name: '02:00',
//         throughput: 2000
//     },
//     {
//         name: '03:00',
//         throughput: 2780
//     },
//     {
//         name: '04:00',
//         throughput: 1890
//     },
//     {
//         name: '05:00',
//         throughput: 2390
//     },
//     {
//         name: '06:00',
//         throughput: 3490
//     },
//     {
//         name: '07:00',
//         throughput: 4000
//     },
//     {
//         name: '08:00',
//         throughput: 3000
//     },
//     {
//         name: '09:00',
//         throughput: 2000
//     },
//     {
//         name: '10:00',
//         throughput: 2780
//     },
//     {
//         name: '11:00',
//         throughput: 1890
//     },
//     {
//         name: '12:00',
//         throughput: 2390
//     },
//     {
//         name: '13:00',
//         throughput: 4000
//     }
// ];

const data2 = [
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
    const [throughputType, setThroughputType] = useState('Write');
    const [selectedComponent, setSelectedComponent] = useState('broker-1');
    return (
        <div className="overview-components-wrapper throughput-overview-container">
            {/* <div className="coming-soon-wrapper">
                <img src={comingSoonBox} width={40} height={70} alt="comingSoonBox" />
                <p>Coming soon</p>
            </div> */}
            <div className="overview-components-header throughput-header">
                <div className="throughput-header-side">
                    <p> Throughput</p>
                    <Segmented options={['Write', 'Read']} onChange={(e) => setThroughputType(e)} />
                </div>
                <SelectThroughput value={selectedComponent} options={[{ name: 'broker-1' }, { name: 'broker-2' }]} onChange={(e) => setSelectedComponent(e)} />
                {/* <ThroughputInterval /> */}
            </div>
            <div className="throughput-chart">
                <ResponsiveContainer>
                    {/* <LineChart
                        data={data}
                        margin={{
                            top: 30,
                            right: 0,
                            left: 0,
                            bottom: 20
                        }}
                    >
                        <CartesianGrid strokeDasharray="6 3" stroke="#f5f5f5" />
                        <XAxis style={axisStyle} dataKey="name" />
                        <YAxis style={axisStyle} />
                        <Tooltip content={<CustomTooltip />} />
                        <Line type="monotone" dataKey="throughput1" stroke="#8884d8" activeDot={{ r: 8 }} />
                        <Line type="monotone" dataKey="throughput2" stroke="#82ca9d" />
                        <Line type="monotone" dataKey="throughput3" stroke="#61DFC6" />
                        
                    </LineChart> */}
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
                        <Area type="monotone" dataKey="throughput1" stroke="#6557FF" fill="url(#colorThroughput)" />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default Throughput;
