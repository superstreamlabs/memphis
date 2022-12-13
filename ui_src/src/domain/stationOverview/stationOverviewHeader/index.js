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

import React, { useContext, useEffect, useState } from 'react';
import { Add, FiberManualRecord, InfoOutlined } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import { Segmented } from 'antd';

import { convertBytes, convertSecondsToDate, numberWithCommas } from '../../../services/valueConvertor';
import deleteWrapperIcon from '../../../assets/images/deleteWrapperIcon.svg';
import averageMesIcon from '../../../assets/images/averageMesIcon.svg';
import DeleteItemsModal from '../../../components/deleteItemsModal';
import awaitingIcon from '../../../assets/images/awaitingIcon.svg';
import TooltipComponent from '../../../components/tooltip/tooltip';
import UpdateSchemaModal from '../components/updateSchemaModal';
import deleteIcon from '../../../assets/images/deleteIcon.svg';
import VersionBadge from '../../../components/versionBadge';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import BackIcon from '../../../assets/images/backIcon.svg';
import UseSchemaModal from '../components/useSchemaModal';
import { httpRequest } from '../../../services/http';
import SdkExample from '../components/sdkExsample';
import TagsList from '../../../components/tagList';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import Modal from '../../../components/modal';
import Auditing from '../components/auditing';
import pathDomains from '../../../router';
import { StationStoreContext } from '..';
import ProtocolExample from '../components/protocolExsample';
import SegmentButton from '../../../components/segmentButton';

