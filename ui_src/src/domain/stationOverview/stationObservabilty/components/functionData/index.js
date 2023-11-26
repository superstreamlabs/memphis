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

import React, { useEffect, useState, useContext } from 'react';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Editor, { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import { StationStoreContext } from '../../../';
import CustomTabs from '../../../../../components/Tabs';
import FunctionLogs from '../functionLogs';
import { ReactComponent as MetricsIcon } from '../../../../../assets/images/metricsIcon.svg';
import { ReactComponent as MetricsClockIcon } from '../../../../../assets/images/metricsClockIcon.svg';
import { ReactComponent as MetricsErrorIcon } from '../../../../../assets/images/metricsErrorIcon.svg';
import { parsingDate, messageParser } from '../../../../../services/valueConvertor';
import Spinner from '../../../../../components/spinner';
const tabValuesList = ['Information', 'Logs', 'Dead-letter'];
loader.init();
loader.config({ monaco });

const FunctionData = ({ functionDetails }) => {
    const [tabValue, setTabValue] = useState('Information');
    const [attachedFunctionDlsMsgs, setAttachedFunctionDlsMsgs] = useState([]);
    const [selectedMsg, setSelectedMsg] = useState(null);
    const [messageDetails, setMessageDetails] = useState({});
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [loadMessageData, setLoadMessageData] = useState(false);

    useEffect(() => {
        tabValue === tabValuesList[2] && getAttachedFunctionDlsMsgs();
    }, [tabValue]);

    useEffect(() => {
        if (attachedFunctionDlsMsgs?.length > 0) {
            setSelectedMsg(attachedFunctionDlsMsgs[0]);
        }
    }, [attachedFunctionDlsMsgs]);

    const getMessageDetails = async () => {
        setMessageDetails({});
        setLoadMessageData(true);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_MESSAGE_DETAILS}?dls_type=functions&station_name=${stationState?.stationMetaData?.name}&is_dls=true&partition_number=${stationState?.stationPartition}&message_id=${selectedMsg?.id}&message_seq=${selectedMsg?.message_seq}&function_id=${functionDetails?.function?.id}&row_number=-1`
            );
            setMessageDetails(data);
        } catch (error) {
            setLoadMessageData(false);
        }
    };

    useEffect(() => {
        getMessageDetails();
    }, [selectedMsg]);

    const getAttachedFunctionDlsMsgs = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_ATTACHED_FUNCTION_DLS_MSG}?function_id=${functionDetails?.function?.id}`);
            setAttachedFunctionDlsMsgs(data?.dls_messages);
        } catch (e) {
            return;
        }
    };

    return (
        <div className="function-data-container">
            <CustomTabs tabs={tabValuesList} size={'small'} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
            {tabValue === tabValuesList[0] && (
                <div className="metrics-wrapper">
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Total invocations</div>
                            <div className="metrics-body-subtitle">{functionDetails?.metrics?.total_invocations?.toLocaleString() || 0}</div>
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
                                {functionDetails?.metrics?.average_processing_time}
                                <span>/ms</span>
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
                            <div className="metrics-body-subtitle">{functionDetails?.metrics?.error_rate}%</div>
                        </div>
                    </div>
                </div>
            )}
            {tabValue === tabValuesList[1] && <FunctionLogs functionId={functionDetails?.function?.id} />}
            {tabValue === tabValuesList[2] && (
                <dls is="x3d">
                    {attachedFunctionDlsMsgs && attachedFunctionDlsMsgs?.length > 0 ? (
                        <>
                            <list is="x3d">
                                <div className="msg-item-header">
                                    <label className="date">Time</label>
                                    <label className="text">Text</label>
                                </div>
                                <div className="messages-list">
                                    {attachedFunctionDlsMsgs?.map((message, index) => {
                                        return (
                                            <div
                                                className={`msg-item ${index % 2 === 0 ? 'even' : 'odd'} ${selectedMsg?.id === message?.id ? 'selected' : undefined}`}
                                                onClick={() => setSelectedMsg(message)}
                                                key={`dls-${index}`}
                                            >
                                                <label className="date">{parsingDate(message?.message?.time_sent, true, true)}</label>
                                                <label className="text">{messageParser('json', message?.message?.data)}</label>
                                            </div>
                                        );
                                    })}
                                </div>
                            </list>
                            <preview is="x3d">
                                {selectedMsg && loadMessageData && (
                                    <div className="loading">
                                        <Spinner />
                                    </div>
                                )}
                                {selectedMsg && !loadMessageData && (
                                    <Editor
                                        options={{
                                            minimap: { enabled: false },
                                            scrollbar: { verticalScrollbarSize: 0, horizontalScrollbarSize: 0 },
                                            scrollBeyondLastLine: false,
                                            roundedSelection: false,
                                            formatOnPaste: true,
                                            formatOnType: true,
                                            fontSize: '12px',
                                            fontFamily: 'Inter',
                                            lineNumbers: 'off',
                                            readOnly: true
                                        }}
                                        className="editor-message"
                                        language={'json'}
                                        height="calc(100%)"
                                        width="calc(100%)"
                                        value={JSON.stringify(messageDetails?.message, null, 2)}
                                    />
                                )}
                            </preview>
                        </>
                    ) : (
                        <div className="no-messages">
                            <p>No messages to show</p>
                        </div>
                    )}
                </dls>
            )}
        </div>
    );
};

export default FunctionData;
