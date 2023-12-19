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

import React, { useContext, useEffect, useRef, useState } from 'react';
import { LOCAL_STORAGE_ENV, LOCAL_STORAGE_REST_GW_HOST, LOCAL_STORAGE_REST_GW_PORT } from '../../../../const/localStorageConsts';
import { Space, Popover } from 'antd';
import { Virtuoso } from 'react-virtuoso';
import { FiPlayCircle } from 'react-icons/fi';
import { ReactComponent as WaitingProducerIcon } from '../../../../assets/images/waitingProducer.svg';
import { ReactComponent as WaitingConsumerIcon } from '../../../../assets/images/waitingConsumer.svg';
import { ReactComponent as PlayVideoIcon } from '../../../../assets/images/playVideoIcon.svg';
import { ReactComponent as PurplePlus } from '../../../../assets/images/purplePlus.svg';
import { ReactComponent as ProducerIcon } from '../../../../assets/images/producerIcon.svg';
import { ReactComponent as ConnectIcon } from '../../../../assets/images/connectIcon.svg';
import { IoPlayCircleOutline, IoRemoveCircleOutline, IoPause, IoWarning } from 'react-icons/io5';
import { HiDotsVertical } from 'react-icons/hi';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import { ReactComponent as UnsupportedIcon } from '../../../../assets/images/unsupported.svg';
import StatusIndication from '../../../../components/indication';
import SdkExample from '../../../../components/sdkExample';
import CustomCollapse from '../components/customCollapse';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import GenerateTokenModal from '../../../stationOverview/components/generateTokenModal';
import { StationStoreContext } from '../..';
import ProduceMessages from '../../../../components/produceMessages';
import ConnectorModal from '../../../../components/connectorModal';
import ConnectorError from '../../../../components/connectorError';
import { ReactComponent as ErrorModalIcon } from '../../../../assets/images/errorModal.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Spinner from '../../../../components/spinner';
import TooltipComponent from '../../../../components/tooltip/tooltip';
import { isCloud } from '../../../../services/valueConvertor';
import { sendTrace } from '../../../../services/genericServices';

const overlayStylesConnectors = {
    borderRadius: '8px',
    width: '230px',
    paddingTop: '5px',
    paddingBottom: '5px',
    marginBottom: '10px'
};

const overlayStyleConnectors = { borderRadius: '8px', width: '150px', paddingTop: '5px', paddingBottom: '5px' };

const MenuItem = ({ name, onClick, icon, loader }) => {
    return (
        <div className="item-wrapper-connectors" onClick={onClick}>
            <span className="item-name">
                {icon}
                <label>{name}</label>
            </span>
            {loader && <Spinner alt="loading" />}
        </div>
    );
};

