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
import { Add, FiberManualRecord, InfoOutlined } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import { MinusOutlined } from '@ant-design/icons';

import { convertBytes, convertSecondsToDate, isCloud, replicasConvertor } from '../../../services/valueConvertor';
import deleteWrapperIcon from '../../../assets/images/deleteWrapperIcon.svg';
import averageMesIcon from '../../../assets/images/averageMesIcon.svg';
import stopUsingIcon from '../../../assets/images/stopUsingIcon.svg';
import schemaIconActive from '../../../assets/images/schemaIconActive.svg';
import DeleteItemsModal from '../../../components/deleteItemsModal';
import PartitionsFilter from '../../../components/partitionsFilter';

import awaitingIcon from '../../../assets/images/awaitingIcon.svg';
import TooltipComponent from '../../../components/tooltip/tooltip';
import redirectIcon from '../../../assets/images/redirectIcon.svg';
import OverflowTip from '../../../components/tooltip/overflowtip';
import UpdateSchemaModal from '../components/updateSchemaModal';
import deleteIcon from '../../../assets/images/deleteIcon.svg';
import ActiveBadge from '../../../components/activeBadge';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import BackIcon from '../../../assets/images/backIcon.svg';
import UseSchemaModal from '../components/useSchemaModal';
import SdkExample from '../../../components/sdkExsample';
import { httpRequest } from '../../../services/http';
import TagsList from '../../../components/tagList';
import Button from '../../../components/button';
import Modal from '../../../components/modal';
import Auditing from '../components/auditing';
import AsyncTasks from '../../../components/asyncTasks';
import pathDomains from '../../../router';
import { StationStoreContext } from '..';
import { TIERED_STORAGE_UPLOAD_INTERVAL } from '../../../const/localStorageConsts';

