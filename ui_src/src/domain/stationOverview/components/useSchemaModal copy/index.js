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

import React, { useEffect, useState } from 'react';

import searchIcon from '../../../../assets/images/searchIcon.svg';
import { httpRequest } from '../../../../../services/http';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import SearchInput from '../../../../../components/searchInput';

const UseSchemaModal = () => {
    const [schemaList, setSchemasList] = useState([]);
    const [copyOfSchemaList, setCopyOfSchemaList] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);

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

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    return (
        <div className="use-schema-modal-container">
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
            />
            <div className="schemas-list">
                {schemaList?.map((schema) => {
                    return (
                        <div>
                            <p>{schema.name}</p>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default UseSchemaModal;
