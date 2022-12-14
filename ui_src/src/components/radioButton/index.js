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

import { Radio, Space } from 'antd';
import React from 'react';

const RadioButton = ({ options = [], radioValue, onChange, onClick, optionType, disabled, vertical, fontFamily, radioWrapper, labelType, height, radioStyle }) => {
    const handleChange = (e) => {
        onChange(e);
    };

    const fieldProps = {
        onChange: handleChange,
        value: radioValue
    };

    return (
        <div className="radio-button">
            <Radio.Group
                {...fieldProps}
                className={vertical ? 'radio-group gr-vertical' : 'radio-group'}
                optionType={optionType ? optionType : null}
                disabled={disabled}
                defaultValue={radioValue || options[0]?.value}
            >
                {options.map((option) => (
                    <div
                        key={option.value}
                        style={{ height: height }}
                        className={labelType ? (radioValue === option.value ? 'label-type radio-value' : 'label-type') : radioWrapper || 'radio-wrapper'}
                        onClick={() => (labelType ? onClick(option.value) : '')}
                    >
                        <span
                            className={labelType ? (radioValue === option.value ? 'radio-style radio-selected' : 'radio-style') : `label ${radioStyle}`}
                            style={{ fontFamily: fontFamily }}
                        >
                            <Radio key={option.id} value={option.value} disabled={option.disabled || false}>
                                <p className="label-option-text"> {option.label}</p>
                            </Radio>
                        </span>
                        {option.description && <span className="des">{option.description}</span>}
                    </div>
                ))}
            </Radio.Group>
        </div>
    );
};

export default RadioButton;
