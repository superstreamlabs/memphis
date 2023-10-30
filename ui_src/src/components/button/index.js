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
    htmlType = 'submit',
    type = 'primary',
    fontFamily = 'Inter',
    tooltip
}) => {
    const handleClick = (e) => {
        onClick(e);
    };

    const borderRadius = getBorderRadius(radiusType);
    const color = getFontColor(colorType);
    const background = getBackgroundColor(backgroundColorType);
    const borderColor = border ? getBorderColor(border) : background;
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
        htmlType: htmlType,
        type: type,
        style: {
            borderRadius,
            color,
            background,
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
                <ButtonDesign {...fieldProps} className={disabled && 'noHover'}>
                    <span style={{ fontFamily: fontFamily }}>{placeholder}</span>
                </ButtonDesign>
            </TooltipComponent>
        </div>
    );
};

export default Button;
