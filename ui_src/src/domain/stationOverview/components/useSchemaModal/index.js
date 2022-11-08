// Credit for The NATS.IO Authors
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

import placeholderSchema from '../../../../assets/images/placeholderSchema.svg';
import stopUsingIcon from '../../../../assets/images/stopUsingIcon.svg';
import DeleteItemsModal from '../../../../components/deleteItemsModal';
import searchIcon from '../../../../assets/images/searchIcon.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import SearchInput from '../../../../components/searchInput';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import SchemaItem from './schemaItem';
import { Context } from '../../../../hooks/store';
import { useHistory } from 'react-router-dom';
import pathDomains from '../../../../router';

const UseSchemaModal = ({ stationName, handleSetSchema, schemaSelected, close }) => {
    const [state, dispatch] = useContext(Context);

    const [schemaList, setSchemasList] = useState([]);
    const [copyOfSchemaList, setCopyOfSchemaList] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [selected, setSelected] = useState(schemaSelected);
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
            const results = schemaList.filter((schema) => schema?.name?.toLowerCase().includes(searchInput));
            setSchemasList(results);
        } else {
            setSchemasList(copyOfSchemaList);
        }
    }, [searchInput]);

    const useSchema = async () => {
        try {
            setUseschemaLoading(true);
            const data = await httpRequest('POST', ApiEndpoints.USE_SCHEMA, { station_name: stationName, schema_name: selected });
            if (data) {
                handleSetSchema(data);
                setUseschemaLoading(false);
            }
        } catch (error) {}
        setUseschemaLoading(false);
    };

    const handleStopUseSchema = async () => {
        try {
            setUseschemaLoading(true);
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_SCHEMA_FROM_STATION, { station_name: stationName });
            if (data) {
                handleSetSchema(data);
                setUseschemaLoading(false);
                setDeleteModal(false);
            }
        } catch (error) {}
        setUseschemaLoading(false);
    };

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const createNew = () => {
        dispatch({ type: 'SET_CREATE_SCHEMA', payload: true });
        history.push(pathDomains.schemas);
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
                        iconComponent={<img alt="search tag" src={searchIcon} />}
                        onChange={handleSearch}
                        value={searchInput}
                        width="100%"
                        height="35px"
                    />
                    <div className="schemas-list">
                        {schemaList?.map((schema) => {
                            return (
                                <SchemaItem
                                    schema={schema}
                                    schemaSelected={schemaSelected}
                                    selected={selected}
                                    handleSelectedItem={(id) => setSelected(id)}
                                    handleStopUseSchema={() => setDeleteModal(true)}
                                />
                            );
                        })}
                    </div>
                    <div className="buttons">
                        <Button
                            width="100%"
                            height="35px"
                            placeholder="Attach"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="13px"
                            fontFamily="InterSemiBold"
                            disabled={selected === schemaSelected || selected === ''}
                            isLoading={useschemaLoading}
                            onClick={useSchema}
                        />
                    </div>
                </>
            )}
            {!isLoading && schemaList?.length === 0 && (
                <div className="no-schema-to-display">
                    <img src={placeholderSchema} width="50" alt="placeholderSchema" />
                    <p className="title">No Schema exist</p>
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
                />
            </Modal>
        </div>
    );
};

export default UseSchemaModal;