const ProduceConsumList = ({ producer }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [selectedRowIndex, setSelectedRowIndex] = useState(0);
    const [producersList, setProducersList] = useState([]);
    const [connectorsSourceList, setConnectorsSourceList] = useState([]);
    const [connectorsSinkList, setConnectorsSinkList] = useState([]);
    const [cgsList, setCgsList] = useState([]);
    const [openProduceMessages, setOpenProduceMessages] = useState(false);
    const [cgDetails, setCgDetails] = useState([]);
    const [openCreateProducer, setOpenCreateProducer] = useState(false);
    const [openCreateConsumer, setOpenCreateConsumer] = useState(false);
    const [generateModal, setGenerateModal] = useState(false);
    const produceMessagesRef = useRef(null);
    const [produceloading, setProduceLoading] = useState(false);
    const [openNoConsumer, setOpenNoConsumer] = useState(false);
    const [activeConsumerList, setActiveConsumerList] = useState([]);
    const [openProducerPopover, setOpenProducerPopover] = useState(false);
    const [openConnectorPopover, setOpenConnectorPopover] = useState(false);
    const [openConnectorPopoverItem, setOpenConnectorPopoverItem] = useState(null);
    const [selectedConnector, setSelectedConnector] = useState(null);
    const [openConnectorModal, setOpenConnectorModal] = useState(false);
    const [openConnectorError, setOpenConnectorError] = useState(false);
    const [loading, setLoader] = useState(false);

    const producerItemsList = [
        {
            action: 'Produce Synthetic Data',
            onClick: () => {
                setOpenProduceMessages(true);
                setOpenProducerPopover(false);
            }
        },
        {
            action: 'Develop a Producer',
            onClick: () => {
                setOpenCreateProducer(true);
                setOpenProducerPopover(false);
            }
        },
        {
            action: 'Produce using REST',
            onClick: () => {
                setGenerateModal(true);
                setOpenProducerPopover(false);
            }
        },
        {
            action: 'Add a Source',
            onClick: () => {
                setOpenConnectorModal(true);
                setOpenProducerPopover(false);
            }
        }
    ];

    const consumeItemsList = [
        {
            action: 'Develop a Consumer',
            onClick: () => {
                setOpenCreateConsumer(true);
                setOpenProducerPopover(false);
            }
        },
        {
            action: 'Consume using REST',
            onClick: () => {
                setGenerateModal(true);
                setOpenProducerPopover(false);
            }
        },
        {
            action: 'Add a Sink',
            onClick: () => {
                setOpenConnectorModal(true);
                setOpenProducerPopover(false);
            }
        }
    ];

    const removeConnector = async (type) => {
        setLoader(true);
        try {
            await httpRequest('POST', ApiEndpoints.REMOVE_CONNECTOR, {
                connector_id: selectedConnector?.id
            });
            if (type === 'source') {
                let newConnecorList = [...connectorsSourceList];
                newConnecorList.splice(openConnectorPopoverItem, 1);
                setConnectorsSourceList(newConnecorList);
            }
            if (type === 'sink') {
                let newConnecorList = [...connectorsSinkList];
                newConnecorList.splice(openConnectorPopoverItem, 1);
                setConnectorsSinkList(newConnecorList);
            }
            sendTrace('removeConnector', {
                type: type,
                connector: type === 'source' ? connectorsSourceList[openConnectorPopoverItem]?.type : connectorsSinkList[openConnectorPopoverItem]?.type,
                name: type === 'source' ? connectorsSourceList[openConnectorPopoverItem]?.name : connectorsSinkList[openConnectorPopoverItem]?.name
            });
            setLoader(false);
            setOpenConnectorPopover(false);
        } catch (error) {
            setLoader(false);
        }
    };

    const startConnector = async (type) => {
        setLoader(true);
        try {
            await httpRequest('POST', ApiEndpoints.START_CONNECTOR, {
                connector_id: selectedConnector?.id
            });
            if (type === 'source') {
                let newConnecorList = [...connectorsSourceList];
                newConnecorList[openConnectorPopoverItem].is_active = true;
                setConnectorsSourceList(newConnecorList);
            }
            if (type === 'sink') {
                let newConnecorList = [...connectorsSinkList];
                newConnecorList[openConnectorPopoverItem].is_active = true;
                setConnectorsSinkList(newConnecorList);
            }
            sendTrace('startConnector', {
                type: type,
                connector: type === 'source' ? connectorsSourceList[openConnectorPopoverItem]?.type : connectorsSinkList[openConnectorPopoverItem]?.type,
                name: type === 'source' ? connectorsSourceList[openConnectorPopoverItem]?.name : connectorsSinkList[openConnectorPopoverItem]?.name
            });
            setLoader(false);
            setOpenConnectorPopover(false);
        } catch (error) {
            setLoader(false);
        }
    };
    const stopConnector = async (type) => {
        setLoader(true);
        try {
            await httpRequest('POST', ApiEndpoints.STOP_CONNECTOR, {
                connector_id: selectedConnector?.id
            });
            if (type === 'source') {
                let newConnecorList = [...connectorsSourceList];
                newConnecorList[openConnectorPopoverItem].is_active = false;
                setConnectorsSourceList(newConnecorList);
            }
            if (type === 'sink') {
                let newConnecorList = [...connectorsSinkList];
                newConnecorList[openConnectorPopoverItem].is_active = false;
                setConnectorsSinkList(newConnecorList);
            }
            sendTrace('stopConnector', {
                type: type,
                connector: type === 'source' ? connectorsSourceList[openConnectorPopoverItem]?.type : connectorsSinkList[openConnectorPopoverItem]?.type,
                name: type === 'source' ? connectorsSourceList[openConnectorPopoverItem]?.name : connectorsSinkList[openConnectorPopoverItem]?.name
            });
            setLoader(false);
            setOpenConnectorPopover(false);
        } catch (error) {
            setLoader(false);
        }
    };

    const handleNewConnector = (connector, source) => {
        source === 'source' ? setConnectorsSourceList([...connectorsSourceList, connector]) : setConnectorsSinkList([...connectorsSinkList, connector]);
    };

    useEffect(() => {
        if (producer) {
            let [result, activeConsumers] = concatFunction('producer', stationState?.stationSocketData);
            setProducersList(result);
            setConnectorsSourceList(stationState?.stationSocketData?.source_connectors);
            setActiveConsumerList(activeConsumers);
        } else {
            let result = concatFunction('cgs', stationState?.stationSocketData);
            setCgsList(result);
            setConnectorsSinkList(stationState?.stationSocketData?.sink_connectors);
        }
    }, [stationState?.stationSocketData]);

    useEffect(() => {
        arrangeData(selectedRowIndex);
    }, [producersList, cgsList]);

    const concatFunction = (type, data) => {
        let connected = [];
        let deleted = [];
        let disconnected = [];
        let concatArrays = [];
        let activeConsumers = [];
        if (type === 'producer') {
            connected = data?.connected_producers || [];
            deleted = data?.deleted_producers || [];
            disconnected = data?.disconnected_producers || [];
            concatArrays = connected.concat(disconnected);
            concatArrays = concatArrays.concat(deleted);
            activeConsumers = data?.connected_cgs || [];
            disconnected = data?.disconnected_cgs || [];
            activeConsumers = activeConsumers.concat(disconnected);
            return [concatArrays, activeConsumers];
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
        arrangeData(rowIndex);
    };

    const arrangeData = (rowIndex) => {
        let concatAllConsumers = concatFunction('consumers', cgsList[rowIndex]);
        let consumersDetails = [];
        concatAllConsumers.map((row, index) => {
            let consumer = {
                name: row.name,
                count: row.count,
                is_active: row.is_active,
                is_deleted: row.is_deleted
            };
            consumersDetails.push(consumer);
        });
        let cgDetails = {
            details: [
                {
                    name: 'Unacknowledged messages',
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

    const restGWHost =
        localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
            ? `http://localhost:${localStorage.getItem(LOCAL_STORAGE_REST_GW_PORT)}`
            : localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST);
    const countProducers = (producersList) => {
        const connected = producersList?.reduce((accumulator, item) => accumulator + item.connected_producers_count, 0);
        const disconnected = producersList?.reduce((accumulator, item) => accumulator + item.disconnected_producers_count, 0);
        return connected + disconnected;
    };

    return (
        <div className="station-observabilty-side">
            <div className="pubSub-list-container">
                <div className="header">
                    {producer && (
                        <span className="poduce-consume-header">
                            <p className="title">
                                <TooltipComponent text="max allowed producers" placement="right">
                                    <>
                                        Sources ({(producersList?.length > 0 && countProducers(producersList)) || 0}
                                        {isCloud() && '/' + stationState?.stationSocketData?.max_amount_of_allowed_producers})
                                    </>
                                </TooltipComponent>
                            </p>
                            <Popover
                                overlayInnerStyle={overlayStylesConnectors}
                                placement="bottomLeft"
                                content={producerItemsList?.map((item, index) => (
                                    <MenuItem key={`${index}-${item.action}`} name={item.action} onClick={item.onClick} />
                                ))}
                                trigger="click"
                                onOpenChange={() => setOpenProducerPopover(!openProducerPopover)}
                                open={openProducerPopover}
                            >
                                <PurplePlus alt="Add source" className="add" />
                            </Popover>
                        </span>
                    )}
                    {!producer && (
                        <span className="poduce-consume-header">
                            <p className="title">
                                Consumer groups {(cgsList?.length > 0 || connectorsSinkList?.length > 0) && `(${cgsList?.length + connectorsSinkList?.length})`}
                            </p>
                            <Popover
                                overlayInnerStyle={overlayStylesConnectors}
                                placement="bottomLeft"
                                content={consumeItemsList?.map((item, index) => (
                                    <MenuItem key={`${index}-${item.action}`} name={item.action} onClick={item.onClick} />
                                ))}
                                trigger="click"
                                onOpenChange={() => setOpenProducerPopover(!openProducerPopover)}
                                open={openProducerPopover}
                            >
                                <PurplePlus alt="Add source" className="add" />
                            </Popover>
                        </span>
                    )}
                </div>
                {producer && (producersList?.length > 0 || connectorsSourceList?.length > 0) && (
                    <div className="coulmns-table">
                        <span style={{ width: '100px' }}>Name</span>
                        <span style={{ width: '100px' }}>Count</span>
                        <span style={{ width: '35px' }}>Status</span>
                        <span style={{ width: '20px' }}></span>
                    </div>
                )}
                {!producer && (cgsList.length > 0 || connectorsSinkList?.length > 0) && (
                    <div className="coulmns-table">
                        <span style={{ width: '60px' }}>Name</span>
                        <span style={{ width: '100px', textAlign: 'center' }}>Unacknowledged</span>
                        <span style={{ width: '80px', textAlign: 'center' }}>Unprocessed</span>
                        <span style={{ width: '35px', textAlign: 'center' }}>Status</span>
                        <span style={{ width: '20px' }}></span>
                    </div>
                )}
                {(producersList?.length > 0 || connectorsSourceList?.length > 0 || connectorsSinkList?.length > 0 || cgsList?.length > 0) && (
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
                            {producer && (producersList?.length > 0 || connectorsSourceList?.length > 0) && (
                                <Virtuoso
                                    data={[...producersList, ...connectorsSourceList]}
                                    overscan={100}
                                    itemContent={(index, row) => (
                                        <div className={returnClassName(index, row?.is_deleted)} key={index} onClick={() => onSelectedRow(index, 'producer')}>
                                            <span className="connector-name">
                                                {row?.connector_connection_id ? (
                                                    <TooltipComponent text="connector">
                                                        <ConnectIcon />
                                                    </TooltipComponent>
                                                ) : (
                                                    <TooltipComponent text="producer">
                                                        <ProducerIcon />
                                                    </TooltipComponent>
                                                )}
                                                <OverflowTip text={row.name} width={'80px'}>
                                                    {row.name}
                                                </OverflowTip>
                                            </span>
                                            <div style={{ width: '92px', maxWidth: '100%' }}>
                                                {row?.connector_connection_id ? (
                                                    'N/A'
                                                ) : (
                                                    <TooltipComponent text="connected | disconnected">
                                                        {row.connected_producers_count + ' | ' + row.disconnected_producers_count}
                                                    </TooltipComponent>
                                                )}
                                            </div>
                                            <span className="status-icon" style={{ width: '38px' }}>
                                                <StatusIndication is_active={row.is_active} is_deleted={row.is_active} />
                                            </span>
                                            <Popover
                                                overlayInnerStyle={overlayStyleConnectors}
                                                placement="bottomLeft"
                                                onOpenChange={() => {
                                                    setOpenConnectorPopoverItem(index);
                                                    setSelectedConnector(row);
                                                    setOpenConnectorPopover(!openConnectorPopover);
                                                }}
                                                open={openConnectorPopover && openConnectorPopoverItem === index}
                                                content={
                                                    <>
                                                        <MenuItem
                                                            name={row?.is_active ? 'Pause' : 'Play'}
                                                            onClick={() => (row?.is_active ? stopConnector('source') : startConnector('source'))}
                                                            icon={row?.is_active ? <IoPause /> : <IoPlayCircleOutline />}
                                                        />
                                                        <MenuItem
                                                            name={'Disconnect'}
                                                            onClick={() => {
                                                                removeConnector('source');
                                                            }}
                                                            icon={<IoRemoveCircleOutline />}
                                                            loader={loading && openConnectorPopoverItem === index}
                                                        />
                                                        <MenuItem
                                                            name={'Erros'}
                                                            onClick={() => {
                                                                setOpenConnectorError(true);
                                                                setOpenConnectorPopover(false);
                                                            }}
                                                            icon={<IoWarning />}
                                                        />
                                                    </>
                                                }
                                                trigger="click"
                                            >
                                                <div
                                                    className={`connector-options${openConnectorPopover && openConnectorPopoverItem === index ? '-selected' : ''}`}
                                                    onClick={(e) => row?.connector_connection_id && e.stopPropagation()}
                                                >
                                                    {row?.connector_connection_id && <HiDotsVertical alt="actions" />}
                                                </div>
                                            </Popover>
                                        </div>
                                    )}
                                />
                            )}
                            {!producer && (cgsList?.length > 0 || connectorsSinkList?.length > 0) && (
                                <Virtuoso
                                    data={[...cgsList, ...connectorsSinkList]}
                                    overscan={100}
                                    itemContent={(index, row) => (
                                        <div className={returnClassName(index, row?.is_deleted)} key={index} onClick={() => onSelectedRow(index, 'consumer')}>
                                            <span className="connector-name">
                                                {row?.connector_connection_id ? (
                                                    <TooltipComponent text="connector">
                                                        <ConnectIcon />
                                                    </TooltipComponent>
                                                ) : (
                                                    <TooltipComponent text="consumer">
                                                        <ProducerIcon />
                                                    </TooltipComponent>
                                                )}
                                                <OverflowTip text={row.name} width={'80px'}>
                                                    {row.name}
                                                </OverflowTip>
                                            </span>
                                            <OverflowTip
                                                text={row?.poison_messages?.toLocaleString()}
                                                width={'80px'}
                                                textAlign={'center'}
                                                textColor={row?.poison_messages > 0 ? '#F7685B' : null}
                                            >
                                                {row?.poison_messages?.toLocaleString()}
                                            </OverflowTip>
                                            <OverflowTip text={row?.unprocessed_messages?.toLocaleString()} width={'80px'} textAlign={'center'}>
                                                {row?.unprocessed_messages?.toLocaleString()}
                                            </OverflowTip>
                                            <span className="status-icon" style={{ width: '35px' }}>
                                                <StatusIndication is_active={row?.is_active} is_deleted={row?.is_deleted} />
                                            </span>
                                            <Popover
                                                overlayInnerStyle={overlayStyleConnectors}
                                                placement="bottomLeft"
                                                onOpenChange={() => {
                                                    setOpenConnectorPopoverItem(index);
                                                    setSelectedConnector(row);
                                                    setOpenConnectorPopover(!openConnectorPopover);
                                                }}
                                                open={openConnectorPopover && openConnectorPopoverItem === index}
                                                content={
                                                    <>
                                                        <MenuItem
                                                            name={row?.is_active ? 'Pause' : 'Play'}
                                                            onClick={() => (row?.is_active ? stopConnector('sink') : startConnector('sink'))}
                                                            icon={row?.is_active ? <IoPause /> : <IoPlayCircleOutline />}
                                                        />
                                                        <MenuItem name={'Disconnect'} onClick={() => removeConnector('sink')} icon={<IoRemoveCircleOutline />} />
                                                        <MenuItem
                                                            name={'Erros'}
                                                            onClick={() => {
                                                                setOpenConnectorError(true);
                                                                setOpenConnectorPopover(false);
                                                            }}
                                                            icon={<IoWarning />}
                                                        />
                                                    </>
                                                }
                                                trigger="click"
                                            >
                                                <div
                                                    className="connector-options"
                                                    onClick={(e) => {
                                                        if (row?.connector_connection_id) e.stopPropagation();
                                                    }}
                                                >
                                                    {row?.connector_connection_id && <HiDotsVertical alt="actions" />}
                                                </div>
                                            </Popover>
                                        </div>
                                    )}
                                />
                            )}
                        </div>
                        <div style={{ marginRight: '10px' }} id={producer ? 'producer-details' : 'consumer-details'}>
                            {producer && (producersList?.length > 0 || connectorsSourceList?.length > 0)}
                            {!producer && cgsList?.length > 0 && (
                                <Space direction="vertical">
                                    <CustomCollapse header="Details" status={false} defaultOpen={true} data={cgDetails.details} />
                                    <CustomCollapse header="Consumers" data={cgDetails.consumers} consumerList={true} />
                                </Space>
                            )}
                        </div>
                    </div>
                )}
                {((producer && producersList?.length === 0 && connectorsSourceList?.length === 0) ||
                    (!producer && cgsList?.length === 0 && connectorsSinkList?.length === 0)) && (
                    <div className="waiting-placeholder">
                        {producer ? <WaitingProducerIcon width={62} alt="producer" /> : <WaitingConsumerIcon width={62} alt="producer" />}
                        <p>{`No ${producer ? 'producers' : 'consumers'} yet`}</p>
                        {producer && (
                            <span className="des">A producer represents the originating application or service responsible for sending messages to a station</span>
                        )}
                        {!producer && <span className="des">A consumer group is a group of clients responsible for retrieving messages from a station</span>}
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
                            <UnsupportedIcon />
                            <p>Some features are limited to Memphis SDK only</p>
                            <Button
                                className="open-sdk"
                                width="200px"
                                height="37px"
                                placeholder="Create your Memphis client"
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
            <Modal width="1200px" height="780px" clickOutside={() => setOpenCreateConsumer(false)} open={openCreateConsumer} displayButtons={false}>
                <SdkExample withHeader={true} showTabs={false} stationName={stationState?.stationMetaData?.name} consumer={true} />
            </Modal>
            <Modal
                width="1200px"
                height="780px"
                clickOutside={() => {
                    setOpenCreateProducer(false);
                }}
                open={openCreateProducer}
                displayButtons={false}
            >
                <SdkExample withHeader={true} showTabs={false} stationName={stationState?.stationMetaData?.name} />
            </Modal>
            <Modal
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <PlayVideoIcon className="headerImage" alt="stationImg" />
                        </div>
                        <p>Produce a message</p>
                        <label>Produce a message through the Console.</label>
                    </div>
                }
                className={'modal-wrapper produce-modal'}
                width="550px"
                height="60vh"
                clickOutside={() => {
                    setOpenProduceMessages(false);
                }}
                open={openProduceMessages}
                displayButtons={true}
                rBtnText={
                    <div className="action-button">
                        <FiPlayCircle />
                        Produce
                    </div>
                }
                rBtnClick={() => {
                    if (activeConsumerList.length === 0 && stationState?.stationMetaData?.retention_type === 'ack_based') {
                        setOpenNoConsumer(true);
                    } else {
                        produceMessagesRef.current();
                    }
                }}
                lBtnClick={() => setOpenProduceMessages(false)}
                lBtnText={'Cancel'}
                isLoading={produceloading}
                keyListener={false}
            >
                <ProduceMessages
                    stationName={stationState?.stationMetaData?.name}
                    setLoading={(e) => setProduceLoading(e)}
                    produceMessagesRef={produceMessagesRef}
                    cancel={() => setOpenProduceMessages(false)}
                />
            </Modal>
            <Modal
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <ErrorModalIcon width={45} height={45} />
                        </div>
                    </div>
                }
                className={'modal-wrapper produce-modal'}
                width="403px"
                clickOutside={() => {
                    setOpenNoConsumer(false);
                }}
                open={openNoConsumer}
                displayButtons={true}
                rBtnText={
                    <div className="action-button">
                        <FiPlayCircle />
                        Produce
                    </div>
                }
                rBtnClick={() => {
                    produceMessagesRef.current();
                    setOpenNoConsumer(false);
                }}
                lBtnClick={() => setOpenNoConsumer(false)}
                lBtnText={'Cancel'}
                isLoading={produceloading}
                keyListener={false}
            >
                <p className="no-consumer-message--p">The message will not be stored</p>
                <label className="no-consumer-message--label">When using ack-based retention, a message will not be stored if no consumers are connected.</label>
            </Modal>
            <ConnectorModal
                open={openConnectorModal}
                clickOutside={() => setOpenConnectorModal(false)}
                newConnecor={(connector, source) => handleNewConnector(connector, source)}
                source={producer}
            />
            <Modal
                header={<span className="rest-title">{`${producer ? 'Produce' : 'Consume'} using REST`}</span>}
                displayButtons={false}
                width="460px"
                clickOutside={() => setGenerateModal(false)}
                open={generateModal}
                className="generate-modal"
            >
                <GenerateTokenModal host={restGWHost} restProducer stationName={stationState?.stationMetaData?.name} close={() => setGenerateModal(false)} />
            </Modal>
            <ConnectorError open={openConnectorError} clickOutside={() => setOpenConnectorError(false)} connectorId={selectedConnector?.id} />
        </div>
    );
};

export default ProduceConsumList;
