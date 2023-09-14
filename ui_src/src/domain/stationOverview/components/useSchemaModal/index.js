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
import pathDomains from '../../../../router';
import { StationStoreContext } from '../..';
import SchemaItem from './schemaItem';

const UseSchemaModal = ({ stationName, handleSetSchema, close }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [detachLoader, setDetachLoader] = useState(false);
    const [schemaList, setSchemasList] = useState([]);
    const [copyOfSchemaList, setCopyOfSchemaList] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [selected, setSelected] = useState();
    const [useschemaLoading, setUseschemaLoading] = useState(false);
    const [deleteModal, setDeleteModal] = useState(false);
    const history = useHistory();

    const getAllSchema = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            if (data) {
                setSchemasList(data);
                setCopyOfSchemaList(data);
            }
        } catch (error) {}
        setIsLoading(false);
    };

    useEffect(() => {
        getAllSchema();
    }, []);

    useEffect(() => {
        if (searchInput.length > 1) {
            const results = schemaList.filter((schema) => schema?.name?.toLowerCase()?.includes(searchInput));
            setSchemasList(results);
        } else {
            setSchemasList(copyOfSchemaList);
        }
    }, [searchInput]);

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
        history.push(`${pathDomains.schemaverse}/create`);
    };
    return (
        <div className="use-schema-modal-container">
            {!isLoading && schemaList?.length > 0 && (
                <>
                    <SearchInput
                        placeholder="Search schema"
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
                        {schemaList?.map((schema) => {
                            return (
                                <SchemaItem
                                    key={schema.name}
                                    schema={schema}
                                    selected={selected}
                                    handleSelectedItem={(id) => setSelected(id)}
                                    handleStopUseSchema={() => setDeleteModal(true)}
                                />
                            );
                        })}
                    </div>
                    <div className="buttons">
                        <div className="add-schema" onClick={() => createNew()}>
                            <AddRounded />
                            <p>Add new schema</p>
                        </div>
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
                    </div>
                </>
            )}
            {!isLoading && schemaList?.length === 0 && (
                <div className="no-schema-to-display">
                    <PlaceholderSchemaIcon width={50} alt="placeholderSchema" />
                    <p className="title">No schemas yet</p>
                    <p className="sub-title">Get started by creating your first schema</p>
                    <Button
                        className="modal-btn"
                        width="160px"
                        height="34px"
                        placeholder="Create new schema"
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
        </div>
    );
};

export default UseSchemaModal;
