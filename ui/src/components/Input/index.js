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
        fontSize
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
        value
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
