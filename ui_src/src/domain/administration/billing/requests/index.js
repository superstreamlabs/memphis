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
import React from 'react';
import { Divider } from 'antd';
import TotalRequests from '../../../../assets/images/setting/totalRequests.svg'
import TotalPayment from '../../../../assets/images/setting/totalPayment.svg'
import AvgMsgSize from '../../../../assets/images/setting/avgMsgSize.svg'
import Consumed from '../../../../assets/images/setting/consumed.svg'
import Redeliver from '../../../../assets/images/setting/redeliver.svg'
import DeadLetter from '../../../../assets/images/setting/deadLetter.svg'
import Storage from '../../../../assets/images/setting/storage.svg'

function Requests() {
    const getNextPaymentDate = () => {
        const today = new Date();
        const nextMonth = (today.getMonth()+1)%12+1
        const year = today.getFullYear() + (nextMonth===1 ? 1: 0);
        today.setMonth(nextMonth - 1);
        return `01 ${today.toLocaleString('en-US', {month: 'long'})} ${year}` 
    }
   
    const val = 43279
    return (
        <div className="requests-container">
           <div className="header-preferences">
                <div className="header">
                    <p className="main-header">Requests</p>
                    <p className="memphis-label">We will keep an eye on your data streams and alert.</p>
                </div>
            </div>
           <div className='usage-header-section'>
                <div className='requests-summary'>
                    <div className='requests-total'>
                        <img src={TotalRequests} alt="TotalRequests"/> 
                        <span className='requests-data'>
                            <label className="requests-title">Total requests</label>
                            <label className="requests-value">{val.toLocaleString("en-US")}</label>
                        </span>
                        
                    </div>
                    <Divider/>
                    <div className='requests-total'>
                        <img src={AvgMsgSize} alt="AvgMsgSize"/> 
                        <span className='requests-data'>
                            <label className="requests-title">Average message size</label>
                            <label className="requests-value">{val.toLocaleString("en-US")}</label>
                        </span>
                        
                    </div>
                </div>
                <div className='total-payment'>
                    <div className='total-payment-header'>
                        <span>
                            <p className='total-ammount'>Total Payment</p>
                            <p className='next-billing'>Next billing date is {getNextPaymentDate()}</p>
                        </span>
                        <label className="requests-value">${val.toLocaleString("en-US")}</label>
                    </div>
                    <Divider/>
                    <span className='billing-item'>
                        <p className='item'>Subtotal</p>
                        <p className='ammount'>$2,425.00</p>
                    </span>
                    <span className='billing-item'>
                        <p className='item'>Other Fees</p>
                        <p className='ammount'>$0.00</p>
                    </span>
                    <span className='billing-item'>
                        <p className='item'>Discount</p>
                        <p className='ammount'>-</p>
                    </span>
                    <span className='billing-item'>
                        <p className='item'></p>
                        <p className='ammount'>$55.00</p>
                    </span>
                    <Divider/>
                    <div className='total-payment-footer'>
                        <span className='promo-code-section'>
                        <p className='next-billing'>Promo Code</p>
                        {/* <Input/> */}
                        </span>
                        <p className='download-invoice'>Download  Invoice</p>
                    </div>
                </div>
            </div>
            <div className='usage-details'>
                <div className='requests-panel'>
                    <div className='requests-item'>
                        <div className='circle-img'>
                        <img src={Consumed} alt='Consumed'/>
                        </div>
                            
                            <div>
                                <label className='request-type'>Consumed</label>
                                <label className='request-description'>Contrary to popular belief, Lorem Ipsum</label>
                            </div>
                    </div>
                    <label className="requests-value">{val.toLocaleString("en-US")}</label>
                </div>
                <div className='requests-panel'>
                    <div className='requests-item'>
                        <div className='circle-img'>
                        <img src={Redeliver} alt='Consumed'/>
                        </div>
                            
                            <div>
                                <label className='request-type'>Redeliver</label>
                                <label className='request-description'>Contrary to popular belief, Lorem Ipsum</label>
                            </div>
                            
                    </div>
                    <label className="requests-value">{val.toLocaleString("en-US")}</label>
                </div>
                <div className='requests-panel'>
                    <div className='requests-item'>
                        <div className='circle-img'>
                        <img src={Storage} alt='Storage'/>
                        </div>
                            
                            <div>
                                <label className='request-type'>Storage</label>
                                <label className='request-description'>Contrary to popular belief, Lorem Ipsum</label>
                            </div>
                    </div>
                    <label className="requests-value">{val.toLocaleString("en-US")}</label>
                </div>
                <div className='requests-panel'>
                    <div className='requests-item'>
                        <div className='circle-img'>
                        <img src={DeadLetter} alt='Consumed'/>
                        </div>
                            
                            <div>
                                <label className='request-type'>Dead Letters</label>
                                <label className='request-description'>Contrary to popular belief, Lorem Ipsum</label>
                            </div>
                    </div>
                    <label className="requests-value">{val.toLocaleString("en-US")}</label>
                </div> 
            </div>
        </div>
    );
}
export default Requests;