const StationOverviewHeader = () => {
    const [state, dispatch] = useContext(Context);
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [modalDeleteIsOpen, modalDeleteFlip] = useState(false);
    const history = useHistory();
    const [retentionValue, setRetentionValue] = useState('');
    const [sdkModal, setSdkModal] = useState(false);
    const [auditModal, setAuditModal] = useState(false);
    const [useSchemaModal, setUseSchemaModal] = useState(false);
    const [updateSchemaModal, setUpdateSchemaModal] = useState(false);
    const [segment, setSegment] = useState('Sdk');
    const [deleteLoader, setDeleteLoader] = useState(false);

    useEffect(() => {
        switch (stationState?.stationMetaData?.retention_type) {
            case 'message_age_sec':
                setRetentionValue(convertSecondsToDate(stationState?.stationMetaData?.retention_value));
                break;
            case 'bytes':
                setRetentionValue(`${stationState?.stationMetaData?.retention_value} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${stationState?.stationMetaData?.retention_value} messages`);
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
            setDeleteLoader(false);
            modalDeleteFlip(false);
            returnToStaionsList();
        } catch (error) {
            setDeleteLoader(false);
            modalDeleteFlip(false);
        }
    };

    return (
        <div className="station-overview-header">
            <div className="title-wrapper">
                <div className="station-details">
                    <div className="station-name">
                        <img src={BackIcon} onClick={() => returnToStaionsList()} alt="backIcon" />
                        <h1>{stationState?.stationMetaData?.name}</h1>
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
                        Created by {stationState?.stationMetaData?.created_by_user} at {stationState?.stationMetaData?.creation_date}
                    </span>
                </div>
                <div className="station-buttons">
                    <div className="station-actions" onClick={() => modalDeleteFlip(true)}>
                        <div className="action">
                            <img src={deleteIcon} alt="redirectIcon" />
                        </div>
                    </div>
                </div>
            </div>
            <div className="details">
                <div className="main-details">
                    <div className="left-side">
                        <p>
                            <b>Retention:</b> {retentionValue}
                        </p>
                        <p>
                            <b>Replicas:</b> {stationState?.stationMetaData?.replicas}
                        </p>
                        <p>
                            <b>Storage Type:</b> {stationState?.stationMetaData?.storage_type}
                        </p>
                    </div>
                    {stationState?.stationSocketData?.schema === undefined || Object.keys(stationState?.stationSocketData?.schema).length === 0 ? (
                        <div className="schema-details sd-center">
                            <div className="add-new">
                                <Button
                                    width="120px"
                                    height="25px"
                                    placeholder={
                                        <div className="use-schema-button">
                                            <Add />
                                            <p>Attach schema</p>
                                        </div>
                                    }
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    onClick={() => setUseSchemaModal(true)}
                                />
                            </div>
                        </div>
                    ) : (
                        <div className="schema-details sd-flex">
                            <div className="title-and-badge">
                                <p className="title">Schema</p>
                                {stationState?.stationSocketData?.schema?.updates_available && <VersionBadge content="Updates available" active={false} />}
                                {!stationState?.stationSocketData?.schema?.updates_available && <VersionBadge content="Updated" active={true} />}
                            </div>
                            <div className="name-and-version">
                                <p>{stationState?.stationSocketData?.schema?.name}</p>
                                <FiberManualRecord />
                                <p>v{stationState?.stationSocketData?.schema?.version_number}</p>
                            </div>
                            <div className="buttons">
                                <Button
                                    width="80px"
                                    minWidth="80px"
                                    height="16px"
                                    placeholder="Edit / Detach"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="10px"
                                    fontFamily="InterMedium"
                                    onClick={() => setUseSchemaModal(true)}
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
                                        onClick={() => setUpdateSchemaModal(true)}
                                    />
                                )}
                            </div>
                        </div>
                    )}
                </div>
                <div className="icons-wrapper">
                    <div className="details-wrapper">
                        <div className="icon">
                            <img src={awaitingIcon} width={22} height={44} alt="awaitingIcon" />
                        </div>
                        <div className="more-details">
                            <p className="title">Total messages</p>
                            <p className="number">{numberWithCommas(stationState?.stationSocketData?.total_messages) || 0}</p>
                        </div>
                    </div>
                    <div className="details-wrapper average">
                        <div className="icon">
                            <img src={averageMesIcon} width={24} height={24} alt="averageMesIcon" />
                        </div>
                        <div className="more-details">
                            <p className="title">Av. message size</p>
                            <TooltipComponent text="Gross size. Payload + headers + Memphis metadata">
                                <p className="number">{convertBytes(stationState?.stationSocketData?.average_message_size)}</p>
                            </TooltipComponent>
                        </div>
                    </div>
                    {/* <div className="details-wrapper">
                        <div className="icon">
                            <img src={memoryIcon} width={24} height={24} alt="memoryIcon" />
                        </div>
                        <div className="more-details">
                            <p className="number">20Mb/80Mb</p>
                            <Progress showInfo={false} status={(20 / 80) * 100 > 60 ? 'exception' : 'success'} percent={(20 / 80) * 100} size="small" />
                            <p className="title">Mem</p>
                        </div>
                    </div> */}
                    {/* <div className="details-wrapper">
                        <div className="icon">
                            <img src={cpuIcon} width={22} height={22} alt="cpuIcon" />
                        </div>
                        <div className="more-details">
                            <p className="number">50%</p>
                            <Progress showInfo={false} status={(35 / 100) * 100 > 60 ? 'exception' : 'success'} percent={(35 / 100) * 100} size="small" />
                            <p className="title">CPU</p>
                        </div>
                    </div> */}
                    {/* <div className="details-wrapper">
                        <div className="icon">
                            <img src={storageIcon} width={30} height={30} alt="storageIcon" />
                        </div>
                        <div className="more-details">
                            <p className="number">{60}Mb/100Mb</p>
                            <Progress showInfo={false} status={(60 / 100) * 100 > 60 ? 'exception' : 'success'} percent={(60 / 100) * 100} size="small" />
                            <p className="title">Storage</p>
                        </div>
                    </div> */}
                </div>
                <div className="info-buttons">
                    <div className="sdk">
                        <p>Code example</p>
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
                    header={
                        <div className="sdk-header">
                            <p className="title">Code example</p>
                            <SegmentButton options={['Sdk', 'Protocol']} onChange={(e) => setSegment(e)} />
                        </div>
                    }
                    width="710px"
                    clickOutside={() => {
                        setSdkModal(false);
                        setSegment('Sdk');
                    }}
                    open={sdkModal}
                    displayButtons={false}
                >
                    {segment === 'Sdk' && <SdkExample />}
                    {segment === 'Protocol' && <ProtocolExample />}
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
                    header="Attach schema"
                    displayButtons={false}
                    height="400px"
                    width="352px"
                    clickOutside={() => setUseSchemaModal(false)}
                    open={useSchemaModal}
                    hr={true}
                    className="use-schema-modal"
                >
                    <UseSchemaModal
                        schemaSelected={stationState?.stationSocketData?.schema?.name || ''}
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
                    width="550px"
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
            </div>
        </div>
    );
};

export default StationOverviewHeader;
