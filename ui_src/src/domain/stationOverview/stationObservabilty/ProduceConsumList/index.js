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

import './style.scss';

import React, { useContext, useEffect, useState } from 'react';
import { Space } from 'antd';
import { Virtuoso } from 'react-virtuoso';

import waitingProducer from '../../../../assets/images/waitingProducer.svg';
import waitingConsumer from '../../../../assets/images/waitingConsumer.svg';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import unsupported from '../../../../assets/images/unsupported.svg';
import StatusIndication from '../../../../components/indication';
import SdkExample from '../../../../components/sdkExsample';
import CustomCollapse from '../components/customCollapse';
import MultiCollapse from '../components/multiCollapse';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import { StationStoreContext } from '../..';

const ProduceConsumList = ({ producer }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [selectedRowIndex, setSelectedRowIndex] = useState(0);
    const [producersList, setProducersList] = useState([]);
    const [cgsList, setCgsList] = useState([]);
    const [producerDetails, setProducerDetails] = useState([]);
    const [cgDetails, setCgDetails] = useState([]);
    const [openCreateProducer, setOpenCreateProducer] = useState(false);
    const [openCreateConsumer, setOpenCreateConsumer] = useState(false);

    useEffect(() => {
        if (producer) {
            let result = concatFunction('producer', stationState?.stationSocketData);
            setProducersList(result);
        } else {
            let result = concatFunction('cgs', stationState?.stationSocketData);
            setCgsList(result);
        }
    }, [stationState?.stationSocketData]);

    useEffect(() => {
        arrangeData('producer', selectedRowIndex);
        arrangeData('cgs', selectedRowIndex);
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
            ];
            setProducerDetails(details);
        } else {
            let concatAllConsumers = concatFunction('consumers', cgsList[rowIndex]);
            let consumersDetails = [];
            concatAllConsumers.map((row, index) => {
                let consumer = {
                    name: row.name,
                    count: row.count,
                    is_active: row.is_active,
                    is_deleted: row.is_deleted,
                };
                consumersDetails.push(consumer);
            });
            let cgDetails = {
                details: [
                    {
                        name: 'Unacked messages',
                        value: cgsList[rowIndex]?.poison_messages?.toLocaleString()
                    },
                    {
                        name: 'Unprocessed messages',
                        value: cgsList[rowIndex]?.unprocessed_messages?.toLocaleString()
                    },
                    {
                        name: 'In process message',
                        value: cgsList[rowIndex]?.in_process_messages?.toLocaleString()
                    },
                    {
                        name: 'Max ack time',
                        value: `${cgsList[rowIndex]?.max_ack_time_ms?.toLocaleString()}ms`
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
        <div>
            <div className="pubSub-list-container">
                {' '}
                <div className="header">
                    {producer && <p className="title">Producers {producersList?.length > 0 && `(${producersList?.length})`}</p>}
                    {!producer && <p className="title">Consumer groups {cgsList?.length > 0 && `(${cgsList?.length})`}</p>}
                </div>
                {producer && producersList?.length > 0 && (
                    <div className="coulmns-table">
                        <span style={{ width: '100px' }}>Name</span>
                        <span style={{ width: '100px' }}>Count</span>
                        <span style={{ width: '35px' }}>Status</span>
                    </div>
                )}
                {!producer && cgsList.length > 0 && (
                    <div className="coulmns-table">
                        <span style={{ width: '80px' }}>Name</span>
                        <span style={{ width: '80px', textAlign: 'center' }}>Unacked</span>
                        <span style={{ width: '80px', textAlign: 'center' }}>Unprocessed</span>
                        <span style={{ width: '35px', textAlign: 'center' }}>Status</span>
                    </div>
                )}
                {(producersList?.length > 0 || cgsList?.length > 0) && (
                    <div className="rows-wrapper">
                        <div
                            className="list-container"
                            style={{
                                height: `calc(100% - ${
                                    producer
                                        ? document.getElementById('producer-details')?.offsetHeight + 3 + 'px'
                                        : document.getElementById('consumer-details')?.offsetHeight + 5 + 'px'
                                })`
                            }}
                        >
                            {producer && producersList?.length > 0 && (
                                <Virtuoso
                                    data={producersList}
                                    overscan={100}
                                    itemContent={(index, row) => (
                                        <div className={returnClassName(index, row.is_deleted)} key={index} onClick={() => onSelectedRow(index, 'producer')}>
                                            <OverflowTip text={row.name} width={'100px'}>
                                                {row.name}
                                            </OverflowTip>
                                            <OverflowTip text={row.count} width={'70px'}>
                                                {row.count}
                                            </OverflowTip>
                                            <span className="status-icon" style={{ width: '38px' }}>
                                                <StatusIndication is_active={row.is_active} is_deleted={row.is_active} />
                                            </span>
                                        </div>
                                    )}
                                />
                            )}
                            {!producer && cgsList?.length > 0 && (
                                <Virtuoso
                                    data={cgsList}
                                    overscan={100}
                                    itemContent={(index, row) => (
                                        <div className={returnClassName(index, row.is_deleted)} key={index} onClick={() => onSelectedRow(index, 'consumer')}>
                                            <OverflowTip text={row.name} width={'80px'}>
                                                {row.name}
                                            </OverflowTip>
                                            <OverflowTip
                                                text={row.poison_messages.toLocaleString()}
                                                width={'80px'}
                                                textAlign={'center'}
                                                textColor={row.poison_messages > 0 ? '#F7685B' : null}
                                            >
                                                {row.poison_messages.toLocaleString()}
                                            </OverflowTip>
                                            <OverflowTip text={row.unprocessed_messages.toLocaleString()} width={'80px'} textAlign={'center'}>
                                                {row.unprocessed_messages.toLocaleString()}
                                            </OverflowTip>
                                            <span className="status-icon" style={{ width: '35px' }}>
                                                <StatusIndication is_active={row.is_active} is_deleted={row.is_deleted} />
                                            </span>
                                        </div>
                                    )}
                                />
                            )}
                        </div>
                        <div style={{ marginRight: '10px' }} id={producer ? 'producer-details' : 'consumer-details'}>
                            {producer && producersList?.length > 0 }
                            {!producer && cgsList?.length > 0 && (
                                <Space direction="vertical">
                                    <CustomCollapse header="Details" status={false} defaultOpen={true} data={cgDetails.details} />
                                    <CustomCollapse header="Consumers" data={cgDetails.consumers} consumerList={true} />
                                </Space>
                            )}
                        </div>
                    </div>
                )}
                {((producer && producersList?.length === 0) || (!producer && cgsList?.length === 0)) && (
                    <div className="waiting-placeholder">
                        <img width={62} src={producer ? waitingProducer : waitingConsumer} alt="producer" />
                        <p>Waiting for the 1st {producer ? 'producer' : 'consumer'}</p>
                        {producer && <span className="des">A producer is the source application that pushes data to the station</span>}
                        {!producer && <span className="des">Consumer groups are a pool of consumers that divide the work of consuming and processing data</span>}
                        <Button
                            className="open-sdk"
                            width="200px"
                            height="37px"
                            placeholder={`Create your first ${producer ? 'producer' : 'consumer'}`}
                            colorType={'black'}
                            radiusType="circle"
                            border={'gray-light'}
                            backgroundColorType={'none'}
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => (producer ? setOpenCreateProducer(true) : setOpenCreateConsumer(true))}
                        />
                    </div>
                )}
                {!stationState?.stationMetaData?.is_native && (
                    <div className="unsupported-placeholder">
                        <div className="placeholder-wrapper">
                            <img src={unsupported} alt="unsupported" />
                            <p>For the full Memphis experience, Memphis SDK is needed</p>
                            <Button
                                className="open-sdk"
                                width="200px"
                                height="37px"
                                placeholder="View Memphis SDK's"
                                colorType={'white'}
                                radiusType="circle"
                                border={'none'}
                                backgroundColorType={'purple'}
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                onClick={() => (producer ? setOpenCreateProducer(true) : setOpenCreateConsumer(true))}
                            />
                        </div>
                    </div>
                )}
            </div>
            <Modal width="710px" height="700px" clickOutside={() => setOpenCreateConsumer(false)} open={openCreateConsumer} displayButtons={false}>
                <SdkExample withHeader={true} showTabs={false} consumer={true} stationName={stationState?.stationMetaData?.name} />
            </Modal>
            <Modal
                width="710px"
                height="700px"
                clickOutside={() => {
                    setOpenCreateProducer(false);
                }}
                open={openCreateProducer}
                displayButtons={false}
            >
                <SdkExample withHeader={true} showTabs={false} stationName={stationState?.stationMetaData?.name} />
            </Modal>
        </div>
    );
};

export default ProduceConsumList;
