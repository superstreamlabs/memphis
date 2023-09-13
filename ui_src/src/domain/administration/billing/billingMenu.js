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

import React from 'react';
import { ReactComponent as RequestsColorIcon } from '../../../assets/images/setting/requestsColor.svg';
import { ReactComponent as RequestsGrayIcon } from '../../../assets/images/setting/requestsGray.svg';

import '../style.scss';

function BillingMenu({ selectedMenuItem, setMenuItem }) {
    return (
        <div className="side-menu">
            <p className="header">Billing</p>
            <p className="sub-header">View your monthly usage</p>
            <div className={selectedMenuItem === 'usage' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('usage')}>
                {selectedMenuItem === 'usage' ? <RequestsColorIcon alt="usage report" /> : <RequestsGrayIcon alt="usage report" />}
                Usage Report
            </div>
            <div className={selectedMenuItem === 'payments' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('payments')}>
                <img src={selectedMenuItem === 'payments' ? PaymentsColor : PaymentsGray} alt="usage report" />
                Payments
            </div>
        </div>
    );
}
export default BillingMenu;
