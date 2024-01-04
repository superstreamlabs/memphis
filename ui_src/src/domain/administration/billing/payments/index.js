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

import React, { Fragment, useEffect, useState, useMemo } from 'react';
import { IoClose } from 'react-icons/io5';
import { BiDollar } from 'react-icons/bi';
import { ReactComponent as BillingModalIcon } from 'assets/images/billinigAlertIcon.svg';
import { ReactComponent as BillingIcon } from 'assets/images/dollarIcon.svg';
import { ReactComponent as ThreeDotsIcon } from 'assets/images/3dotsIcon.svg';
import { ReactComponent as EditIcon } from 'assets/images/editIcon.svg';
import { Popover } from 'antd';
import TotalPayment from './components/totalPayment';
import Button from 'components/button';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import Input from 'components/Input';
import Modal from 'components/modal';
import { showMessages } from 'services/genericServices';
import { Form } from 'antd';
import { LOCAL_STORAGE_USER_NAME } from 'const/localStorageConsts';

function Payments() {
    const [isOpen, setIsOpen] = useState(false);
    const [alertLoading, setAlertLoading] = useState(false);
    const [formFields, setFormFields] = useState({});
    const [creationForm] = Form.useForm();
    const [isOptionsOpen, setIsOptionsOpen] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [isFormValid, setIsFormValid] = useState(false);

    useEffect(() => {
        getBillingAlert();
    }, []);

    const getBillingAlert = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('GET', ApiEndpoints.GET_BILLING_ALERT);
            if (data) {
                setFormFields({ ...data, current_price: data.current_price || 0 });
            }
        } catch (err) {
            console.error(err);
        } finally {
            setIsLoading(false);
        }
    };
    const billingAlertButtonText = useMemo(() => {
        let res = 'Set billing alert';
        if (formFields?.current_price > formFields?.amount) {
            res = `Your billing limit is crossed by $${formFields?.current_price - formFields?.amount} `;
        } else if (formFields?.amount && formFields?.current_price === formFields?.amount && formFields?.amount !== 0) {
            res = `Your billing on the limit $${formFields.amount}`;
        } else if (formFields?.current_price < formFields?.amount) {
            res = `Your billing is under by $${formFields.amount - formFields?.current_price}`;
        }
        return res;
    }, [formFields]);
    const handleFormChange = (_, allFields) => {
        const emailValid = /^[a-zA-Z0-9._%]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(allFields.email);
        const amountValid = allFields.amount && !isNaN(allFields.amount);

        setIsFormValid(emailValid && amountValid);
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
                email: creationForm.getFieldValue('email'),
                amount: creationForm.getFieldValue('amount')
            });
            if (data) {
                setIsOpen(false);
                showMessages('success', 'Billing alert updated successfully');
                getBillingAlert();
            }
        } catch (err) {
            console.error(err);
        } finally {
            setAlertLoading(false);
        }
    };

    const unsetBillingAlert = async () => {
        setIsLoading(true);
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_BILLING_ALERT, {
                email: formFields?.email,
                amount: -1
            });
            if (data) {
                showMessages('success', 'Billing alert unset successfully');
                getBillingAlert();
            }
        } catch (err) {
        } finally {
            setIsLoading(false);
        }
    };

    const getPopoverContent = () => {
        return (
            <ul className="popover-ul">
                <li
                    className="list-item"
                    onClick={() => {
                        setIsOpen(true);
                        setIsOptionsOpen(false);
                    }}
                >
                    <EditIcon /> <span>Edit</span>
                </li>
                <li
                    className="list-item"
                    onClick={async () => {
                        setIsOptionsOpen(false);
                        setFormFields({});
                        creationForm.resetFields(['email', 'amount']);
                        await unsetBillingAlert();
                    }}
                >
                    <IoClose /> <span>Unset</span>
                </li>
            </ul>
        );
    };
    return (
        <div className="payments-container">
            <div className="header-preferences">
                <div className="header">
                    <div className="header-flex">
                        <p className="main-header">Payments</p>
                        <div className="actions">
                            <Button
                                className="modal-btn"
                                width="fit-content"
                                height="32px"
                                placeholder={
                                    <div className="billinig-alert-button">
                                        <BillingIcon />
                                        <p className="label"> {billingAlertButtonText}</p>
                                    </div>
                                }
                                colorType="purple"
                                radiusType="circle"
                                backgroundColorType="purple-light"
                                border="purple"
                                fontSize="14px"
                                fontFamily="InterMedium"
                                onClick={() => {
                                    setIsOpen(true);
                                }}
                                isLoading={isLoading}
                                disabled={formFields?.amount > formFields?.base_price}
                            />
                            <Popover
                                placement="bottomRight"
                                content={getPopoverContent()}
                                trigger={'click'}
                                onOpenChange={() => setIsOptionsOpen(!isOptionsOpen)}
                                open={formFields?.amount ? isOptionsOpen : false}
                            >
                                <Button
                                    height="32px"
                                    width="32px"
                                    placeholder={<ThreeDotsIcon />}
                                    colorType="purple"
                                    radiusType="circle"
                                    backgroundColorType="purple-light"
                                    border="purple"
                                    fontSize="14px"
                                    fontFamily="InterMedium"
                                    onClick={() => null}
                                    disabled={formFields?.amount ? false : true}
                                />
                            </Popover>
                        </div>
                    </div>
                    <p className="memphis-label">This section is for managing your payment methods and invoices.</p>
                </div>
            </div>
            <TotalPayment />
            <Modal
                width={'350px'}
                className="modal-wrapper billing-alert-modal"
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <BillingModalIcon />
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
                    <Form name="form" form={creationForm} autoComplete="off" onFinish={onFinish} className="billing-alert-form" onValuesChange={handleFormChange}>
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
                                initialValue={formFields?.email || localStorage.getItem(LOCAL_STORAGE_USER_NAME)}
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
                                    onBlur={(e) => creationForm.setFieldValue('email', e.target.value)}
                                    onChange={(e) => creationForm.setFieldValue('email', e.target.value)}
                                    value={creationForm.getFieldValue('email')}
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
                                initialValue={formFields?.amount || ''}
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
                                    onBlur={(e) => creationForm.setFieldValue('amount', parseFloat(e.target.value))}
                                    onChange={(e) => creationForm.setFieldValue('amount', parseFloat(e.target.value))}
                                    value={creationForm.getFieldValue('amount')}
                                    suffixIconComponent={<BiDollar size={24} />}
                                />
                            </Form.Item>
                        </div>

                        <div className="form-button">
                            <Button
                                className="modal-btn"
                                width="100%"
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
                                width="100%"
                                height="32px"
                                placeholder="Set Alert"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType={'purple'}
                                fontSize="12px"
                                fontWeight="600"
                                disabled={!isFormValid}
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
