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

import { Button as ButtonDesign } from 'antd';
import React from 'react';

import { getBorderRadius, getFontColor, getBackgroundColor, getBoxShadows } from '../../utils/styleTemplates';

const Button = (props) => {
    const {
        width,
        height,
        placeholder,
        radiusType,
        colorType,
        onClick,
        backgroundColorType,
        fontSize,
        fontWeight,
        disabled,
        margin,
        isLoading,
        padding,
        textAlign,
        minWidth,
        marginBottom,
        marginTop,
        marginRight,
        boxShadowStyle,
        minHeight,
        zIndex,
        border,
        alignSelf
    } = props;

    const handleClick = (e) => {
        onClick(e);
    };

    const borderRadius = getBorderRadius(radiusType);
    const color = getFontColor(colorType);
    const backgroundColor = getBackgroundColor(backgroundColorType);
    const borderColor = border ? getBackgroundColor(border) : backgroundColor;
    const opacity = disabled ? '0.5' : '1';
    const boxShadow = getBoxShadows(boxShadowStyle);

    const styleButtonContainer = {
        margin: margin,
        textAlign: textAlign,
        marginBottom: marginBottom,
        marginTop: marginTop,
        marginRight: marginRight,
        alignSelf: alignSelf
    };

    const fieldProps = {
        onClick: handleClick,
        disabled,
        style: {
            borderRadius,
            color,
            backgroundColor,
            width,
            height,
            borderColor,
            fontSize,
            fontWeight,
            opacity,
            minHeight: minHeight,
            minWidth: minWidth || '60px',
            padding,
            zIndex: zIndex,
            boxShadow
        },
        loading: isLoading
    };

    return (
        <div className="button-container" style={styleButtonContainer}>
            <ButtonDesign {...fieldProps} type="primary" htmlType="submit">
                {placeholder}
            </ButtonDesign>
        </div>
    );
};

export default Button;
