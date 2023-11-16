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
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import { Context } from '../../../../../hooks/store';
import Modal from '../../../../../components/modal';
import { StationStoreContext } from '../../../';
import { ReactComponent as AddFunctionIcon } from '../../../../../assets/images/addFunction.svg';
import { ReactComponent as PlusIcon } from '../../../../../assets/images/plusIcon.svg';
import { ReactComponent as ProcessedIcon } from '../../../../../assets/images/processIcon.svg';
import { ReactComponent as GitIcon } from '../../../../../assets/images/gitIcon.svg';
import { ReactComponent as CodeGrayIcon } from '../../../../../assets/images/codeGrayIcon.svg';
import { ReactComponent as CloseIcon } from '../../../../../assets/images/close.svg';
import { ReactComponent as MetricsIcon } from '../../../../../assets/images/metricsIcon.svg';
import { ReactComponent as MetricsClockIcon } from '../../../../../assets/images/metricsClockIcon.svg';
import { ReactComponent as MetricsErrorIcon } from '../../../../../assets/images/metricsErrorIcon.svg';
import { Tabs } from 'antd';
import dataPassLineLottie from '../../../../../assets/lotties/dataPassLine.json';
import dataPassLineEmptyLottie from '../../../../../assets/lotties/dataPassLineEmpty.json';
import Lottie from 'lottie-react';
import FunctionCard from '../functionCard';
import FunctionsModal from '../functionsModal';
import FunctionLogs from '../functionLogs';
import OverflowTip from '../../../../../components/tooltip/overflowtip';
import { StringCodec, JSONCodec } from 'nats.ws';

let sub;

