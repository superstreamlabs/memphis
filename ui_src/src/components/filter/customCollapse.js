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
import { Collapse } from 'antd';
import { Checkbox } from 'antd';
import { Divider } from 'antd';

import CollapseArrow from '../../assets/images/collapseArrow.svg';
import Button from '../button';
import DatePicker from '../datePicker';
import RadioButton from '../radioButton';
import { filterType, labelType } from '../../const/filterConsts';

const { Panel } = Collapse;

const CustomCollapse = ({ data, onCheck, cancel, confirm }) => {
    const [activeKey, setActiveKey] = useState(['0']);
    const onChange = (key) => {
        setActiveKey(key);
    };

    const drawComponent = (filterGroup, filterGroupIndex) => {
        switch (filterGroup.filterType) {
            case filterType.CHECKBOX:
                return drawCheckBox(filterGroup, filterGroupIndex);
            case filterType.DATE:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="" key={filterField.name}>
                            <label>{filterField.label}</label>
                            <DatePicker
                                placeholder="Type your name"
                                type="text"
                                radiusType="semi-round"
                                colorType="gray"
                                backgroundColorType="none"
                                borderColorType="red"
                                width="200px"
                                minWidth="200px"
                                onChange={(e) => console.log(e)}
                            />
                        </div>
                    );
                });
            case filterType.RADIOBUTTON:
                return (
                    <RadioButton
                        fontFamily="InterSemiBold"
                        options={filterGroup.fields.map((item, id) => {
                            return { id: id, value: id, label: item.name };
                        })}
                        radioValue={filterGroup.fields[0].label}
                        onChange={(e) => onCheck(filterGroupIndex, e.target.value)}
                    />
                );
        }
    };

    const drawCheckBox = (filterGroup, filterGroupIndex) => {
        switch (filterGroup.labelType) {
            case labelType.BADGE:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="label-container" key={filterField.name}>
                            <Checkbox checked={filterField.checked} onChange={() => onCheck(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />

                            <label className="label" style={{ color: filterField.color, backgroundColor: filterField.background }}>
                                {filterField.name}
                            </label>
                        </div>
                    );
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
        }
    };

    return (
        <Collapse ghost defaultActiveKey={['0']} onChange={onChange} className="custom-collapse-filter">
            <div className="collapse-header">Filter</div>
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
                    width={'100px'}
                    height="26px"
                    placeholder={'Cancle'}
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType={'white'}
                    border={'gray'}
                    fontSize="12px"
                    fontWeight="bold"
                    htmlType="submit"
                    onClick={cancel}
                />
                <Button
                    width={'100px'}
                    height="26px"
                    placeholder={'Confirm'}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType={'purple'}
                    fontSize="12px"
                    fontWeight="bold"
                    htmlType="submit"
                    onClick={confirm}
                />
            </div>
        </Collapse>
    );
};
export default CustomCollapse;
