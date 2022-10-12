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

import React, { useEffect, useContext, useState } from 'react';
import { SearchOutlined } from '@ant-design/icons';

import placeholderSchema from '../../../../assets/images/placeholderSchema.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import SearchInput from '../../../../components/searchInput';
import { httpRequest } from '../../../../services/http';
import Loader from '../../../../components/loader';
import Button from '../../../../components/button';
import { Context } from '../../../../hooks/store';
import SchemaBox from '../schemaBox';

function SchemaList({ createNew }) {
    const [state, dispatch] = useContext(Context);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [schemaList, setSchemaList] = useState([]);
    const [isLoading, setisLoading] = useState(false);

    const getAllSchemas = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            setSchemaList(data);
            setisLoading(false);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
        getAllSchemas();
    }, []);

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(schemaList.map((li) => li.name));
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

    const handleDeleteSelected = () => {
        setisLoading(true);
        isCheck.forEach(async (name) => {
            try {
                const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_SCHEMA, {
                    schema_name: name
                });
                if (data) {
                    setSchemaList(schemaList.filter((schema) => schema.name !== name));
                    setIsCheck(isCheck.filter((item) => item !== name));
                    setTimeout(() => {
                        setisLoading(false);
                    }, 500);
                }
            } catch (error) {
                setisLoading(false);
            }
        });
    };
    return (
        <div className="schema-container">
            <h1 className="main-header-h1">Schema</h1>
            <div className="action-section">
                {isCheck?.length > 0 && (
                    <Button
                        width="131px"
                        height="34px"
                        placeholder="Delete Selected"
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        onClick={() => handleDeleteSelected()}
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
                    iconComponent={<SearchOutlined />}
                    // onChange={handleSearch}
                    // value={searchInput}
                />
                <Button
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
                />
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
            <div className="schema-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {schemaList.map((schema, index) => {
                    return <SchemaBox key={index} schema={schema} isCheck={isCheck.includes(schema.name)} handleCheckedClick={handleCheckedClick} />;
                })}
                {!isLoading && schemaList.length === 0 && (
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
            </div>
        </div>
    );
}

export default SchemaList;
