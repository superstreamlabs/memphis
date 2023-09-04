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

import React, { Fragment, useEffect, useState, useContext } from 'react';

import { Context } from '../../../../hooks/store';
import billinigAlertIcon from '../../../../assets/images/billinigAlertIcon.svg';
import TotalPayment from './components/totalPayment';
import Button from '../../../../components/button';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Input from '../../../../components/Input';
import Modal from '../../../../components/modal';
import { showMessages } from '../../../../services/genericServices';
import { Form } from 'antd';

function Payments() {
    const [state, dispatch] = useContext(Context);
    const [isOpen, setIsOpen] = useState(false);
    const [amount, setAmount] = useState('');
    const [alertLoading, setAlertLoading] = useState(false);
    const [formFields, setFormFields] = useState({});
    const [creationForm] = Form.useForm();

    const isFreePlan = state?.monitor_data?.billing_details?.is_free_plan;

    useEffect(() => {
        getBillingAlert();
    }, []);

    const getBillingAlert = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_BILLING_ALERT);
            if (data) {
                setFormFields(data);
            }
        } catch (err) {
            console.error(err);
        }
    };

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const onFinish = async () => {
        const fieldsValue = await creationForm.validateFields();
        if (fieldsValue?.errorFields) {
            return;
        } else updateBillingAlert();
    };

    const updateBillingAlert = async () => {
        setAlertLoading(true);
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_BILLING_ALERT, {
                ...formFields
            });
            if (data) {
                setIsOpen(false);
                showMessages('success', 'Billing Alert Updated Successfully');
            }
        } catch (err) {
            console.error(err);
        } finally {
            setAlertLoading(false);
        }
    };
    return (
        <div className="payments-container">
            <div className="header-preferences">
                <div className="header">
                    <div className="header-flex">
                        <p className="main-header">Payments</p>
                        {!!!isFreePlan && (
                            <Button
                                className="modal-btn"
                                width="fit-content"
                                height="32px"
                                placeholder={
                                    <div className="billinig-alert-button">
                                        <img src={billinigAlertIcon} alt="billinigAlertIcon" />
                                        <label>{Object.keys(formFields).length > 0 ? 'Edit Billing Alert' : 'Set Billing Alert'}</label>
                                    </div>
                                }
                                colorType="black"
                                radiusType="circle"
                                backgroundColorType="none"
                                border="gray-light"
                                fontSize="14px"
                                fontFamily="InterMedium"
                                onClick={() => {
                                    setIsOpen(true);
                                }}
                            />
                        )}
                    </div>
                    <p className="memphis-label">This section is for managing your payment methods and invoices.</p>
                </div>
            </div>
            <TotalPayment />
            <Modal
                className="modal-wrapper billing-alert-modal"
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <img src={billinigAlertIcon} alt="billinigAlertIcon" />
                        </div>
                        <p>Billing Alert</p>
                        <label>We will notify you when your cost passes the defined amount</label>
                    </div>
                }
                displayButtons={false}
                open={isOpen}
                clickOutside={() => setIsOpen(false)}
            >
                <Fragment>
                    <Form name="form" form={creationForm} autoComplete="off" onFinish={onFinish} className="billing-alert-form">
                        <div className="form-field">
                            <p className="field-title">Enter your email address</p>
                            <Form.Item
                                name="email"
                                rules={[
                                    {
                                        required: true,
                                        message: 'Please input email!'
                                    },
                                    {
                                        message: 'Please enter a valid email address!',
                                        pattern: /^[a-zA-Z0-9._%]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
                                    }
                                ]}
                                initialValue={formFields.email}
                            >
                                <Input
                                    placeholder="Type email"
                                    type="text"
                                    radiusType="semi-round"
                                    maxLength={60}
                                    colorType="black"
                                    backgroundColorType="none"
                                    borderColorType="gray"
                                    height="40px"
                                    fontSize="12px"
                                    onBlur={(e) => updateFormState('email', e.target.value)}
                                    onChange={(e) => updateFormState('email', e.target.value)}
                                    value={formFields.email}
                                />
                            </Form.Item>
                        </div>
                        <div className="form-field">
                            <p>Set your billing amount</p>
                            <Form.Item
                                name="amount"
                                rules={[
                                    {
                                        required: true,
                                        message: 'Please input amount!'
                                    }
                                ]}
                                initialValue={formFields.amount}
                            >
                                <Input
                                    placeholder="Type amount"
                                    type="number"
                                    radiusType="semi-round"
                                    maxLength={60}
                                    colorType="black"
                                    backgroundColorType="none"
                                    borderColorType="gray"
                                    height="40px"
                                    fontSize="12px"
                                    onBlur={(e) => updateFormState('amount', parseFloat(e.target.value))}
                                    onChange={(e) => updateFormState('amount', parseFloat(e.target.value))}
                                    value={formFields.amount}
                                />
                            </Form.Item>
                        </div>

                        <div className="form-button">
                            <Button
                                className="modal-btn"
                                width="10vw"
                                height="32px"
                                placeholder="Close"
                                colorType="black"
                                radiusType="circle"
                                border="gray-light"
                                backgroundColorType={'white'}
                                fontSize="12px"
                                fontWeight="600"
                                onClick={() => {
                                    setIsOpen(false);
                                }}
                            />
                            <Button
                                className="modal-btn"
                                width="10vw"
                                height="32px"
                                placeholder="Set Alert"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType={'purple'}
                                fontSize="12px"
                                fontWeight="600"
                                disabled={
                                    !creationForm.isFieldsTouched(['email', 'amount']) || creationForm.getFieldsError().filter(({ errors }) => errors.length).length > 0
                                }
                                onClick={() => {
                                    updateBillingAlert();
                                }}
                                isLoading={alertLoading}
                            />
                        </div>
                    </Form>
                </Fragment>
            </Modal>
        </div>
    );
}
export default Payments;
