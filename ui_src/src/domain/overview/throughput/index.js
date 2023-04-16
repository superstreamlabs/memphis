// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useState, useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { Line } from 'react-chartjs-2';
import { Chart } from 'chart.js';
import 'chartjs-plugin-streaming';
import moment from 'moment';
import { convertBytes, convertNumberToK } from '../../../services/valueConvertor';
import SelectThroughput from '../../../components/selectThroughput';
import SegmentButton from '../../../components/segmentButton';
import Loader from '../../../components/loader';
import DataNotFound from '../../../assets/images/dataNotFound.svg';
import pathDomains from '../../../router';

import { Context } from '../../../hooks/store';
import { PauseRounded, PlayArrowRounded } from '@material-ui/icons';
import RadioButton from '../../../components/radioButton';

const throughputTypes = [
    {
        id: 1,
        value: 'Bytes per second',
        label: 'Bytes per second'
    },
    {
        id: 2,
        value: 'Messages per second',
        label: 'Messages per second'
    }
];

const gradient = (chartInstance) => {
    const ctx = chartInstance.chart.ctx;
    const gradient = ctx.createLinearGradient(0, 0, 0, 300);
    gradient.addColorStop(0, 'rgba(101, 87, 255, 0.1)');
    gradient.addColorStop(0.45, 'rgba(226, 223, 255, 0.1)');
    gradient.addColorStop(0.75, 'rgba(241, 240, 255, 0.1)');
    gradient.addColorStop(1, 'rgba(255, 255, 255, 0.1)');
    return gradient;
};

const getDataset = (dsName, readWrite, hidden) => {
    return {
        label: `${readWrite} ${dsName}`,
        borderColor: '#6557FF',
        borderWidth: 1,
        backgroundColor: gradient,
        fill: true,
        lineTension: 0,
        data: [],
        hidden: hidden,
        pointRadius: 0
    };
};

