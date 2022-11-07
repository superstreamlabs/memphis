// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

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
