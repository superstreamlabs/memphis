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

import { Button as ButtonDesign } from 'antd';
import React from 'react';

import { getBorderRadius, getFontColor, getBackgroundColor, getBoxShadows, getBorderColor } from '../../utils/styleTemplates';
import TooltipComponent from '../tooltip/tooltip';

const Button = ({
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
    marginLeft,
    boxShadowStyle,
    minHeight,
    zIndex,
    border,
    alignSelf,
    fontFamily = 'Inter',
    tooltip
}) => {
    const handleClick = (e) => {
        onClick(e);
    };

    const borderRadius = getBorderRadius(radiusType);
    const color = getFontColor(colorType);
    const backgroundColor = getBackgroundColor(backgroundColorType);
    const borderColor = border ? getBorderColor(border) : backgroundColor;
    const opacity = disabled ? '0.5' : '1';
    const boxShadow = boxShadowStyle ? getBoxShadows(boxShadowStyle) : 'none';
    const styleButtonContainer = {
        margin: margin,
        textAlign: textAlign,
        marginBottom: marginBottom,
        marginTop: marginTop,
        marginRight: marginRight,
        marginLeft: marginLeft,
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
            fontFamily,
            opacity,
            minHeight: minHeight,
            minWidth: minWidth || '60px',
            boxShadow,
            padding,
            zIndex: zIndex,
            lineHeight: fontSize
        },
        loading: isLoading
    };

    return (
        <div className="button-container" style={styleButtonContainer}>
            <TooltipComponent text={tooltip}>
                <ButtonDesign {...fieldProps} type="primary" htmlType="submit" className={disabled && 'noHover'}>
                    <span style={{ fontFamily: fontFamily }}>{placeholder}</span>
                </ButtonDesign>
            </TooltipComponent>
        </div>
    );
};

export default Button;
