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

import CollapseArrow from '../../assets/images/collapseArrow.svg';
import StatusIndication from '../../components/indication';
import Copy from '../../components/copy';
import Button from '../button';

const { Panel } = Collapse;

const CustomCollapse = (props) => {
    const [activeKey, setActiveKey] = useState(props.defaultOpen ? ['1'] : []);
    const onChange = (key) => {
        setActiveKey(key);
    };

    useEffect(() => {
        // console.log(props.data);
    }, []);

    return (
        <Collapse ghost defaultActiveKey={activeKey} onChange={onChange} className="filter-collapse">
            <p className="filter-header">Filter</p>
            {props.data.map((filterGroup, filterGroupIndex = 0) => {
                return (
                    <Panel
                        showArrow={false}
                        header={
                            <div className="collapse-header">
                                <p className="title">{filterGroup.value}</p>
                                <img className={activeKey[0] === '1' ? 'collapse-arrow open' : 'collapse-arrow close'} src={CollapseArrow} alt="collapse-arrow" />
                            </div>
                        }
                        key={filterGroup.name}
                    >
                        <div className="collapse-body">
                            {filterGroup.type === 'label' &&
                                filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                                    return (
                                        <div className="label" key={filterField.name}>
                                            <Checkbox checked={filterField.checked} onChange={() => props.onCheck(filterGroupIndex, filterFieldIndex)} name="checkedG" />

                                            <label style={{ color: filterField.color, backgroundColor: filterField.background }}>{filterField.name}</label>
                                        </div>
                                    );
                                })}
                            {filterGroup.type === 'circle' &&
                                filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                                    return (
                                        <div className="circle" key={filterField.name}>
                                            <Checkbox checked={filterField.checked} onChange={() => props.onCheck(filterGroupIndex, filterFieldIndex)} name="checkedG" />
                                            <p className="circle-letter" style={{ backgroundColor: filterField.color }}>
                                                {filterField.name[0]}
                                            </p>
                                            <label>{filterField.name}</label>
                                        </div>
                                    );
                                })}
                        </div>
                    </Panel>
                );
            })}
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
                    onClick={() => props.cancel()}
                    // isLoading={getStartedState?.isLoading}
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
                    onClick={() => props.confirm()}
                    // isLoading={getStartedState?.isLoading}
                />
            </div>
        </Collapse>
    );
};

export default CustomCollapse;
