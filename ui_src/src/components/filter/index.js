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

import React, { createContext, useEffect, useReducer } from 'react';

import Reducer from './hooks/reducer';

import './style.scss';

import filterImg from '../../assets/images/filter.svg';
import CustomCollapse from './customCollapse';
import { Popover } from 'antd';
import { filterType } from '../../const/filterConsts';
const initialState = {
    isOpen: false,
    counter: 0,
    filterFields: [],
    applied: false
};

const Filter = ({ filterFields, filtersUpdated, height }) => {
    const [filterState, filterDispatch] = useReducer(Reducer, initialState);

    useEffect(() => {
        const filter = filterFields;
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filter });
    }, []);

    const flipOpen = () => {
        filterDispatch({ type: 'SET_IS_OPEN', payload: !filterState.isOpen });
    };

    const handleApply = () => {
        filterDispatch({ type: 'SET_APPLY', payload: true });
        let filterTerms = [];
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
        filtersUpdated(filterTerms);
        flipOpen();
    };

    const handleClear = () => {
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
        filterDispatch({ type: 'SET_COUNTER', payload: 0 });
    };

    const handleCancel = () => {
        handleClear();
        filtersUpdated([]);
        filterDispatch({ type: 'SET_IS_OPEN', payload: false });
    };

    const content = <CustomCollapse header="Details" data={filterState?.filterFields} cancel={handleCancel} apply={handleApply} clear={handleClear} />;

    return (
        <FilterStoreContext.Provider value={[filterState, filterDispatch]}>
            <Popover className="filter-menu" placement="bottomLeft" content={content} trigger="click" onClick={() => flipOpen()} open={filterState.isOpen}>
                <div className="filter-container" style={{ height: height }}>
                    <img src={filterImg} width="25" height="25" alt="filter" />
                    Filters
                    {filterState?.apply && filterState?.counter > 0 && <div className="filter-counter">{filterState?.counter}</div>}
                </div>
            </Popover>
        </FilterStoreContext.Provider>
    );
};
export const FilterStoreContext = createContext({});
export default Filter;
