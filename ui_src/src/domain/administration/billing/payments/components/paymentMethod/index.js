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

import React, { useState } from 'react';
import { ReactComponent as CreditCardIcon } from '../../../../../../assets/images/setting/credit-card.svg';
import { ReactComponent as AddPaymentIcon } from '../../../../../../assets/images/setting/addPayment.svg';
import Button from '../../../../../../components/button';

import './style.scss';

function PaymentMethod() {
    const [paymentMethods, setPaymentMethods] = useState(['method1']);
    return (
        <div className="payments-section-card">
            {paymentMethods.length > 0 ? (
                <>
                    <div className="payment-method-top">
                        <div>
                            <p className="payment-method">Payment method</p>
                            <p className="payment-method-description">Change how you want to pay for your plan.</p>
                        </div>
                        <label className="view-cards">View Cards</label>
                    </div>
                    <div className="payment-method-bottom">
                        <div className="credit-card-bottom">
                            <CreditCardIcon alt="credit-card-img" />
                            <div>
                                <p>**** **** **** 4956</p>
                                <p>Debit Card</p>
                            </div>
                        </div>

                        <Button
                            className="modal-btn"
                            width="83px"
                            height="32px"
                            placeholder="Update"
                            disabled={false}
                            colorType="navy"
                            radiusType="semi-round"
                            border="gray"
                            backgroundColorType={'white'}
                            fontSize="12px"
                            fontWeight="600"
                            isLoading={false}
                            onClick={() => {
                                console.log('hi');
                            }}
                        />
                    </div>
                </>
            ) : (
                <div className="no-payment-method">
                    <AddPaymentIcon className="add-payment" alt="add-payment" />
                    <div>
                        <p className="payment-method-title">Payment Method</p>
                        <p className="payment-method-description">Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece o</p>
                        <p className="add-new-method">Add New Payment Method</p>
                    </div>
                </div>
            )}
        </div>
    );
}

export default PaymentMethod;
