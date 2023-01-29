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

import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import React, { useEffect, useState, useContext } from 'react';

import { Context } from '../../../hooks/store';
import SelectThroughput from '../../../components/selectThroughput';
import SegmentButton from '../../../components/segmentButton';
import { convertBytes } from '../../../services/valueConvertor';

const axisStyle = {
    fontSize: '12px',
    fontFamily: 'InterSemiBold',
    margin: '0px'
};

const Throughput = () => {
    const [state, dispatch] = useContext(Context);
    const [throughputType, setThroughputType] = useState('write');
    const [selectedComponent, setSelectedComponent] = useState('total');
    const [selectOptions, setSelectOptions] = useState([]);
    const [dataRead, setDataRead] = useState([]);
    const [dataWrite, setDataWrite] = useState([]);
    const [timeDiffIndicator, setTimeDiffIndicator] = useState(false);

    const CustomTooltip = ({ active, payload, label }) => {
        if (active && payload && payload.length) {
            return (
                <div className="custom-tooltip">
                    <p className="throughput-type">
                        {selectedComponent} {throughputType}
                    </p>
                    <p>{`time: ${label}`}</p>
                    <p>
                        {payload[0].dataKey}: {convertBytes(payload[0].value)}/s
                    </p>
                </div>
            );
        }
        return null;
    };

    const getTimeDiff = (before, after) => {
        const diff = Math.abs(after - before);
        const minutes = Math.floor(diff / 1000 / 60);
        return minutes;
    };

    useEffect(() => {
        const current = new Date();
        const time = current.toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        });
        let write = { time: time, date: current };
        let read = { time: time, date: current };
        let components = [];
        state?.monitor_data?.brokers_throughput?.forEach((element) => {
            const elementName = element.name;
            components.push({ name: elementName });
            write[elementName] = element.write;
            read[elementName] = element.read;
        });
        components.sort(function (a, b) {
            if (a.name === 'total') return -1;
            if (b.name === 'total') return 1;
            let nameA = a.name.toUpperCase();
            let nameB = b.name.toUpperCase();
            if (nameA < nameB) {
                return -1;
            }
            if (nameA > nameB) {
                return 1;
            }
            return 0;
        });
        setSelectOptions(components);
        if (!timeDiffIndicator && dataWrite.length > 0) {
            if (getTimeDiff(current, dataWrite[0].date) >= 10) {
                setTimeDiffIndicator(true);
            }
        }
        if (timeDiffIndicator) {
            setDataRead([...dataRead.slice(1), read]);
            setDataWrite([...dataWrite.slice(1), write]);
        } else {
            setDataRead([...dataRead, read]);
            setDataWrite([...dataWrite, write]);
        }
    }, [state?.monitor_data?.brokers_throughput]);

    return (
        <div className="overview-components-wrapper throughput-overview-container">
            <div className="overview-components-header throughput-header">
                <div className="throughput-header-side">
                    <p>Live throughput</p>
                    <SegmentButton options={['write', 'read']} onChange={(e) => setThroughputType(e)} />
                </div>
                <SelectThroughput value={selectedComponent || 'total'} options={selectOptions} onChange={(e) => setSelectedComponent(e)} />
            </div>
            <div className="throughput-chart">
                <ResponsiveContainer>
                    <AreaChart
                        data={throughputType === 'write' ? dataWrite : dataRead}
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
                        <XAxis style={axisStyle} dataKey="time" minTickGap={200} />
                        <YAxis width={85} style={axisStyle} dataKey={selectedComponent} tickFormatter={(tickItem) => `${convertBytes(tickItem, true)}/s`} />
                        <Tooltip content={<CustomTooltip />} />
                        <Area type="monotone" dataKey={selectedComponent} stroke="#6557FF" fill="url(#colorThroughput)" />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default Throughput;
