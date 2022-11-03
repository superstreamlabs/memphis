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

import { ArrowDropDownRounded } from '@material-ui/icons';
import SchemaIconSelect from '../../assets/images/schemaIconSelect.svg';
import { parsingDate } from '../../services/valueConvertor';
const { Option } = Select;

const SelectSchema = ({ options, onChange, value, placeholder }) => {
    const handleChange = (e) => {
        onChange(e);
    };

    return (
        <div className="select-schema-container">
            <Select
                className="select"
                placeholder={placeholder}
                value={value}
                bordered={false}
                suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                onChange={handleChange}
                placement="bottomRight"
                popupClassName="select-schema-options"
            >
                {options?.map((schema) => {
                    return (
                        <Option key={schema?.id} value={schema?.name}>
                            <div className="schema-details">
                                <img src={SchemaIconSelect} alt="SchemaIconSelect" height="20px" />
                                <p className="schema-name">{schema?.name}</p>
                            </div>
                            <p className="created-by">
                                {schema?.type} &#8226; {parsingDate(schema?.creation_date)}
                            </p>
                        </Option>
                    );
                })}
            </Select>
        </div>
    );
};

export default SelectSchema;
