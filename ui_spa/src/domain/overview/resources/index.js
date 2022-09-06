// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
                <img src={comingSoonBox} width={40} height={70} />
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
