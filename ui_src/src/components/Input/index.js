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

import { Input as InputDesign } from 'antd';
import React from 'react';

import { getBorderRadius, getFontColor, getBackgroundColor, getBorderColor, getBoxShadows } from '../../utils/styleTemplates';

const Input = (props) => {
    const {
        placeholder,
        type,
        height,
        width,
        radiusType,
        colorType,
        backgroundColorType,
        onBlur,
        onChange,
        iconComponent,
        borderColorType,
        boxShadowsType,
        disabled,
        numberOfRows,
        value,
        opacity,
        id,
        minWidth,
        fontSize,
        onPressEnter,
        autoFocus = false
    } = props;

    const handleBlurChange = (e) => (onBlur ? onBlur(e) : '');
    const handleChange = (e) => (onChange ? onChange(e) : '');

    const { TextArea } = InputDesign;

    const borderRadius = getBorderRadius(radiusType);
    const color = getFontColor(colorType);
    const backgroundColor = getBackgroundColor(backgroundColorType);
    const borderColor = getBorderColor(borderColorType);
    const boxShadow = getBoxShadows(boxShadowsType);
    const rows = numberOfRows ? Number(numberOfRows) : 1;

    const fieldProps = {
        type,
        placeholder,
        onBlur: handleBlurChange,
        onChange: handleChange,
        id,
        style: {
            width,
            height,
            borderRadius,
            color,
            backgroundColor,
            borderColor,
            boxShadow,
            resize: 'none',
            opacity,
            minWidth: minWidth || '100px',
            fontSize: fontSize || '16px'
        },
        disabled,
        value,
        autoFocus,
        onPressEnter
    };
    const suffix = iconComponent !== undefined ? <div className="icon">{iconComponent}</div> : <span />;

    return (
        <div className="input-component-container">
            {type === 'textArea' ? (
                <div className="textarea-container">
                    <TextArea {...fieldProps} autoSize={{ minRows: rows, maxRows: rows }} />
                </div>
            ) : (
                <div className="input-container">
                    {type === 'password' && <InputDesign.Password {...fieldProps} prefix={suffix}></InputDesign.Password>}
                    {(type === 'text' || type === 'email' || type === 'number') && iconComponent !== undefined && (
                        <InputDesign {...fieldProps} prefix={suffix}></InputDesign>
                    )}
                    {(type === 'text' || type === 'email' || type === 'number') && iconComponent === undefined && <InputDesign {...fieldProps}></InputDesign>}
                </div>
            )}
        </div>
    );
};

export default Input;
