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

import React from 'react';
import ApexCharts from 'apexcharts';
import ReactApexChart from 'react-apexcharts';

export default function ApexChart(props) {
    const series = [(props.data.usage / props.data.total) * 100];

    const options = {
        chart: {
            // height: 350,
            type: 'radialBar'
        },
        fill: {
            type: 'solid',
            colors: ['#5A4FE5']
        },
        plotOptions: {
            radialBar: {
                hollow: {
                    size: 60
                },
                dataLabels: {
                    name: {
                        show: true,
                        fontSize: '12px',
                        fontFamily: undefined,
                        fontWeight: 400,
                        color: '#1D1D1D',
                        offsetY: 5
                    },
                    value: {
                        show: false
                    }
                }
            }
        },
        labels: [props.data.resource]
    };

    return (
        <div className="chart-pie">
            <ReactApexChart options={options} series={series} type="radialBar" width={150} />
        </div>
    );
}
