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
import { message } from 'antd';

import React, { useContext, useEffect, useState } from 'react';
import { InfoOutlined } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import { Checkbox, Space } from 'antd';

import { convertBytes, numberWithCommas, parsingDate } from '../../../../services/valueConvertor';
import waitingMessages from '../../../../assets/images/waitingMessages.svg';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import Journey from '../../../../assets/images/journey.svg';
import CustomCollapse from '../components/customCollapse';
import MultiCollapse from '../components/multiCollapse';
import { httpRequest } from '../../../../services/http';
import CustomTabs from '../../../../components/Tabs';
import Button from '../../../../components/button';
import { StationStoreContext } from '../..';
import pathDomains from '../../../../router';

const Messages = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [selectedRowIndex, setSelectedRowIndex] = useState(0);
    const [isCheck, setIsCheck] = useState([]);
    const [messageDetails, setMessageDetails] = useState({});
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [resendProcced, setResendProcced] = useState(false);
    const [ignoreProcced, setIgnoreProcced] = useState(false);
    const [loadMessageData, setLoadMessageData] = useState(false);
    const url = window.location.href;
    const stationName = url.split('stations/')[1];

    const [tabValue, setTabValue] = useState('0');
    const tabs = ['All', 'Dead-letter'];
    const history = useHistory();

    useEffect(() => {
        if (stationState?.stationSocketData?.messages?.length > 0 && Object.keys(messageDetails).length === 0) {
            getMessageDetails(false, null, stationState?.stationSocketData?.messages[0]?.message_seq);
        } else if (tabValue === '0' && stationState?.stationSocketData?.messages?.length > 0) {
            getMessageDetails(false, null, stationState?.stationSocketData?.messages[0]?.message_seq);
        }
        if (tabValue === '1' && stationState?.stationSocketData?.poison_messages?.length > 0) {
            getMessageDetails(true, stationState?.stationSocketData?.poison_messages[0]?._id, null);
        }
    }, [stationState?.stationSocketData?.messages, stationState?.stationSocketData?.poison_messages]);

    const getMessageDetails = async (isPoisonMessage, messageId = null, message_seq = null) => {
        setLoadMessageData(true);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_MESSAGE_DETAILS}?station_name=${stationName}&is_poison_message=${isPoisonMessage}&message_id=${messageId}&message_seq=${message_seq}`
            );
            arrangeData(data);
            setTimeout(() => {
                setLoadMessageData(false);
            }, 1000);
        } catch (error) {
            setLoadMessageData(false);
        }
    };

    const arrangeData = (data) => {
        let poisionedCGs = [];
        if (data) {
            data.poisoned_cgs.map((row, index) => {
                let cg = {
                    name: row.cg_name,
                    is_active: row.is_active,
                    is_deleted: row.is_deleted,
                    details: [
                        {
                            name: 'Poison messages',
                            value: numberWithCommas(row?.total_poison_messages)
                        },
                        {
                            name: 'Unprocessed messages',
                            value: numberWithCommas(row?.unprocessed_messages)
                        },
                        {
                            name: 'In process message',
                            value: numberWithCommas(row?.in_process_messages)
                        },
                        {
                            name: 'Max ack time',
                            value: `${numberWithCommas(row?.max_ack_time_ms)}ms`
                        },
                        {
                            name: 'Max message deliveries',
                            value: row?.max_msg_deliveries
                        }
                    ]
                };
                poisionedCGs.push(cg);
            });
            let messageDetails = {
                id: data._id ?? null,
                messageSeq: data.message_seq,
                details: [
                    {
                        name: 'Message size',
                        value: convertBytes(data.message?.size)
                    },
                    {
                        name: 'Time sent',
                        value: parsingDate(data.message?.time_sent)
                    }
                ],
                producer: {
                    is_active: data.producer?.is_active,
                    is_deleted: data.producer?.is_deleted,
                    details: [
                        {
                            name: 'Name',
                            value: data.producer?.name
                        },
                        {
                            name: 'User',
                            value: data.producer?.created_by_user
                        },
                        {
                            name: 'IP',
                            value: data.producer?.client_address
                        }
                    ]
                },
                message: data.message?.data,
                poisionedCGs: poisionedCGs
            };
            setMessageDetails(messageDetails);
        }
    };

    const onSelectedRow = (isPoisonMessage, id, rowIndex) => {
        setSelectedRowIndex(rowIndex);
        getMessageDetails(isPoisonMessage, isPoisonMessage ? id : null, isPoisonMessage ? null : id);
    };

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(stationState?.stationSocketData?.poison_messages.map((li) => li._id));
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        setIsCheck([...isCheck, id]);
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
        }
        if (isCheck.length === 1 && !checked) {
            setIsCheckAll(false);
        }
    };

    const handleChangeMenuItem = (newValue) => {
        if (newValue === '0' && stationState?.stationSocketData?.messages?.length > 0) {
            getMessageDetails(false, null, stationState?.stationSocketData?.messages[0]?.message_seq);
        }
        if (newValue === '1' && stationState?.stationSocketData?.poison_messages?.length > 0) {
            getMessageDetails(true, stationState?.stationSocketData?.poison_messages[0]?._id, null);
        }
        setTabValue(newValue);
        setSelectedRowIndex(0);
    };

    const handleAck = async () => {
        setIgnoreProcced(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.ACK_POISION_MESSAGE}`, { poison_message_ids: isCheck });
            let poisions = stationState?.stationSocketData?.poison_messages;
            isCheck.map((messageId, index) => {
                poisions = poisions?.filter((item) => {
                    return item._id !== messageId;
                });
            });
            setTimeout(() => {
                setIgnoreProcced(false);
                stationDispatch({ type: 'SET_POISINS_MESSAGES', payload: poisions });
                setIsCheck([]);
                setIsCheckAll(false);
            }, 1500);
        } catch (error) {
            setIgnoreProcced(false);
        }
    };

    const handleResend = async () => {
        setResendProcced(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.RESEND_POISION_MESSAGE_JOURNEY}`, { poison_message_ids: isCheck });
            setTimeout(() => {
                setResendProcced(false);
                message.success({
                    key: 'memphisSuccessMessage',
                    content: isCheck.length === 1 ? 'The message was sent successfully' : 'The messages were sent successfully',
                    duration: 5,
                    style: { cursor: 'pointer' },
                    onClick: () => message.destroy('memphisSuccessMessage')
                });
                setIsCheck([]);
                setIsCheckAll(false);
            }, 1500);
        } catch (error) {
            setResendProcced(false);
        }
    };

    return (
        <div className="messages-container">
            <div className="header">
                <div className="left-side">
                    <p className="title">Station</p>
                    {tabValue === '0' && stationState?.stationSocketData?.messages?.length > 0 && (
                        <div className="messages-amount">
                            <InfoOutlined />
                            <p>Showing last {stationState?.stationSocketData?.messages?.length} messages</p>
                        </div>
                    )}
                    {tabValue === '1' && stationState?.stationSocketData?.poison_messages?.length > 0 && (
                        <div className="messages-amount">
                            <InfoOutlined />
                            <p>Showing last {stationState?.stationSocketData?.poison_messages?.length} messages</p>
                        </div>
                    )}
                </div>
                {tabValue === '1' && (
                    <div className="right-side">
                        <Button
                            width="80px"
                            height="32px"
                            placeholder="Drop"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            disabled={isCheck.length === 0}
                            isLoading={ignoreProcced}
                            onClick={() => handleAck()}
                        />
                        <Button
                            width="100px"
                            height="32px"
                            placeholder="Resend"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            disabled={isCheck.length === 0}
                            isLoading={resendProcced}
                            onClick={() => handleResend()}
                        />
                    </div>
                )}
            </div>
            <div className="tabs">
                <CustomTabs
                    value={tabValue}
                    onChange={handleChangeMenuItem}
                    tabs={tabs}
                    disabled={stationState?.stationSocketData?.poison_messages?.length === 0}
                ></CustomTabs>
            </div>
            {tabValue === '0' && stationState?.stationSocketData?.messages?.length > 0 && (
                <div className="list-wrapper">
                    <div className="coulmns-table">
                        <div className="left-coulmn all">
                            <p>Message</p>
                        </div>
                        <p className="right-coulmn">Details</p>
                    </div>
                    <div className="list">
                        <div className="rows-wrapper all">
                            {stationState?.stationSocketData?.messages?.map((message, id) => {
                                return (
                                    <div
                                        className={selectedRowIndex === id ? 'message-row selected' : 'message-row'}
                                        key={id}
                                        onClick={() => onSelectedRow(false, message.message_seq, id)}
                                    >
                                        <OverflowTip text={message?.data} width={'300px'}>
                                            {message?.data}
                                        </OverflowTip>
                                    </div>
                                );
                            })}
                        </div>
                        <div className="message-wrapper">
                            <div className="row-data">
                                <Space direction="vertical">
                                    <CustomCollapse header="Producer" status={true} data={messageDetails.producer} />
                                    <MultiCollapse header="Failed CGs" defaultOpen={true} data={messageDetails.poisionedCGs} />
                                    <CustomCollapse status={false} header="Details" data={messageDetails.details} />
                                    <CustomCollapse status={false} header="Payload" defaultOpen={true} data={messageDetails.message} message={true} />
                                </Space>
                            </div>
                        </div>
                    </div>
                </div>
            )}
            {tabValue === '1' && stationState?.stationSocketData?.poison_messages?.length > 0 && (
                <div className="list-wrapper">
                    <div className="coulmns-table">
                        <div className="left-coulmn">
                            <Checkbox checked={isCheckAll} id="selectAll" onChange={onCheckedAll} name="selectAll" />
                            <p>Message</p>
                        </div>
                        <p className="right-coulmn">Details</p>
                    </div>
                    <div className="list">
                        <div className="rows-wrapper">
                            {stationState?.stationSocketData?.poison_messages?.map((message, id) => {
                                return (
                                    <div
                                        className={selectedRowIndex === id ? 'message-row selected' : 'message-row'}
                                        key={id}
                                        onClick={() => onSelectedRow(true, message._id, id)}
                                    >
                                        {tabValue === '1' && (
                                            <Checkbox
                                                key={message._id}
                                                checked={isCheck.includes(message._id)}
                                                id={message._id}
                                                onChange={handleCheckedClick}
                                                name={message._id}
                                            />
                                        )}
                                        <OverflowTip text={message.message.data} width={'300px'}>
                                            {message.message.data}
                                        </OverflowTip>
                                    </div>
                                );
                            })}
                        </div>
                        <div className="message-wrapper">
                            <div className="row-data">
                                <Space direction="vertical">
                                    <CustomCollapse header="Producer" status={true} data={messageDetails.producer} />
                                    <MultiCollapse header="Failed CGs" defaultOpen={true} data={messageDetails.poisionedCGs} />
                                    <CustomCollapse status={false} header="Details" data={messageDetails.details} />
                                    <CustomCollapse status={false} header="Payload" defaultOpen={true} data={messageDetails.message} message={true} />
                                </Space>
                            </div>
                            <Button
                                width="96%"
                                height="40px"
                                placeholder={
                                    <div className="botton-title">
                                        <img src={Journey} alt="Journey" />
                                        <p>Message Journey</p>
                                    </div>
                                }
                                colorType="black"
                                radiusType="semi-round"
                                backgroundColorType="orange"
                                fontSize="12px"
                                fontWeight="600"
                                onClick={() => history.push(`${window.location.pathname}/${messageDetails.id}`)}
                            />
                        </div>
                    </div>
                </div>
            )}
            {tabValue === '0' && stationState?.stationSocketData?.messages === null && (
                <div className="waiting-placeholder">
                    <img width={100} src={waitingMessages} />
                    <p>No messages yet</p>
                    <span className="des">Create your 1st producer and start producing data.</span>
                    {process.env.REACT_APP_SANDBOX_ENV && (
                        <Button
                            className="open-sdk"
                            width="110px"
                            height="37px"
                            placeholder="Explore demo"
                            colorType={'white'}
                            radiusType="circle"
                            border={'none'}
                            backgroundColorType={'purple'}
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => history.go(`${pathDomains.stations}/demo-app`)}
                        />
                    )}
                </div>
            )}
            {tabValue === '1' && stationState?.stationSocketData?.poison_messages?.length === 0 && (
                <div className="empty-messages">
                    <p>Congrats, No messages in your station's dead-letter, yet ;)</p>
                </div>
            )}
        </div>
    );
};

export default Messages;
