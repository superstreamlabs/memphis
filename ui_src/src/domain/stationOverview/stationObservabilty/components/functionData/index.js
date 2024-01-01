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
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import Editor, { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import { StationStoreContext } from '../../../';
import CustomTabs from 'components/Tabs';
import FunctionLogs from '../functionLogs';
import FunctionInformation from '../functionInformation';
import { ReactComponent as MetricsIcon } from 'assets/images/metricsIcon.svg';
import { ReactComponent as MetricsClockIcon } from 'assets/images/metricsClockIcon.svg';
import { ReactComponent as MetricsErrorIcon } from 'assets/images/metricsErrorIcon.svg';
import { ReactComponent as OrderingIcon } from 'assets/images/orderingIcon.svg';
import { ReactComponent as GitIcon } from 'assets/images/gitIcon.svg';
import { ReactComponent as CodeGrayIcon } from 'assets/images/codeGrayIcon.svg';
import { ReactComponent as PurpleQuestionMark } from 'assets/images/purpleQuestionMark.svg';
import { parsingDate, messageParser } from 'services/valueConvertor';
import Spinner from 'components/spinner';
import Drawer from "components/drawer";
import { IoClose } from 'react-icons/io5';
import OverflowTip from 'components/tooltip/overflowtip';

const tabValuesList = ['Monitoring', 'Information', 'Dead-letter', 'Logs'];
loader.init();
loader.config({ monaco });

const FunctionData = ({ open, onClose, setOpenFunctionDetails, functionDetails }) => {
    const [tabValue, setTabValue] = useState('Information');
    const [attachedFunctionDlsMsgs, setAttachedFunctionDlsMsgs] = useState([]);
    const [selectedMsg, setSelectedMsg] = useState(null);
    const [messageDetails, setMessageDetails] = useState({});
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [loadMessageData, setLoadMessageData] = useState(false);

    useEffect(() => {
        if (open) {
            setTabValue('Monitoring');
        }
    }, [open]);

    useEffect(() => {
        tabValue === tabValuesList[2] && getAttachedFunctionDlsMsgs();
    }, [tabValue]);

    useEffect(() => {
        if (attachedFunctionDlsMsgs?.length > 0) {
            setSelectedMsg(attachedFunctionDlsMsgs[0]);
        }
    }, [attachedFunctionDlsMsgs]);

    const getMessageDetails = async () => {
        setLoadMessageData(true);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_MESSAGE_DETAILS}?dls_type=functions&station_name=${stationState?.stationMetaData?.name}&is_dls=true&partition_number=${stationState?.stationPartition}&message_id=${selectedMsg?.id}&message_seq=${selectedMsg?.message_seq}&function_id=${functionDetails?.function?.id}&row_number=-1`
            );
            arrangeData(data);
        } catch (error) {
            setLoadMessageData(false);
        }
    };

    const arrangeData = (data) => {
        let updatedData = {};
        updatedData['function_name'] = data?.function_name;
        updatedData['message'] = data?.message;
        updatedData['message']['data'] = messageParser('string', data?.message?.data);
        updatedData['producer'] = data?.producer;
        updatedData['error'] = data?.validation_error;
        setMessageDetails(updatedData);
        setLoadMessageData(false);
    };

    useEffect(() => {
        selectedMsg && getMessageDetails();
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
        <Drawer
            placement="bottom"
            open={open}
            height={'300px'}
            onClose={onClose}
            closeIcon={<IoClose style={{ color: '#D1D1D1', width: '25px', height: '25px' }} />}
            maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
            headerStyle={{ padding: '0px' }}
            bodyStyle={{ padding: '0 20px' }}
            destroyOnClose={true}
            title={
                <div className="ms-function-details-top">
                    <div className="left">
                        <OverflowTip text={functionDetails?.function?.function_name}>
                            <span>{functionDetails?.function?.function_name}</span>
                        </OverflowTip>
                        <div className="ms-function-details-badge">
                            <GitIcon />
                            <OverflowTip text={functionDetails?.function?.repo}>{functionDetails?.function?.repo}</OverflowTip>
                        </div>
                        <div className="ms-function-details-badge">
                            <CodeGrayIcon />
                            {functionDetails?.function?.language}
                        </div>
                    </div>
                    <div className="right">
                        <PurpleQuestionMark className="info-icon" alt="Integration info" onClick={setOpenFunctionDetails} />
                    </div>
                </div>
            }
        >
            <div className="function-data-container">
                <CustomTabs tabs={tabValuesList} size={'small'} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
                {tabValue === tabValuesList[0] && (
                    <div className="metrics-wrapper">
                        <div className="metrics">
                            <MetricsIcon width="25" height="25" />
                            <div className="metrics-body">
                                <div className="metrics-body-title">Total invocations</div>
                                <div className="metrics-body-subtitle">{functionDetails?.metrics?.total_invocations?.toLocaleString() || 0}</div>
                            </div>
                        </div>

                        <div className="metrics">
                            <MetricsClockIcon width="25" height="25" />
                            <div className="metrics-body">
                                <div className="metrics-body-title">Av. Processing time</div>
                                <div className="metrics-body-subtitle">
                                    {functionDetails?.metrics?.average_processing_time}
                                    <span className="ms">/ms</span>
                                </div>
                            </div>
                        </div>
                        <div className="metrics">
                            <MetricsErrorIcon width="25" height="25" />
                            <div className="metrics-body">
                                <div className="metrics-body-title">Error rate</div>
                                <div className="metrics-body-subtitle">{functionDetails?.metrics?.error_rate}%</div>
                            </div>
                        </div>
                        <div className="metrics">
                            <OrderingIcon width="25" height="25" />
                            <div className="metrics-body">
                                <div className="metrics-body-title">Ordering</div>
                                <div className="metrics-body-subtitle">{functionDetails?.ordering_matter ? 'Yes' : 'No'}</div>
                            </div>
                        </div>
                    </div>
                )}
                {tabValue === tabValuesList[1] && <FunctionInformation inputs={functionDetails?.function?.inputs || []} />}
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
                                            value={JSON.stringify(messageDetails, null, 2)}
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
                {tabValue === tabValuesList[3] && <FunctionLogs functionId={functionDetails?.function?.id} />}
            </div>
        </Drawer>
    );
};

export default FunctionData;
