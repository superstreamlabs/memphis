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

import React from 'react';
import { DatePicker } from 'antd';
import CalendarIcon from '../../assets/images/Calendar.svg';

const DatePickerComponent = ({ width, height, minWidth, onChange, placeholder }) => {
    return (
        <div className="date-picker-container">
            <DatePicker
                onChange={(date, dateString) => (dateString ? onChange(date._d) : onChange(''))}
                placeholder={placeholder}
                suffixIcon={<img src={CalendarIcon} />}
                popupClassName="date-picker-popup"
                style={{
                    height: height,
                    width: width,
                    minWidth: minWidth || '100px',
                    fontSize: '10px',
                    border: '1px solid #D8D8D8',
                    borderRadius: '4px',
                    zIndex: 9999
                }}
            />
        </div>
    );
};

export default DatePickerComponent;