const StationOverviewHeader = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [updateSchemaModal, setUpdateSchemaModal] = useState(false);
    const [modalDeleteIsOpen, modalDeleteFlip] = useState(false);
    const [useSchemaModal, setUseSchemaModal] = useState(false);
    const [retentionValue, setRetentionValue] = useState('');
    const [detachLoader, setDetachLoader] = useState(false);
    const [deleteLoader, setDeleteLoader] = useState(false);
    const [deleteModal, setDeleteModal] = useState(false);
    const [auditModal, setAuditModal] = useState(false);
    const [sdkModal, setSdkModal] = useState(false);
    const history = useHistory();

    useEffect(() => {
        switch (stationState?.stationMetaData?.retention_type) {
            case 'message_age_sec':
                setRetentionValue(convertSecondsToDate(stationState?.stationMetaData?.retention_value));
                break;
            case 'bytes':
                setRetentionValue(`${stationState?.stationMetaData?.retention_value?.toLocaleString()} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${stationState?.stationMetaData?.retention_value?.toLocaleString()} messages`);
                break;
            case 'ack_based':
                setRetentionValue('Ack based');
                break;
            default:
                break;
        }
    }, [stationState?.stationMetaData?.retention_type]);

    const returnToStaionsList = () => {
        history.push(pathDomains.stations);
    };

    const updateTags = (tags) => {
        stationDispatch({ type: 'SET_TAGS', payload: tags });
    };

    const removeTag = async (tagName) => {
        try {
            await httpRequest('DELETE', `${ApiEndpoints.REMOVE_TAG}`, { name: tagName, entity_type: 'station', entity_name: stationState?.stationMetaData?.name });
            let tags = stationState?.stationSocketData?.tags;
            let updatedTags = tags.filter((tag) => tag.name !== tagName);
            stationDispatch({ type: 'SET_TAGS', payload: updatedTags });
        } catch (error) {}
    };

    const setSchema = (schema) => {
        stationDispatch({ type: 'SET_SCHEMA', payload: schema });
    };

    const handleDeleteStation = async () => {
        setDeleteLoader(true);
        try {
            await httpRequest('DELETE', ApiEndpoints.REMOVE_STATION, {
                station_names: [stationState?.stationMetaData?.name]
            });
            returnToStaionsList();
            setDeleteLoader(false);
            modalDeleteFlip(false);
        } catch (error) {
            setDeleteLoader(false);
            modalDeleteFlip(false);
        }
    };

    const handleStopUseSchema = async () => {
        setDetachLoader(true);
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_SCHEMA_FROM_STATION, { station_name: stationState?.stationMetaData?.name });
            if (data) {
                setSchema(data);
                setDeleteModal(false);
                setDetachLoader(false);
            }
        } catch (error) {
            setDetachLoader(false);
            setDeleteModal(false);
        }
    };

    return (
        <div className="station-overview-header">
            <div className="title-wrapper">
                <div className="station-details">
                    <div className="station-name">
                        <img src={BackIcon} onClick={() => returnToStaionsList()} alt="backIcon" />
                        <OverflowTip text={stationState?.stationMetaData?.name} className="station-name-overlow" maxWidth={'350px'} textAlign={'center'}>
                            {stationState?.stationMetaData?.name}
                        </OverflowTip>
                        <TagsList
                            tagsToShow={3}
                            className="tags-list"
                            tags={stationState?.stationSocketData?.tags}
                            addNew={true}
                            editable={true}
                            handleDelete={(tag) => removeTag(tag)}
                            entityType={'station'}
                            entityName={stationState?.stationMetaData?.name}
                            handleTagsUpdate={(tags) => {
                                updateTags(tags);
                            }}
                        />
                    </div>
                    <span className="created-by">
                        Created by <b>{stationState?.stationMetaData?.created_by_username}</b> at {stationState?.stationMetaData?.created_at}{' '}
                        {!stationState?.stationMetaData?.is_native && '(non-native)'}
                    </span>
                </div>
                <div className="station-buttons">
                    {stationState?.stationMetaData?.partitions_number > 1 && (
                        <PartitionsFilter partitions_number={stationState?.stationMetaData?.partitions_number || 0} />
                    )}

                    <AsyncTasks height={'32px'} />
                    <Button
                        width="100px"
                        height="32px"
                        placeholder="Delete station"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        onClick={() => modalDeleteFlip(true)}
                    />
                </div>
            </div>
            <div className="details">
                <div className="main-details">
                    <div className="left-side">
                        <p>
                            <b>Retention:</b> {retentionValue}
                        </p>
                        <div className="storage-section">
                            {!isCloud() && (
                                <p>
                                    <b>Replicas:</b> {replicasConvertor(stationState?.stationMetaData?.replicas, false)}
                                </p>
                            )}
                            <p>
                                <b>Partitions: </b>
                                {stationState?.stationMetaData?.partitions_number === 0 ? 1 : stationState?.stationMetaData?.partitions_number}
                            </p>
                        </div>
                        <div className="storage-section">
                            <p>
                                <b>Local Storage:</b> {stationState?.stationMetaData?.storage_type}
                            </p>
                            <p>
                                <b>Remote Storage:</b> {stationState?.stationMetaData?.tiered_storage_enabled ? 'S3' : <MinusOutlined style={{ color: '#2E2C34' }} />}
                                {stationState?.stationMetaData?.tiered_storage_enabled && (
                                    <TooltipComponent text={`Iterate every ${localStorage.getItem(TIERED_STORAGE_UPLOAD_INTERVAL)} seconds.`} minWidth="35px">
                                        <InfoOutlined />
                                    </TooltipComponent>
                                )}
                            </p>
                        </div>
                    </div>
                </div>
                <div className="icons-wrapper">
                    <div className="details-wrapper">
                        <div className="icon">
                            <img src={schemaIconActive} width={22} height={44} alt="schemaIconActive" />
                        </div>
                        <div className="more-details schema-box">
                            <div className="schema-header">
                                <div className="schema-version">
                                    <p className="schema-title">Schema</p>
                                    {stationState?.stationSocketData?.schema !== undefined && Object.keys(stationState?.stationSocketData?.schema).length !== 0 && (
                                        <div className="schema-details sd-flex">
                                            {stationState?.stationSocketData?.schema?.updates_available && <ActiveBadge content="Updates available" active={false} />}
                                            {!stationState?.stationSocketData?.schema?.updates_available && <ActiveBadge content="Updated" active={true} />}
                                        </div>
                                    )}
                                </div>
                                {stationState?.stationSocketData?.schema !== undefined && Object.keys(stationState?.stationSocketData?.schema).length !== 0 && (
                                    <img
                                        src={redirectIcon}
                                        width={15}
                                        height={15}
                                        alt="redirectIcon"
                                        onClick={() => history.push(`${pathDomains.schemaverse}/list/${stationState?.stationSocketData?.schema?.name}`)}
                                    />
                                )}
                            </div>
                            {stationState?.stationSocketData?.schema !== undefined && Object.keys(stationState?.stationSocketData?.schema).length !== 0 && (
                                <div className="name-and-version">
                                    <p>{stationState?.stationSocketData?.schema?.name}</p>
                                    <FiberManualRecord />
                                    <p>v{stationState?.stationSocketData?.schema?.version_number}</p>
                                </div>
                            )}
                            {stationState?.stationSocketData?.schema === undefined ||
                                (Object.keys(stationState?.stationSocketData?.schema).length === 0 ? (
                                    <>
                                        <div className="add-new">
                                            <Button
                                                width="120px"
                                                height="25px"
                                                placeholder={
                                                    <div className="use-schema-button">
                                                        <Add />
                                                        <p>Enforce schema</p>
                                                    </div>
                                                }
                                                tooltip={!stationState?.stationMetaData?.is_native && 'Supported only by using Memphis SDKs'}
                                                colorType="white"
                                                radiusType="circle"
                                                backgroundColorType="purple"
                                                fontSize="12px"
                                                fontFamily="InterSemiBold"
                                                disabled={!stationState?.stationMetaData?.is_native}
                                                onClick={() => setUseSchemaModal(true)}
                                            />
                                        </div>
                                    </>
                                ) : (
                                    <div className="buttons">
                                        <Button
                                            width="80px"
                                            minWidth="80px"
                                            height="16px"
                                            placeholder="Detach"
                                            colorType="white"
                                            radiusType="circle"
                                            backgroundColorType="purple"
                                            fontSize="10px"
                                            fontFamily="InterMedium"
                                            boxShadowStyle="float"
                                            onClick={() => setDeleteModal(true)}
                                        />
                                        {stationState?.stationSocketData?.schema?.updates_available && (
                                            <Button
                                                width="80px"
                                                height="16px"
                                                placeholder="Update now"
                                                colorType="white"
                                                radiusType="circle"
                                                backgroundColorType="purple"
                                                fontSize="10px"
                                                fontFamily="InterMedium"
                                                boxShadowStyle="float"
                                                onClick={() => setUpdateSchemaModal(true)}
                                            />
                                        )}
                                    </div>
                                ))}
                        </div>
                    </div>
                    <div className="details-wrapper middle">
                        <div className="icon">
                            <img src={awaitingIcon} width={22} height={44} alt="awaitingIcon" />
                        </div>
                        <div className="more-details">
                            <p className="title">Total messages</p>
                            <p className="number">{stationState?.stationSocketData?.total_messages?.toLocaleString() || 0}</p>
                        </div>
                    </div>
                    <div className="details-wrapper pointer">
                        <div className="icon">
                            <img src={averageMesIcon} width={24} height={24} alt="averageMesIcon" />
                        </div>
                        <div className="more-details ">
                            <p className="title">Av. message size</p>
                            <TooltipComponent text="Gross size. Payload + headers + Memphis metadata">
                                <p className="number">{convertBytes(stationState?.stationSocketData?.average_message_size)}</p>
                            </TooltipComponent>
                        </div>
                    </div>
                </div>
                <div className="info-buttons">
                    <div className="sdk">
                        <p>Code examples</p>
                        <span
                            onClick={() => {
                                setSdkModal(true);
                            }}
                        >
                            View details {'>'}
                        </span>
                    </div>
                    <div className="audit">
                        <p>Audit</p>
                        <span onClick={() => setAuditModal(true)}>View details {'>'}</span>
                    </div>
                </div>
                <Modal
                    width="710px"
                    height="720px"
                    clickOutside={() => {
                        setSdkModal(false);
                    }}
                    open={sdkModal}
                    displayButtons={false}
                >
                    <SdkExample stationName={stationState?.stationMetaData?.name} withHeader={true} />
                </Modal>
                <Modal
                    header={
                        <div className="audit-header">
                            <p className="title">Audit</p>
                            <div className="msg">
                                <InfoOutlined />
                                <p>Showing last 5 days</p>
                            </div>
                        </div>
                    }
                    displayButtons={false}
                    height="300px"
                    width="800px"
                    clickOutside={() => setAuditModal(false)}
                    open={auditModal}
                    hr={false}
                >
                    <Auditing />
                </Modal>
                <Modal
                    header="Enforce schema"
                    displayButtons={false}
                    height="400px"
                    width="352px"
                    clickOutside={() => setUseSchemaModal(false)}
                    open={useSchemaModal}
                    hr={true}
                    className="use-schema-modal"
                >
                    <UseSchemaModal
                        stationName={stationState?.stationMetaData?.name}
                        handleSetSchema={(schema) => {
                            setSchema(schema);
                            setUseSchemaModal(false);
                        }}
                        close={() => setUseSchemaModal(false)}
                    />
                </Modal>
                <Modal
                    header="Update schema"
                    displayButtons={false}
                    height="650px"
                    width="650px"
                    clickOutside={() => setUpdateSchemaModal(false)}
                    open={updateSchemaModal}
                    className="update-schema-modal"
                >
                    <UpdateSchemaModal
                        schemaSelected={stationState?.stationSocketData?.schema?.name}
                        stationName={stationState?.stationMetaData?.name}
                        dispatch={(schema) => {
                            setSchema(schema);
                            setUpdateSchemaModal(false);
                        }}
                        close={() => setUpdateSchemaModal(false)}
                    />
                </Modal>
                <Modal
                    header={<img src={deleteWrapperIcon} alt="deleteWrapperIcon" />}
                    width="520px"
                    height="240px"
                    displayButtons={false}
                    clickOutside={() => modalDeleteFlip(false)}
                    open={modalDeleteIsOpen}
                >
                    <DeleteItemsModal
                        title="Are you sure you want to delete this station?"
                        desc="Deleting this station means it will be permanently deleted."
                        buttontxt="I understand, delete the station"
                        handleDeleteSelected={handleDeleteStation}
                        loader={deleteLoader}
                    />
                </Modal>
                <Modal
                    header={<img src={stopUsingIcon} alt="stopUsingIcon" />}
                    width="520px"
                    height="240px"
                    displayButtons={false}
                    clickOutside={() => setDeleteModal(false)}
                    open={deleteModal}
                >
                    <DeleteItemsModal
                        title="Are you sure you want to detach schema from the station?"
                        desc="Detaching schema might interrupt producers from producing data"
                        buttontxt="I understand, detach schema"
                        textToConfirm="detach"
                        handleDeleteSelected={handleStopUseSchema}
                        loader={detachLoader}
                    />
                </Modal>
            </div>
        </div>
    );
};

export default StationOverviewHeader;