const FunctionsOverview = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [currentFunction, setCurrentFunction] = useState(null);
    const [functionDetails, setFunctionDetails] = useState(null);
    const [openFunctionsModal, setOpenFunctionsModal] = useState(false);
    const [socketOn, setSocketOn] = useState(false);
    const [state, dispatch] = useContext(Context);

    useEffect(() => {
        if (socketOn) {
            getFunctionsOverview();
        }
    }, [stationState?.stationPartition || stationState?.stationMetaData?.name]);

    const startListen = async () => {
        const jc = JSONCodec();
        const sc = StringCodec();

        const listenForUpdates = async () => {
            try {
                if (sub) {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: data });
                        if (!socketOn) {
                            setSocketOn(true);
                        }
                    }
                }
            } catch (err) {
                console.error(`Error receiving data updates for station overview:`, err);
            }
        };

        try {
            const rawBrokerName = await state.socket?.request(
                `$memphis_ws_subs.get_functions_overview.${stationState?.stationMetaData?.name}.${stationState?.stationPartition || -1}`,
                sc.encode('SUB')
            );
            if (rawBrokerName) {
                const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                sub = state.socket?.subscribe(
                    `$memphis_ws_pubs.get_functions_overview.${stationState?.stationMetaData?.name}.${stationState?.stationPartition || -1}.${brokerName}`
                );
                listenForUpdates();
            }
        } catch (err) {
            console.error('Error subscribing to station overview data:', err);
        }
    };

    const stopListen = async () => {
        if (sub) {
            try {
                await sub.unsubscribe();
            } catch (err) {
                console.error('Error unsubscribing from station overview data:', err);
            }
        }
    };

    useEffect(() => {
        if (state.socket) {
            startListen();
        }
        return () => {
            stopListen();
        };
    }, [state.socket, stationState?.stationMetaData?.name]);

    useEffect(() => {
        if (sub && socketOn) {
            stopListen();
            startListen();
        }
    }, [stationState?.stationPartition, stationState?.stationMetaData?.name]);

    const getFunctionsOverview = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_FUNCTIONS_OVERVIEW}?station_name=${stationState?.stationMetaData?.name}`);
            stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: data });
            setOpenFunctionsModal(false);
        } catch (e) {
            return;
        }
    };

    const getFunctionDetails = async () => {
        try {
            const response = await httpRequest('GET', `${ApiEndpoints.GET_FUNCTION_DETAILS}?function_id=${currentFunction?.id}`);
            setFunctionDetails(response);
        } catch (e) {
            return;
        }
    };

    useEffect(() => {
        getFunctionsOverview();
    }, []);

    useEffect(() => {
        currentFunction && getFunctionDetails();
    }, [currentFunction]);

    const handleAddFunction = async (requestBody) => {
        requestBody.station_name = stationState?.stationMetaData?.name;
        requestBody.partition = stationState?.stationMetaData?.partitions_number;
        try {
            await httpRequest('POST', ApiEndpoints.ADD_FUNCTION, requestBody);
            getFunctionsOverview();
        } catch (e) {
        } finally {
            setFunctionDetails(null);
        }
    };

    const changeActivition = async (funcId, flag) => {
        let functionsOveriew = { ...stationState?.stationFunctions };
        functionsOveriew?.functions?.forEach((item) => {
            if (item.id === funcId) {
                item.activated = flag;
            }
        });
        stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: functionsOveriew });
    };

    const handleDeleteFunction = async () => {
        setCurrentFunction(null);
    };

    const items = [
        {
            key: '1',
            label: 'Metrics',
            children: (
                <div className="metrics-wrapper">
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Total invocations</div>
                            <div className="metrics-body-subtitle">{functionDetails?.total_invocations?.toLocaleString() || 0}</div>
                        </div>
                    </div>
                    <div className="metrics-divider"></div>
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsClockIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Av. Processing time</div>
                            <div className="metrics-body-subtitle">
                                {functionDetails?.avg_processing_time}
                                <span>/sec</span>
                            </div>
                        </div>
                    </div>
                    <div className="metrics-divider"></div>
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsErrorIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Error rate</div>
                            <div className="metrics-body-subtitle">{functionDetails?.error_rate}%</div>
                        </div>
                    </div>
                </div>
            )
        },
        {
            key: '2',
            label: 'Logs',
            children: <FunctionLogs functionId={currentFunction?.id} />
        }
    ];

    const statisticsData = [
        { name: 'Awaiting msgs', data: stationState?.stationFunctions?.total_awaiting_messages?.toLocaleString() },
        { name: 'Processed msgs', data: stationState?.stationFunctions?.total_processed_messages?.toLocaleString() },
        { name: 'Total invocations', data: stationState?.stationFunctions?.total_invocations?.toLocaleString() || 0 },
        { name: 'Avg Error rate', data: stationState?.stationFunctions?.average_error_rate },
        {
            name: 'Ordering',
            data: stationState?.stationFunctions?.functions?.length > 0 ? (stationState?.stationFunctions?.functions[0]?.ordering_matter ? 'Yes' : 'No') : 'N/A'
        }
    ];
    return (
        <div className="station-function-overview">
            <functions-header is="x3d">
                {statisticsData?.map((item) => (
                    <div className="statistics-box">
                        <div className="statistics-box-title">{item?.name}</div>
                        <div className="statistics-box-number">{item?.data?.toLocaleString()}</div>
                    </div>
                ))}
            </functions-header>
            <functions-list is="x3d">
                <div className="tab-functions">
                    <div className="tab-functions-inner">
                        <div className="tab-functions-inner-line">
                            {stationState?.stationMetaData?.is_native ? (
                                <>
                                    {stationState?.stationSocketData?.connected_producers?.length === 0 && (
                                        <>
                                            <Lottie animationData={dataPassLineEmptyLottie} loop={true} />
                                            <Lottie animationData={dataPassLineEmptyLottie} loop={true} />
                                        </>
                                    )}
                                    {stationState?.stationSocketData?.connected_producers?.length > 0 && (
                                        <>
                                            <Lottie animationData={dataPassLineLottie} loop={true} />
                                            <Lottie animationData={dataPassLineLottie} loop={true} />
                                        </>
                                    )}
                                </>
                            ) : (
                                <>
                                    <Lottie animationData={dataPassLineEmptyLottie} loop={true} />
                                    <Lottie animationData={dataPassLineEmptyLottie} loop={true} />
                                </>
                            )}
                        </div>
                        {stationState?.stationFunctions?.functions && stationState?.stationFunctions?.functions.length === 1 && <div></div>}
                        {stationState?.stationFunctions?.functions && stationState?.stationFunctions?.functions?.length > 0 && (
                            <div className="function-overview">
                                <div className="tab-functions-inner-left">
                                    <div className="tab-functions-inner-cards">
                                        {stationState?.stationFunctions?.functions?.map((functionItem, index) => (
                                            <FunctionCard
                                                functionItem={functionItem}
                                                stationName={stationState?.stationMetaData?.name}
                                                partiotionNumber={stationState?.stationMetaData?.partitions_number}
                                                onClick={() => setCurrentFunction(functionItem)}
                                                updatedFunctionList={(data) => stationDispatch({ type: 'UPDATE_FUNCTION_LIST', payload: data?.functions['1'] })}
                                                key={`function-tab-2-${functionItem.id}`}
                                                changeActivition={(e) => changeActivition(functionItem.id, e)}
                                                onDeleteFunction={() => handleDeleteFunction(index)}
                                                selected={currentFunction?.id === functionItem.id}
                                            />
                                        ))}
                                        <div
                                            className="tab-functions-inner-add"
                                            onClick={() => {
                                                setOpenFunctionsModal(true);
                                            }}
                                        >
                                            <PlusIcon />
                                        </div>
                                    </div>
                                    <div className="tab-functions-inner-right">
                                        <div className="processed">
                                            <div className="processed-title">Processed</div>
                                            <ProcessedIcon />
                                            <span>{stationState?.stationFunctions?.total_processed_messages?.toLocaleString() || 0}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                        {stationState?.stationFunctions?.functions?.length === 0 && (
                            <div className="functions-empty-wrap">
                                <div className="functions-empty" onClick={() => setOpenFunctionsModal(true)}>
                                    <AddFunctionIcon />
                                    <span>Add Function</span>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </functions-list>
            {currentFunction && functionDetails && (
                <div className={`ms-function-details ${currentFunction ? 'ms-function-details-entered' : 'ms-function-details-exited'}`}>
                    <div className="ms-function-details-top">
                        <div className="left">
                            <OverflowTip text={functionDetails?.function_name}>
                                <span>{functionDetails?.function_name}</span>
                            </OverflowTip>
                            <div className="ms-function-details-badge">
                                <GitIcon />
                                <OverflowTip text={functionDetails?.repo}>{functionDetails?.repo}</OverflowTip>
                            </div>
                            <div className="ms-function-details-badge">
                                <CodeGrayIcon />
                                {functionDetails?.language}
                            </div>
                        </div>
                        <div className="right">
                            <CloseIcon
                                onClick={() => {
                                    setTimeout(() => {
                                        setCurrentFunction(null);
                                    }, 500);
                                }}
                            />
                        </div>
                    </div>
                    <div className="ms-function-details-body">
                        <div className="tabs-container">
                            <Tabs defaultActiveKey="1" items={items} size="small" />
                        </div>
                    </div>
                </div>
            )}
            <Modal
                open={openFunctionsModal}
                clickOutside={() => setOpenFunctionsModal(false)}
                displayButtons={false}
                className="ms-function-details-modal"
                height="95vh"
                width="1200px"
            >
                <FunctionsModal
                    applyFunction={(requestBody) => {
                        handleAddFunction(requestBody);
                    }}
                />
            </Modal>
        </div>
    );
};

export default FunctionsOverview;
