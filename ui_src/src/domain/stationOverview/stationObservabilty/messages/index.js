// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';
import { message } from 'antd';

import React, { useContext, useEffect, useState } from 'react';
import { InfoOutlined } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import { Checkbox, Space } from 'antd';

import { convertBytes, msToUnits, numberWithCommas, parsingDate } from '../../../../services/valueConvertor';
import waitingMessages from '../../../../assets/images/waitingMessages.svg';
import deadLetterPlaceholder from '../../../../assets/images/deadLetterPlaceholder.svg';
import idempotencyIcon from '../../../../assets/images/idempotencyIcon.svg';
import dlsEnableIcon from '../../../../assets/images/dls_enable_icon.svg';
import followersImg from '../../../../assets/images/followersDetails.svg';
import TooltipComponent from '../../../../components/tooltip/tooltip';
import leaderImg from '../../../../assets/images/leaderDetails.svg';
import CheckboxComponent from '../../../../components/checkBox';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import Journey from '../../../../assets/images/journey.svg';
import CustomCollapse from '../components/customCollapse';
import DetailBox from '../../../../components/detailBox';
import DlsConfig from '../../../../components/dlsConfig';
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
    const [indeterminate, setIndeterminate] = useState(false);

    const url = window.location.href;
    const stationName = url.split('stations/')[1];

    const [tabValue, setTabValue] = useState('All');
    const [subTabValue, setSubTabValue] = useState('Poison');
    const tabs = ['All', 'Dead-letter', 'Details'];
    const subTabs = ['Poison', 'Schemaverse'];
    const history = useHistory();

    useEffect(() => {
        if (stationState?.stationSocketData?.messages?.length > 0 && (Object.keys(messageDetails).length === 0 || tabValue === 'All') && selectedRowIndex === 0) {
            getMessageDetails(false, null, stationState?.stationSocketData?.messages[0]?.message_seq, false);
        }
        if (tabValue === 'Dead-letter' && stationState?.stationSocketData?.poison_messages?.length > 0 && selectedRowIndex === 0) {
            getMessageDetails(true, stationState?.stationSocketData?.poison_messages[0]?._id, null, false);
        }
    }, [stationState?.stationSocketData?.messages, stationState?.stationSocketData?.poison_messages]);

    const getMessageDetails = async (isPoisonMessage, messageId = null, message_seq = null, loadMessage) => {
        setLoadMessageData(loadMessage);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_MESSAGE_DETAILS}?station_name=${stationName}&is_poison_message=${isPoisonMessage}&message_id=${encodeURIComponent(
                    messageId
                )}&message_seq=${message_seq}`
            );
            arrangeData(data);
        } catch (error) {}
        setLoadMessageData(false);
    };

    const arrangeData = (data) => {
        let poisonedCGs = [];
        if (data) {
            data?.poisoned_cgs?.map((row, index) => {
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
                poisonedCGs.push(cg);
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
                    is_active: data?.producer?.is_active,
                    is_deleted: data?.producer?.is_deleted,
                    details: [
                        {
                            name: 'Name',
                            value: data.producer?.name || ''
                        },
                        {
                            name: 'User',
                            value: data.producer?.created_by_user || ''
                        },
                        {
                            name: 'IP',
                            value: data.producer?.client_address || ''
                        }
                    ]
                },
                message: data.message?.data,
                headers: data.message?.headers || {},
                poisonedCGs: poisonedCGs
            };
            setMessageDetails(messageDetails);
        }
    };

    const onSelectedRow = (isPoisonMessage, id, rowIndex) => {
        setSelectedRowIndex(rowIndex);
        getMessageDetails(isPoisonMessage, isPoisonMessage ? id : null, isPoisonMessage ? null : id, false);
    };

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(stationState?.stationSocketData?.poison_messages.map((li) => li._id));
        setIndeterminate(false);
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        let checkedList = [];
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
            checkedList = isCheck.filter((item) => item !== id);
        }
        if (checked) {
            checkedList = [...isCheck, id];
            setIsCheck(checkedList);
        }
        setIsCheckAll(checkedList.length === stationState?.stationSocketData?.poison_messages?.length);
        setIndeterminate(!!checkedList.length && checkedList.length < stationState?.stationSocketData?.poison_messages?.length);
    };

    const handleChangeMenuItem = (newValue) => {
        if (newValue === 'All' && stationState?.stationSocketData?.messages?.length > 0) {
            getMessageDetails(false, null, stationState?.stationSocketData?.messages[0]?.message_seq, true);
        }
        if (newValue === 'Dead-letter' && stationState?.stationSocketData?.poison_messages?.length > 0) {
            getMessageDetails(true, stationState?.stationSocketData?.poison_messages[0]?._id, null, true);
        }
        setTabValue(newValue);
        setSelectedRowIndex(0);
    };

    const handleChangeSubMenuItem = (newValue) => {
        setSubTabValue(newValue);
    };

    const handleAck = async () => {
        setIgnoreProcced(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.ACK_POISON_MESSAGE}`, { poison_message_ids: isCheck });
            let poisons = stationState?.stationSocketData?.poison_messages;
            isCheck.map((messageId, index) => {
                poisons = poisons?.filter((item) => {
                    return item._id !== messageId;
                });
            });
            setTimeout(() => {
                setIgnoreProcced(false);
                stationDispatch({ type: 'SET_POISINS_MESSAGES', payload: poisons });
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
            await httpRequest('POST', `${ApiEndpoints.RESEND_POISON_MESSAGE_JOURNEY}`, { poison_message_ids: isCheck });
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
                    {tabValue === 'All' && stationState?.stationSocketData?.messages?.length > 0 && (
                        <div className="messages-amount">
                            <InfoOutlined />
                            <p>Showing last {stationState?.stationSocketData?.messages?.length} messages</p>
                        </div>
                    )}
                    {tabValue === 'Dead-letter' && stationState?.stationSocketData?.poison_messages?.length > 0 && (
                        <div className="messages-amount">
                            <InfoOutlined />
                            <p>Showing last {stationState?.stationSocketData?.poison_messages?.length} messages</p>
                        </div>
                    )}
                </div>
                {tabValue === 'Dead-letter' && subTabValue === 'Poison' && stationState?.stationSocketData?.poison_messages?.length > 0 && (
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
                            disabled={isCheck.length === 0 || !stationState?.stationMetaData.is_native}
                            tooltip={!stationState?.stationMetaData.is_native && 'Not supported without using the native Memphis SDK’s'}
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
                    length={
                        (stationState?.stationSocketData?.poison_messages?.length > 0 || stationState?.stationSocketData?.schema_fail_messages?.length > 0) && [
                            null,
                            (stationState?.stationSocketData?.poison_messages?.length || 0) + (stationState?.stationSocketData?.schema_fail_messages?.length || 0)
                        ]
                    }
                ></CustomTabs>
            </div>
            {tabValue === 'Dead-letter' && (
                <div className="tabs">
                    <CustomTabs
                        value={subTabValue}
                        onChange={handleChangeSubMenuItem}
                        tabs={subTabs}
                        length={stationState?.stationSocketData?.poison_messages?.length > 0 && [stationState?.stationSocketData?.poison_messages?.length, null]}
                    ></CustomTabs>
                </div>
            )}
            {tabValue === 'All' && stationState?.stationSocketData?.messages?.length > 0 && (
                <div className="list-wrapper msg-list">
                    <div className="coulmns-table">
                        <div className="left-coulmn all">
                            <p>Messages</p>
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
                                        <span className="preview-message">{message?.data}</span>
                                    </div>
                                );
                            })}
                        </div>
                        <div className="message-wrapper">
                            <div className="row-data">
                                <Space direction="vertical">
                                    <CustomCollapse
                                        collapsible={!stationState?.stationMetaData?.is_native}
                                        tooltip={!stationState?.stationMetaData?.is_native && 'Not supported without using the native Memphis SDK’s'}
                                        header="Producer"
                                        status={true}
                                        data={messageDetails?.producer}
                                    />

                                    <MultiCollapse
                                        header="Failed CGs"
                                        tooltip={!stationState?.stationMetaData?.is_native && 'Not supported without using the native Memphis SDK’s'}
                                        defaultOpen={true}
                                        data={messageDetails?.poisonedCGs}
                                    />
                                    <CustomCollapse status={false} header="Metadata" data={messageDetails?.details} />
                                    <CustomCollapse status={false} header="Headers" defaultOpen={false} data={messageDetails?.headers} message={true} />
                                    <CustomCollapse status={false} header="Payload" defaultOpen={true} data={messageDetails?.message} message={true} />
                                </Space>
                            </div>
                        </div>
                    </div>
                </div>
            )}
            {tabValue === 'Dead-letter' && subTabValue === 'Poison' && stationState?.stationSocketData?.poison_messages?.length > 0 && (
                //
                <div className="list-wrapper dls-list">
                    <div className="coulmns-table">
                        <div className="left-coulmn">
                            <CheckboxComponent indeterminate={indeterminate} checked={isCheckAll} id={'selectAll'} onChange={onCheckedAll} name={'selectAll'} />
                            <p>Messages</p>
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
                                        {tabValue === 'Dead-letter' && (
                                            <CheckboxComponent
                                                checked={isCheck.includes(message._id)}
                                                id={message._id}
                                                onChange={handleCheckedClick}
                                                name={message._id}
                                            />
                                        )}
                                        <span className="preview-message">{message?.message.data}</span>
                                    </div>
                                );
                            })}
                        </div>
                        <div className="message-wrapper">
                            <div className="row-data">
                                <Space direction="vertical">
                                    {stationState?.stationMetaData.is_native && <CustomCollapse header="Producer" status={true} data={messageDetails?.producer} />}
                                    <MultiCollapse header="Failed CGs" defaultOpen={true} data={messageDetails?.poisonedCGs} />
                                    <CustomCollapse status={false} header="Metadata" data={messageDetails?.details} />
                                    <CustomCollapse status={false} header="Headers" defaultOpen={false} data={messageDetails?.headers} message={true} />
                                    <CustomCollapse status={false} header="Payload" defaultOpen={true} data={messageDetails?.message} message={true} />
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
                                tooltip={!stationState?.stationMetaData.is_native && 'Not supported without using the native Memphis SDK’s'}
                                disabled={!stationState?.stationMetaData.is_native}
                                onClick={() => history.push(`${window.location.pathname}/${messageDetails.id}`)}
                            />
                        </div>
                    </div>
                </div>
            )}
            {tabValue === 'All' && stationState?.stationSocketData?.messages === null && (
                <div className="waiting-placeholder msg-plc">
                    <img width={100} src={waitingMessages} alt="waitingMessages" />
                    <p>No messages yet</p>
                    <span className="des">Create your 1st producer and start producing data</span>
                    {process.env.REACT_APP_SANDBOX_ENV && stationName !== 'demo-app' && (
                        <a className="explore-button" href={`${pathDomains.stations}/demo-app`} target="_parent">
                            Explore demo
                        </a>
                    )}
                </div>
            )}
            {tabValue === 'Dead-letter' &&
                ((subTabValue === 'Poison' && stationState?.stationSocketData?.poison_messages?.length === 0) ||
                    (subTabValue === 'Schemaverse' && stationState?.stationSocketData?.schema_fail_messages?.length === 0)) && (
                    <div className="waiting-placeholder msg-plc">
                        <img width={100} src={deadLetterPlaceholder} alt="waitingMessages" />
                        <p>Hooray! No messages</p>
                    </div>
                )}
            {tabValue === 'Details' && (
                <div className="details">
                    <DetailBox
                        img={leaderImg}
                        title={'Leader'}
                        desc={
                            <span>
                                The current leader of this station.{' '}
                                <a href="https://docs.memphis.dev/memphis/memphis/concepts/station#leaders-and-followers" target="_blank">
                                    Learn More
                                </a>
                            </span>
                        }
                        data={[stationState?.stationSocketData?.leader]}
                    />
                    {stationState?.stationSocketData?.followers?.length > 0 && (
                        <DetailBox
                            img={followersImg}
                            title={'Followers'}
                            desc={
                                <span>
                                    The brokers that contain a replica of this station and in case of failure will replace the leader.{' '}
                                    <a href="https://docs.memphis.dev/memphis/memphis/concepts/station#leaders-and-followers" target="_blank">
                                        Learn More
                                    </a>
                                </span>
                            }
                            data={stationState?.stationSocketData?.followers}
                        />
                    )}
                    <DetailBox img={dlsEnableIcon} title={'DLS configuration'} desc="By which event, messages will be stored in the dead-letter station.">
                        <DlsConfig />
                    </DetailBox>
                    <DetailBox
                        img={idempotencyIcon}
                        title={'Idempotency'}
                        desc={
                            <span>
                                Ensures messages will be produced once.{' '}
                                <a href="https://docs.memphis.dev/memphis/memphis/concepts/idempotency" target="_blank">
                                    Learn More
                                </a>
                            </span>
                        }
                        data={[msToUnits(stationState?.stationSocketData?.idempotency_window_in_ms)]}
                    />
                </div>
            )}
        </div>
    );
};

export default Messages;
