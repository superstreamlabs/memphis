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

import { InputNumber } from 'antd';
import React from 'react';

import ArrowDropDownRounded from '@material-ui/icons/ArrowDropDownRounded';
import ArrowDropUpRounded from '@material-ui/icons/ArrowDropUpRounded';

const InputNumberComponent = ({ min, max, onChange, value, placeholder, disabled }) => {
    const handleChange = (e) => {
        onChange(e);
    };

    return (
        <InputNumber
            bordered={false}
            min={min}
            max={max}
            keyboard={true}
            onChange={(e) => handleChange(e)}
            value={value}
            placeholder={placeholder}
            disabled={disabled}
            className="input-number-wrapper"
            controls={{ downIcon: <ArrowDropDownRounded />, upIcon: <ArrowDropUpRounded /> }}
        />
    );
};

export default InputNumberComponent;
