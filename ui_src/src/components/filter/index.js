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

import React, { createContext, useEffect, useReducer, useState } from 'react';

import Reducer from './hooks/reducer';

import './style.scss';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import filterImg from '../../assets/images/filter.svg';
import CustomCollapse from './customCollapse';
import { Popover } from 'antd';
import { filterType, labelType } from '../../const/filterConsts';
import { CircleLetterColor } from '../../const/circleLetterColor';

import Button from '../button';

const initialState = {
    isOpen: false,
    counter: 0,
    filterFields: []
};

const Filter = ({ filterComponent, stateRef, filtersUpdated, height }) => {
    const [filterState, filterDispatch] = useReducer(Reducer, initialState);
    const [tagList, setTagList] = useState([]);
    const [filterFields, setFilterFields] = useState([]);
    const [filterTerms, setFilterTerms] = useState([]);

    useEffect(() => {
        buildFilter();
    }, []);

    const buildFilter = () => {
        switch (filterComponent) {
            case 'stations':
                getTags();
                getFilterData(stateRef.current[0]);
                return;
            default:
                return;
        }
    };

    useEffect(() => {
        getTagsFilter();
    }, [tagList.length > 0]);

    useEffect(() => {
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filterFields });
    }, [filterFields]);

    useEffect(() => {
        handleFilter();
    }, [filterTerms]);

    const getFilterData = (stations) => {
        filterFields.findIndex((x) => x.name === 'created') === -1 && getCreatedByFilter(stations);
        filterFields.findIndex((x) => x.name === 'storage') === -1 && getStorageTypeFilter();
    };

    const getCreatedByFilter = (stations) => {
        let createdBy = [];
        stations.forEach((item) => {
            createdBy.push(item.station.created_by_user);
        });
        const created = [...new Set(createdBy)].map((user) => {
            return {
                name: user,
                color: CircleLetterColor[user[0].toUpperCase()],
                checked: false
            };
        });
        const cratedFilter = {
            name: 'created',
            value: 'Created By',
            labelType: labelType.CIRCLEDLETTER,
            filterType: filterType.CHECKBOX,
            fields: created
        };
        let filteredFields = filterFields;
        filteredFields.push(cratedFilter);
        setFilterFields(filteredFields);
    };
    const getStorageTypeFilter = () => {
        const storageTypeFilter = {
            name: 'storage',
            value: 'Storage Type',
            filterType: filterType.RADIOBUTTON,
            radioValue: -1,
            fields: [{ name: 'Memory' }, { name: 'File' }]
        };
        let filteredFields = filterFields;
        filteredFields.push(storageTypeFilter);
        setFilterFields(filteredFields);
    };

    const getTags = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_USED_TAGS}`);
            if (res) setTagList(res);
        } catch (err) {
            return;
        }
    };

    const getTagsFilter = () => {
        const fields = tagList.map((tag) => {
            return {
                name: tag.name,
                color: tag.color,
                checked: false
            };
        });
        const tagFilter = {
            name: 'tags',
            value: 'Tags',
            labelType: labelType.BADGE,
            filterType: filterType.CHECKBOX,
            fields: fields
        };
        let filteredFields = filterFields;
        const tagLocation = filterFields.findIndex((x) => x.name === 'tags');
        tagLocation === -1 ? filteredFields.splice(0, 0, tagFilter) : filteredFields.splice(tagLocation, 1, tagFilter);
        setFilterFields(filteredFields);
    };

    const flipOpen = () => {
        filterDispatch({ type: 'SET_IS_OPEN', payload: !filterState.isOpen });
    };

    const handleFilter = () => {
        switch (filterComponent) {
            case 'stations':
                let objTags = [];
                let objCreated = [];
                let objStorage = [];
                try {
                    objTags = filterTerms.find((o) => o.name === 'tags').fields.map((element) => element.toLowerCase());
                } catch {}
                try {
                    objCreated = filterTerms.find((o) => o.name === 'created').fields.map((element) => element.toLowerCase());
                } catch {}
                try {
                    objStorage = filterTerms.find((o) => o.name === 'storage').fields.map((element) => element.toLowerCase());
                } catch {}
                const data = stateRef.current[1]
                    .filter((item) => (objTags.length > 0 ? item.tags.some((tag) => objTags.includes(tag.name)) : !item.tags.some((tag) => objTags.includes(tag.name))))
                    .filter((item) => (objCreated.length > 0 ? objCreated.includes(item.station.created_by_user) : !objCreated.includes(item.station.created_by_user)))
                    .filter((item) => (objStorage.length > 0 ? objStorage.includes(item.station.storage_type) : !objStorage.includes(item.station.storage_type)));
                filtersUpdated(data);
                return;
            default:
                return;
        }
    };

    const handleApply = () => {
        let filterTerms = [];
        console.log(filterState?.filterFields);
        filterState?.filterFields.forEach((element) => {
            let term = {
                name: element.name,
                fields: []
            };
            if (element.filterType === filterType.CHECKBOX) {
                element.fields.forEach((field) => {
                    if (field.checked) {
                        let t = term.fields;
                        t.push(field.name);
                        term.fields = t;
                    }
                });
            } else if (element.filterType === filterType.RADIOBUTTON && element.radioValue !== -1) {
                let t = [];
                t.push(element.fields[element.radioValue].name);
                term.fields = t;
            } else {
                element.fields.forEach((field) => {
                    if (field.value !== '') {
                        let t = term.fields;
                        let d = {};
                        d[field.name] = field.value;
                        t.push(d);
                        term.fields = t;
                    }
                });
            }
            if (term.fields.length > 0) filterTerms.push(term);
        });
        setFilterTerms(filterTerms);
        flipOpen();
    };

    const handleClear = () => {
        filterDispatch({ type: 'SET_COUNTER', payload: 0 });
        let filter = filterFields;
        filter.map((filterGroup) => {
            switch (filterGroup.filterType) {
                case filterType.CHECKBOX:
                    filterGroup.fields.map((field) => (field.checked = false));
                case filterType.DATE:
                    filterGroup.fields.map((field) => (field.value = ''));
                case filterType.RADIOBUTTON:
                    filterGroup.radioValue = -1;
            }
        });
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filter });
    };

    const handleCancel = () => {
        filterDispatch({ type: 'SET_IS_OPEN', payload: false });
    };

    const content = <CustomCollapse header="Details" data={filterState?.filterFields} cancel={handleCancel} apply={handleApply} clear={handleClear} />;

    return (
        <FilterStoreContext.Provider value={[filterState, filterDispatch]}>
            <Popover className="filter-menu" placement="bottomLeft" content={content} trigger="click" onClick={() => flipOpen()} open={filterState.isOpen}>
                <Button
                    className="modal-btn"
                    width="110px"
                    height={height}
                    placeholder={
                        <div className="filter-container">
                            <img src={filterImg} width="25" alt="filter" />
                            <label className="filter-title">Filters</label>
                            {filterState?.apply && filterState?.counter > 0 && <div className="filter-counter">{filterState?.counter}</div>}
                        </div>
                    }
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="14px"
                    fontWeight="bold"
                    boxShadowStyle="login-input"
                    onClick={() => {}}
                />
            </Popover>
        </FilterStoreContext.Provider>
    );
};
export const FilterStoreContext = createContext({});
export default Filter;
