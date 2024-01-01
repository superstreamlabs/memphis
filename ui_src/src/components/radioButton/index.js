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

import { Radio } from 'antd';
import React from 'react';
import TooltipComponent from 'components/tooltip/tooltip';
import CloudOnly from 'components/cloudOnly';

const RadioButton = ({ options = [], radioValue, onChange, optionType, disabled, vertical, fontFamily, radioWrapper, labelType, height, radioStyle }) => {
    const handleChange = (e) => {
        onChange(e);
    };

    const fieldProps = {
        onChange: handleChange,
        value: radioValue
    };

    return (
        <div className="radio-button">
            <Radio.Group
                {...fieldProps}
                className={vertical ? 'radio-group gr-vertical' : 'radio-group'}
                optionType={optionType ? optionType : null}
                disabled={disabled}
                defaultValue={radioValue || options[0]?.value}
            >
                {options.map((option) =>
                    option.tooltip ? (
                        <TooltipComponent key={option.value} text={option.tooltip}>
                            <div
                                key={option.value}
                                style={{ height: height }}
                                className={labelType ? (radioValue === option.value ? 'label-type radio-value' : 'label-type') : radioWrapper || 'radio-wrapper'}
                            >
                                <span
                                    className={labelType ? (radioValue === option.value ? 'radio-style radio-selected' : 'radio-style') : `label ${radioStyle}`}
                                    style={{ fontFamily: fontFamily }}
                                >
                                    <Radio key={option.id} value={option.value} disabled={option.disabled || false}>
                                        <p className="label-option-text"> {option.label}</p>
                                    </Radio>
                                </span>
                            </div>
                        </TooltipComponent>
                    ) : (
                        <div
                            key={option.value}
                            style={{ height: height }}
                            className={labelType ? (radioValue === option.value ? 'label-type radio-value' : 'label-type') : radioWrapper || 'radio-wrapper'}
                        >
                            {option.onlyCloud && <CloudOnly />}
                            <span
                                className={labelType ? (radioValue === option.value ? 'radio-style radio-selected' : 'radio-style') : `label ${radioStyle}`}
                                style={{ fontFamily: fontFamily }}
                            >
                                <Radio key={option.id} value={option.value} disabled={option.disabled || false}>
                                    <p className="label-option-text"> {option.label}</p>
                                </Radio>
                            </span>
                            {option.description && <span className="des">{option.description}</span>}
                        </div>
                    )
                )}
            </Radio.Group>
        </div>
    );
};

export default RadioButton;
