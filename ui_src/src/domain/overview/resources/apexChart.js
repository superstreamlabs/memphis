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
