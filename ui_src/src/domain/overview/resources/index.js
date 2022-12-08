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

const data = [
    {
        name: '30-34',
        uv: 31.47,
        pv: 2400,
        fill: '#ffffff'
    },

    {
        name: '30-34',
        uv: 15.69,
        pv: 1398,
        fill: '#8dd1e1'
    }
];

const style = {
    top: '50%',
    right: 0,
    transform: 'translate(0, -50%)',
    lineHeight: '24px'
};

const Resources = () => {
    const [resourcesTotal, setResources] = useState([
        { resource: 'CPU', usage: 50, total: 100, units: 'Mb' },
        { resource: 'Mem', usage: 75, total: 100, units: 'Mb' },
        { resource: 'Storage', usage: 25, total: 100, units: 'Mb' }
    ]);

    return (
        <div className="overview-wrapper resources-container">
            {/* <div className="coming-soon-wrapper">
                <img src={comingSoonBox} width={40} height={70} alt="comingSoonBox" />
                <p>Coming soon</p>
            </div> */}
            <p className="overview-components-header">Resources</p>
            <div className="charts-wrapper">
                {
                    <ResponsiveContainer width="100%" height="100%">
                        <RadialBarChart cx="50%" cy="50%" innerRadius="10%" outerRadius="80%" barSize={10} data={data}>
                            <RadialBar cornerRadius={50} minAngle={15} background dataKey="uv" />
                            {/* <Legend iconSize={10} layout="vertical" verticalAlign="middle" wrapperStyle={style} /> */}
                        </RadialBarChart>
                    </ResponsiveContainer>
                }
                {/* {resourcesTotal &&
                    resourcesTotal.map((res) => {
                        return (
                            <PieChart width={730} height={250}>
                                <Pie data={data01} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={50} fill="#8884d8" />
                                <Pie data={data02} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={60} outerRadius={80} fill="#82ca9d" label />
                            </PieChart>
                            // <div className="resource" key={res.resource}>
                            //     <p className="chart-data">{`${res.usage}${res.units}/${res.total}${res.units}`}</p>
                            // </div>
                        );
                    })} */}
            </div>
        </div>
    );
};

export default Resources;
