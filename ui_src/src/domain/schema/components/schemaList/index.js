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

import React, { useEffect, useContext, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { ReactComponent as PlaceholderSchema } from 'assets/images/placeholderSchema.svg';
import { ReactComponent as DeleteWrapperIcon } from 'assets/images/deleteWrapperIcon.svg';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import { useGetAllowedActions } from 'services/genericServices';
import Loader from 'components/loader';
import Button from 'components/button';
import Filter from 'components/filter';
import { Context } from 'hooks/store';
import Modal from 'components/modal';
import SchemaBox from '../schemaBox';
import { filterArray, isCloud } from 'services/valueConvertor';
import DeleteItemsModal from 'components/deleteItemsModal';
import { useHistory } from 'react-router-dom';
import pathDomains from 'router';

function SchemaList({ createNew }) {
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [isLoading, setisLoading] = useState(true);
    const [deleteModal, setDeleteModal] = useState(false);
    const [deleteLoader, setDeleteLoader] = useState(false);
    const location = useLocation();
    const getAllowedActions = useGetAllowedActions();
    useEffect(() => {
        getAllSchemas();
        return () => {
            dispatch({ type: 'SET_SCHEMA_LIST', payload: [] });
            dispatch({ type: 'SET_STATION_FILTERED_LIST', payload: [] });
        };
    }, []);

    useEffect(() => {
        location?.create && createNew(true);
    }, [location]);

    const getAllSchemas = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            dispatch({ type: 'SET_SCHEMA_LIST', payload: data });
            dispatch({ type: 'SET_STATION_FILTERED_LIST', payload: data });
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (error) {
            setisLoading(false);
        }
    };

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(state.schemaFilteredList?.map((li) => li.name));
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

    const handleDeleteSelected = async () => {
        setDeleteLoader(true);
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_SCHEMA, {
                schema_names: isCheck
            });
            if (data) {
                dispatch({ type: 'SET_SCHEMA_LIST', payload: filterArray(state.schemaFilteredList, isCheck) });
                setIsCheck([]);
                setIsCheckAll(false);
            }
        } catch (error) {
        } finally {
            setDeleteLoader(false);
            setDeleteModal(false);
            isCloud() && getAllowedActions();
        }
    };

    const createNewSchema = () => {
        history.push(`${pathDomains.schemaverse}/create`);
        createNew(true);
    };

    return (
        <div className="schema-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <label className="main-header-h1">
                        Schemaverse <label className="length-list">{state.schemaFilteredList?.length > 0 && `(${state.schemaFilteredList?.length})`}</label>
                    </label>
                    <span className="memphis-label">A modern approach to schema enforcement and increased data quality!</span>
                </div>
                <div className="action-section">
                    <Button
                        height="34px"
                        placeholder={`Delete selected (${isCheck?.length})`}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        boxShadowStyle="float"
                        disabled={isCheck?.length === 0}
                        isVisible={isCheck?.length !== 0}
                        onClick={() => setDeleteModal(true)}
                    />
                    <Button
                        width="131px"
                        height="34px"
                        placeholder={isCheckAll ? 'Unselect all' : 'Select all'}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        boxShadowStyle="float"
                        disabled={state?.schemaFilteredList?.length === 0}
                        onClick={() => onCheckedAll()}
                    />
                    <Filter filterComponent="schemaverse" height="34px" hideElement="search" />
                    <Button
                        width="160px"
                        height="34px"
                        placeholder={'Create a new schema'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        boxShadowStyle="float"
                        aria-haspopup="true"
                        onClick={createNewSchema}
                    />
                </div>
            </div>

            <div className="schema-list-top">
                <Filter filterComponent="schemaverse" height="34px" hideElement="filter" />
            </div>

            <div className="schema-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading &&
                    state.schemaFilteredList?.map((schema, index) => {
                        return <SchemaBox key={index} schemaBox={schema} isCheck={isCheck?.includes(schema.name)} handleCheckedClick={handleCheckedClick} />;
                    })}
                {!isLoading && state.schemaList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <PlaceholderSchema alt="placeholderSchema" width={100} height={100} />
                        <p className="title">No schemas yet</p>
                        <p className="sub-title">Get started by creating your first schema</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder="Create a new schema"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            onClick={createNewSchema}
                        />
                    </div>
                )}
                {!isLoading && state.schemaList?.length > 0 && state.schemaFilteredList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <PlaceholderSchema alt="placeholderSchema" width={100} height={100} />
                        <p className="title">No schemas found</p>
                        <p className="sub-title">Please try to search again</p>
                    </div>
                )}
            </div>
            <Modal
                header={<DeleteWrapperIcon alt="deleteWrapperIcon" />}
                width="520px"
                height="240px"
                displayButtons={false}
                clickOutside={() => setDeleteModal(false)}
                open={deleteModal}
            >
                <DeleteItemsModal
                    title="Are you sure you want to delete the selected schemas?"
                    desc="Deleting these schemas means they will be permanently deleted."
                    buttontxt="I understand, delete the selected schemas"
                    handleDeleteSelected={handleDeleteSelected}
                    loader={deleteLoader}
                />
            </Modal>
        </div>
    );
}

export default SchemaList;
