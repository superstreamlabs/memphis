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

import React, { useState, useEffect } from 'react';
import { Form } from 'antd';
import Input from '../../../../../components/Input';
import Modal from '../../../../../components/modal';
import TooltipComponent from '../../../../../components/tooltip/tooltip';
import { ReactComponent as CheckShieldIcon } from '../../../../../assets/images/checkShieldIcon.svg';
import { ReactComponent as LockIcon } from '../../../../../assets/images/lockIcon.svg';

const FunctionInputsModal = ({ open, clickOutside, rBtnClick, rBtnText, clickedFunction, handleInputsChange }) => {
    const [inputs, setInputs] = useState([]);
    const [inputsForm] = Form.useForm();

    const handleChange = async (e, index) => {
        const newInputs = [...inputs];
        newInputs[index].value = e.target.value;
        setInputs(newInputs);
        handleInputsChange(newInputs);
    };

    useEffect(() => {
        inputsForm.resetFields();
    }, [open]);

    useEffect(() => {
        setInputs(clickedFunction?.inputs);
    }, [clickedFunction]);

    const onFinish = async () => {
        try {
            await inputsForm.validateFields();
            rBtnClick();
        } catch (e) {
            return;
        }
    };

    return (
        <Modal
            width={'1000px'}
            header={
                <div className="modal-header">
                    <div className="header-img-container">
                        <CheckShieldIcon />
                    </div>
                    <span className="flex-label">
                        <p>Function Inputs Variables</p>
                        <TooltipComponent text="The values entered in the following fields will be encrypted at-rest using AES256 encryption">
                            <LockIcon alt="secure data" />
                        </TooltipComponent>
                    </span>
                </div>
            }
            open={open}
            clickOutside={() => clickOutside()}
            rBtnClick={onFinish}
            rBtnText={rBtnText}
        >
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
                    <Form name="form" form={inputsForm}>
                        {inputs?.map((input, index) => (
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
                                <Form.Item
                                    name={input?.name}
                                    validateTrigger="onChange"
                                    rules={[
                                        {
                                            required: true,
                                            message: `Please input ${input?.name}!`
                                        }
                                    ]}
                                >
                                    <Input
                                        placeholder={'Value'}
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
                                </Form.Item>
                            </span>
                        ))}
                    </Form>
                </div>
            </div>
        </Modal>
    );
};

export default FunctionInputsModal;
