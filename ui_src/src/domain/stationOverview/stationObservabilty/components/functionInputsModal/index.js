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

import { useState } from 'react';
import Input from '../../../../../components/Input';

const FunctionInputsModal = ({ functionInputsChange }) => {
    const [inputs, setInputs] = useState([
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' },
        { name: 'API_KEY', type: 'string' }
    ]);

    const handleChange = (e, index) => {
        const newInputs = [...inputs];
        newInputs[index].value = e.target.value;
        setInputs(newInputs);
        functionInputsChange(newInputs);
    };

    return (
        <div className="function-inputs-modal">
            <span className="info">
                <p>These variables serve as dynamic placeholders, holding different values crucial for the function’s logic.</p>
                <p>
                    Functions Inputs can be configured in the{' '}
                    <label className="link" onClick={() => window.open('https://docs.memphis.dev/memphis/functions', '_blank')}>
                        memphis.yaml
                    </label>
                </p>
            </span>
            <div className="inputs-container">
                {inputs.map((input, index) => (
                    <span className="input-row" key={`${input?.name}${index}`}>
                        <Input
                            placeholder={input?.name}
                            type="text"
                            radiusType="semi-round"
                            colorType="gray"
                            backgroundColorType="light-gray"
                            borderColorType="gray"
                            height="40px"
                            width="240px"
                            value={input?.name}
                            disabled
                        />
                        <Input
                            placeholder={input?.type}
                            type="text"
                            radiusType="semi-round"
                            colorType="black"
                            backgroundColorType="none"
                            borderColorType="gray"
                            height="40px"
                            width="240px"
                            onBlur={(e) => handleChange(e, index)}
                            onChange={(e) => handleChange(e, index)}
                            value={input?.value}
                        />
                    </span>
                ))}
            </div>
        </div>
    );
};

export default FunctionInputsModal;
