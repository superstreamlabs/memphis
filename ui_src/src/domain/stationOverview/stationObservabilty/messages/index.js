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
import { InfoOutlined } from '@material-ui/icons';

import { DEAD_LETTERED_MESSAGES_RETENTION_IN_HOURS } from '../../../../const/localStorageConsts';
import { ReactComponent as DeadLetterPlaceholderIcon } from '../../../../assets/images/deadLetterPlaceholder.svg';
import { isCloud, messageParser, msToUnits } from '../../../../services/valueConvertor';
import { ReactComponent as PurgeWrapperIcon } from '../../../../assets/images/purgeWrapperIcon.svg';
import { ReactComponent as WaitingMessagesIcon } from '../../../../assets/images/waitingMessages.svg';
import { ReactComponent as IdempotencyIcon } from '../../../../assets/images/idempotencyIcon.svg';
import { ReactComponent as DlsEnableIcon } from '../../../../assets/images/dls_enable_icon.svg';
import { ReactComponent as FollowersIcon } from '../../../../assets/images/followersDetails.svg';
import TooltipComponent from '../../../../components/tooltip/tooltip';
import { ReactComponent as LeaderIcon } from '../../../../assets/images/leaderDetails.svg';
import PurgeStationModal from '../components/purgeStationModal';
import CheckboxComponent from '../../../../components/checkBox';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import MessageDetails from '../components/messageDetails';
import DetailBox from '../../../../components/detailBox';
import DlsConfig from '../../../../components/dlsConfig';
import { httpRequest } from '../../../../services/http';
import { ReactComponent as PurgeIcon } from '../../../../assets/images/purge.svg';
import CustomTabs from '../../../../components/Tabs';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import { StationStoreContext } from '../..';
import { Virtuoso } from 'react-virtuoso';
import { showMessages } from '../../../../services/genericServices';
import { ReactComponent as UpRightArrow } from '../../../../assets/images/upRightCorner.svg';
import { ReactComponent as DisconnectIcon } from '../../../../assets/images/disconnectDls.svg';
import UseSchemaModal from '../../components/useSchemaModal';
import DeleteItemsModal from '../../../../components/deleteItemsModal';
import { ReactComponent as DisableIcon } from '../../../../assets/images/disableIcon.svg';

