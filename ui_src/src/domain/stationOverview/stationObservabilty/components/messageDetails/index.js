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
import Lottie from 'lottie-react';
import { Space } from 'antd';

import { convertBytes, parsingDate } from '../../../../../services/valueConvertor';
import { ReactComponent as AttachedPlaceholderIcon } from '../../../../../assets/images/attachedPlaceholder.svg';
import animationData from '../../../../../assets/lotties/MemphisGif.json';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { ReactComponent as JourneyIcon } from '../../../../../assets/images/journey.svg';
import { httpRequest } from '../../../../../services/http';
import Button from '../../../../../components/button';
import { StationStoreContext } from '../../..';
import CustomCollapse from '../customCollapse';
import MultiCollapse from '../multiCollapse';
import StatusIndication from '../../../../../components/indication';
import { Place } from '@material-ui/icons';

const MessageDetails = ({ isDls, isFailedSchemaMessage = false }) => {
    const url = window.location.href;
    const stationName = url.split('stations/')[1];
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [messageDetails, setMessageDetails] = useState({});
    const [loadMessageData, setLoadMessageData] = useState(false);

    const history = useHistory();

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
                    isFailedSchemaMessage ? 'schema' : 'poison'
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
                validationError: data.validation_error || ''
            };
            setMessageDetails(messageDetails);
        }
    };

    const loader = () => {
        return (
            <div className="memphis-gif">
                <Lottie animationData={animationData} loop={true} />
            </div>
        );
    };

    return (
        <>
            <div className={`message-wrapper ${isDls && !isFailedSchemaMessage && 'message-wrapper-dls'}`}>
                {loadMessageData ? (
                    loader()
                ) : stationState?.selectedRowId && Object.keys(messageDetails).length > 0 ? (
                    <>
                        <div className="row-data">
                            <Space direction="vertical">
                                {messageDetails?.validationError !== '' && (
                                    <CustomCollapse status={false} header="Validation error" data={messageDetails?.validationError} message={true} />
                                )}
                                <div className="info-box">
                                    <div>
                                        <span className="title">Producer name</span>
                                        <span className="content">{messageDetails?.producer?.details[0].value}</span>
                                    </div>
                                    <StatusIndication is_active={messageDetails?.producer.is_active} />
                                </div>

                                {!isFailedSchemaMessage && (
                                    <MultiCollapse
                                        header="Failed CGs"
                                        tooltip={!stationState?.stationMetaData?.is_native && 'Not supported without using the native Memphis SDK’s'}
                                        defaultOpen={false}
                                        data={messageDetails?.poisonedCGs}
                                    />
                                )}
                                <CustomCollapse status={false} header="Metadata" data={messageDetails?.details} />

                                <CustomCollapse status={false} header="Headers" defaultOpen={false} data={messageDetails?.headers} message={true} />
                                <CustomCollapse
                                    status={false}
                                    header="Payload"
                                    defaultOpen={true}
                                    data={messageDetails?.message}
                                    message={true}
                                    schemaType={stationState?.schemaType}
                                />
                            </Space>
                        </div>
                        {isDls && !isFailedSchemaMessage && (
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
                    </>
                ) : (
                    <div className="placeholder">
                        <AttachedPlaceholderIcon />
                        <p>No message selected</p>
                    </div>
                )}
            </div>
        </>
    );
};

export default MessageDetails;
