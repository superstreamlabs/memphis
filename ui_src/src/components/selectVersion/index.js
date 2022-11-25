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

import { ArrowDropDownRounded, FiberManualRecord } from '@material-ui/icons';
import VersionBadge from '../versionBadge';
import { parsingDateWithotTime } from '../../services/valueConvertor';

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
                suffixIcon={<ArrowDropDownRounded className="drop-sown-icon" />}
                onChange={handleChange}
                placement="bottomRight"
                popupClassName="select-version-options"
            >
                {options?.map((option, index) => {
                    return (
                        <Option key={option?.id} value={option?.version_number}>
                            <div className="schema-name">
                                <p className="label">Version {option?.version_number}</p>
                                {option.active && (
                                    <>
                                        <VersionBadge content="Active" active={true} />
                                    </>
                                )}
                            </div>
                            <div className="scheme-details">
                                <p className="created-by">Created by {option?.created_by_user}</p>
                                <FiberManualRecord />
                                <p className="created-at">{parsingDateWithotTime(option?.creation_date)}</p>
                            </div>
                        </Option>
                    );
                })}
            </Select>
        </div>
    );
};

export default SelectVersion;
