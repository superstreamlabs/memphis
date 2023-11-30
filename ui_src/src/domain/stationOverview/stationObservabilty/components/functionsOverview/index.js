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
import { StationStoreContext } from '../../../';
import { ReactComponent as AddFunctionIcon } from '../../../../../assets/images/addFunction.svg';
import { ReactComponent as PlusIcon } from '../../../../../assets/images/plusIcon.svg';
import { ReactComponent as ProcessedIcon } from '../../../../../assets/images/processIcon.svg';
import { IoClose } from 'react-icons/io5';
import { Drawer } from 'antd';
import dataPassLineLottie from '../../../../../assets/lotties/dataPassLine.json';
import dataPassLineEmptyLottie from '../../../../../assets/lotties/dataPassLineEmpty.json';
import Lottie from 'lottie-react';
import FunctionCard from '../functionCard';
import FunctionsModal from '../functionsModal';
import FunctionData from '../functionData';
import FunctionDetails from '../../../../functions/components/functionDetails';
import OverflowTip from '../../../../../components/tooltip/overflowtip';
import { StringCodec, JSONCodec } from 'nats.ws';
import Spinner from '../../../../../components/spinner';

let sub;

const FunctionsOverview = ({ referredFunction, dismissFunction, moveToGenralView }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [currentFunction, setCurrentFunction] = useState(null);
    const [functionDetails, setFunctionDetails] = useState(null);
    const [openFunctionsModal, setOpenFunctionsModal] = useState(false);
    const [openFunctionDetails, setOpenFunctionDetails] = useState(false);
    const [openBottomDetails, setOpenBottomDetails] = useState(false);
    const [socketOn, setSocketOn] = useState(false);
    const [isLoading, setLoading] = useState(false);
    const [state, dispatch] = useContext(Context);

    useEffect(() => {
        getFunctionsOverview();
        referredFunction && setOpenFunctionsModal(true);
    }, []);

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
        stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: [] });
        getFunctionsOverview();
    }, [stationState?.stationPartition]);

    useEffect(() => {
        if (sub && socketOn) {
            stopListen();
            startListen();
        }
    }, [stationState?.stationPartition, stationState?.stationMetaData?.name]);

    const getFunctionsOverview = async () => {
        setLoading(true);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_FUNCTIONS_OVERVIEW}?station_name=${stationState?.stationMetaData?.name}&partition=${stationState?.stationPartition || -1}`
            );
            stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: data });
            setLoading(false);
        } catch (e) {
            setLoading(false);
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
        currentFunction && getFunctionDetails();
    }, [currentFunction]);

    const handleAddFunction = async (requestBody) => {
        requestBody.station_name = stationState?.stationMetaData?.name;
        requestBody.partition = stationState?.stationPartition || -1;
        try {
            await httpRequest('POST', ApiEndpoints.ADD_FUNCTION, requestBody);
            getFunctionsOverview();
        } catch (e) {
        } finally {
            setFunctionDetails(null);
            setOpenFunctionsModal(false);
            dismissFunction();
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

    const isDataMoving = stationState?.stationSocketData?.connected_producers?.length > 0 && stationState?.stationMetaData?.is_native;

    const statisticsData = [
        { name: 'Awaiting msgs', data: stationState?.stationFunctions?.total_awaiting_messages?.toLocaleString() },
        { name: 'In process', data: stationState?.stationFunctions?.total_processed_messages?.toLocaleString() },
        { name: 'Total invocations', data: stationState?.stationFunctions?.total_invocations?.toLocaleString() },
        { name: 'Avg Error rate', data: stationState?.stationFunctions?.average_error_rate },
        {
            name: 'Ordering',
            data: stationState?.stationFunctions?.functions?.length > 0 ? (stationState?.stationFunctions?.functions[0]?.ordering_matter ? 'Yes' : 'No') : 'N/A'
        }
    ];
    return (
        <div className="station-function-overview">
            <functions-header is="x3d">
                {statisticsData?.map((item, index) => (
                    <div className="statistics-box" key={`${item?.name}-${index}`}>
                        <div className="statistics-box-title">{item?.name}</div>
                        <div className="statistics-box-number">{item?.data?.toLocaleString()}</div>
                    </div>
                ))}
            </functions-header>
            <functions-list is="x3d">
                <div className="tab-functions">
                    <div className="tab-functions-inner">
                        <div className="tab-functions-inner-line">
                            <Lottie animationData={isDataMoving ? dataPassLineLottie : dataPassLineEmptyLottie} loop={true} />
                            <Lottie animationData={isDataMoving ? dataPassLineLottie : dataPassLineEmptyLottie} loop={true} />
                        </div>
                        {isLoading && (
                            <div className="loading">
                                <Spinner />
                            </div>
                        )}
                        {stationState?.stationFunctions?.functions && stationState?.stationFunctions?.functions?.length > 0 && (
                            <div className="function-overview">
                                <div className="tab-functions-inner-left">
                                    <div className={stationState?.stationFunctions?.functions?.length < 2 ? `tab-functions-inner-one-card` : `tab-functions-inner-cards`}>
                                        {stationState?.stationFunctions?.functions?.map((functionItem, index) => (
                                            <FunctionCard
                                                functionItem={functionItem}
                                                stationName={stationState?.stationMetaData?.name}
                                                partiotionNumber={stationState?.stationPartition || -1}
                                                onClick={() => {
                                                    setCurrentFunction(functionItem);
                                                    setOpenBottomDetails(true);
                                                }}
                                                onClickMenu={() => setCurrentFunction(functionItem)}
                                                updatedFunctionList={(data) => stationDispatch({ type: 'UPDATE_FUNCTION_LIST', payload: data?.functions['1'] })}
                                                key={`${functionItem?.id}-${index}`}
                                                changeActivition={(e) => changeActivition(functionItem?.id, e)}
                                                onDeleteFunction={() => handleDeleteFunction(index)}
                                                selected={currentFunction?.id === functionItem?.id}
                                            />
                                        ))}
                                        <div className="tab-functions-inner-add" onClick={() => setOpenFunctionsModal(true)}>
                                            <PlusIcon />
                                        </div>
                                    </div>
                                    <div className="tab-functions-inner-right">
                                        <div className="processed" onClick={moveToGenralView}>
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
            <FunctionsModal
                applyFunction={(requestBody) => handleAddFunction(requestBody)}
                open={openFunctionsModal}
                clickOutside={() => {
                    setOpenFunctionsModal(false);
                    dismissFunction();
                }}
                referredFunction={referredFunction}
            />
            <Drawer
                placement="right"
                size={'large'}
                className="function-drawer"
                onClose={() => setOpenFunctionDetails(false)}
                destroyOnClose={true}
                open={openFunctionDetails}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<IoClose style={{ color: '#D1D1D1', width: '25px', height: '25px' }} />}
            >
                <FunctionDetails selectedFunction={currentFunction} integrated={true} stationView />
            </Drawer>
            <FunctionData
                open={openBottomDetails}
                onClose={() => setOpenBottomDetails(false)}
                functionDetails={functionDetails}
                setOpenFunctionDetails={() => setOpenFunctionDetails(true)}
            />
        </div>
    );
};

export default FunctionsOverview;
