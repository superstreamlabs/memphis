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
import ApexChart from './apexChart';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';

const Resources = () => {
    const [resourcesTotal, setResources] = useState([
        { resource: 'CPU', usage: 50, total: 100, units: 'Mb' },
        { resource: 'Mem', usage: 75, total: 100, units: 'Mb' },
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
                {resourcesTotal &&
                    resourcesTotal.map((res) => {
                        return (
                            <div className="resource" key={res.resource}>
                                <ApexChart data={res} className="chart" />
                                <p className="chart-data">{`${res.usage}${res.units}/${res.total}${res.units}`}</p>
                            </div>
                        );
                    })}
            </div>
        </div>
    );
};

export default Resources;
