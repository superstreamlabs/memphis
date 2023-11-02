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

import React, { useContext, useEffect, useMemo, useState, useRef } from 'react';
import { AddRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';

import { ReactComponent as PlaceholderSchemaIcon } from '../../../../assets/images/placeholderSchema.svg';
import { ReactComponent as StopUsingIcon } from '../../../../assets/images/stopUsingIcon.svg';
import DeleteItemsModal from '../../../../components/deleteItemsModal';
import { ReactComponent as SearchIcon } from '../../../../assets/images/searchIcon.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import SearchInput from '../../../../components/searchInput';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import LearnMore from '../../../../components/learnMore';
import pathDomains from '../../../../router';
import { StationStoreContext } from '../..';
import SchemaItem from './schemaItem';
import { ReactComponent as StationIcon } from '../../../../assets/images/stationsIconActive.svg';
import CreateStationForm from '../../../../components/createStationForm';

const UseSchemaModal = ({ stationName, handleSetSchema, close, type = 'schema' }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const createStationRef = useRef(null);

    const [detachLoader, setDetachLoader] = useState(false);
    const [schemaList, setSchemasList] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [selected, setSelected] = useState();
    const [useschemaLoading, setUseschemaLoading] = useState(false);
    const [deleteModal, setDeleteModal] = useState(false);
    const [newStationModal, setNewStationModal] = useState(false);
    const [creatingProsessd, setCreatingProsessd] = useState(false);

    const history = useHistory();

    const getAllSchema = async () => {
        try {
            setIsLoading(true);
            const data = type === 'dls' ? await httpRequest('GET', ApiEndpoints.GET_ALL_STATIONS) : await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            if (data) {
                if (type === 'dls') {
                    setSchemasList(data?.filter((station) => station.name !== stationName));
                } else {
                    setSchemasList(data);
                }
            }
        } catch (error) {}
        setIsLoading(false);
    };

    useEffect(() => {
        getAllSchema();
    }, [creatingProsessd]);

    const listOfValues = useMemo(() => {
        return searchInput.length > 0 ? schemaList.filter((schema) => schema?.name?.toLowerCase()?.includes(searchInput)) : schemaList;
    }, [searchInput, schemaList]);

    const useSchema = async () => {
        try {
            setUseschemaLoading(true);
            const data = await httpRequest('POST', ApiEndpoints.USE_SCHEMA, { station_names: [stationName], schema_name: selected });
            if (data) {
                handleSetSchema(data);
                stationDispatch({ type: 'SET_SCHEMA_TYPE', payload: data.schema_type });
                setUseschemaLoading(false);
            }
        } catch (error) {
            setUseschemaLoading(false);
        }
    };

    const handleStopUseSchema = async () => {
        setDetachLoader(true);
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_SCHEMA_FROM_STATION, { station_name: stationName });
            if (data) {
                handleSetSchema(data);
                setDetachLoader(false);
                setDeleteModal(false);
            }
        } catch (error) {
            setDeleteModal(false);
            setDetachLoader(false);
        }
    };

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const createNew = () => {
        if (type === 'dls') {
            setNewStationModal(true);
        } else {
            history.push(`${pathDomains.schemaverse}/create`);
        }
    };
    return (
        <div className="use-schema-modal-container">
            {!isLoading && schemaList?.length > 0 && (
                <>
                    <SearchInput
                        placeholder={`Search ${type === 'dls' ? 'station' : 'schema'}`}
                        colorType="navy"
                        backgroundColorType="none"
                        borderRadiusType="circle"
                        borderColorType="search-input"
                        iconComponent={<SearchIcon />}
                        onChange={handleSearch}
                        value={searchInput}
                        width="100%"
                        height="35px"
                    />
                    <div className="schemas-list">
                        {listOfValues?.map((schema) => {
                            return (
                                <SchemaItem
                                    key={schema.name}
                                    schema={schema}
                                    selected={selected}
                                    handleSelectedItem={(id) => setSelected(id)}
                                    handleStopUseSchema={() => setDeleteModal(true)}
                                    type={type}
                                />
                            );
                        })}
                    </div>
                    <div className="buttons">
                        <div className="add-schema" onClick={() => createNew()}>
                            <AddRounded />
                            <p>Add new {type === 'dls' ? 'station' : 'schema'}</p>
                        </div>
                        {type === 'dls' ? (
                            <div className="btn-container">
                                <Button
                                    width="101px"
                                    height="35px"
                                    placeholder="Close"
                                    border="gray-light"
                                    colorType="black"
                                    radiusType="circle"
                                    backgroundColorType="white"
                                    fontSize="13px"
                                    fontFamily="InterSemiBold"
                                    onClick={close}
                                />
                                <Button
                                    width="101px"
                                    height="35px"
                                    placeholder="Consume"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="13px"
                                    fontFamily="InterSemiBold"
                                    disabled={selected === ''}
                                    isLoading={useschemaLoading}
                                    onClick={() => {
                                        setUseschemaLoading(true);
                                        handleSetSchema(selected);
                                        setUseschemaLoading(false);
                                    }}
                                />
                            </div>
                        ) : (
                            <Button
                                width="100%"
                                height="35px"
                                placeholder="Enforce"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="13px"
                                fontFamily="InterSemiBold"
                                disabled={selected === ''}
                                isLoading={useschemaLoading}
                                onClick={useSchema}
                            />
                        )}
                    </div>
                </>
            )}
            {!isLoading && schemaList?.length === 0 && (
                <div className="no-schema-to-display">
                    {type === 'dls' ? <StationIcon width={50} height={50} /> : <PlaceholderSchemaIcon width={50} alt="placeholderSchema" />}

                    <p className="title">{type === 'dls' ? 'No stations yet' : ' No schemas yet'}</p>
                    <p className="sub-title">
                        Get started by creating your first
                        {type === 'dls' ? ' station ' : ' schema'}
                    </p>
                    <Button
                        className="modal-btn"
                        width="160px"
                        height="34px"
                        placeholder={`Create new ${type === 'dls' ? 'station' : 'schema'}`}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontFamily="InterSemiBold"
                        aria-controls="usecse-menu"
                        aria-haspopup="true"
                        onClick={() => createNew()}
                    />
                </div>
            )}
            <Modal
                header={<StopUsingIcon alt="stopUsingIcon" />}
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
            <Modal
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <StationIcon className="headerImage" alt="stationImg" />
                        </div>
                        <p>Create a new station</p>
                        <label>
                            A station is a distributed unit that stores the produced data{' '}
                            <LearnMore url="https://docs.memphis.dev/memphis/memphis-broker/concepts/station" />
                        </label>
                    </div>
                }
                height="58vh"
                width="1020px"
                rBtnText="Create"
                lBtnText="Cancel"
                lBtnClick={() => {
                    setNewStationModal(false);
                }}
                rBtnClick={() => {
                    createStationRef.current();
                    setNewStationModal(false);
                }}
                clickOutside={() => setNewStationModal(false)}
                open={newStationModal}
                isLoading={creatingProsessd}
            >
                <CreateStationForm createStationFormRef={createStationRef} setLoading={(e) => setCreatingProsessd(e)} noRedirect={true} />
            </Modal>
        </div>
    );
};

export default UseSchemaModal;
