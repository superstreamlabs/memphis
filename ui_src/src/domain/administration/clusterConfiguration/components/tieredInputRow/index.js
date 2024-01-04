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

import React, { useState, useEffect } from 'react';
import Input from 'components/Input';
import SelectComponent from 'components/select';
import { tieredStorageTimeValidator } from 'services/valueConvertor';

function TieredInputRow({ title, desc, value, onChanges, img }) {
    const [inputValue, setInputValue] = useState(value);
    const tsTimeOptions = ['Seconds', 'Minutes'];
    const [tsTimeType, setTsTimeType] = useState(tsTimeOptions[0]);
    const [error, setError] = useState('');

    const onChange = (newValue) => {
        let val = Number(newValue);
        if (tsTimeType === tsTimeOptions[1]) {
            val = val * 60;
        }
        let status = tieredStorageTimeValidator(val, tsTimeType);
        setError(status);
        onChanges(val, status);
        setInputValue(newValue);
    };
    const onChangeType = (type) => {
        let val = Number(inputValue);
        setTsTimeType(type);
        if (type === tsTimeOptions[1]) {
            val = val * 60;
        }
        let status = tieredStorageTimeValidator(val);
        setError(status);
        onChanges(val, status);
    };

    return (
        <div className="configuration-list-container">
            <div className="name">
                <img src={img} alt="ConfImg2" />
                <div>
                    <p className="conf-name">{title}</p>
                    <label className="conf-description">{desc}</label>
                </div>
            </div>
            <div className="input">
                <div className="input-and-error">
                    <Input
                        value={inputValue}
                        type="number"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        height="38px"
                        onChange={(e) => {
                            onChange(e.target.value);
                        }}
                        width="14vw"
                        minWidth="20px"
                    />
                    <div className="error">{error}</div>
                </div>
                <SelectComponent
                    colorType="black"
                    backgroundColorType="none"
                    fontFamily="Inter"
                    borderColorType="gray"
                    radiusType="semi-round"
                    height="38px"
                    popupClassName="select-options"
                    options={tsTimeOptions}
                    value={tsTimeType}
                    onChange={(e) => {
                        onChangeType(e);
                    }}
                    width="14vw"
                    minWidth="20px"
                />
            </div>
        </div>
    );
}

export default TieredInputRow;
