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

import { Input } from 'antd';
import React from 'react';

import { getFontColor, getBackgroundColor, getBorderRadius, getBorderColor, getBoxShadows } from '../../utils/styleTemplates';

const SearchInput = (props) => {
    const {
        placeholder,
        height,
        width,
        colorType,
        backgroundColorType,
        onChange,
        iconComponent,
        borderRadiusType,
        borderBottom,
        borderColorType,
        boxShadowsType,
        value,
        onPressEnter,
        className
    } = props;

    const handleChange = (e) => onChange(e);
    const handlePressEnter = (e) => onPressEnter(e);

    const color = getFontColor(colorType);
    const backgroundColor = getBackgroundColor(backgroundColorType);
    const borderRadius = getBorderRadius(borderRadiusType);
    const padding = 0;
    const borderColor = getBorderColor(borderColorType);
    const boxShadow = getBoxShadows(boxShadowsType);

    const fieldProps = {
        placeholder,
        onChange: handleChange,
        onPressEnter: handlePressEnter,
        style: { width, height, color, backgroundColor, padding, borderBottom, borderRadius, borderColor, boxShadow },
        value
    };

    return (
        <div className="search-input-container">
            <Input {...fieldProps} bordered={false} prefix={<div className="search-icon">{iconComponent}</div>}></Input>
        </div>
    );
};

export default SearchInput;
