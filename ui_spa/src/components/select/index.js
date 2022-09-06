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

import { getFontColor, getBackgroundColor, getBorderColor, getBoxShadows, getBorderRadius } from '../../utils/styleTemplates';
import Arrow from '../../assets/images/arrow.svg';

const { Option } = Select;

const SelectComponent = (props) => {
    const {
        options = [],
        width,
        onChange,
        colorType,
        value,
        backgroundColorType,
        borderColorType,
        dropdownClassName,
        boxShadowsType,
        radiusType,
        size,
        dropdownStyle,
        height,
        customOptions,
        disabled
    } = props;

    const handleChange = (e) => {
        onChange(e);
    };

    const color = getFontColor(colorType);
    const backgroundColor = getBackgroundColor(backgroundColorType);
    const borderColor = getBorderColor(borderColorType);
    const boxShadow = getBoxShadows(boxShadowsType);
    const borderRadius = getBorderRadius(radiusType);

    const fieldProps = {
        onChange: handleChange,
        disabled,
        style: {
            width,
            color,
            backgroundColor,
            boxShadow,
            borderColor,
            borderRadius,
            height: height || '40px'
        }
    };

    return (
        <div className="select-container">
            <Select
                {...fieldProps}
                className="select"
                size={size}
                dropdownClassName={dropdownClassName}
                value={value}
                suffixIcon={<img src={Arrow} alt="select-arrow" />}
                dropdownStyle={dropdownStyle}
            >
                {customOptions && options}
                {!customOptions &&
                    options.map((option) => (
                        <Option key={option?.id || option} disabled={option?.disabled || false}>
                            {option?.name || option}
                        </Option>
                    ))}
            </Select>
        </div>
    );
};

export default SelectComponent;
