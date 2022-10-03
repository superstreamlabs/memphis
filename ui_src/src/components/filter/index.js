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

import React, { useState, useEffect } from 'react';
import MenuItem from '@material-ui/core/MenuItem';
import Popover from '@material-ui/core/Popover';
import filterImg from '../../assets/images/filter.svg';
import { MailOutlined, AppstoreOutlined, SettingOutlined } from '@ant-design/icons';
import { DownOutlined } from '@ant-design/icons';
import CustomCollapse from './customCollapse';
import { Dropdown, Menu, Space } from 'antd';
// import { getBorderRadius, getFontColor, getBackgroundColor, getBorderColor, getBoxShadows } from '../../utils/styleTemplates';

const Filter = (props) => {
    // const {
    //     placeholder,
    //     type,
    //     height,
    //     width,
    //     radiusType,
    //     colorType,
    //     backgroundColorType,
    //     onBlur,
    //     onChange,
    //     iconComponent,
    //     borderColorType,
    //     boxShadowsType,
    //     disabled,
    //     numberOfRows,
    //     value,
    //     opacity,
    //     id,
    //     minWidth,
    //     fontSize
    // } = props;

    // const handleChange = (e) => (onChange ? onChange(e) : '');
    const [anchorEl, setAnchorEl] = useState(null);
    const [filtersConter, setFilterCounter] = useState(0);
    const open = Boolean(anchorEl);

    const handleClickMenu = (event) => {
        setAnchorEl(event.currentTarget);
    };

    const handleCloseMenu = () => {
        setAnchorEl(null);
    };

    const [filterFields, setFilterFields] = useState([
        {
            name: 'tags',
            value: 'Tags',
            type: 'label',
            fields: [
                {
                    name: 'Github',
                    color: '#00A5FF',
                    background: 'rgba(0, 165, 255, 0.1)',
                    checked: false
                },
                {
                    name: 'Mixpod',
                    color: '#FFA043',
                    background: 'rgba(255, 160, 67, 0.1)',
                    checked: false
                },
                {
                    name: '2022',
                    color: '#5542F6',
                    background: 'rgba(85, 66, 246, 0.1)',
                    checked: false
                },
                {
                    name: 'Success',
                    color: '#20C9AC',
                    background: 'rgba(32, 201, 172, 0.1)',
                    checked: false
                }
            ]
        },
        {
            name: 'created',
            value: 'Created By',
            type: 'circle',
            fields: [
                {
                    name: 'sveta@memphis.dev',
                    color: '#FFC633',
                    checked: false
                },
                {
                    name: 'root',
                    color: 'yellowGreen',
                    checked: false
                }
            ]
        }
    ]);

    const handleCheck = (filterGroup, filterField) => {
        let data = filterFields;
        data[filterGroup].fields[filterField].checked = !data[filterGroup].fields[filterField].checked;
        setFilterFields(data);
    };

    const handleConfirm = () => {
        let counter = 0;
        filterFields.forEach((element) => {
            element.fields.forEach((field) => {
                if (field.checked) counter++;
            });
        });
        setFilterCounter(counter);
        props.filtersUpdated(filterFields);
        handleCloseMenu();
    };

    const handleCancel = () => {
        let data = filterFields;
        data.forEach((element) => {
            element.fields.forEach((field) => {
                field.checked = false;
            });
        });
        setFilterFields(data);
        handleCloseMenu();
    };
    return (
        <div className="filter-container">
            <img
                src={filterImg}
                width="25"
                height="25"
                alt="filter"
                onClick={(e) => {
                    e.preventDefault();
                    handleClickMenu(e);
                }}
            />
            Filters
            {filtersConter > 0 && <div className="filter-counter">{filtersConter}</div>}
            <Popover id="long-menu" classes={{ paper: 'Menu c' }} anchorEl={anchorEl} onClose={handleCloseMenu} open={open}>
                <CustomCollapse
                    header="Details"
                    data={filterFields}
                    onCheck={(filterGroup, filterField) => handleCheck(filterGroup, filterField)}
                    cancel={handleCancel}
                    confirm={handleConfirm}
                    defaultOpen={true}
                />
            </Popover>
        </div>
    );
};

export default Filter;
