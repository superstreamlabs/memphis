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

import { Select } from 'antd';
import React from 'react';

import { getFontColor, getBackgroundColor, getBorderColor, getBoxShadows, getBorderRadius } from '../../utils/styleTemplates';
import ArrowDropDownRounded from '@material-ui/icons/ArrowDropDownRounded';

const { Option } = Select;

const SelectComponent = ({
    options = [],
    width,
    onChange,
    colorType,
    value,
    backgroundColorType,
    borderColorType,
    popupClassName,
    boxShadowsType,
    radiusType,
    dropdownStyle,
    height,
    disabled,
    iconColor,
    fontSize,
    fontFamily,
    placeholder
}) => {
    const handleChange = (e) => {
        onChange(e);
    };

    const color = getFontColor(colorType);
    const backgroundColor = getBackgroundColor(backgroundColorType);
    const borderColor = getBorderColor(borderColorType);
    const boxShadow = getBoxShadows(boxShadowsType);
    const borderRadius = getBorderRadius(radiusType);
    const dropIconColor = getFontColor(iconColor || 'black');

    const fieldProps = {
        onChange: handleChange,
        disabled,
        placeholder,
        style: {
            width,
            color,
            backgroundColor,
            boxShadow,
            borderColor,
            borderRadius,
            height: height || '40px',
            fontFamily: fontFamily || 'Inter',
            fontSize: fontSize || '14px',
            overflowY: 'auto'
        }
    };

    return (
        <div className="select-container">
            <Select
                {...fieldProps}
                className="select"
                popupClassName={popupClassName}
                value={value}
                suffixIcon={<ArrowDropDownRounded style={{ color: dropIconColor }} />}
                dropdownStyle={dropdownStyle}
                listHeight={options.length > 10 ? 340 : options.length * 40}
            >
                {options.map((option) => (
                    <Option key={option?.id || option?.name || option} disabled={option?.disabled || false}>
                        <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                            <span style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
                                <img src={option?.icon} alt={option?.name} /> <label>{option?.name || option}</label>
                            </span>
                            {option?.comment && <span className="comment">{option?.comment}</span>}
                            {option?.soon && <span className="coming-soon-select">Coming soon</span>}
                        </span>
                    </Option>
                ))}
            </Select>
        </div>
    );
};

export default SelectComponent;
