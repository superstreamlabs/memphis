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

import React, { useState, useEffect } from 'react';
import filterImg from '../../assets/images/filter.svg';
import CustomCollapse from './customCollapse';
import { Popover } from 'antd';
import { filterType, labelType } from '../../const/filterConsts';

const Filter = ({ filterFields, filtersUpdated, height }) => {
    const [open, setIsOpen] = useState(false);
    const [filtersConter, setFilterCounter] = useState(0);
    const [filter, setFilterFields] = useState(null);
    const [counter, setCounter] = useState(0);

    useEffect(() => {
        setFilterFields(filterFields);
    }, []);

    const handleCheck = (filterGroup, filterField, value) => {
        let updatedCounter = counter;
        let data = filter;
        if (filterFields[filterGroup].filterType === filterType.RADIOBUTTON) {
            if (data[filterGroup].radioValue === 0) updatedCounter++;
            data[filterGroup].radioValue = filterField;
        } else if (filterFields[filterGroup].filterType === filterType.DATE) {
            if (data[filterGroup].fields[filterField].value === '' && value !== '') updatedCounter++;
            else if (data[filterGroup].fields[filterField].value !== '' && value === '') updatedCounter--;
            data[filterGroup].fields[filterField].value = value;
        } else {
            if (data[filterGroup].fields[filterField].checked) updatedCounter--;
            else updatedCounter++;
            data[filterGroup].fields[filterField].checked = !data[filterGroup].fields[filterField].checked;
        }
        setCounter(updatedCounter);
        setFilterFields(data);
    };

    const handleApply = () => {
        let filterTerms = [];
        filter.forEach((element) => {
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
        setFilterCounter(counter);
        filtersUpdated(filterTerms);
        setIsOpen(false);
    };

    const handleCancel = () => {
        setIsOpen(false);
    };

    const handleClear = () => {
        setFilterFields(filterFields);
        setCounter(0);
        setFilterCounter(0);
    };

    const content = (
        <CustomCollapse
            header="Details"
            data={filter}
            filterCount={counter}
            onCheck={(filterGroup, filterField, value) => handleCheck(filterGroup, filterField, value)}
            cancel={handleCancel}
            apply={handleApply}
            clear={handleClear}
            defaultOpen={true}
        />
    );

    return (
        <Popover className="filter-menu" placement="bottomLeft" content={content} trigger="click" onClick={() => setIsOpen(!open)} open={open}>
            <div className="filter-container" style={{ height: height }}>
                <img src={filterImg} width="25" height="25" alt="filter" />
                Filters
                {filtersConter > 0 && <div className="filter-counter">{filtersConter}</div>}
            </div>
        </Popover>
    );
};

export default Filter;