const Messages = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [selectedRowIndex, setSelectedRowIndex] = useState(null);
    const [selectedRowPartition, setSelectedRowPartition] = useState(null);
    const [modalPurgeIsOpen, modalPurgeFlip] = useState(false);
    const [resendProcced, setResendProcced] = useState(false);
    const [ignoreProcced, setIgnoreProcced] = useState(false);
    const [userScrolled, setUserScrolled] = useState(false);
    const [subTabValue, setSubTabValue] = useState('Unacked');
    const [tabValue, setTabValue] = useState('Messages');
    const [isCheck, setIsCheck] = useState([]);
    const [useDlsModal, setUseDlsModal] = useState(false);
    const [disableModal, setDisableModal] = useState(false);
    const [disableLoader, setDisableLoader] = useState(false);

    const dls = stationState?.stationMetaData?.dls_station === '' ? null : stationState?.stationMetaData?.dls_station;
    const tabs = ['Messages', 'Dead-letter', 'Details'];
    const subTabs = [
        { name: 'Unacked', disabled: false },
        { name: 'Schema violation', disabled: !stationState?.stationMetaData?.is_native }
    ];
    const url = window.location.href;
    const stationName = url.split('stations/')[1];

    const onSelectedRow = (id, partition) => {
        setUserScrolled(false);
        setSelectedRowIndex(id);
        setSelectedRowPartition(partition);
        stationDispatch({ type: 'SET_SELECTED_ROW_ID', payload: id });
        stationDispatch({ type: 'SET_SELECTED_ROW_PARTITION', payload: partition });
    };

    const setDls = (dls) => {
        stationDispatch({ type: 'SET_DLS', payload: dls });
    };

    const handleSetDls = async (dls) => {
        try {
            await httpRequest('POST', ApiEndpoints.ATTACH_DLS, { name: dls, station_names: [stationState?.stationMetaData?.name] });
            setDls(dls);
            setUseDlsModal(false);
        } catch (error) {
            setUseDlsModal(false);
        }
    };

    const handleDetachDls = async () => {
        setDisableLoader(true);
        try {
            await httpRequest('DELETE', ApiEndpoints.DETACH_DLS, { name: dls, station_names: [stationState?.stationMetaData?.name] });
            setDls('');
            setDisableModal(false);
            setDisableLoader(false);
        } catch (error) {
            setDisableLoader(false);
            setDisableModal(false);
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
    };

    const handleChangeMenuItem = (newValue) => {
        stationDispatch({ type: 'SET_SELECTED_ROW_ID', payload: null });
        stationDispatch({ type: 'SET_SELECTED_ROW_PARTITION', payload: null });
        setSelectedRowIndex(null);
        setIsCheck([]);

        setTabValue(newValue);
        subTabValue === 'Schema violation' && setSubTabValue('Unacked');
    };

    useEffect(() => {
        if (selectedRowIndex && !userScrolled) {
            const element = document.getElementById(selectedRowIndex);
            if (element) {
                element.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }
        }
    }, []);

    const handleChangeSubMenuItem = (newValue) => {
        stationDispatch({ type: 'SET_SELECTED_ROW_ID', payload: null });
        stationDispatch({ type: 'SET_SELECTED_ROW_PARTITION', payload: null });
        setSelectedRowIndex(null);
        setSubTabValue(newValue);
        setIsCheck([]);
    };

    const handleDrop = async () => {
        setIgnoreProcced(true);
        let messages;
        try {
            if (tabValue === tabs[0]) {
                const message_seqs = isCheck.map((item) => {
                    return { message_seq: Number(item.split('_')[0]), partition_number: Number(item.split('_')[1]) };
                });
                await httpRequest('DELETE', `${ApiEndpoints.REMOVE_MESSAGES}`, { station_name: stationName, messages: message_seqs });
                messages = stationState?.stationSocketData?.messages;
                isCheck.map((messageId, index) => {
                    messages = messages?.filter((item) => {
                        return `${item.message_seq}_${item.partition}` !== messageId;
                    });
                });
            } else {
                await httpRequest('POST', `${ApiEndpoints.DROP_DLS_MESSAGE}`, {
                    dls_type: subTabValue === subTabs[0].name ? 'poison' : 'schema',
                    dls_message_ids: isCheck,
                    station_name: stationName
                });
                messages = subTabValue === subTabs[0].name ? stationState?.stationSocketData?.poison_messages : stationState?.stationSocketData?.schema_failed_messages;
                isCheck.map((messageId, index) => {
                    messages = messages?.filter((item) => {
                        return item.id !== messageId;
                    });
                });
            }
            setTimeout(() => {
                setIgnoreProcced(false);
                tabValue === tabs[0]
                    ? stationDispatch({ type: 'SET_MESSAGES', payload: messages })
                    : subTabValue === subTabs[0].name
                    ? stationDispatch({ type: 'SET_POISON_MESSAGES', payload: messages })
                    : stationDispatch({ type: 'SET_FAILED_MESSAGES', payload: messages });
                stationDispatch({ type: 'SET_SELECTED_ROW_ID', payload: null });
                stationDispatch({ type: 'SET_SELECTED_ROW_PARTITION', payload: null });
                setSelectedRowIndex(null);
                setIsCheck([]);
            }, 1500);
        } catch (error) {
            setIgnoreProcced(false);
        }
    };

    const handleResend = async () => {
        setResendProcced(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.RESEND_POISON_MESSAGE_JOURNEY}`, { poison_message_ids: isCheck, station_name: stationName });
            if (isCheck.length > 0) {
                setTimeout(() => {
                    setResendProcced(false);
                    showMessages('success', isCheck.length === 1 ? 'The message was sent successfully' : 'The messages were sent successfully');
                    setIsCheck([]);
                }, 1500);
            } else {
                showMessages('success', 'All DLS messages are being resent asynchronously. We will let you know upon completion');
                setTimeout(() => {
                    setResendProcced(false);
                    setIsCheck([]);
                }, 3500);
            }
        } catch (error) {
            setResendProcced(false);
        }
    };

    const handleScroll = () => {
        setUserScrolled(true);
    };

    const listGenerator = (index, message) => {
        const id = tabValue === tabs[1] ? message?.id : message?.message_seq;
        const partition = tabValue === tabs[1] ? null : message?.partition;
        return (
            <div className={index % 2 === 0 ? 'even' : 'odd'}>
                <CheckboxComponent
                    className="check-box-message"
                    checked={isCheck?.includes(partition ? `${id}_${partition}` : id)}
                    id={partition ? `${id}_${partition}` : id}
                    onChange={handleCheckedClick}
                    name={partition ? `${id}_${partition}` : id}
                />
                <div
                    className={selectedRowIndex === id && selectedRowPartition === partition ? 'row-message selected' : 'row-message'}
                    key={id}
                    id={id}
                    onClick={() => onSelectedRow(id, partition)}
                >
                    {selectedRowIndex === id && selectedRowPartition === partition && <div className="hr-selected"></div>}
                    <span className="preview-message">
                        {tabValue === tabs[1] ? messageParser('string', message?.message?.data) : messageParser('string', message?.data)}
                    </span>
                </div>
            </div>
        );
    };

    const listGeneratorWrapper = () => {
        let isDls = tabValue === tabs[1];
        return (
            <div className={isDls ? 'list-wrapper dls-list' : 'list-wrapper msg-list'}>
                <div className="coulmns-table">
                    <div className="left-coulmn-wrapper">
                        <p className="left-coulmn">Messages</p>
                        <TooltipComponent text={`DLS retention is ${localStorage.getItem(DEAD_LETTERED_MESSAGES_RETENTION_IN_HOURS)} hours.`} minWidth="35px">
                            <InfoOutlined />
                        </TooltipComponent>
                    </div>
                    <p className="right-coulmn">Information</p>
                </div>
                <div className="list">
                    <div className="rows-wrapper">
                        <Virtuoso
                            data={
                                !isDls
                                    ? stationState?.stationSocketData?.messages
                                    : subTabValue === 'Unacked'
                                    ? stationState?.stationSocketData?.poison_messages
                                    : stationState?.stationSocketData?.schema_failed_messages
                            }
                            onScroll={() => handleScroll()}
                            overscan={100}
                            itemContent={(index, message) => listGenerator(index, message)}
                        />
                    </div>
                    <MessageDetails isDls={isDls} isFailedSchemaMessage={subTabValue === 'Schema violation'} />
                </div>
            </div>
        );
    };

    const showLastMsg = () => {
        let amount = 0;
        if (tabValue === tabs[0] && stationState?.stationSocketData?.messages?.length > 0) amount = stationState?.stationSocketData?.messages?.length;
        else if (tabValue === tabs[1] && subTabValue === subTabs[0].name && stationState?.stationSocketData?.poison_messages?.length > 0)
            amount = stationState?.stationSocketData?.poison_messages?.length;
        else if (tabValue === tabs[1] && subTabValue === subTabs[1].name && stationState?.stationSocketData?.schema_failed_messages?.length > 0)
            amount = stationState?.stationSocketData?.schema_failed_messages?.length;
        return (
            amount > 0 && (
                <div className="messages-amount">
                    <InfoOutlined />
                    <p>
                        Showing last {amount?.toLocaleString()} out of{' '}
                        {tabValue === tabs[0]
                            ? stationState?.stationSocketData?.total_messages?.toLocaleString()
                            : stationState?.stationSocketData?.total_dls_messages?.toLocaleString()}{' '}
                        messages
                    </p>
                </div>
            )
        );
    };

    const getDescriptin = () => {
        if (stationState?.stationSocketData?.connected_producers?.length > 0 || stationState?.stationSocketData?.disconnected_producers?.length > 0) {
            if (
                stationState?.stationMetaData?.retention_type === 'ack_based' &&
                stationState?.stationSocketData?.disconnected_cgs?.length === 0 &&
                stationState?.stationSocketData?.connected_cgs?.length === 0
            ) {
                return 'When retention is ack-based, messages will be auto-deleted if no consumers are connected to the station';
            } else {
                return 'Start / Continue producing data.';
            }
        } else {
            return 'Create your 1st producer and start producing data.';
        }
    };

    return (
        <div className="messages-container">
            <div className="header">
                <div className="left-side">
                    <p className="title">Station</p>
                    {showLastMsg()}
                </div>
                <div className="right-side">
                    {((tabValue === tabs[0] && stationState?.stationSocketData?.messages?.length > 0) ||
                        (tabValue === tabs[1] &&
                            ((subTabValue === subTabs[0].name && stationState?.stationSocketData?.poison_messages?.length > 0) ||
                                (subTabValue === subTabs[1].name && stationState?.stationSocketData?.schema_failed_messages?.length > 0)))) && (
                        <Button
                            width="80px"
                            height="32px"
                            placeholder={isCheck.length === 0 ? 'Purge' : `Drop (${isCheck.length})`}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            isLoading={ignoreProcced}
                            onClick={() => (isCheck.length === 0 ? modalPurgeFlip(true) : handleDrop())}
                        />
                    )}
                    {tabValue === 'Dead-letter' && subTabValue === 'Unacked' && stationState?.stationSocketData?.poison_messages?.length > 0 && (
                        <Button
                            width="95px"
                            height="32px"
                            placeholder={isCheck.length === 0 ? 'Resend all' : `Resend (${isCheck.length})`}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            disabled={resendProcced || stationState?.stationSocketData?.resend_disabled || !stationState?.stationMetaData?.is_native}
                            isLoading={resendProcced && isCheck.length > 0}
                            tooltip={!stationState?.stationMetaData?.is_native && 'Supported only by using Memphis SDKs'}
                            onClick={() => handleResend()}
                        />
                    )}
                </div>
            </div>
            <div className="tabs">
                <CustomTabs
                    value={tabValue}
                    onChange={handleChangeMenuItem}
                    tabs={tabs}
                    length={[
                        null,
                        stationState?.stationSocketData?.poison_messages?.length || stationState?.stationSocketData?.schema_failed_messages?.length || null,
                        null
                    ]}
                    icon
                />
            </div>
            {tabValue === tabs[1] && (
                <div className="tabs">
                    <CustomTabs
                        defaultValue
                        value={subTabValue}
                        onChange={handleChangeSubMenuItem}
                        tabs={subTabs}
                        length={[
                            stationState?.stationSocketData?.poison_messages?.length || null,
                            stationState?.stationSocketData?.schema_failed_messages?.length || null
                        ]}
                        tooltip={[null, !stationState?.stationMetaData?.is_native && 'Supported only by using Memphis SDKs']}
                    />
                </div>
            )}
            {tabValue === tabs[0] && stationState?.stationSocketData?.messages?.length > 0 && listGeneratorWrapper()}
            {tabValue === tabs[1] && subTabValue === subTabs[0].name && stationState?.stationSocketData?.poison_messages?.length > 0 && listGeneratorWrapper()}
            {tabValue === tabs[1] && subTabValue === subTabs[1].name && stationState?.stationSocketData?.schema_failed_messages?.length > 0 && listGeneratorWrapper()}

            {tabValue === tabs[0] && (stationState?.stationSocketData?.messages === null || stationState?.stationSocketData?.messages?.length === 0) && (
                <div className="waiting-placeholder msg-plc">
                    <WaitingMessagesIcon width={100} alt="waitingMessages" />
                    <p>No messages</p>
                    <span className="des">{getDescriptin()}</span>
                </div>
            )}
            {tabValue === tabs[1] &&
                ((subTabValue === 'Unacked' && stationState?.stationSocketData?.poison_messages?.length === 0) ||
                    (subTabValue === 'Schema violation' && stationState?.stationSocketData?.schema_failed_messages?.length === 0)) && (
                    <div className="waiting-placeholder msg-plc">
                        <DeadLetterPlaceholderIcon width={80} alt="waitingMessages" />
                        <p>Hooray! No messages</p>
                    </div>
                )}
            {tabValue === tabs[2] && (
                <div className="details">
                    <DetailBox
                        icon={<DlsEnableIcon width={24} alt="dlsEnableIcon" />}
                        title={
                            <>
                                <span>Dead-letter station configuration</span>
                                <Button
                                    width="130px"
                                    height="25px"
                                    placeholder={
                                        <div className="use-dls-button">
                                            {dls ? <DisconnectIcon /> : <UpRightArrow />}
                                            <p>{dls ? 'Disable' : 'Enable'} Consumption</p>
                                        </div>
                                    }
                                    colorType={dls ? 'white' : 'black'}
                                    radiusType="circle"
                                    backgroundColorType={dls ? 'red' : 'orange'}
                                    fontSize="10px"
                                    fontFamily="InterSemiBold"
                                    fontWeight={600}
                                    disabled={!stationState?.stationMetaData?.is_native}
                                    onClick={() => (dls ? setDisableModal(true) : setUseDlsModal(true))}
                                />
                            </>
                        }
                        desc="Triggers for storing messages in the dead-letter station."
                        rightSection={false}
                    >
                        <DlsConfig />
                    </DetailBox>
                    <DetailBox
                        icon={<PurgeIcon width={24} alt="purgeIcon" />}
                        title={'Purge'}
                        desc="Clean station from messages."
                        data={[
                            <Button
                                width="80px"
                                height="25px"
                                placeholder="Purge"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontWeight="600"
                                disabled={stationState?.stationSocketData?.total_dls_messages === 0 && stationState?.stationSocketData?.total_messages === 0}
                                onClick={() => modalPurgeFlip(true)}
                            />
                        ]}
                    ></DetailBox>
                    {!isCloud() && stationState?.stationPartition !== -1 && (
                        <DetailBox
                            icon={<LeaderIcon width={24} alt="leaderIcon" />}
                            title={'Leader'}
                            desc={
                                <span>
                                    The current leader of this station.{' '}
                                    <a href="https://docs.memphis.dev/memphis/memphis/concepts/station#leaders-and-followers" target="_blank">
                                        Learn more
                                    </a>
                                </span>
                            }
                            data={[stationState?.stationSocketData?.leader]}
                        />
                    )}
                    {stationState?.stationSocketData?.followers?.length > 0 && !isCloud() && stationState?.stationPartition !== -1 && (
                        <DetailBox
                            icon={<FollowersIcon width={24} alt="followersImg" />}
                            title={'Followers'}
                            desc={
                                <span>
                                    The brokers that contain a replica of this station and in case of failure will replace the leader.{' '}
                                    <a href="https://docs.memphis.dev/memphis/memphis/concepts/station#leaders-and-followers" target="_blank">
                                        Learn more
                                    </a>
                                </span>
                            }
                            data={stationState?.stationSocketData?.followers}
                        />
                    )}

                    <DetailBox
                        icon={<IdempotencyIcon width={24} alt="idempotencyIcon" />}
                        title={'Idempotency'}
                        desc={
                            <span>
                                Ensures messages with the same "msg-id" value will be produced only once for the configured time.{' '}
                                <a href="https://docs.memphis.dev/memphis/memphis/concepts/idempotency" target="_blank">
                                    Learn more
                                </a>
                            </span>
                        }
                        data={[msToUnits(stationState?.stationSocketData?.idempotency_window_in_ms)]}
                    />
                </div>
            )}
            <Modal
                header={<PurgeWrapperIcon alt="deleteWrapperIcon" />}
                width="460px"
                height="320px"
                displayButtons={false}
                clickOutside={() => modalPurgeFlip(false)}
                open={modalPurgeIsOpen}
            >
                <PurgeStationModal
                    title="Purge"
                    desc="This action will clean the station from messages."
                    stationName={stationName}
                    cancel={() => modalPurgeFlip(false)}
                    msgsDisabled={stationState?.stationSocketData?.total_messages === 0}
                    dlsDisabled={stationState?.stationSocketData?.total_dls_messages === 0}
                />
            </Modal>
            <Modal
                header={
                    <div className="modal-header">
                        <p>Consume via another station</p>
                        <label>Only new messages will be able to be consumed.</label>
                    </div>
                }
                displayButtons={false}
                height="400px"
                width="352px"
                clickOutside={() => setUseDlsModal(false)}
                open={useDlsModal}
                hr={true}
                className="use-schema-modal"
            >
                <UseSchemaModal stationName={stationState?.stationMetaData?.name} handleSetSchema={handleSetDls} type="dls" close={() => setUseDlsModal(false)} />
            </Modal>
            <Modal
                header={<DisableIcon alt="stopUsingIcon" />}
                width="520px"
                height="240px"
                displayButtons={false}
                clickOutside={() => setDisableModal(false)}
                open={disableModal}
            >
                <DeleteItemsModal
                    title="Disabling dead-letter consumption will stop pushing new dead-letter messages"
                    desc={
                        <span>
                            Station <strong>{stationState?.stationMetaData?.name}</strong> will be disconnected from <strong>{dls} </strong>.
                        </span>
                    }
                    buttontxt="I understand, disable consumption"
                    textToConfirm="disable"
                    handleDeleteSelected={handleDetachDls}
                    loader={disableLoader}
                />
            </Modal>
        </div>
    );
};

export default Messages;
