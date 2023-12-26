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
import { useHistory } from 'react-router-dom';
import { convertBytes, parsingDate, messageParser } from '../../../../../services/valueConvertor';
import { httpRequest } from '../../../../../services/http';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { ReactComponent as JourneyIcon } from '../../../../../assets/images/journey.svg';
import { CiViewList } from 'react-icons/ci';
import SegmentButton from '../../../../../components/segmentButton';
import StatusIndication from '../../../../../components/indication';
import Spinner from '../../../../../components/spinner';
import Button from '../../../../../components/button';
import Copy from '../../../../../components/copy';
import { LOCAL_STORAGE_MSG_PARSER } from '../../../../../const/localStorageConsts';
import Editor, { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import MultiCollapse from '../multiCollapse';
import { StationStoreContext } from '../../..';
import { Drawer } from 'antd';

loader.init();
loader.config({ monaco });

const MessageDetails = ({ open, isDls, unselect, isFailedSchemaMessage = false, isFailedFunctionMessage = false }) => {
    const url = window.location.href;
    const stationName = url.split('stations/')[1];
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [messageDetails, setMessageDetails] = useState({});
    const [loadMessageData, setLoadMessageData] = useState(false);
    const [payloadType, setPayloadType] = useState('string');

    const history = useHistory();

    useEffect(() => {
        if (localStorage.getItem(LOCAL_STORAGE_MSG_PARSER) !== null) setPayloadType(localStorage.getItem(LOCAL_STORAGE_MSG_PARSER));
    }, []);

    useEffect(() => {
        if (Object.keys(messageDetails).length !== 0) {
            setLoadMessageData(false);
        }
        return () => {};
    }, [messageDetails]);

    useEffect(() => {
        if ((isDls && stationState?.selectedRowId && stationState?.selectedRowPartition && !loadMessageData) || (stationState?.selectedRowId && !loadMessageData)) {
            getMessageDetails(stationState?.selectedRowId, stationState?.selectedRowPartition === 0 ? -1 : stationState?.selectedRowPartition);
        }
    }, [stationState?.selectedRowId, stationState?.selectedRowPartition]);

    const getMessageDetails = async (selectedRow, selectedRowPartition) => {
        setMessageDetails({});
        setLoadMessageData(true);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_MESSAGE_DETAILS}?dls_type=${
                    isFailedFunctionMessage ? 'functions' : isFailedSchemaMessage ? 'schema' : 'poison'
                }&station_name=${stationName}&is_dls=${isDls}&partition_number=${selectedRowPartition}&message_id=${isDls ? parseInt(selectedRow) : -1}&message_seq=${
                    isDls ? -1 : selectedRow
                }`
            );
            arrangeData(data);
        } catch (error) {
            setLoadMessageData(false);
        }
    };

    const arrangeData = async (data) => {
        let poisonedCGs = [];
        if (data) {
            data?.poisoned_cgs?.map((row) => {
                let cg = {
                    name: row.cg_name,
                    is_active: row.is_active,
                    is_deleted: row.is_deleted,
                    details: [
                        {
                            name: 'Unacked messages',
                            value: row?.total_poison_messages?.toLocaleString()
                        },
                        {
                            name: 'Unprocessed messages',
                            value: row?.unprocessed_messages?.toLocaleString()
                        },
                        {
                            name: 'In process message',
                            value: row?.in_process_messages?.toLocaleString()
                        },
                        {
                            name: 'Max ack time',
                            value: `${row?.max_ack_time_ms?.toLocaleString()}ms`
                        },
                        {
                            name: 'Max message deliveries',
                            value: row?.max_msg_deliveries
                        }
                    ]
                };

                poisonedCGs.push(cg);
            });
            let messageDetails = {
                id: data.id ?? null,
                message_seq: data.message_seq,
                details: [
                    {
                        name: 'Message size',
                        value: convertBytes(data.message?.size)
                    },
                    {
                        name: 'Time sent',
                        value: parsingDate(data.message?.time_sent, true)
                    }
                ],
                producer: {
                    is_active: data?.producer?.is_active,
                    is_deleted: data?.producer?.is_deleted,
                    details: [
                        {
                            name: 'Name',
                            value: data.producer?.name || ''
                        },
                        {
                            name: 'User',
                            value: data.producer?.created_by_username || ''
                        },
                        {
                            name: 'IP',
                            value: data.producer?.client_address || ''
                        }
                    ]
                },
                message: data.message?.data,
                headers: data.message?.headers || {},
                poisonedCGs: poisonedCGs,
                validationError: data.validation_error || '',
                function_name: data.function_name || ''
            };
            setMessageDetails(messageDetails);
        }
    };

    const generateEditor = (langCode, value) => {
        return (
            <>
                <Editor
                    options={{
                        minimap: { enabled: false },
                        scrollbar: { verticalScrollbarSize: 0, horizontalScrollbarSize: 0 },
                        scrollBeyondLastLine: false,
                        roundedSelection: false,
                        formatOnPaste: true,
                        formatOnType: true,
                        readOnly: true,
                        fontSize: '12px',
                        fontFamily: 'Inter',
                        height: '100%',
                        minHeight: 'fit-content',
                        lineNumbers: 'off',
                        glyphMargin: false,
                        lineDecorationsWidth: 0,
                        lineNumbersMinChars: 0
                    }}
                    language={langCode}
                    height="100%"
                    width="100%"
                    value={value}
                />
                <Copy data={value} />
            </>
        );
    };

    const loader = <Spinner fontSize={60} style={{ display: 'flex', justifyContent: 'center' }} />;

    const messageDetailsItem = (props) => {
        const { title, value, showIsActive, is_active, headers, details, payload, cg } = props;
        const keysArray = headers ? Object.keys(value) : null;
        return (
            <div className="message-detail-item">
                <span className="title">
                    <CiViewList style={{ color: '#6557FF' }} />
                    {title}
                </span>
                {!headers && !details && !payload && !cg && (
                    <span className="content">
                        {value} {showIsActive && <StatusIndication is_active={is_active} />}
                        <Copy data={value} />
                    </span>
                )}
                {headers &&
                    keysArray &&
                    keysArray.map((item) => (
                        <span key={item} className="content">
                            <label>{item}</label>
                            <label className="val">{value[item]}</label>
                            <Copy data={value[item]} />
                        </span>
                    ))}
                {details &&
                    value?.map((item) => (
                        <span key={item.name} className="content">
                            <label>{item.name}</label>
                            <label className="val">{item.value}</label>
                            <Copy data={`${item.name} ${item.value}`} />
                        </span>
                    ))}
                {payload && (
                    <>
                        <SegmentButton
                            value={payloadType || 'string'}
                            options={['string', 'bytes', 'json', 'protobuf']}
                            onChange={(e) => {
                                setPayloadType(e);
                                localStorage.setItem(LOCAL_STORAGE_MSG_PARSER, e);
                            }}
                        />
                        <span className="content content-json">{generateEditor('json', messageParser(payloadType, value))}</span>
                    </>
                )}

                {cg && (
                    <MultiCollapse
                        header="Failed CGs"
                        tooltip={!stationState?.stationMetaData?.is_native && 'Not supported without using the native Memphis SDK’s'}
                        defaultOpen={false}
                        data={messageDetails?.poisonedCGs}
                    />
                )}
            </div>
        );
    };

    return (
        <div onClick={(e) => e.stopPropagation()}>
            <Drawer
                placement="right"
                title="Message details"
                onClose={() => {
                    stationDispatch({ type: 'SET_SELECTED_ROW_ID', payload: null });
                    unselect();
                }}
                destroyOnClose={true}
                width="450px"
                open={open}
                mask={false}
                bodyStyle={{ height: '100%', position: 'relative' }}
            >
                {loadMessageData && loader}
                {!loadMessageData && stationState?.selectedRowId && Object.keys(messageDetails).length > 0 && (
                    <div className={`message-wrapper ${isDls && !isFailedSchemaMessage && !isFailedFunctionMessage ? 'message-wrapper-dls' : undefined}`}>
                        <span className={`${isDls && !isFailedSchemaMessage && !isFailedFunctionMessage ? 'content-wrapper-dls' : 'content-wrapper'}`}>
                            {isFailedFunctionMessage &&
                                messageDetails?.validationError !== '' &&
                                messageDetailsItem({ title: 'Error', value: messageDetails?.validationError, error: true })}
                            {!isFailedFunctionMessage &&
                                messageDetails?.validationError !== '' &&
                                messageDetailsItem({ title: 'Validation error', value: messageDetails?.validationError, error: true })}
                            {isFailedFunctionMessage &&
                                messageDetails?.function_name !== '' &&
                                messageDetailsItem({ title: 'Function', value: messageDetails?.function_name })}
                            {messageDetailsItem({
                                title: 'Producer name',
                                value: messageDetails?.producer?.details[0].value,
                                showIsActive: true,
                                is_active: messageDetails?.producer.is_active
                            })}

                            {isDls &&
                                !isFailedSchemaMessage &&
                                !isFailedFunctionMessage &&
                                messageDetailsItem({ title: 'Failed CGs', value: messageDetails?.poisonedCGs, cg: true })}
                            {messageDetailsItem({ title: 'Metadata', value: messageDetails?.details, details: true })}
                            {messageDetails?.headers &&
                                Object.keys(messageDetails?.headers)?.length > 0 &&
                                messageDetailsItem({ title: 'Headers', value: messageDetails?.headers, headers: true })}
                            {messageDetailsItem({ title: 'Payload', value: messageDetails?.message, payload: true })}
                        </span>
                        {isDls && !isFailedSchemaMessage && !isFailedFunctionMessage && (
                            <Button
                                width="96%"
                                height="40px"
                                placeholder={
                                    <div className="botton-title">
                                        <JourneyIcon alt="Journey" />
                                        <p>Message Journey</p>
                                    </div>
                                }
                                colorType="black"
                                radiusType="semi-round"
                                backgroundColorType="orange"
                                fontSize="12px"
                                fontWeight="600"
                                tooltip={!stationState?.stationMetaData?.is_native && 'Not supported without using the native Memphis SDK’s'}
                                disabled={!stationState?.stationMetaData?.is_native || !messageDetails?.id}
                                onClick={() => history.push(`${window.location.pathname}/${messageDetails?.id}`)}
                            />
                        )}
                    </div>
                )}
            </Drawer>
        </div>
    );
};

export default MessageDetails;
