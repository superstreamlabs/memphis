// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { useState, useEffect } from 'react';
import Button from '../../../../../components/button';

import { ReactComponent as CheckIcon } from '../../../../../assets/images/checkIcon.svg';

const FunctionsApplyModal = ({ onCancel, onApply, successText }) => {
    const [selected, setSelected] = useState(0);

    const handleSelect = (index) => {
        setSelected(index);
    };
    const content = [
        {
            title: 'No',
            value: false
        },
        {
            title: 'Yes',
            value: true
        }
    ];
    return (
        <>
            <div className="select-section">
                {content.map((item, index) => (
                    <div className={`modal-card ${selected === index ? 'selected' : undefined}`} onClick={() => handleSelect(index)}>
                        <div className="modal-card__top-row">
                            <div className="modal-card__title">{item.title}</div>
                            {[selected].includes(index) ? <CheckIcon /> : <div className="empty-circle" />}
                        </div>
                        <div className="modal-card__subtitle">{item.subtitle}</div>
                    </div>
                ))}
            </div>
            <div className="modal-buttons">
                <Button
                    width="100%"
                    height="34px"
                    placeholder="Cancel"
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    border={'gray'}
                    fontSize="12px"
                    fontWeight="600"
                    onClick={onCancel}
                />
                <Button
                    width="100%"
                    height="34px"
                    placeholder={successText}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontWeight="600"
                    onClick={() => onApply(content[selected]?.value)}
                />
            </div>
        </>
    );
};

export default FunctionsApplyModal;
