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
import Consumed from '../../../../assets/images/setting/consumed.svg'
import Redeliver from '../../../../assets/images/setting/redeliver.svg'
import DeadLetter from '../../../../assets/images/setting/deadLetter.svg'

function Requests() {
    const val = 43279
    return (
        <div className="requests-container">
            <div className="header-preferences">
                <div className="header">
                    <p className="main-header">Requests</p>
                    <p className="memphis-label">We will keep an eye on your data streams and alert.</p>
                </div>
            </div>
            <div className='requests-panel summary'>
                <div className='requests-panel-item'>
                    <div>
                        <img src={TotalRequests} alt="TotalRequests"/> 
                        <label className="requests-title">Total requests</label>
                    </div>
                    <labe className="requests-value">{val.toLocaleString("en-US")}</labe>
                </div>
                
                <Divider type='vertical'/>
                <div className='requests-panel-item'>
                    <div>
                        <img src={TotalPayment} alt="TotalPayment"/> 
                        <label className="requests-title">Total payment</label>
                    </div>
                    <labe className="requests-value">${val.toLocaleString("en-US")}</labe>
                </div>
                <Divider type='vertical'/>
                <div className='requests-panel-item'>
                    <div>
                        <img src={TotalPayment} alt="PricePerRequest"/> 
                        <label className="requests-title">Price per request</label>
                    </div>
                    <labe className="requests-value">${val.toLocaleString("en-US")}</labe>
                </div>
            </div>
            <div className='requests-panel'>
                <div className='requests-item'>
                    <div className='circle-img'>
                    <img src={Consumed} alt='Consumed'/>
                    </div>
                        
                        <div>
                            <label>Consumed</label>
                            <label>Contrary to popular belief, Lorem Ipsum</label>
                        </div>
                </div>
                <labe className="requests-value value">{val.toLocaleString("en-US")}</labe>
            </div>
            <div className='requests-panel'>
                <div className='requests-item'>
                    <div className='circle-img'>
                    <img src={Redeliver} alt='Consumed'/>
                    </div>
                        
                        <div>
                            <label>Redeliver</label>
                            <label>Contrary to popular belief, Lorem Ipsum</label>
                        </div>
                        
                </div>
                <labe className="requests-value value">{val.toLocaleString("en-US")}</labe>
            </div>
            <div className='requests-panel'>
                <div className='requests-item'>
                    <div className='circle-img'>
                    <img src={DeadLetter} alt='Consumed'/>
                    </div>
                        
                        <div>
                            <label>Dead Letters</label>
                            <label>Contrary to popular belief, Lorem Ipsum</label>
                        </div>
                </div>
                <labe className="requests-value value">{val.toLocaleString("en-US")}</labe>
            </div>
        </div>
    );
}
export default Requests;

