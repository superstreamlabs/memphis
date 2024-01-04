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

import React, { useState } from 'react';
import { DatePicker } from 'antd';
import { ReactComponent as CalendarIcon } from 'assets/images/Calendar.svg';
const DatePickerComponent = ({ width, height, minWidth, onChange, placeholder, picker, dateFrom }) => {
    const [disabledMonths, setDisabledMonths] = useState([]);

    const disabledDate = (current) => {
        const startingDate = dateFrom ? new Date(dateFrom) : new Date('2023-06-01');
        const disabledBefore = current && current < startingDate;
        const disabledAfter = current && current > new Date();

        return disabledBefore || disabledAfter;
    };

    const onOpenChange = (open) => {
        if (open) {
            const disabledMonths = [];
            let currentDate = new Date();
            while (currentDate > new Date('2023-06-01')) {
                disabledMonths.push(currentDate);
                currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() - 1);
            }
            setDisabledMonths(disabledMonths);
        }
    };
    return (
        <div className="date-picker-container">
            <DatePicker
                onChange={(date, dateString) => onChange(date._d)}
                placeholder={placeholder}
                suffixIcon={<CalendarIcon />}
                popupClassName="date-picker-popup"
                picker={picker}
                allowClear={false}
                style={{
                    height: height,
                    width: width,
                    minWidth: minWidth || '100px',
                    fontSize: '10px',
                    border: '1px solid #D8D8D8',
                    boxShadow: '0px 1px 3px rgba(0, 0, 0, 0.12)',
                    borderRadius: '32px',
                    zIndex: 9999
                }}
                disabledDate={picker === 'month' && disabledDate}
                onOpenChange={onOpenChange}
                disabledMonths={disabledMonths}
            />
        </div>
    );
};

export default DatePickerComponent;
