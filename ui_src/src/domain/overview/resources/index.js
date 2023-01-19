// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
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

import React, { useState } from 'react';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';
import { PieChart, Pie, Label, ResponsiveContainer } from 'recharts';
import { Divider } from '@material-ui/core';

const getData = (resource) => {
    return [
        { name: `${resource.resource}total`, value: resource.total - resource.usage, fill: 'transparent' },
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
            <div className="coming-soon-wrapper">
                <img src={comingSoonBox} width={40} height={70} alt="comingSoonBox" />
                <p>Coming soon</p>
            </div>
            <p className="overview-components-header">Resources</p>
            <div className="charts-wrapper">
                {resourcesTotal?.length > 0 &&
                    resourcesTotal.map((resource, index) => (
                        <>
                            <div className="resource">
                                <ResponsiveContainer height={'100%'} width={'40%'}>
                                    <PieChart>
                                        <Pie
                                            dataKey="value"
                                            startAngle={-270}
                                            data={[{ name: `total`, value: resource.total, fill: '#F5F5F5' }]}
                                            stroke=""
                                            innerRadius={'60%'}
                                        ></Pie>
                                        <Pie cornerRadius={40} dataKey="value" stroke="" data={getData(resource)} startAngle={-270} innerRadius={'60%'}>
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
                    ))}
            </div>
        </div>
    );
};

export default Resources;
