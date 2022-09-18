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

import React, { useContext, useEffect, useState } from 'react';
import { Space } from 'antd';

import { numberWithCommas } from '../../../../services/valueConvertor';
import waitingProducer from '../../../../assets/images/waitingProducer.svg';
import waitingConsumer from '../../../../assets/images/waitingConsumer.svg';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import Modal from '../../../../components/modal';
import StatusIndication from '../../../../components/indication';
import CustomCollapse from '../components/customCollapse';
import MultiCollapse from '../components/multiCollapse';
import { StationStoreContext } from '../..';
import Button from '../../../../components/button';
import SdkExample from '../../sdkExsample';

const ProduceConsumList = (props) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [selectedRowIndex, setSelectedRowIndex] = useState(0);
    const [producersList, setProducersList] = useState([]);
    const [cgsList, setCgsList] = useState([]);
    const [producerDetails, setProducerDetails] = useState([]);
    const [cgDetails, setCgDetails] = useState([]);
    const [openCreateProducer, setOpenCreateProducer] = useState(false);
    const [openCreateConsumer, setOpenCreateConsumer] = useState(false);

    useEffect(() => {
        if (props.producer) {
            let result = concatFunction('producer', stationState?.stationSocketData);
            setProducersList(result);
        } else {
            let result = concatFunction('cgs', stationState?.stationSocketData);
            setCgsList(result);
        }
    }, [stationState?.stationSocketData]);

    useEffect(() => {
        arrangeData('producer', 0);
        arrangeData('cgs', 0);
    }, [producersList, cgsList]);

    const concatFunction = (type, data) => {
        let connected = [];
        let deleted = [];
        let disconnected = [];
        let concatArrays = [];
        if (type === 'producer') {
            connected = data?.connected_producers || [];
            deleted = data?.deleted_producers || [];
            disconnected = data?.disconnected_producers || [];
            concatArrays = connected.concat(disconnected);
            concatArrays = concatArrays.concat(deleted);
            return concatArrays;
        } else if (type === 'cgs') {
            connected = data?.connected_cgs || [];
            disconnected = data?.disconnected_cgs || [];
            deleted = data?.deleted_cgs || [];
            concatArrays = connected.concat(disconnected);
            concatArrays = concatArrays.concat(deleted);
            return concatArrays;
        } else {
            connected = data?.connected_consumers || [];
            disconnected = data?.disconnected_consumers || [];
            deleted = data?.deleted_consumers || [];
            concatArrays = connected.concat(disconnected);
            concatArrays = concatArrays.concat(deleted);
            return concatArrays;
        }
    };

    const onSelectedRow = (rowIndex, type) => {
        setSelectedRowIndex(rowIndex);
        arrangeData(type, rowIndex);
    };

    const arrangeData = (type, rowIndex) => {
        if (type === 'producer') {
            let details = [
                {
                    name: 'Name',
                    value: producersList[rowIndex]?.name
                },
                {
                    name: 'User',
                    value: producersList[rowIndex]?.created_by_user
                },
                {
                    name: 'IP',
                    value: producersList[rowIndex]?.client_address
                }
            ];
            setProducerDetails(details);
        } else {
            let concatAllConsumers = concatFunction('consumers', cgsList[rowIndex]);
            let consumersDetails = [];
            concatAllConsumers.map((row, index) => {
                let consumer = {
                    name: row.name,
                    is_active: row.is_active,
                    is_deleted: row.is_deleted,
                    details: [
                        {
                            name: 'User',
                            value: row.created_by_user
                        },
                        {
                            name: 'IP',
                            value: row.client_address
                        }
                    ]
                };
                consumersDetails.push(consumer);
            });
            let cgDetails = {
                details: [
                    {
                        name: 'Poison messages',
                        value: numberWithCommas(cgsList[rowIndex]?.poison_messages)
                    },
                    {
                        name: 'Unprocessed messages',
                        value: numberWithCommas(cgsList[rowIndex]?.unprocessed_messages)
                    },
                    {
                        name: 'In process message',
                        value: numberWithCommas(cgsList[rowIndex]?.in_process_messages)
                    },
                    {
                        name: 'Max ack time',
                        value: `${numberWithCommas(cgsList[rowIndex]?.max_ack_time_ms)}ms`
                    },
                    {
                        name: 'Max message deliveries',
                        value: cgsList[rowIndex]?.max_msg_deliveries
                    }
                ],
                consumers: consumersDetails
            };
            setCgDetails(cgDetails);
        }
    };

    const returnClassName = (index, is_deleted) => {
        if (selectedRowIndex === index) {
            if (is_deleted) {
                return 'pubSub-row selected deleted';
            } else return 'pubSub-row selected';
        } else if (is_deleted) {
            return 'pubSub-row deleted';
        } else return 'pubSub-row';
    };

    return (
        <div className="pubSub-list-container">
            <div className="header">
                <p className="title">{props.producer ? 'Producers' : 'Consumer groups'}</p>
                {/* <p className="add-connector-button">{props.producer ? 'Add producer' : 'Add consumer'}</p> */}
            </div>
            {props.producer && producersList?.length > 0 && (
                <div className="coulmns-table">
                    <span style={{ width: '100px' }}>Name</span>
                    <span style={{ width: '80px' }}>User</span>
                    <span style={{ width: '35px' }}>Status</span>
                </div>
            )}
            {!props.producer && cgsList.length > 0 && (
                <div className="coulmns-table">
                    <span style={{ width: '75px' }}>Name</span>
                    <span style={{ width: '65px', textAlign: 'center' }}>Poison</span>
                    <span style={{ width: '75px', textAlign: 'center' }}>Unprocessed</span>
                    <span style={{ width: '35px', textAlign: 'center' }}>Status</span>
                </div>
            )}

            <div className="rows-wrapper">
                <div className="list-container">
                    {props.producer &&
                        producersList?.length > 0 &&
                        producersList?.map((row, index) => {
                            return (
                                <div className={returnClassName(index, row.is_deleted)} key={index} onClick={() => onSelectedRow(index, 'producer')}>
                                    <OverflowTip text={row.name} width={'100px'}>
                                        {row.name}
                                    </OverflowTip>
                                    <OverflowTip text={row.created_by_user} width={'80px'}>
                                        {row.created_by_user}
                                    </OverflowTip>
                                    <span className="status-icon" style={{ width: '38px' }}>
                                        <StatusIndication is_active={row.is_active} is_deleted={row.is_deleted} />
                                    </span>
                                </div>
                            );
                        })}
                    {!props.producer &&
                        cgsList?.length > 0 &&
                        cgsList?.map((row, index) => {
                            return (
                                <div className={returnClassName(index, row.is_deleted)} key={index} onClick={() => onSelectedRow(index, 'consumer')}>
                                    <OverflowTip text={row.name} width={'75px'}>
                                        {row.name}
                                    </OverflowTip>
                                    <OverflowTip text={row.poison_messages} width={'60px'} textAlign={'center'} textColor={row.poison_messages > 0 ? '#F7685B' : null}>
                                        {row.poison_messages}
                                    </OverflowTip>
                                    <OverflowTip text={row.unprocessed_messages} width={'75px'} textAlign={'center'}>
                                        {row.unprocessed_messages}
                                    </OverflowTip>
                                    <span className="status-icon" style={{ width: '38px' }}>
                                        <StatusIndication is_active={row.is_active} is_deleted={row.is_deleted} />
                                    </span>
                                </div>
                            );
                        })}
                </div>
                <div style={{ marginRight: '10px' }}>
                    {props.producer && producersList?.length > 0 && <CustomCollapse header="Details" defaultOpen={true} data={producerDetails} />}
                    {!props.producer && cgsList?.length > 0 && (
                        <Space direction="vertical">
                            <CustomCollapse status={false} header="Details" defaultOpen={true} data={cgDetails.details} />
                            <MultiCollapse header="Consumers" data={cgDetails.consumers} />
                        </Space>
                    )}
                </div>
            </div>
            {((props.producer && producersList?.length === 0) || (!props.producer && cgsList?.length === 0)) && (
                <div className="waiting-placeholder">
                    <img width={62} src={props.producer ? waitingProducer : waitingConsumer} />
                    <p>Waiting for the 1st {props.producer ? 'producer' : 'consumer'}</p>
                    {props.producer && <span className="des">A producer is the source application that pushes data to the station</span>}
                    {!props.producer && <span className="des">Consumer groups are a pool of consumers that divide the work of consuming and processing data</span>}
                    <Button
                        className="open-sdk"
                        width="200px"
                        height="37px"
                        placeholder={`Create your first ${props.producer ? 'producer' : 'consumer'}`}
                        colorType={'black'}
                        radiusType="circle"
                        border={'gray-light'}
                        backgroundColorType={'none'}
                        fontSize="12px"
                        fontFamily="InterSemiBold"
                        onClick={() => (props.producer ? setOpenCreateProducer(true) : setOpenCreateConsumer(true))}
                    />
                </div>
            )}
            <Modal header="SDK" width="710px" clickOutside={() => setOpenCreateConsumer(false)} open={openCreateConsumer} displayButtons={false}>
                <SdkExample showTabs={false} consumer={true} />
            </Modal>
            <Modal header="SDK" width="710px" clickOutside={() => setOpenCreateProducer(false)} open={openCreateProducer} displayButtons={false}>
                <SdkExample showTabs={false} />
            </Modal>
        </div>
    );
};

export default ProduceConsumList;
