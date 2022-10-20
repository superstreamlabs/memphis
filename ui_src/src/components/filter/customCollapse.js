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

import React, { useState } from 'react';
import { Collapse, Tag } from 'antd';
import { Checkbox } from 'antd';
import { Divider } from 'antd';

import CollapseArrow from '../../assets/images/collapseArrow.svg';
import Button from '../button';
import DatePicker from '../datePicker';
import RadioButton from '../radioButton';
import { filterType, labelType } from '../../const/filterConsts';

const { Panel } = Collapse;

const CustomCollapse = ({ data, filterCount, onCheck, cancel, apply, clear }) => {
    const [activeKey, setActiveKey] = useState(['0']);
    const onChange = (key) => {
        setActiveKey(key);
    };

    function drawComponent(filterGroup, filterGroupIndex) {
        switch (filterGroup.filterType) {
            case filterType.CHECKBOX:
                return drawCheckBox(filterGroup, filterGroupIndex);
            case filterType.DATE:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="date-container" key={filterField.name}>
                            <label>{filterField.label}</label>
                            <DatePicker
                                type="text"
                                radiusType="semi-round"
                                colorType="gray"
                                backgroundColorType="none"
                                borderColorType="red"
                                width="240px"
                                minWidth="200px"
                                onChange={(e) => onCheck(filterGroupIndex, filterFieldIndex, e)}
                            />
                        </div>
                    );
                });
            case filterType.RADIOBUTTON:
                return (
                    <RadioButton
                        filter
                        fontFamily="InterSemiBold"
                        options={filterGroup.fields.map((item, id) => {
                            return { id: id, value: id, label: item.name };
                        })}
                        radioValue={filterGroup.radioValue}
                        onChange={(e) => onCheck(filterGroupIndex, e.target.value)}
                    />
                );
        }
    }

    const drawCheckBox = (filterGroup, filterGroupIndex) => {
        switch (filterGroup.labelType) {
            case labelType.BADGE:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    // if (filterFieldIndex < 4)
                    return (
                        <div className="label-container" key={filterField.name}>
                            <Checkbox checked={filterField.checked} onChange={() => onCheck(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                            <Tag color={filterField.color}>{filterField.name}</Tag>
                        </div>
                    );
                    // else if (filterFieldIndex === 4) return <p className="show-more">Show All...</p>;
                });
            case labelType.CIRCLEDLETTER:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="circle-container" key={filterField.name}>
                            <Checkbox checked={filterField.checked} onChange={() => onCheck(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                            <p className="circle-letter" style={{ backgroundColor: filterField.color }}>
                                {filterField.name[0].toUpperCase()}
                            </p>
                            <label>{filterField.name}</label>
                        </div>
                    );
                });
            default:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="default-checkbox" key={filterField.name}>
                            <Checkbox checked={filterField.checked} onChange={() => onCheck(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                            <label>{filterField.name}</label>
                        </div>
                    );
                });
        }
    };

    return (
        <Collapse ghost defaultActiveKey={['0']} onChange={onChange} className="custom-collapse-filter">
            <div className="collapse-header">
                <div className="header-name-counter">
                    <label>Filter</label>
                    {filterCount > 0 && <div className="filter-counter">{filterCount}</div>}
                </div>
                <label className="clear" onClick={clear}>
                    Clear All
                </label>
            </div>
            {data.map((filterGroup, filterGroupIndex = 0) => (
                <Panel
                    header={
                        <div className="filter-header">
                            <label className="title">{filterGroup.value}</label>
                            <img
                                className={activeKey?.includes(filterGroupIndex.toString()) ? 'collapse-arrow open' : 'collapse-arrow'}
                                src={CollapseArrow}
                                alt="collapse-arrow"
                            />
                        </div>
                    }
                    key={`${filterGroupIndex}`}
                    showArrow={false}
                >
                    {drawComponent(filterGroup, filterGroupIndex)}
                    {filterGroupIndex + 1 < data?.length && <Divider />}
                </Panel>
            ))}
            <div className="collapse-footer">
                <Button
                    width="110px"
                    height="26px"
                    placeholder="Cancel"
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType={'white'}
                    border={'gray'}
                    fontSize="12px"
                    fontWeight="bold"
                    onClick={cancel}
                />
                <Button
                    width="110px"
                    height="26px"
                    placeholder="Apply"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType={'purple'}
                    fontSize="12px"
                    fontWeight="bold"
                    onClick={apply}
                />
            </div>
        </Collapse>
    );
};
export default CustomCollapse;
