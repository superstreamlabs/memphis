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

import { ArrowDropDownRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import React, { useContext } from 'react';
import { Select } from 'antd';

import SchemaIconSelect from '../../assets/images/schemaIconSelect.svg';
import placeholderSchema from '../../assets/images/placeholderSchema.svg';
import { parsingDate } from '../../services/valueConvertor';
import { Context } from '../../hooks/store';
import Button from '../button';
import pathDomains from '../../router';

const { Option } = Select;

const SelectSchema = ({ options, onChange, value, placeholder }) => {
    const history = useHistory();
    const [state, dispatch] = useContext(Context);

    const handleChange = (e) => {
        onChange(e);
    };

    const createNew = () => {
        dispatch({ type: 'SET_CREATE_SCHEMA', payload: true });
        history.push(pathDomains.schemas);
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
                notFoundContent={
                    <div className="no-schema-to-display">
                        <img src={placeholderSchema} width="50" height="50" alt="placeholderSchema" />
                        <p className="title">No schema found</p>
                        <p className="sub-title">Get started by creating your first schema</p>
                        <Button
                            className="modal-btn"
                            width="120px"
                            height="34px"
                            placeholder="Create schema"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            onClick={() => createNew()}
                        />
                    </div>
                }
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
