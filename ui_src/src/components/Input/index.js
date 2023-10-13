// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

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
        autoFocus = false,
        maxLength,
        suffixIconComponent
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
        maxLength: maxLength || null,
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
    const prefix = iconComponent !== undefined ? <div className="icon ">{iconComponent}</div> : null;
    const suffix = suffixIconComponent !== undefined ? <div className="icon ">{suffixIconComponent}</div> : null;
    return (
        <div className="input-component-container">
            {type === 'textArea' ? (
                <div className="textarea-container">
                    <TextArea {...fieldProps} autoSize={{ minRows: rows, maxRows: rows }} />
                </div>
            ) : (
                <div className="input-container">
                    {type === 'password' && <InputDesign.Password {...fieldProps} prefix={prefix} suffix={suffix}></InputDesign.Password>}
                    {(type === 'text' || type === 'email' || type === 'number') && <InputDesign {...fieldProps} prefix={prefix} suffix={suffix}></InputDesign>}
                </div>
            )}
        </div>
    );
};

export default Input;
