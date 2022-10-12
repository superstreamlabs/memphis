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

import { Select } from 'antd';
import React from 'react';

import Arrow from '../../assets/images/arrow.svg';
import { FiberManualRecord } from '@material-ui/icons';

const { Option } = Select;

const SelectVersion = ({ options, onChange, value }) => {
    const handleChange = (e) => {
        onChange(e);
    };

    return (
        <div className="select-version-container">
            <Select
                className="select"
                value={value}
                bordered={false}
                suffixIcon={<img src={Arrow} alt="select-arrow" />}
                onChange={handleChange}
                placement="bottomRight"
                popupClassName="select-version-options"
            >
                {options?.map((option, index) => {
                    return (
                        <Option key={option?.id} value={option?.version_number}>
                            <p className="schema-name">Version {option?.version_number}</p>
                            <div className="scheme-details">
                                <p>Created by {option?.created_by_user}</p>
                                {(option.active || index === 0) && <FiberManualRecord />}
                                {option.active && <p>Current</p>}
                                {index === 0 && <p>Latest</p>}
                            </div>
                        </Option>
                    );
                })}
            </Select>
        </div>
    );
};

export default SelectVersion;
