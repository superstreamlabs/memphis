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

import { Radio } from 'antd';
import React from 'react';

const RadioButton = (props) => {
    const { options = [], radioValue, onChange, optionType, disabled, fontFamily } = props;

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
                className="radio-group"
                optionType={optionType ? optionType : null}
                disabled={disabled}
                defaultValue={radioValue || options[0]?.value}
            >
                {options.map((option) => (
                    <div
                        style={{
                            background: props.labelType && radioValue === option.value ? 'rgba(101, 87, 255, 0.1)' : '',
                            width: props.labelType ? '300px' : '',
                            border:
                                props.labelType && radioValue === option.value
                                    ? '1px solid #6557FF'
                                    : props.labelType && radioValue !== option.value
                                    ? '1px solid #EAECF0'
                                    : '',
                            padding: props.labelType ? '10px' : '',
                            marginRight: props.labelType ? '10px' : '',
                            borderRadius: props.labelType ? '8px' : '',
                            cursor: 'pointer'
                        }}
                        onClick={() => (props.labelType ? props.onClick(option.value) : '')}
                    >
                        <Radio key={option.id} value={option.value}>
                            <span
                                style={{
                                    fontFamily: fontFamily,
                                    color: props.labelType && radioValue === option.value ? '#6557FF' : '',
                                    fontSize: props.labelType ? '14px' : ''
                                }}
                            >
                                {option.label}
                            </span>
                        </Radio>
                    </div>
                ))}
            </Radio.Group>
        </div>
    );
};

export default RadioButton;
