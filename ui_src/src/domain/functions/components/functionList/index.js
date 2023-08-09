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

import React, { useEffect, useContext, useState } from 'react';
import placeholderFunctions from '../../../../assets/images/placeholderFunctions.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Loader from '../../../../components/loader';
import Button from '../../../../components/button';
import Filter from '../../../../components/filter';
import { Context } from '../../../../hooks/store';
import { useHistory } from 'react-router-dom';
import FunctionBox from '../functionBox';
import Modal from '../../../../components/modal';
import GitHubIntegration from '../../../administration/integrations/components/gitHubIntegration';
import { JSONCodec, StringCodec } from 'nats.ws';

function FunctionList() {
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [isLoading, setisLoading] = useState(true);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrated, setIntegrated] = useState(false);

    useEffect(() => {
        getAllFunctions();
        return () => {
            dispatch({ type: 'SET_FUNCTION_LIST', payload: [] });
            dispatch({ type: 'SET_FUNCTION_FILTERED_LIST', payload: [] });
        };
    }, []);

    useEffect(() => {
        const sc = StringCodec();
        const jc = JSONCodec();
        let sub;
        const subscribeToFunctions = async () => {
            try {
                const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.get_all_functions`, sc.encode('SUB'));

                if (rawBrokerName) {
                    const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                    sub = state.socket?.subscribe(`$memphis_ws_pubs.get_all_functions.${brokerName}`);
                    listenForUpdates();
                }
            } catch (err) {
                console.error('Error subscribing to overview data:', err);
            }
        };

        const listenForUpdates = async () => {
            try {
                if (sub) {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        setIntegrated(data.scm_integrated);
                        dispatch({ type: 'SET_FUNCTION_LIST', payload: data?.functions });
                        dispatch({ type: 'SET_FUNCTION_FILTERED_LIST', payload: data?.functions });
                    }
                }
            } catch (err) {
                console.error('Error receiving overview data updates:', err);
            }
        };

        subscribeToFunctions();

        return () => {
            if (sub) {
                try {
                    sub.unsubscribe();
                } catch (err) {
                    console.error('Error unsubscribing from overview data:', err);
                }
            }
        };
    }, [state.socket]);

    const getAllFunctions = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_FUNCTIONS);
            setIntegrated(data.scm_integrated);
            dispatch({ type: 'SET_FUNCTION_LIST', payload: data?.functions });
            dispatch({ type: 'SET_FUNCTION_FILTERED_LIST', payload: data?.functions });
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (error) {
            setisLoading(false);
        }
    };

    const fetchFunctions = async () => {
        getAllFunctions();
        modalFlip(false);
    };

    return (
        <div className="schema-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <label className="main-header-h1">
                        Functions <label className="length-list">{state.functionFilteredList?.length > 0 && `(${state.functionFilteredList?.length})`}</label>
                    </label>
                    <span className="memphis-label">Serverless functions to process ingested events "on the fly"</span>
                </div>
                <div className="action-section">
                    {/* <Filter filterComponent="function" height="34px" /> */}
                    {/* <Button
                        width="160px"
                        height="34px"
                        placeholder={'Create from blank'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        boxShadowStyle="float"
                        aria-haspopup="true"
                        onClick={() => {}}
                    /> */}
                </div>
            </div>

            <div className="schema-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading &&
                    state.functionFilteredList?.map((func, index) => {
                        return <FunctionBox key={index} funcDetails={func} />;
                    })}
                {!isLoading && state.functionList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={placeholderFunctions} width="150" alt="placeholderFunctions" />
                        <p className="title">No functions yet</p>
                        <p className="sub-title">Functions will start to sync and appear once an integration with a git repository is established.</p>
                        {!integrated && (
                            <Button
                                className="modal-btn"
                                width="160px"
                                height="34px"
                                placeholder="Start to integrate"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                aria-controls="usecse-menu"
                                aria-haspopup="true"
                                onClick={() => modalFlip(true)}
                            />
                        )}
                    </div>
                )}
                {!isLoading && state.functionList?.length > 0 && state.functionFilteredList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={placeholderFunctions} width="150" alt="placeholderFunctions" />
                        <p className="title">No functions found</p>
                        <p className="sub-title">Please try to search again</p>
                    </div>
                )}
            </div>
            <Modal className="integration-modal" height="95vh" width="720px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                <GitHubIntegration
                    close={(data) => {
                        if (Object.keys(data).length > 0) {
                            fetchFunctions();
                        } else {
                            modalFlip(false);
                        }
                    }}
                    value={{}}
                />
            </Modal>
        </div>
    );
}

export default FunctionList;
