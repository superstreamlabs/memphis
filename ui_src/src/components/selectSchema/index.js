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
import ArrowDropDownRounded from '@material-ui/icons/ArrowDropDownRounded';
import SchemaIconSelect from '../../assets/images/schemaIconSelect.svg';

const { Option } = Select;

const SelectSchema = ({
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
    size,
    dropdownStyle,
    height,
    customOptions,
    disabled,
    iconColor,
    fontSize,
    fontFamily
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

    React.useEffect(() => {
        console.log(options);
    }, []);

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
            height: height || '40px',
            fontFamily: fontFamily || 'InterMedium',
            fontSize: fontSize || '14px'
        }
    };

    return (
        <Select
            className="select"
            value={value}
            // bordered={false}
            suffixIcon={<ArrowDropDownRounded className="drop-sown-icon" />}
            onChange={handleChange}
            placement="bottomRight"
            popupClassName="select-version-options"
        >
            {options.map((schema) => {
                return (
                    <Option key={schema?.id || schema?.name} value={schema?.name} disabled={schema?.disabled || false}>
                        <div className="scheme-details-container">
                            <div className="schema-details">
                                <img src={SchemaIconSelect} alt="SchemaIconSelect" height="20px" /> <p className="schema-name">Version {schema?.name}</p>
                            </div>

                            <p className="created-by">
                                {schema?.type} &#8226; {schema?.creation_date}
                            </p>
                        </div>
                    </Option>

                    // <Option key={schema?.id || schema?.name} value={schema?.name} disabled={schema?.disabled || false}>
                    //     {/* <div> */}
                    //     <div className="select-header">
                    //         <img src={SchemaIconSelect} alt="schemaIcon" height="25px" /> <p>{schema?.name}</p>
                    //     </div>
                    //     <p className="sub-title">{schema?.name}</p>
                    //     {/* </div> */}
                    // </Option>
                );
            })}
        </Select>
    );
};

export default SelectSchema;
