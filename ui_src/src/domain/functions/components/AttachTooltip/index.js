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
import Button from '../../../../components/button';
import { BsFillCloudyFill } from 'react-icons/bs';
import { isCloud } from '../../../../services/valueConvertor';
import { Popover } from 'antd';

const steps = [{ name: 'Step 1' }, { name: 'Step 2' }, { name: 'Step 3' }];

const content = (
    <div className="attach-function">
        <div className="info">
            {steps.map((step, index) => (
                <div className="step-container" key={index}>
                    <div className="step-header">
                        <div className="icon">{index + 1}</div>
                        <div className="step-name">{step.name}</div>
                    </div>
                </div>
            ))}
        </div>
    </div>
);
const AttachTooltip = ({ disabled }) => {
    const [open, setOpen] = useState(false);

    return (
        <Popover
            placement="bottomLeft"
            title={'To attach a function please perform the following steps:'}
            content={content}
            trigger="click"
            overlayClassName="attach-function-popover"
            open={open}
            onOpenChange={(open) => setOpen(open)}
        >
            <Button
                placeholder={
                    <div className="code-btn">
                        {!isCloud() && <BsFillCloudyFill />}
                        <label>Attach</label>
                    </div>
                }
                width={'100px'}
                backgroundColorType={'purple'}
                colorType={'white'}
                radiusType={'circle'}
                fontSize="12px"
                fontFamily="InterSemiBold"
                onClick={() => setOpen(true)}
                disabled={disabled}
            />
        </Popover>
    );
};

export default AttachTooltip;