function Throughput() {
    const [state, dispatch] = useContext(Context);
    const [throughputAction, setThroughputAction] = useState('write');
    const [selectedComponent, setSelectedComponent] = useState('total');
    const [selectOptions, setSelectOptions] = useState([]);
    const [dataSamples, setDataSamples] = useState({});
    const [data, setData] = useState({});
    const [loading, setLoading] = useState(false);
    const [stop, setstop] = useState(false);
    // const [socketFailIndicator, setSocketFailIndicator] = useState(false);
    const history = useHistory();
    const [throughputType, setThroughputType] = useState(throughputTypes[0].value);
    const [optionChart, setOptionsChart] = useState({});

    // Chart.plugins.register({
    //     afterDraw: function (chart) {
    //         if (data?.datasets?.length == 0) {
    //             !socketFailIndicator && setSocketFailIndicator(true);
    //         } else socketFailIndicator && setSocketFailIndicator(false);
    //     }
    // });

    const yAxesOptions = [
        {
            gridLines: {
                display: true,
                borderDash: [3, 3]
            },
            ticks: {
                beginAtZero: true,
                callback: function (value) {
                    return throughputType === throughputTypes[0].value ? `${convertBytes(value, true)}/s` : `${convertNumberToK(value)} Msgs/s`;
                },
                maxTicksLimit: 5,
                min: 0,
                suggestedMax: 45 * 1024
            }
        }
    ];

    const initiateDataState = () => {
        let dataSets = [];
        selectOptions.forEach((selectOption, i) => {
            dataSets.push(getDataset(selectOption.name, 'bytes_write', throughputAction === 'write' && throughputType === throughputTypes[0].value ? false : true));
            dataSets.push(getDataset(selectOption.name, 'bytes_read', throughputAction !== 'write' && throughputType === throughputTypes[0].value ? false : true));
            dataSets.push(getDataset(selectOption.name, 'msgs_write', throughputAction === 'write' && throughputType === throughputTypes[1].value ? false : true));
            dataSets.push(getDataset(selectOption.name, 'msgs_read', throughputAction !== 'write' && throughputType === throughputTypes[1].value ? false : true));
        });
        setData({ datasets: dataSets });
    };

    useEffect(() => {
        setOptionsChart(options);
    }, [throughputType]);

    useEffect(() => {
        if (data?.datasets?.length === 0 && selectOptions?.length > 0) initiateDataState();
    }, [selectOptions]);

    useEffect(() => {
        const foundItemIndex = selectOptions?.findIndex((item) => item.name === selectedComponent);
        if (foundItemIndex === -1) return;
        setLoader();
        for (let i = 0; i < selectOptions?.length; i++) {
            if (i === foundItemIndex) {
                data.datasets[4 * i].hidden = throughputAction === 'write' ? (throughputType === throughputTypes[0].value ? false : true) : true;
                data.datasets[4 * i + 1].hidden = throughputAction !== 'write' ? (throughputType === throughputTypes[0].value ? false : true) : true;
                data.datasets[4 * i + 2].hidden = throughputAction === 'write' ? (throughputType === throughputTypes[1].value ? false : true) : true;
                data.datasets[4 * i + 3].hidden = throughputAction !== 'write' ? (throughputType === throughputTypes[1].value ? false : true) : true;
            } else {
                data.datasets[4 * i].hidden = true;
                data.datasets[4 * i + 1].hidden = true;
                data.datasets[4 * i + 2].hidden = true;
                data.datasets[4 * i + 3].hidden = true;
            }
        }
    }, [throughputAction, selectedComponent, throughputType]);

    const getSelectComponentList = () => {
        const components = state?.monitor_data?.brokers_throughput
            ?.map((element) => {
                return { name: element.name };
            })
            .sort(function (a, b) {
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
    };

    useEffect(() => {
        if (Object.keys(dataSamples).length > 0) {
            selectOptions.length === 0 && getSelectComponentList();
            state?.monitor_data?.brokers_throughput?.forEach((component) => {
                let updatedDataSamples = { ...dataSamples };
                updatedDataSamples[component.name].bytes_write = [...updatedDataSamples[component.name]?.bytes_write, ...component.bytes_write];
                updatedDataSamples[component.name].bytes_read = [...updatedDataSamples[component.name]?.bytes_read, ...component.bytes_read];
                updatedDataSamples[component.name].msgs_write = [...updatedDataSamples[component.name]?.msgs_write, ...component.msgs_write];
                updatedDataSamples[component.name].msgs_read = [...updatedDataSamples[component.name]?.msgs_read, ...component.msgs_read];

                setDataSamples(updatedDataSamples);
            });
        } else {
            let sampleObject = {};
            state?.monitor_data?.brokers_throughput?.forEach((component) => {
                const componentName = component.name;
                sampleObject[componentName] = {
                    bytes_read: component.bytes_read,
                    bytes_write: component.bytes_write,
                    msgs_read: component.msgs_read,
                    msgs_write: component.msgs_write
                };
            });
            setDataSamples(sampleObject);
        }
    }, [state?.monitor_data?.brokers_throughput]);

    const getValue = (type, select) => {
        let updatedDataSamples = { ...dataSamples };
        let value;
        if (throughputType === throughputTypes[0].value) {
            if (throughputAction === 'write') {
                value = dataSamples[select]?.bytes_write[0]?.write;
                updatedDataSamples[select]?.bytes_write.shift();
            } else {
                value = dataSamples[select]?.bytes_read[0]?.read;
                updatedDataSamples[select]?.bytes_read.shift();
            }
        } else {
            if (throughputAction === 'write') {
                value = dataSamples[select]?.msgs_write[0]?.write;
                updatedDataSamples[select]?.msgs_write.shift();
            } else {
                value = dataSamples[select]?.msgs_read[0]?.read;
                updatedDataSamples[select]?.msgs_read.shift();
            }
        }

        setDataSamples(updatedDataSamples);
        return value;
    };

    const updateData = (chart) => {
        for (let i = 0; i < selectOptions?.length; i++) {
            chart.data?.datasets[4 * i]?.data?.push({
                x: moment(),
                y: getValue('bytes_write', selectOptions[i].name)
            });
            chart.data?.datasets[4 * i + 1]?.data?.push({
                x: moment(),
                y: getValue('bytes_read', selectOptions[i].name)
            });
            chart.data?.datasets[4 * i + 2]?.data?.push({
                x: moment(),
                y: getValue('msgs_write', selectOptions[i].name)
            });
            chart.data?.datasets[4 * i + 3]?.data?.push({
                x: moment(),
                y: getValue('msgs_read', selectOptions[i].name)
            });
        }
    };

    const setLoader = () => {
        setLoading(true);
        setTimeout(function () {
            setLoading(false);
        }, 1000);
    };

    const options = {
        plugins: {
            streaming: {
                pause: stop,
                frameRate: 5
            }
        },
        animation: false,
        legend: { display: false },
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        hover: { mode: 'top', intersect: true },
        tooltips: {
            mode: 'index',
            intersect: false,
            displayColors: false,
            callbacks: {
                title: () => {
                    return `${selectedComponent.charAt(0).toUpperCase() + selectedComponent.slice(1)} - ${throughputAction}`;
                },
                label: (tooltipItem) => {
                    return `${tooltipItem.label}`;
                },
                afterLabel: (tooltipItem) => {
                    return throughputType === throughputTypes[0].value
                        ? `Throughput: ${convertBytes(tooltipItem.yLabel, true)}/s`
                        : `Throughput: ${convertNumberToK(tooltipItem.yLabel)} Msgs/s`;
                }
            },
            backgroundColor: 'rgba(0, 0, 0, 0.75)',
            cornerRadius: 3,
            titleFontSize: 14,
            titleFontFamily: 'InterMedium',
            titleFontStyle: 'normal',
            bodyFontFamily: 'InterMedium',
            bodyFontSize: 14,
            bodySpacing: 6
        },

        elements: { line: { tension: 0.5 }, point: { borderWidth: 1, radius: 1, backgroundColor: 'rgba(0,0,0,0)' } },
        scales: {
            xAxes: [
                {
                    type: 'realtime',
                    distribution: 'linear',
                    realtime: {
                        refresh: 1000,
                        onRefresh: function (chart) {
                            if (data?.datasets?.length !== 0) {
                                updateData(chart);
                            }
                        },
                        delay: 1000,
                        duration: 300000,
                        time: { displayFormat: 'h:mm:ss' }
                    },
                    gridLines: { display: true },
                    ticks: { stepSize: 4, autoSkip: false }
                }
            ],
            yAxes: yAxesOptions
        }
    };

    return (
        <div className="overview-components-wrapper throughput-overview-container">
            <div className="overview-components-header throughput-header">
                <div className="throughput-header-side">
                    <p>Throughput</p>
                    <SegmentButton options={['write', 'read']} onChange={(e) => setThroughputAction(e)} />
                </div>
                <div className="throughput-actions">
                    <div className="play-pause-btn" onClick={() => setstop(!stop)}>
                        {stop ? <PlayArrowRounded /> : <PauseRounded />}
                    </div>
                    <SelectThroughput value={selectedComponent || 'total'} options={selectOptions} onChange={(e) => setSelectedComponent(e)} />
                </div>
            </div>
            <div className="external-monitoring">
                <label>For historical monitoring, please connect Memphis to an external </label>
                <label className="link-to-integrations" onClick={() => history.push(`${pathDomains.administration}/integrations`)}>
                    monitoring tool
                </label>
            </div>
            <div className="throughput-chart">
                {loading && <Loader />}
                {/* {socketFailIndicator && (
                    <div className="failed-socket">
                        <img src={DataNotFound} alt="Data not found" />
                        <p className="title">No data found</p>
                    </div>
                )} */}
                <Line id="test" data={data} options={optionChart} />
            </div>
            <div className="wrapper-radio-button">
                <RadioButton
                    vertical={false}
                    options={throughputTypes}
                    radioValue={throughputType}
                    onChange={(e) => setThroughputType(e.target.value)}
                />
            </div>
        </div>
    );
}

export default Throughput;
