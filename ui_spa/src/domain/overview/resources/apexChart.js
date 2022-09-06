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
