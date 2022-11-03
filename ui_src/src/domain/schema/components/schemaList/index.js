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

import React, { useEffect, useContext, useState, useCallback } from 'react';
import { useHistory } from 'react-router-dom';

import placeholderSchema from '../../../../assets/images/placeholderSchema.svg';
import deleteWrapperIcon from '../../../../assets/images/deleteWrapperIcon.svg';
import searchIcon from '../../../../assets/images/searchIcon.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import SearchInput from '../../../../components/searchInput';
import { httpRequest } from '../../../../services/http';
import Loader from '../../../../components/loader';
import Button from '../../../../components/button';
import { Context } from '../../../../hooks/store';
import Modal from '../../../../components/modal';
import pathDomains from '../../../../router';
import SchemaBox from '../schemaBox';
import { filterArray } from '../../../../services/valueConvertor';
import DeleteItemsModal from '../../../../components/deleteItemsModal';

function SchemaList() {
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [isLoading, setisLoading] = useState(true);
    const [deleteModal, setDeleteModal] = useState(false);
    const [searchInput, setSearchInput] = useState('');

    useEffect(() => {
        getAllSchemas();
        return () => {
            dispatch({ type: 'SET_DOMAIN_LIST', payload: [] });
            dispatch({ type: 'SET_FILTERED_LIST', payload: [] });
        };
    }, []);

    const getAllSchemas = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            dispatch({ type: 'SET_DOMAIN_LIST', payload: data });
            dispatch({ type: 'SET_FILTERED_LIST', payload: data });
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
        if (searchInput?.length >= 2) {
            dispatch({ type: 'SET_FILTERED_LIST', payload: state?.domainList.filter((schema) => schema.name.includes(searchInput)) });
        } else {
            dispatch({ type: 'SET_FILTERED_LIST', payload: state?.domainList });
        }
    }, [searchInput, state?.domainList]);

    const handleRegisterToSchema = useCallback(() => {
        state.socket?.emit('get_all_schemas_data');
    }, [state.socket]);

    useEffect(() => {
        state.socket?.on('schemas_overview_data', (data) => {
            dispatch({ type: 'SET_DOMAIN_LIST', payload: data });
        });

        state.socket?.on('error', (error) => {
            history.push(pathDomains.overview);
        });

        setTimeout(() => {
            handleRegisterToSchema();
        }, 1000);

        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(state.filteredList.map((li) => li.name));
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
        setisLoading(true);
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_SCHEMA, {
                schema_names: isCheck
            });
            if (data) {
                dispatch({ type: 'SET_DOMAIN_LIST', payload: filterArray(state.filteredList, isCheck) });
                setIsCheck([]);
                setisLoading(false);
            }
        } catch (error) {
            setisLoading(false);
        }
        setDeleteModal(false);
    };

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const createNew = () => {
        dispatch({ type: 'SET_CREATE_SCHEMA', payload: true });
    };

    return (
        <div className="schema-container">
            <div className="header-wraper">
                <label className="main-header-h1">
                    Schemas <label className="length-list">{state.filteredList?.length > 0 && `(${state.filteredList?.length})`}</label>
                </label>
                <div className="action-section">
                    <Button
                        width="131px"
                        height="34px"
                        placeholder={`Delete selected (${isCheck?.length})`}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        disabled={isCheck?.length === 0}
                        onClick={() => setDeleteModal(true)}
                    />

                    {state.filteredList?.length > 1 && (
                        <Button
                            width="131px"
                            height="34px"
                            placeholder="Select all"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            fontSize="12px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => onCheckedAll()}
                        />
                    )}
                    <SearchInput
                        placeholder="Search schema"
                        colorType="navy"
                        backgroundColorType="gray-dark"
                        width="288px"
                        height="34px"
                        borderRadiusType="circle"
                        borderColorType="none"
                        boxShadowsType="none"
                        iconComponent={<img src={searchIcon} alt="searchIcon" />}
                        onChange={handleSearch}
                        value={searchInput}
                    />
                    {/* <Button
                        width="111px"
                        height="34px"
                        placeholder={'Filters'}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        // onClick={() => addUserModalFlip(true)}
                    /> */}
                    {/* <Button
                        width="81px"
                        height="34px"
                        placeholder={'Sort'}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        // onClick={() => addUserModalFlip(true)}
                    /> */}
                    <Button
                        width="160px"
                        height="34px"
                        placeholder={'Create from blank'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        onClick={() => createNew()}
                    />
                    {/* <Button
                        width="145px"
                        height="34px"
                        placeholder={'Import schema'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        // onClick={() => createNew()}
                    /> */}
                </div>
            </div>
            <div className="schema-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading &&
                    state.filteredList?.map((schema, index) => {
                        return <SchemaBox key={index} schema={schema} isCheck={isCheck.includes(schema.name)} handleCheckedClick={handleCheckedClick} />;
                    })}
                {!isLoading && state.domainList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={placeholderSchema} width="100" height="100" alt="placeholderSchema" />
                        <p className="title">No Schema found</p>
                        <p className="sub-title">Get started by creating your first schema</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder="Create from blank"
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
                {!isLoading && state.domainList?.length > 0 && state.filteredList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={placeholderSchema} width="100" height="100" alt="placeholderSchema" />
                        <p className="title">No Schema found</p>
                        <p className="sub-title">Please try to search again</p>
                    </div>
                )}
            </div>
            <Modal
                header={<img src={deleteWrapperIcon} alt="deleteWrapperIcon" />}
                width="520px"
                height="240px"
                displayButtons={false}
                clickOutside={() => setDeleteModal(false)}
                open={deleteModal}
            >
                <DeleteItemsModal
                    title="Are you sure you want to delete the selected schemas?"
                    desc="Deleting these schemas means they will be permanently deleted."
                    buttontxt="I understand this consequences, delete the selected schemas"
                    handleDeleteSelected={handleDeleteSelected}
                />
            </Modal>
        </div>
    );
}

export default SchemaList;
